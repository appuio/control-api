package organization

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/endpoints/filters"
)

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=./mock/$GOFILE
type ContextAuthorizer interface {
	authorizer.Authorizer
	AuthorizeContext(ctx context.Context) (authorizer.Decision, string, error)
}

type rbacAuthorizer struct {
	Authorizer authorizer.Authorizer
}

var rbacGV = metav1.GroupVersion{
	Group:   "rbac.appuio.io",
	Version: "v1",
}

func (a rbacAuthorizer) Authorize(ctx context.Context, attr authorizer.Attributes) (authorizer.Decision, string, error) {
	if attr.GetResource() != "organizations" {
		return authorizer.DecisionNoOpinion, "Unkown resource", fmt.Errorf("unkown resource %q", attr.GetResource())
	}
	return a.Authorizer.Authorize(ctx, authorizer.AttributesRecord{
		User:            attr.GetUser(),
		Verb:            attr.GetVerb(),
		Name:            attr.GetName(),
		Namespace:       attr.GetName(),
		APIGroup:        rbacGV.Group,
		APIVersion:      rbacGV.Version,
		Resource:        attr.GetResource(),
		Subresource:     attr.GetSubresource(),
		ResourceRequest: attr.IsResourceRequest(),
		Path:            attr.GetPath(),
	})
}

func (a rbacAuthorizer) AuthorizeContext(ctx context.Context) (authorizer.Decision, string, error) {
	attr, err := filters.GetAuthorizerAttributes(ctx)
	if err != nil {
		return authorizer.DecisionNoOpinion, "Failed to parse request", err
	}
	return a.Authorize(ctx, attr)
}
