package controllers

import (
	"context"
	"math/rand"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	orgv1 "github.com/appuio/control-api/apis/organization/v1"
)

// OrgBillingEntityNameCacheController reconciles OrganizationMembers resources
type OrgBillingEntityNameCacheController struct {
	client.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme

	RefreshInterval time.Duration
	RefreshJitter   time.Duration
}

//+kubebuilder:rbac:groups="organization.appuio.io",resources=organizations,verbs=get;list;watch
//+kubebuilder:rbac:groups="organization.appuio.io",resources=organizations/status,verbs=get;update;patch

// Reconcile reacts on changes of users and mirrors these changes to Keycloak
func (r *OrgBillingEntityNameCacheController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(4).WithValues("request", req).Info("Reconciling")

	var org orgv1.Organization
	if err := r.Get(ctx, req.NamespacedName, &org); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !org.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if org.Spec.BillingEntityRef == "" {
		return ctrl.Result{}, nil
	}

	var be billingv1.BillingEntity
	err := r.Client.Get(ctx, client.ObjectKey{Name: org.Spec.BillingEntityRef}, &be)
	if err != nil {
		return ctrl.Result{}, err
	}

	org.Status.BillingEntityName = be.Spec.Name
	if err := r.Client.Status().Update(ctx, &org); err != nil {
		return ctrl.Result{}, err
	}

	// We can't watch BillingEntity resources, so we have to requeue
	return ctrl.Result{RequeueAfter: r.requeueAfter()}, nil
}

func (r *OrgBillingEntityNameCacheController) requeueAfter() time.Duration {
	var jitter time.Duration
	if r.RefreshJitter > 0 {
		jitter = time.Duration(rand.Intn(int(r.RefreshJitter)))
	}
	return r.RefreshInterval + jitter
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrgBillingEntityNameCacheController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&orgv1.Organization{}).
		Complete(r)
}
