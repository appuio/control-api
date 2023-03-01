package controllers

import (
	"context"
	"errors"
	"fmt"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"go.uber.org/multierr"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"
)

// InvitationEmailReconciler reconciles invitations and sends invitation emails if appropriate
type InvitationEmailReconciler struct {
	client.Client

	Recorder record.EventRecorder
	Scheme   *runtime.Scheme

	EmailTestMode bool
}

//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations/status,verbs=get;update;patch

//+kubebuilder:rbac:groups=appuio.io,resources=organizationmembers;teams,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings;rolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile reacts to redeemed invitations and sends invitation emails to the user if needed.
func (r *InvitationEmailReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(4).WithValues("request", req).Info("Reconciling")

	inv := userv1.Invitation{}
	if err := r.Get(ctx, req.NamespacedName, &inv); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !inv.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	status := apimeta.FindStatusCondition(inv.Status.Conditions, userv1.ConditionEmailSent)

	if status == nil {
		return ctrl.Result{}, nil
	}

	if apimeta.IsStatusConditionTrue(inv.Status.Conditions, userv1.ConditionEmailSent) {
		return ctrl.Result{}, nil
	}

	// send email

	log.V(0).Info("Sending e-mail")


	var errors []error
	statusHasChanged := false

	if statusHasChanged {
		err := r.Client.Status().Update(ctx, &inv)
		return ctrl.Result{}, multierr.Append(err, multierr.Combine(errors...))
	}
	return ctrl.Result{}, multierr.Combine(errors...)
}

// SetupWithManager sets up the controller with the Manager.
func (r *InvitationEmailReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&userv1.Invitation{}).
		Complete(r)
}