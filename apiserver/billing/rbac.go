package billing

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"go.uber.org/multierr"
	rbacv1 "k8s.io/api/rbac/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/filters"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/appuio/control-api/apiserver/billing/odoostorage"
)

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;delete;patch;update;edit
// +kubebuilder:rbac:groups=rbac.appuio.io,resources=billingentities,verbs=*

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

	rolename := fmt.Sprintf("billingentities-%s-viewer", objName)

	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: rolename,
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
		if deleter, canDelete := c.Storage.(rest.GracefulDeleter); canDelete {
			_, _, err := deleter.Delete(ctx, objName, nil, &metav1.DeleteOptions{DryRun: opts.DryRun})
			return err
		}
		logr.FromContextOrDiscard(ctx).Info("storage does not implement GracefulDeleter, skipping rollback", "object", objName)
		return nil
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
