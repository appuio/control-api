package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	controlv1 "github.com/appuio/control-api/apis/v1"
)

var finalizer = "control-api.appuio.io/finalizer"

// UserReconciler reconciles a User object
type UserReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=appuio.io,resources=users,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=appuio.io,resources=users/status,verbs=get;update;patch

// Reconcile reacts on changes of users and mirrors these changes to Keycloak
func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(4).WithValues("request", req).Info("Reconciling")

	log.V(4).Info("Getting the User...")
	user := controlv1.User{}
	if err := r.Get(ctx, req.NamespacedName, &user); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !user.ObjectMeta.DeletionTimestamp.IsZero() {
		log.V(4).Info("Deleting RBAC...")
		if err := r.removeRBAC(ctx, user); err != nil {
			r.Recorder.Event(&user, "Warning", "DeletionFailed", "Failed to delete RBAC")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, r.removeFinalizer(ctx, &user)
	}
	if err := r.addFinalizer(ctx, &user); err != nil {
		return ctrl.Result{}, err
	}

	log.V(4).Info("Updating RBAC..")
	if err := r.setRBAC(ctx, user); err != nil {
		r.Recorder.Event(&user, "Warning", "UpdateFailed", "Failed to set RBAC")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *UserReconciler) setRBAC(ctx context.Context, user controlv1.User) error {
	return nil
}

func (r *UserReconciler) removeRBAC(ctx context.Context, user controlv1.User) error {
	return nil
}

func (r *UserReconciler) addFinalizer(ctx context.Context, user client.Object) error {
	if controllerutil.ContainsFinalizer(user, finalizer) {
		return nil
	}
	controllerutil.AddFinalizer(user, finalizer)
	return r.Update(ctx, user)
}

func (r *UserReconciler) removeFinalizer(ctx context.Context, user client.Object) error {
	if !controllerutil.ContainsFinalizer(user, finalizer) {
		return nil
	}
	controllerutil.RemoveFinalizer(user, finalizer)
	return r.Update(ctx, user)
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&controlv1.User{}).
		Complete(r)
}
