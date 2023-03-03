package controllers

import (
	"context"

	"go.uber.org/multierr"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	controlv1 "github.com/appuio/control-api/apis/v1"
)

// OrganizationMembersReconciler reconciles OrganizationMembers resources
type OrganizationMembersReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme

	// UserPrefix is the prefix applied to the user in the RoleBinding.subjects.name.
	UserPrefix  string
	MemberRoles []string
}

//+kubebuilder:rbac:groups=appuio.io,resources=organizationmembers,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=appuio.io,resources=organizationmembers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Needed so that we are allowed to delegate common member roles
//+kubebuilder:rbac:groups="rbac.appuio.io",resources=organizations,verbs=get;list;watch;create;delete;patch;update
//+kubebuilder:rbac:groups="organization.appuio.io",resources=organizations,verbs=get;list;watch;create;delete;patch;update
//+kubebuilder:rbac:groups=appuio.io,resources=organizationmembers,verbs=get;list;watch;create;delete;patch;update
//+kubebuilder:rbac:groups="appuio.io",resources=teams,verbs=get;list;watch;create;delete;patch;update
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile reacts on changes of users and mirrors these changes to Keycloak
func (r *OrganizationMembersReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(1).WithValues("request", req).Info("Reconciling")

	memb := controlv1.OrganizationMembers{}
	if err := r.Get(ctx, req.NamespacedName, &memb); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !memb.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	var errGroup error
	for _, role := range r.MemberRoles {
		err := r.putRoleBinding(ctx, memb, role)
		if err != nil {
			errGroup = multierr.Append(errGroup, err)
			r.Recorder.Event(&memb, "Warning", "RBACUpdateFailed", "Failed to set RBAC for Organization members")
		}
	}

	return ctrl.Result{}, errGroup
}

func (r *OrganizationMembersReconciler) putRoleBinding(ctx context.Context, memb controlv1.OrganizationMembers, role string) error {
	rb := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      role,
			Namespace: memb.Namespace,
		},
	}
	op, err := ctrl.CreateOrUpdate(ctx, r.Client, &rb, func() error {
		sub := make([]rbacv1.Subject, len(memb.Spec.UserRefs))
		for i, ur := range memb.Spec.UserRefs {
			sub[i] = rbacv1.Subject{
				APIGroup: rbacv1.GroupName,
				Kind:     "User",
				Name:     r.UserPrefix + ur.Name,
			}
		}
		rb.Subjects = sub
		rb.RoleRef = rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     role,
		}
		return ctrl.SetControllerReference(&memb, &rb, r.Scheme)
	})
	log.FromContext(ctx).V(1).Info("reconcile RoleBinding", "operation", op)
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrganizationMembersReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&controlv1.OrganizationMembers{}).
		Owns(&rbacv1.RoleBinding{}).
		Complete(r)
}
