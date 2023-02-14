package controllers

import (
	"context"
	"time"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	userv1 "github.com/appuio/control-api/apis/user/v1"
)

// InvitationTokenReconciler reconciles invitations and adds a token to the status if required.
type InvitationTokenReconciler struct {
	client.Client

	Recorder record.EventRecorder
	Scheme   *runtime.Scheme

	TokenValidFor time.Duration
}

//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations/status,verbs=get;update;patch

// Reconcile reacts on invitations and adds a token to the status if required.
func (r *InvitationTokenReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(4).WithValues("request", req).Info("Reconciling")

	log.V(4).Info("Getting the User...")
	inv := userv1.Invitation{}
	if err := r.Get(ctx, req.NamespacedName, &inv); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !inv.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if inv.Status.Token != "" {
		return ctrl.Result{}, nil
	}

	inv.Status.Token = uuid.New().String()
	inv.Status.ValidUntil = metav1.NewTime(time.Now().Add(r.TokenValidFor))
	if err := r.Status().Update(ctx, &inv); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InvitationTokenReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&userv1.Invitation{}).
		Complete(r)
}
