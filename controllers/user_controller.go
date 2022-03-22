package controllers

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// UserPrefix is the prefix applied to the user in the ClusterRoleBinding.subjects.name.
	UserPrefix string
}

//+kubebuilder:rbac:groups=appuio.io,resources=users,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=appuio.io,resources=users/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=subjects;rolebindings,verbs=get;list;create;update;patch

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
	if err := r.updateClusterRole(ctx, user); err != nil {
		return err
	}

	return r.updateClusterRoleBinding(ctx, user)
}

func (r *UserReconciler) removeRBAC(ctx context.Context, user controlv1.User) error {
	return nil
}

func (r *UserReconciler) updateClusterRole(ctx context.Context, user controlv1.User) error {
	cr := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName(user),
		},
	}
	op, err := ctrl.CreateOrUpdate(ctx, r.Client, &cr, func() error {
		cr.Rules = []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"appuio.io"},
				Resources:     []string{"users"},
				ResourceNames: []string{user.Name},
				Verbs:         []string{"get", "update", "patch"},
			},
		}
		return ctrl.SetControllerReference(&user, &cr, r.Scheme)
	})
	log.FromContext(ctx).V(4).Info("reconcile ClusterRole", "operation", op)
	return err
}

func (r *UserReconciler) updateClusterRoleBinding(ctx context.Context, user controlv1.User) error {
	crb := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName(user),
		},
	}
	op, err := ctrl.CreateOrUpdate(ctx, r.Client, &crb, func() error {
		crb.Subjects = []rbacv1.Subject{
			{
				APIGroup: rbacv1.GroupName,
				Kind:     "User",
				Name:     r.UserPrefix + user.Name,
			},
		}
		crb.RoleRef = rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     roleName(user),
		}
		return ctrl.SetControllerReference(&user, &crb, r.Scheme)
	})
	log.FromContext(ctx).V(4).Info("reconcile ClusterRoleBinding", "operation", op)
	return err
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

func roleName(user controlv1.User) string {
	return user.Name + "-owner"
}
