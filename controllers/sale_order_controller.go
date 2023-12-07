package controllers

import (
	"context"
	"fmt"

	"go.uber.org/multierr"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	organizationv1 "github.com/appuio/control-api/apis/organization/v1"
	"github.com/appuio/control-api/controllers/saleorder"
)

// SaleOrderReconciler reconciles invitations and adds a token to the status if required.
type SaleOrderReconciler struct {
	client.Client

	Recorder record.EventRecorder
	Scheme   *runtime.Scheme

	SaleOrderStorage saleorder.SaleOrderStorage
}

//+kubebuilder:rbac:groups="rbac.appuio.io",resources=organizations,verbs=get;list;watch
//+kubebuilder:rbac:groups="user.appuio.io",resources=organizations,verbs=get;list;watch
//+kubebuilder:rbac:groups="rbac.appuio.io",resources=organizations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="user.appuio.io",resources=organizations/status,verbs=get;update;patch

// Reconcile reacts to Organizations and creates Sale Orders if necessary
func (r *SaleOrderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(1).WithValues("request", req).Info("Reconciling")

	org := organizationv1.Organization{}
	if err := r.Get(ctx, req.NamespacedName, &org); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if org.Spec.BillingEntityRef == "" {
		return ctrl.Result{}, nil
	}

	if org.Status.SaleOrderName != "" {
		return ctrl.Result{}, nil
	}

	if org.Status.SaleOrderID != "" {
		// ID is present, but Name is not. Update name.
		soName, err := r.SaleOrderStorage.GetSaleOrderName(org)
		if err != nil {
			log.V(0).Error(err, "Error getting sale order name")
			apimeta.SetStatusCondition(&org.Status.Conditions, metav1.Condition{
				Type:    organizationv1.ConditionSaleOrderNameUpdated,
				Status:  metav1.ConditionFalse,
				Reason:  organizationv1.ConditionReasonGetNameFailed,
				Message: err.Error(),
			})
			return ctrl.Result{}, multierr.Append(err, r.Client.Status().Update(ctx, &org))
		}
		apimeta.SetStatusCondition(&org.Status.Conditions, metav1.Condition{
			Type:   organizationv1.ConditionSaleOrderNameUpdated,
			Status: metav1.ConditionTrue,
		})
		org.Status.SaleOrderName = soName
		return ctrl.Result{}, r.Client.Status().Update(ctx, &org)
	}

	// Neither ID nor Name is present. Create new SO.
	soId, err := r.SaleOrderStorage.CreateSaleOrder(org)

	if err != nil {
		log.V(0).Error(err, "Error creating sale order")
		apimeta.SetStatusCondition(&org.Status.Conditions, metav1.Condition{
			Type:    organizationv1.ConditionSaleOrderCreated,
			Status:  metav1.ConditionFalse,
			Reason:  organizationv1.ConditionReasonCreateFailed,
			Message: err.Error(),
		})
		return ctrl.Result{}, multierr.Append(err, r.Client.Status().Update(ctx, &org))
	}

	apimeta.SetStatusCondition(&org.Status.Conditions, metav1.Condition{
		Type:   organizationv1.ConditionSaleOrderCreated,
		Status: metav1.ConditionTrue,
	})

	org.Status.SaleOrderID = fmt.Sprint(soId)
	return ctrl.Result{}, r.Client.Status().Update(ctx, &org)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SaleOrderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&organizationv1.Organization{}).
		Complete(r)
}
