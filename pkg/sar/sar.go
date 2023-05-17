package sar

import (
	"context"
	"errors"
	"fmt"

	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AuthorizeResource checks if the given user is allowed to access the given resource, using SubjectAccessReviews.
func AuthorizeResource(ctx context.Context, c client.Client, user authenticationv1.UserInfo, resource ResourceAttributes) error {
	// I could not find a way to create a SubjectAccessReview object with the client.
	// `no kind "CreateOptions" is registered for the internal version of group "authorization.k8s.io" in scheme`
	// even after installing the authorization scheme.
	rawSAR := &unstructured.Unstructured{
		Object: map[string]any{
			"spec": sarSpec{
				ResourceAttributes: resource,

				User:   user.Username,
				Groups: user.Groups,
				Extra:  user.Extra,
				UID:    user.UID,
			},
		}}
	rawSAR.SetGroupVersionKind(schema.GroupVersionKind{Group: "authorization.k8s.io", Version: "v1", Kind: "SubjectAccessReview"})

	if err := c.Create(ctx, rawSAR); err != nil {
		return fmt.Errorf("failed to create SubjectAccessReview: %w", err)
	}

	allowed, _, err := unstructured.NestedBool(rawSAR.Object, "status", "allowed")
	if err != nil {
		return fmt.Errorf("failed to get SubjectAccessReview status.allowed: %w", err)
	}

	if !allowed {
		return fmt.Errorf("%q is not allowed by %q", resource, user)
	}

	return nil
}

// MOCK_SubjectAccessReviewResponder is a wrapper for client.WithWatch that responds to SubjectAccessReview create requests
// and allows or denies the request based on the AllowedUser name.
type MOCK_SubjectAccessReviewResponder struct {
	client.WithWatch

	AllowedUser string
}

func (r MOCK_SubjectAccessReviewResponder) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if sar, ok := obj.(*unstructured.Unstructured); ok {
		if sar.GetKind() == "SubjectAccessReview" {
			o, ok, err := unstructured.NestedFieldNoCopy(sar.Object, "spec")
			if err != nil {
				return err
			}
			if !ok {
				return errors.New("spec not found")
			}
			s, ok := o.(sarSpec)
			if !ok {
				return errors.New("unknown spec type, might not originate from this package")
			}

			unstructured.SetNestedField(sar.Object, s.User == r.AllowedUser, "status", "allowed")
			return nil
		}
	}
	return r.WithWatch.Create(ctx, obj, opts...)
}

type sarSpec struct {
	ResourceAttributes ResourceAttributes `json:"resourceAttributes"`

	User   string                                 `json:"user,omitempty"`
	Groups []string                               `json:"groups,omitempty"`
	Extra  map[string]authenticationv1.ExtraValue `json:"extra,omitempty"`
	UID    string                                 `json:"uid,omitempty"`
}
