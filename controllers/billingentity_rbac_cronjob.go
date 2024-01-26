package controllers

import (
	"context"
	"fmt"

	"go.uber.org/multierr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	"github.com/appuio/control-api/pkg/billingrbac"
)

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;delete;patch;update;edit
// +kubebuilder:rbac:groups=rbac.appuio.io;billing.appuio.io,resources=billingentities,verbs=*

// BillingEntityRBACCronJob periodically checks billing entities and sends notification emails if appropriate
type BillingEntityRBACCronJob struct {
	client.Client
}

func NewBillingEntityRBACCronJob(client client.Client, eventRecorder record.EventRecorder) BillingEntityRBACCronJob {
	return BillingEntityRBACCronJob{
		Client: client,
	}
}

// Run lists all BillingEntity resources and sends notification emails if needed.
func (r *BillingEntityRBACCronJob) Run(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("BillingEntityRBACCronJob")
	log.Info("Reconciling BillingEntity RBAC")

	list := &billingv1.BillingEntityList{}
	err := r.Client.List(ctx, list)
	if err != nil {
		return fmt.Errorf("could not list billing entities: %w", err)
	}

	var errors []error
	for _, be := range list.Items {
		log := log.WithValues("billingentity", be.Name)
		err := r.reconcile(ctx, &be)
		if err != nil {
			log.Error(err, "could not reconcile billing entity")
			errors = append(errors, err)
		}
	}
	return multierr.Combine(errors...)
}

func (r *BillingEntityRBACCronJob) reconcile(ctx context.Context, be *billingv1.BillingEntity) error {
	ar, arBinding, vr, vrBinding := billingrbac.ClusterRoles(be.Name, billingrbac.ClusterRolesParams{
		AllowSubjectsToViewRole: true,
	})

	arErr := r.Client.Patch(ctx, ar, client.Apply, client.ForceOwnership, client.FieldOwner("control-api"))
	arBinding.Subjects = nil // we don't want to manage the subjects
	arBindingErr := r.Client.Patch(ctx, arBinding, client.Apply, client.ForceOwnership, client.FieldOwner("control-api"))
	vrErr := r.Client.Patch(ctx, vr, client.Apply, client.ForceOwnership, client.FieldOwner("control-api"))
	vrBinding.Subjects = nil // we don't want to manage the subjects
	vrBindingErr := r.Client.Patch(ctx, vrBinding, client.Apply, client.ForceOwnership, client.FieldOwner("control-api"))

	return multierr.Combine(arErr, arBindingErr, vrErr, vrBindingErr)
}
