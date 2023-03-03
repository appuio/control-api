package controllers

import (
	"context"
	"fmt"
	"time"

	//"errors"
	//"time"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	//"go.uber.org/multierr"
	//rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mailbackends "github.com/appuio/control-api/mailbackends"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	//controlv1 "github.com/appuio/control-api/apis/v1"
)

// InvitationEmailReconciler reconciles invitations and sends invitation emails if appropriate
type InvitationEmailReconciler struct {
	client.Client

	Recorder record.EventRecorder
	Scheme   *runtime.Scheme

	MailSender    mailbackends.MailSender
	RetryInterval time.Duration
}

//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations/status,verbs=get;update;patch

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

	if inv.Status.Token == "" {
		return ctrl.Result{}, nil
	}

	status := apimeta.FindStatusCondition(inv.Status.Conditions, userv1.ConditionEmailSent)

	if status == nil {
		// TODO remove after testing
		log.V(0).Info("Adding false email status to invite!")
		apimeta.SetStatusCondition(&inv.Status.Conditions, metav1.Condition{
			Type:   userv1.ConditionEmailSent,
			Status: metav1.ConditionFalse,
		})
		return ctrl.Result{}, r.Client.Status().Update(ctx, &inv)
	}

	if apimeta.IsStatusConditionTrue(inv.Status.Conditions, userv1.ConditionEmailSent) {
		return ctrl.Result{}, nil
	}
	condition := apimeta.FindStatusCondition(inv.Status.Conditions, userv1.ConditionEmailSent)

	email := inv.Spec.Email
	id, err := r.MailSender.Send(ctx, email, inv.Name, inv.Status.Token)

	if err != nil {
		log.V(0).Error(err, "Error in e-mail backend")
		if condition.Reason != err.Error() {
			// Only update status if the error changes
			apimeta.SetStatusCondition(&inv.Status.Conditions, metav1.Condition{
				Type:   userv1.ConditionEmailSent,
				Status: metav1.ConditionFalse,
				Reason: err.Error(),
			})
			return ctrl.Result{}, r.Client.Status().Update(ctx, &inv)
		}
		return ctrl.Result{RequeueAfter: r.RetryInterval}, nil
	}

	if id != "" {
		apimeta.SetStatusCondition(&inv.Status.Conditions, metav1.Condition{
			Type:    userv1.ConditionEmailSent,
			Status:  metav1.ConditionTrue,
			Message: fmt.Sprintf("Message ID: %s", id),
		})
	} else {
		apimeta.SetStatusCondition(&inv.Status.Conditions, metav1.Condition{
			Type:   userv1.ConditionEmailSent,
			Status: metav1.ConditionTrue,
		})
	}

	return ctrl.Result{}, r.Client.Status().Update(ctx, &inv)
}

// SetupWithManager sets up the controller with the Manager.
func (r *InvitationEmailReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&userv1.Invitation{}).
		Complete(r)
}
