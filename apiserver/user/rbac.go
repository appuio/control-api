package user

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"go.uber.org/multierr"
	rbacv1 "k8s.io/api/rbac/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/filters"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/klog/v2"
	"k8s.io/utils/strings"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/appuio/control-api/apiserver/secretstorage"
)

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;delete;patch;update;edit
// +kubebuilder:rbac:groups=rbac.appuio.io;user.appuio.io,resources=invitations,verbs=get;edit;update;patch;delete

// rbacCreatorIsOwner is a wrapper around the Invitation storage that creates a ClusterRole and ClusterRoleBinding
// to make the creator of the Invitation the owner of the Invitation.
type rbacCreatorIsOwner struct {
	secretstorage.ScopedStandardStorage
	client client.Client
}

// Create passes the object to the wrapped storage and creates a ClusterRole and ClusterRoleBinding for the creator of the object using the returned object's name.
func (c *rbacCreatorIsOwner) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, opts *metav1.CreateOptions) (runtime.Object, error) {
	attr, err := filters.GetAuthorizerAttributes(ctx)
	if err != nil {
		return nil, err
	}
	user := attr.GetUser()

	createdObj, err := c.ScopedStandardStorage.Create(ctx, obj, createValidation, opts)
	if err != nil {
		return createdObj, err
	}

	ac := apimeta.NewAccessor()
	objName, err := ac.Name(createdObj)
	if err != nil {
		return createdObj, fmt.Errorf("could not get name of created object: %w", err)
	}

	rolename := roleName(objName)

	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: rolename,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"rbac.appuio.io", "user.appuio.io"},
				Resources:     []string{"invitations"},
				Verbs:         []string{"get", "edit", "update", "patch", "delete"},
				ResourceNames: []string{objName},
			},
		},
	}

	rolebinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: rolename,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "User",
				APIGroup: "rbac.authorization.k8s.io",
				Name:     user.GetName(),
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     rolename,
		},
	}

	rollback := func() error {
		_, _, err := c.ScopedStandardStorage.Delete(ctx, objName, nil, &metav1.DeleteOptions{DryRun: opts.DryRun})
		return err
	}

	err = c.client.Create(ctx, role, &client.CreateOptions{DryRun: opts.DryRun})
	if err != nil {
		rollbackErr := rollback()
		return createdObj, multierr.Append(err, rollbackErr)
	}
	err = c.client.Create(ctx, rolebinding, &client.CreateOptions{DryRun: opts.DryRun})
	if err != nil {
		rollbackErr := rollback()
		roleRollbackErr := c.client.Delete(ctx, role, &client.DeleteOptions{DryRun: opts.DryRun})
		return createdObj, multierr.Combine(err, rollbackErr, roleRollbackErr)
	}

	return createdObj, nil
}

// Delete passes the object to the wrapped storage and deletes the ClusterRole and ClusterRoleBinding associated with the object.
func (c *rbacCreatorIsOwner) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, opts *metav1.DeleteOptions) (runtime.Object, bool, error) {
	deletedObj, im, err := c.ScopedStandardStorage.Delete(ctx, name, deleteValidation, opts)
	if err != nil {
		return deletedObj, im, err
	}

	rolename := roleName(name)
	err1 := c.client.Delete(ctx, &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: rolename,
		},
	}, &client.DeleteOptions{DryRun: opts.DryRun})
	err2 := c.client.Delete(ctx, &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: rolename,
		},
	}, &client.DeleteOptions{DryRun: opts.DryRun})

	if err := multierr.Combine(err1, err2); err != nil {
		klog.FromContext(ctx).Error(err, "failed to clean up RBAC resources")
	}

	return deletedObj, im, nil
}

func (s *rbacCreatorIsOwner) DeleteCollection(ctx context.Context, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions, listOptions *metainternalversion.ListOptions) (runtime.Object, error) {
	return nil, fmt.Errorf("not implemented")
}

func roleName(objName string) string {
	prefix := "invitations-"
	suffix := "-owner"

	if len(prefix)+len(suffix)+len(objName) <= 63 {
		return fmt.Sprintf("%s%s%s", prefix, objName, suffix)
	}

	h := sha1.New()
	h.Write([]byte(objName))
	hsh := strings.ShortenString(hex.EncodeToString(h.Sum(nil)), 7)

	maxLength := 63 - len(prefix) - len(suffix) - len(hsh) - 1
	maxSafe := strings.ShortenString(objName, maxLength)

	return fmt.Sprintf("%s%s-%s%s", prefix, maxSafe, hsh, suffix)
}
