package controllers

import (
	"context"

	"go.uber.org/multierr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	controlv1 "github.com/appuio/control-api/apis/v1"
)

// DefaultOrganizationReconciler reconciles User resources to ensure they have a DefaultOrganization set if applicable.
type DefaultOrganizationReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=appuio.io,resources=organizationmembers,verbs=get;list;watch
//+kubebuilder:rbac:groups=appuio.io,resources=users,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=appuio.io,resources=users/status,verbs=get

// Reconcile reacts on changes of memberships and sets members' default organization if appropriate
func (r *DefaultOrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(1).WithValues("request", req).Info("Reconciling")

	memb := controlv1.OrganizationMembers{}
	if err := r.Get(ctx, req.NamespacedName, &memb); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !memb.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	allMemberships := controlv1.OrganizationMembersList{}
	if err := r.List(ctx, &allMemberships); err != nil {
		return ctrl.Result{}, err
	}

	var errGroup error
	for _, user := range memb.Status.ResolvedUserRefs {
		myOrgs := make([]string, 0)

		for _, membership := range allMemberships.Items {
			for _, membershipUser := range membership.Status.ResolvedUserRefs {
				if user.Name == membershipUser.Name {
					myOrgs = append(myOrgs, membership.Namespace)
					break
				}
			}
		}
		if len(myOrgs) == 1 {
			err := setUserDefaultOrganization(ctx, r.Client, user.Name, myOrgs[0])
			errGroup = multierr.Append(errGroup, err)
		}
	}

	return ctrl.Result{}, errGroup
}

func setUserDefaultOrganization(ctx context.Context, c client.Client, userName string, orgName string) error {
	user := controlv1.User{}
	if err := c.Get(ctx, types.NamespacedName{Name: userName}, &user); err != nil {
		return err
	}

	if user.Spec.Preferences.DefaultOrganizationRef != "" {
		return nil
	}

	user.Spec.Preferences.DefaultOrganizationRef = orgName
	return c.Update(ctx, &user)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DefaultOrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&controlv1.OrganizationMembers{}).
		Complete(r)
}
