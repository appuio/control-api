package controllers

import (
	"context"
	"time"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	userv1 "github.com/appuio/control-api/apis/user/v1"
)

// InvitationCleanupReconciler reconciles invitations, deleting them if appropriate
type InvitationCleanupReconciler struct {
	client.Client

	Recorder record.EventRecorder
	Scheme   *runtime.Scheme

	RedeemedInvitationTTL time.Duration
}

//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations/status,verbs=get
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations/status,verbs=get

// Reconcile reacts on invitations and removes them if required
func (r *InvitationCleanupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(1).WithValues("request", req).Info("Reconciling")

	log.V(1).Info("Getting the Invitation...")
	inv := userv1.Invitation{}
	if err := r.Get(ctx, req.NamespacedName, &inv); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !inv.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if inv.Status.Token == "" {
		// Invitation is not yet valid
		return ctrl.Result{}, nil
	}
	now := metav1.NewTime(time.Now())

	if inv.IsRedeemed() {
		cond := apimeta.FindStatusCondition(inv.Status.Conditions, userv1.ConditionRedeemed)
		ttlExpirationTime := cond.LastTransitionTime.Add(r.RedeemedInvitationTTL)
		if ttlExpirationTime.Before(now.Time) {
			log.V(1).Info("Redeemed Invitation TTL expired - deleting", "ttlExpirationTime", ttlExpirationTime)
			return ctrl.Result{}, r.Delete(ctx, &inv)
		}
		return ctrl.Result{RequeueAfter: ttlExpirationTime.Sub(now.Add(-time.Minute))}, nil
	}

	if inv.Status.ValidUntil.Before(&now) {
		log.V(1).Info("Invitation expired - deleting")
		return ctrl.Result{}, r.Delete(ctx, &inv)
	}

	return ctrl.Result{RequeueAfter: inv.Status.ValidUntil.Sub(now.Add(-time.Minute))}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InvitationCleanupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&userv1.Invitation{}).
		Complete(r)
}
