package billing

import (
	"context"
	"fmt"

	"go.uber.org/multierr"
	rbacv1 "k8s.io/api/rbac/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/filters"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/appuio/control-api/apiserver/billing/odoostorage"
)

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;delete;patch;update;edit
// +kubebuilder:rbac:groups=rbac.appuio.io;billing.appuio.io,resources=billingentities,verbs=*

// createRBACWrapper is a wrapper around the storage that creates a ClusterRole and ClusterRoleBinding for each BillingEntity on creation.
type createRBACWrapper struct {
	odoostorage.Storage
	client client.Client
}

func (c *createRBACWrapper) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, opts *metav1.CreateOptions) (runtime.Object, error) {
	attr, err := filters.GetAuthorizerAttributes(ctx)
	if err != nil {
		return nil, err
	}
	user := attr.GetUser()

	createdObj, err := c.Storage.Create(ctx, obj, createValidation, opts)
	if err != nil {
		return createdObj, err
	}

	ac := apimeta.NewAccessor()
	objName, err := ac.Name(createdObj)
	if err != nil {
		return createdObj, fmt.Errorf("could not get name of created object: %w", err)
	}

	viewRoleName := fmt.Sprintf("billingentities-%s-viewer", objName)
	viewRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: viewRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"rbac.appuio.io"},
				Resources:     []string{"billingentities"},
				Verbs:         []string{"get"},
				ResourceNames: []string{objName},
			},
		},
	}
	viewRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: viewRoleName,
		},
		Subjects: []rbacv1.Subject{},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     viewRoleName,
		},
	}
	adminRoleName := fmt.Sprintf("billingentities-%s-admin", objName)
	adminRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: adminRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"rbac.appuio.io", "billing.appuio.io"},
				Resources:     []string{"billingentities"},
				Verbs:         []string{"get", "patch", "update", "edit"},
				ResourceNames: []string{objName},
			},
			{
				APIGroups:     []string{"rbac.authorization.k8s.io"},
				Resources:     []string{"clusterrolebindings"},
				Verbs:         []string{"get", "edit", "update", "patch"},
				ResourceNames: []string{viewRoleName, adminRoleName},
			},
		},
	}
	adminRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: adminRoleName,
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
			Name:     adminRoleName,
		},
	}

	rollback := func() error {
		if deleter, canDelete := c.Storage.(rest.GracefulDeleter); canDelete {
			_, _, err := deleter.Delete(ctx, objName, nil, &metav1.DeleteOptions{DryRun: opts.DryRun})
			return err
		}
		klog.FromContext(ctx).Info("storage does not implement GracefulDeleter, skipping rollback", "object", objName)
		return nil
	}

	toCreate := []client.Object{viewRole, viewRoleBinding, adminRole, adminRoleBinding}
	created := make([]client.Object, 0, len(toCreate))
	var createErr error
	for _, obj := range toCreate {
		if err := c.client.Create(ctx, obj, &client.CreateOptions{DryRun: opts.DryRun}); err != nil {
			createErr = err
			break
		}
		created = append(created, obj)
	}
	if err := createErr; err != nil {
		for _, obj := range created {
			multierr.AppendInto(&err, c.client.Delete(ctx, obj, &client.DeleteOptions{DryRun: opts.DryRun}))
		}
		return createdObj, multierr.Combine(err, rollback())
	}

	return createdObj, nil
}
