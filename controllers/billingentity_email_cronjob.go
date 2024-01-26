package controllers

import (
	"context"
	"fmt"

	"go.uber.org/multierr"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/appuio/control-api/mailsenders"
	"github.com/prometheus/client_golang/prometheus"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
)

// BillingEntityEmailCronJob periodically checks billing entities and sends notification emails if appropriate
type BillingEntityEmailCronJob struct {
	client.Client

	Recorder record.EventRecorder

	MailSender mailsenders.MailSender

	mailRecipientAddress string

	failureCounter prometheus.Counter
	successCounter prometheus.Counter
}

func NewBillingEntityEmailCronJob(client client.Client, eventRecorder record.EventRecorder, mailSender mailsenders.MailSender, MailRecipientAddress string) BillingEntityEmailCronJob {
	return BillingEntityEmailCronJob{
		Client:               client,
		Recorder:             eventRecorder,
		MailSender:           mailSender,
		mailRecipientAddress: MailRecipientAddress,
		failureCounter:       newFailureCounter("control_api_billingentity_emails"),
		successCounter:       newSuccessCounter("control_api_billingentity_emails"),
	}

}

func (r *BillingEntityEmailCronJob) GetMetrics() prometheus.Collector {
	reg := prometheus.NewRegistry()
	reg.MustRegister(r.failureCounter)
	reg.MustRegister(r.successCounter)
	return reg
}

//+kubebuilder:rbac:groups="rbac.appuio.io",resources=billingentities,verbs=get;list;update;patch
//+kubebuilder:rbac:groups="billing.appuio.io",resources=billingentities,verbs=get;list;update;patch
//+kubebuilder:rbac:groups="rbac.appuio.io",resources=billingentities/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="billing.appuio.io",resources=billingentities/status,verbs=get;update;patch

// Run lists all BillingEntity resources and sends notification emails if needed.
func (r *BillingEntityEmailCronJob) Run(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.V(1).Info("Sending Billing Entity E-Mails...")

	list := &billingv1.BillingEntityList{}

	err := r.Client.List(ctx, list)

	if err != nil {
		return err
	}

	var errors []error
	for _, be := range list.Items {
		if apimeta.FindStatusCondition(be.Status.Conditions, billingv1.ConditionEmailSent) == nil || apimeta.IsStatusConditionFalse(be.Status.Conditions, billingv1.ConditionEmailSent) {
			err = r.sendEmailAndUpdateStatus(ctx, be)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}
	return multierr.Combine(errors...)
}

func (r *BillingEntityEmailCronJob) sendEmailAndUpdateStatus(ctx context.Context, be billingv1.BillingEntity) error {
	log := log.FromContext(ctx)
	id, err := r.MailSender.Send(ctx, r.mailRecipientAddress, be)
	if err != nil {
		log.V(0).Error(err, "Error in e-mail backend")
		r.failureCounter.Add(1)
		apimeta.SetStatusCondition(&be.Status.Conditions, metav1.Condition{
			Type:    billingv1.ConditionEmailSent,
			Status:  metav1.ConditionFalse,
			Reason:  billingv1.ConditionReasonSendFailed,
			Message: err.Error(),
		})
	} else {
		r.successCounter.Add(1)

		var message string
		if id != "" {
			message = fmt.Sprintf("Message ID: %s", id)
		}
		apimeta.SetStatusCondition(&be.Status.Conditions, metav1.Condition{
			Type:    billingv1.ConditionEmailSent,
			Status:  metav1.ConditionTrue,
			Message: message,
		})
	}

	return r.Client.Update(ctx, &be)
}
