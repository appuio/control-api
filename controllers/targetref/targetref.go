package targetref

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"
)

// GetTarget returns the target object for the given TargetRef.
// Returns an error if the target is not supported.
func GetTarget(ctx context.Context, c client.Client, target userv1.TargetRef) (client.Object, error) {
	var obj client.Object
	switch {
	case target.APIGroup == "appuio.io" && target.Kind == "OrganizationMembers":
		obj = &controlv1.OrganizationMembers{}
	case target.APIGroup == "appuio.io" && target.Kind == "Team":
		obj = &controlv1.Team{}
	case target.APIGroup == rbacv1.GroupName && target.Kind == "ClusterRoleBinding":
		obj = &rbacv1.ClusterRoleBinding{}
	case target.APIGroup == rbacv1.GroupName && target.Kind == "RoleBinding":
		obj = &rbacv1.RoleBinding{}
	default:
		return nil, fmt.Errorf("unsupported target %q.%q", target.APIGroup, target.Kind)
	}

	err := c.Get(ctx, client.ObjectKey{Name: target.Name, Namespace: target.Namespace}, obj)
	return obj, err
}

// UserAccessor is an interface for accessing users from objects supported by this project.
type UserAccessor interface {
	// EnsureUser adds the user to the object if it is not already present.
	// Returns true if the user was added.
	// Uses the prefix where applicable.
	EnsureUser(prefix, user string) (added bool)
	// HasUser returns true if the user is present in the object.
	// Uses the prefix where applicable.
	HasUser(prefix, user string) bool
}

// NewUserAccessor returns a UserAccessor for the given object or an error if the object is not supported.
func NewUserAccessor(obj client.Object) (UserAccessor, error) {
	switch o := obj.(type) {
	case *controlv1.OrganizationMembers:
		return &controlv1UserRefAccessor{userRefs: &o.Spec.UserRefs}, nil
	case *controlv1.Team:
		return &controlv1UserRefAccessor{userRefs: &o.Spec.UserRefs}, nil
	case *rbacv1.ClusterRoleBinding:
		return &rbacv1SubjectAccessor{subjects: &o.Subjects}, nil
	case *rbacv1.RoleBinding:
		return &rbacv1SubjectAccessor{subjects: &o.Subjects}, nil
	default:
		return nil, fmt.Errorf("unsupported object %T", obj)
	}
}

type controlv1UserRefAccessor struct {
	userRefs *[]controlv1.UserRef
}

var _ UserAccessor = &controlv1UserRefAccessor{}

func (s *controlv1UserRefAccessor) HasUser(_, user string) bool {
	// Prefix is not used for UserRef
	return isInSlice(*s.userRefs, controlv1.UserRef{Name: user})
}

func (s *controlv1UserRefAccessor) EnsureUser(_, user string) (added bool) {
	// Prefix is not used for UserRef
	*s.userRefs, added = ensure(*s.userRefs, controlv1.UserRef{Name: user})
	return
}

type rbacv1SubjectAccessor struct {
	subjects *[]rbacv1.Subject
}

var _ UserAccessor = &rbacv1SubjectAccessor{}

func (s *rbacv1SubjectAccessor) HasUser(prefix, user string) bool {
	return isInSlice(*s.subjects, newSubject(prefix+user))
}

func (s *rbacv1SubjectAccessor) EnsureUser(prefix, user string) (added bool) {
	*s.subjects, added = ensure(*s.subjects, newSubject(prefix+user))
	return
}

func newSubject(user string) rbacv1.Subject {
	return rbacv1.Subject{
		Kind:     rbacv1.UserKind,
		APIGroup: rbacv1.GroupName,
		Name:     user,
	}
}

func ensure[T comparable](s []T, e T) (ret []T, added bool) {
	for _, v := range s {
		if v == e {
			return s, false
		}
	}
	return append(s, e), true
}

func isInSlice[T comparable](s []T, e T) (found bool) {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
