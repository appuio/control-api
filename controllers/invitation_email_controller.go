package controllers

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/multierr"
	"golang.org/x/time/rate"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/appuio/control-api/mailsenders"

	userv1 "github.com/appuio/control-api/apis/user/v1"
)

const ReasonSendFailed = "SendFailed"

// InvitationEmailReconciler reconciles invitations and sends invitation emails if appropriate
type InvitationEmailReconciler struct {
	client.Client

	Recorder record.EventRecorder
	Scheme   *runtime.Scheme

	MailSender     mailsenders.MailSender
	BaseRetryDelay time.Duration
}

//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations/status,verbs=get;update;patch

// Reconcile reacts to redeemed invitations and sends invitation emails to the user if needed.
func (r *InvitationEmailReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(1).WithValues("request", req).Info("Reconciling")

	inv := userv1.Invitation{}
	if err := r.Get(ctx, req.NamespacedName, &inv); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !inv.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if inv.Status.Token == "" || inv.Spec.Email == "" {
		return ctrl.Result{}, nil
	}

	if apimeta.IsStatusConditionTrue(inv.Status.Conditions, userv1.ConditionEmailSent) {
		return ctrl.Result{}, nil
	}

	email := inv.Spec.Email
	id, err := r.MailSender.Send(ctx, email, inv)
	if err != nil {
		log.V(0).Error(err, "Error in e-mail backend")

		apimeta.SetStatusCondition(&inv.Status.Conditions, metav1.Condition{
			Type:    userv1.ConditionEmailSent,
			Status:  metav1.ConditionFalse,
			Reason:  ReasonSendFailed,
			Message: err.Error(),
		})
		return ctrl.Result{}, multierr.Append(err, r.Client.Status().Update(ctx, &inv))
	}

	var message string
	if id != "" {
		message = fmt.Sprintf("Message ID: %s", id)
	}
	apimeta.SetStatusCondition(&inv.Status.Conditions, metav1.Condition{
		Type:    userv1.ConditionEmailSent,
		Status:  metav1.ConditionTrue,
		Message: message,
	})

	return ctrl.Result{}, r.Client.Status().Update(ctx, &inv)
}

// SetupWithManager sets up the controller with the Manager.
func (r *InvitationEmailReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&userv1.Invitation{}).
		WithOptions(controller.Options{
			RateLimiter: r.rateLimiter(),
		}).
		Complete(r)
}

func (r *InvitationEmailReconciler) rateLimiter() workqueue.RateLimiter {
	// This is the default rate limiter for controllers with higher baseDelay if the reconciliation fails.
	return workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(r.BaseRetryDelay, 5*time.Minute),
		// 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
	)
}
