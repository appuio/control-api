package controllers

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/multierr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	"github.com/appuio/control-api/controllers/targetref"
)

// InvitationRedeemReconciler reconciles invitations and adds a token to the status if required.
type InvitationRedeemReconciler struct {
	client.Client

	Recorder record.EventRecorder
	Scheme   *runtime.Scheme

	UsernamePrefix string
}

//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations/status,verbs=get;update;patch

//+kubebuilder:rbac:groups=appuio.io,resources=organizationmembers;teams,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings;rolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile reacts to redeemed invitations and adds the user to the targets listed in the invitation status.
func (r *InvitationRedeemReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(1).WithValues("request", req).Info("Reconciling")

	inv := userv1.Invitation{}
	if err := r.Get(ctx, req.NamespacedName, &inv); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !inv.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if !inv.IsRedeemed() {
		return ctrl.Result{}, nil
	}

	if inv.Status.RedeemedBy == "" {
		return ctrl.Result{}, errors.New("redeemed invitation has no user")
	}

	var errors []error
	statusHasChanged := false
	for i := range inv.Status.TargetStatuses {
		if inv.Status.TargetStatuses[i].Condition.Status == metav1.ConditionTrue {
			continue
		}

		statusHasChanged = true
		username := strings.TrimPrefix(inv.Status.RedeemedBy, r.UsernamePrefix)
		err := addUserToTarget(ctx, r.Client, username, r.UsernamePrefix, inv.Status.TargetStatuses[i].TargetRef)
		if err != nil {
			errors = append(errors, err)
			inv.Status.TargetStatuses[i].Condition.LastTransitionTime = metav1.Now()
			inv.Status.TargetStatuses[i].Condition.Message = err.Error()
			inv.Status.TargetStatuses[i].Condition.Reason = metav1.StatusFailure
			inv.Status.TargetStatuses[i].Condition.Status = metav1.ConditionFalse
			continue
		}
		inv.Status.TargetStatuses[i].Condition.LastTransitionTime = metav1.Now()
		inv.Status.TargetStatuses[i].Condition.Message = ""
		inv.Status.TargetStatuses[i].Condition.Reason = metav1.StatusSuccess
		inv.Status.TargetStatuses[i].Condition.Status = metav1.ConditionTrue
	}

	if statusHasChanged {
		err := r.Client.Status().Update(ctx, &inv)
		return ctrl.Result{}, multierr.Append(err, multierr.Combine(errors...))
	}
	return ctrl.Result{}, multierr.Combine(errors...)
}

// SetupWithManager sets up the controller with the Manager.
func (r *InvitationRedeemReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&userv1.Invitation{}).
		Complete(r)
}

func addUserToTarget(ctx context.Context, c client.Client, user, prefix string, target userv1.TargetRef) error {
	o, err := targetref.GetTarget(ctx, c, target)
	if err != nil {
		return err
	}

	a, err := targetref.NewUserAccessor(o)
	if err != nil {
		return err
	}

	added := a.EnsureUser(prefix, user)
	if added {
		return c.Update(ctx, o)
	}

	return nil
}
