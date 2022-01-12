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

// rbacAuthorizer processes authorization requests for `organizations.organization.appuio.io` and checks them based on rbac rules for `organizations.rbac.appuio.io`
type rbacAuthorizer struct {
	Authorizer authorizer.Authorizer
}

var rbacGV = metav1.GroupVersion{
	Group:   "rbac.appuio.io",
	Version: "v1",
}

// Authorizer makes an authorization decision based on the Attributes.
// It returns nil when an action is authorized, otherwise it returns an error.
func (a rbacAuthorizer) Authorize(ctx context.Context, attr authorizer.Attributes) error {
	if attr.GetResource() != "organizations" {
		return fmt.Errorf("unkown resource %q", attr.GetResource())
	}
	decision, reason, err := a.Authorizer.Authorize(ctx, authorizer.AttributesRecord{
		User:            attr.GetUser(),
		Verb:            attr.GetVerb(),
		Name:            attr.GetName(),
		Namespace:       attr.GetName(), // We handle cluster wide resources
		APIGroup:        rbacGV.Group,
		APIVersion:      rbacGV.Version,
		Resource:        attr.GetResource(),
		Subresource:     attr.GetSubresource(),
		ResourceRequest: true, // Always a resource request
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

// AuthorizerContext makes an authorization decision based on the Attributes present in the given Context.
// It returns nil when the context contains Attributes and the action is authorized, otherwise it returns an error.
func (a rbacAuthorizer) AuthorizeContext(ctx context.Context) error {
	attr, err := filters.GetAuthorizerAttributes(ctx)
	if err != nil {
		return err
	}
	return a.Authorize(ctx, attr)
}

// AuthorizerVerb makes an authorization decision based on the Attributes present in the given Context, but overriding the verb and object name to the provided values
// It returns nil when the context contains Attributes and the action is authorized, otherwise it returns an error.
func (a rbacAuthorizer) AuthorizeVerb(ctx context.Context, verb string, name string) error {
	attr, err := filters.GetAuthorizerAttributes(ctx)
	if err != nil {
		return err
	}
	return a.Authorize(ctx, authorizer.AttributesRecord{
		User:       attr.GetUser(),
		Verb:       verb,
		Name:       name,
		Namespace:  attr.GetNamespace(),
		APIGroup:   attr.GetAPIGroup(),
		APIVersion: attr.GetAPIVersion(),
		Resource:   attr.GetResource(),
		Path:       attr.GetPath(),
	})
}

// AuthorizerGet makes an authorization decision based on the Attributes present in the given Context, but overriding the verb to `get` and the object name to the provided values
// It returns nil when the context contains Attributes and the action is authorized, otherwise it returns an error.
func (a rbacAuthorizer) AuthorizeGet(ctx context.Context, name string) error {
	return a.AuthorizeVerb(ctx, "get", name)
}
