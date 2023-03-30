package controllers

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/multierr"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	kstrings "k8s.io/utils/strings"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	"github.com/appuio/control-api/controllers/targetref"
)

// InvitationRedeemReconciler reconciles invitations and adds a token to the status if required.
type InvitationRedeemReconciler struct {
	client.Client

	Recorder record.EventRecorder
	Scheme   *runtime.Scheme

	UsernamePrefix string
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
	log.V(1).WithValues("request", req).Info("Reconciling")

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

	if err := r.createRedeemerRole(ctx, &inv); err != nil {
		return ctrl.Result{}, err
	}

	var errors []error
	statusHasChanged := false
	for i := range inv.Status.TargetStatuses {
		if inv.Status.TargetStatuses[i].Condition.Status == metav1.ConditionTrue {
			continue
		}

		statusHasChanged = true
		username := strings.TrimPrefix(inv.Status.RedeemedBy, r.UsernamePrefix)
		err := addUserToTarget(ctx, r.Client, username, r.UsernamePrefix, inv.Status.TargetStatuses[i].TargetRef)
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

	if statusHasChanged {
		err := r.Client.Status().Update(ctx, &inv)
		return ctrl.Result{}, multierr.Append(err, multierr.Combine(errors...))
	}
	return ctrl.Result{}, multierr.Combine(errors...)
}

// SetupWithManager sets up the controller with the Manager.
func (r *InvitationRedeemReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&userv1.Invitation{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&rbacv1.ClusterRole{}).
		Complete(r)
}

func addUserToTarget(ctx context.Context, c client.Client, user, prefix string, target userv1.TargetRef) error {
	o, err := targetref.GetTarget(ctx, c, target)
	if err != nil {
		return err
	}

	a, err := targetref.NewUserAccessor(o)
	if err != nil {
		return err
	}

	added := a.EnsureUser(prefix, user)
	if added {
		return c.Update(ctx, o)
	}

	return nil
}

func (r *InvitationRedeemReconciler) createRedeemerRole(ctx context.Context, inv *userv1.Invitation) error {
	rolename := invRedeemRoleName(inv.Name)

	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: rolename,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"rbac.appuio.io", "user.appuio.io"},
				Resources:     []string{"invitations"},
				Verbs:         []string{"get", "list", "watch"},
				ResourceNames: []string{inv.Name},
			},
		},
	}

	rolebinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: rolename,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "User",
				APIGroup: "rbac.authorization.k8s.io",
				Name:     inv.Status.RedeemedBy,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     rolename,
		},
	}

	for _, o := range [...]client.Object{role, rolebinding} {
		if err := controllerutil.SetControllerReference(inv, o, r.Scheme); err != nil {
			return fmt.Errorf("failed setting controller reference for %T/%s: %w", o, o.GetName(), err)
		}

		if err := r.Create(ctx, o); client.IgnoreAlreadyExists(err) != nil {
			if apierrors.IsAlreadyExists(err) {
				log.FromContext(ctx).Error(err, "object already exists while redeeming invitation", "invitation", inv.Name)
			} else {
				return fmt.Errorf("failed creating %T/%s: %w", o, o.GetName(), err)
			}
		}
	}

	return nil
}

func invRedeemRoleName(objName string) string {
	prefix := "invitations-"
	suffix := "-redeemer"

	if len(prefix)+len(suffix)+len(objName) <= 63 {
		return fmt.Sprintf("%s%s%s", prefix, objName, suffix)
	}

	h := sha1.New()
	h.Write([]byte(objName))
	hsh := kstrings.ShortenString(hex.EncodeToString(h.Sum(nil)), 7)

	maxLength := 63 - len(prefix) - len(suffix) - len(hsh) - 1
	maxSafe := kstrings.ShortenString(objName, maxLength)

	return fmt.Sprintf("%s%s-%s%s", prefix, maxSafe, hsh, suffix)
}
