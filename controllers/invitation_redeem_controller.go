package controllers

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/multierr"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"
)

// InvitationRedeemReconciler reconciles invitations and adds a token to the status if required.
type InvitationRedeemReconciler struct {
	client.Client

	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations/status,verbs=get;update;patch

//+kubebuilder:rbac:groups=appuio.io,resources=organizationmembers;teams,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings;rolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile reacts to redeemed invitations and adds the user to the targets listed in the invitation status.
func (r *InvitationRedeemReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(4).WithValues("request", req).Info("Reconciling")

	inv := userv1.Invitation{}
	if err := r.Get(ctx, req.NamespacedName, &inv); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !inv.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if !inv.IsRedeemed() {
		return ctrl.Result{}, nil
	}

	if inv.Status.RedeemedBy == "" {
		return ctrl.Result{}, errors.New("redeemed invitation has no user")
	}

	var errors []error
	for i := range inv.Status.TargetStatuses {
		if inv.Status.TargetStatuses[i].Condition.Status == metav1.ConditionTrue {
			continue
		}

		err := addUserToTarget(ctx, r.Client, inv.Status.RedeemedBy, inv.Status.TargetStatuses[i].TargetRef)
		if err != nil {
			errors = append(errors, err)
			inv.Status.TargetStatuses[i].Condition.LastTransitionTime = metav1.Now()
			inv.Status.TargetStatuses[i].Condition.Message = err.Error()
			inv.Status.TargetStatuses[i].Condition.Reason = metav1.StatusFailure
			inv.Status.TargetStatuses[i].Condition.Status = metav1.ConditionFalse
			continue
		}
		inv.Status.TargetStatuses[i].Condition.LastTransitionTime = metav1.Now()
		inv.Status.TargetStatuses[i].Condition.Message = ""
		inv.Status.TargetStatuses[i].Condition.Reason = metav1.StatusSuccess
		inv.Status.TargetStatuses[i].Condition.Status = metav1.ConditionTrue
	}

	err := r.Client.Status().Update(ctx, &inv)
	return ctrl.Result{}, multierr.Append(err, multierr.Combine(errors...))
}

// SetupWithManager sets up the controller with the Manager.
func (r *InvitationRedeemReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&userv1.Invitation{}).
		Complete(r)
}

func addUserToTarget(ctx context.Context, c client.Client, user string, target userv1.TargetRef) error {
	switch {
	case target.APIGroup == "appuio.io" && target.Kind == "OrganizationMembers":
		om := controlv1.OrganizationMembers{}
		if err := c.Get(ctx, client.ObjectKey{Name: target.Name, Namespace: target.Namespace}, &om); err != nil {
			return err
		}
		om.Spec.UserRefs, _ = ensure(om.Spec.UserRefs, controlv1.UserRef{Name: user})
		return c.Update(ctx, &om)
	case target.APIGroup == "appuio.io" && target.Kind == "Team":
		te := controlv1.Team{}
		if err := c.Get(ctx, client.ObjectKey{Name: target.Name, Namespace: target.Namespace}, &te); err != nil {
			return err
		}
		te.Spec.UserRefs, _ = ensure(te.Spec.UserRefs, controlv1.UserRef{Name: user})
		return c.Update(ctx, &te)
	case target.APIGroup == rbacv1.GroupName && target.Kind == "ClusterRoleBinding":
		crb := rbacv1.ClusterRoleBinding{}
		if err := c.Get(ctx, client.ObjectKey{Name: target.Name}, &crb); err != nil {
			return err
		}
		crb.Subjects, _ = ensure(crb.Subjects, newSubject(user))
		return c.Update(ctx, &crb)
	case target.APIGroup == rbacv1.GroupName && target.Kind == "RoleBinding":
		rb := rbacv1.RoleBinding{}
		if err := c.Get(ctx, client.ObjectKey{Name: target.Name, Namespace: target.Namespace}, &rb); err != nil {
			return err
		}
		rb.Subjects, _ = ensure(rb.Subjects, newSubject(user))
		return c.Update(ctx, &rb)
	}

	return fmt.Errorf("unsupported target %q.%q", target.APIGroup, target.Kind)
}

// ensure ensures that the given element is present in the given slice.
// If the element is already present, the original slice is returned.
// If the element is not present, a new slice is returned with the element appended.
func ensure[T comparable](s []T, e T) (ret []T, added bool) {
	for _, v := range s {
		if v == e {
			return s, false
		}
	}
	return append(s, e), true
}

func newSubject(user string) rbacv1.Subject {
	return rbacv1.Subject{
		Kind:     rbacv1.UserKind,
		APIGroup: rbacv1.GroupName,
		Name:     user,
	}
}
