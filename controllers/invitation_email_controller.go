package controllers

import (
	"context"
	"fmt"
	"time"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/appuio/control-api/mailsenders"

	userv1 "github.com/appuio/control-api/apis/user/v1"
)

// InvitationEmailReconciler reconciles invitations and sends invitation emails if appropriate
type InvitationEmailReconciler struct {
	client.Client

	Recorder record.EventRecorder
	Scheme   *runtime.Scheme

	MailSender    mailsenders.MailSender
	RetryInterval time.Duration
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

	if inv.Status.Token == "" {
		return ctrl.Result{}, nil
	}

	if apimeta.IsStatusConditionTrue(inv.Status.Conditions, userv1.ConditionEmailSent) {
		return ctrl.Result{}, nil
	}
	condition := apimeta.FindStatusCondition(inv.Status.Conditions, userv1.ConditionEmailSent)

	if condition == nil {
		return ctrl.Result{}, nil
	}

	email := inv.Spec.Email
	id, err := r.MailSender.Send(ctx, email, inv.Name, inv.Status.Token)

	if err != nil {
		log.V(0).Error(err, "Error in e-mail backend")

		// Only update status if the error changes
		if condition.Reason != err.Error() {
			apimeta.SetStatusCondition(&inv.Status.Conditions, metav1.Condition{
				Type:   userv1.ConditionEmailSent,
				Status: metav1.ConditionFalse,
				Reason: err.Error(),
			})
			return ctrl.Result{}, r.Client.Status().Update(ctx, &inv)
		}
		return ctrl.Result{RequeueAfter: r.RetryInterval}, nil
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
		Complete(r)
}
