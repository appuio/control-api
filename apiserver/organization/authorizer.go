package organization

import (
	"context"
	"errors"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/endpoints/filters"
)

//go:generate go run github.com/golang/mock/mockgen -destination=./mock/$GOFILE -package mock_organization k8s.io/apiserver/pkg/authorization/authorizer Authorizer

type rbacAuthorizer struct {
	Authorizer authorizer.Authorizer
}

var rbacGV = metav1.GroupVersion{
	Group:   "rbac.appuio.io",
	Version: "v1",
}

func (a rbacAuthorizer) Authorize(ctx context.Context, attr authorizer.Attributes) error {
	if attr.GetResource() != "organizations" {
		return fmt.Errorf("unkown resource %q", attr.GetResource())
	}
	decision, reason, err := a.Authorizer.Authorize(ctx, authorizer.AttributesRecord{
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

	if err != nil {
		return err
	} else if decision != authorizer.DecisionAllow {
		return apierrors.NewForbidden(schema.GroupResource{
			Group:    attr.GetAPIGroup(),
			Resource: attr.GetResource(),
		}, attr.GetName(), errors.New(reason))
	}
	return nil
}

func (a rbacAuthorizer) AuthorizeContext(ctx context.Context) error {
	attr, err := filters.GetAuthorizerAttributes(ctx)
	if err != nil {
		return err
	}
	return a.Authorize(ctx, attr)
}

func (a rbacAuthorizer) AuthorizeVerb(ctx context.Context, verb string) error {
	attr, err := filters.GetAuthorizerAttributes(ctx)
	if err != nil {
		return err
	}
	return a.Authorize(ctx, authorizer.AttributesRecord{
		User:            attr.GetUser(),
		Verb:            verb,
		Name:            attr.GetName(),
		Namespace:       attr.GetName(),
		APIGroup:        attr.GetAPIGroup(),
		APIVersion:      attr.GetAPIVersion(),
		Resource:        attr.GetResource(),
		Subresource:     attr.GetSubresource(),
		ResourceRequest: attr.IsResourceRequest(),
		Path:            attr.GetPath(),
	})
}
