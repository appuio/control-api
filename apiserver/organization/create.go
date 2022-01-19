package organization

import (
	"context"
	"fmt"
	"strings"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
)

var _ rest.Creater = &organizationStorage{}

func (s *organizationStorage) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	org, ok := obj.(*orgv1.Organization)
	if !ok {
		return nil, fmt.Errorf("not an organization: %#v", obj)
	}
	err := s.authorizer.AuthorizeContext(ctx)
	if err != nil {
		return nil, err
	}

	// Validate Org
	if err := createValidation(ctx, obj); err != nil {
		return nil, err
	}

	return s.create(ctx, org, options)
}

func (s *organizationStorage) create(ctx context.Context, org *orgv1.Organization, options *metav1.CreateOptions) (*orgv1.Organization, error) {
	if err := s.namepaces.CreateNamespace(ctx, org.ToNamespace(), options); err != nil {
		return nil, convertNamespaceError(err)
	}

	if err := s.rbac.CreateRoleBindings(ctx, org.Name); err != nil {
		// rollback
		_, deleteErr := s.namepaces.DeleteNamespace(ctx, org.Name, nil)
		if deleteErr != nil {
			err = fmt.Errorf("%w and failed to clean up namespace: %s", err, deleteErr.Error())
		}
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	orgMembers := newOrganizationMembers(ctx, org.Name, s.usernamePrefix)

	if err := s.members.CreateMembers(ctx, orgMembers); err != nil {
		// rollback
		_, deleteErr := s.namepaces.DeleteNamespace(ctx, org.Name, nil)
		if deleteErr != nil {
			err = fmt.Errorf("%w and failed to clean up namespace: %s", err, deleteErr.Error())
		}
		return nil, fmt.Errorf("failed to create organization: %w", err)

	}

	return org, nil
}

func newOrganizationMembers(ctx context.Context, organization, usernamePrefix string) *controlv1.OrganizationMembers {
	userRefs := []controlv1.UserRef{}
	user, ok := request.UserFrom(ctx)
	if ok {
		userRefs = append(userRefs, controlv1.UserRef{
			ID: strings.TrimPrefix(user.GetName(), usernamePrefix),
		})
	}

	return &controlv1.OrganizationMembers{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "members",
			Namespace: organization,
		},
		Spec: controlv1.OrganizationMembersSpec{
			UserRefs: userRefs,
		},
	}
}
