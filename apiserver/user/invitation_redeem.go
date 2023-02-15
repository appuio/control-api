package user

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	controlv1 "github.com/appuio/control-api/apis/v1"
	"go.uber.org/multierr"
	rbacv1 "k8s.io/api/rbac/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	"github.com/appuio/control-api/apiserver/secretstorage"
)

var _ rest.Connecter = &invitationRedeemer{}
var _ rest.StandardStorage = &invitationRedeemer{}
var _ rest.Scoper = &invitationRedeemer{}

type invitationRedeemer struct {
	secretstorage.ScopedStandardStorage
	client client.Client

	usernamePrefix string
}

func (ir *invitationRedeemer) ConnectMethods() []string {
	return []string{"REDEEM"}
}

func (ir *invitationRedeemer) NewConnectOptions() (runtime.Object, bool, string) {
	// Adds the token from the path to the options under the field "token"
	return &userv1.RedeemOptions{}, true, "token"
}

func (s *invitationRedeemer) Connect(ctx context.Context, name string, options runtime.Object, responder rest.Responder) (http.Handler, error) {
	opts := options.(*userv1.RedeemOptions)

	l := klog.FromContext(ctx).WithName("InvitationRedeemer.Connect")
	l.V(1).Info("called", "name", name, "token", opts.Token)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Might come from the path, so we need to trim the leading slash
		token := strings.TrimLeft(opts.Token, "/")

		inv := &userv1.Invitation{}
		if err := s.client.Get(ctx, client.ObjectKey{Name: name}, inv); err != nil {
			responder.Error(err)
			return
		}

		tokenValid := inv.Status.Token != "" && inv.Status.ValidUntil.After(time.Now())
		if inv.IsRedeemed() || !tokenValid || inv.Status.Token != token {
			unauthorized(responder)
			return
		}

		user, ok := userFrom(ctx, s.usernamePrefix)
		if !ok {
			unauthorized(responder)
			return
		}

		{
			var errs []error
			for _, target := range inv.Spec.TargetRefs {
				err := addUserToTargetGroup(ctx, s.client, user.GetName(), target, true)
				if err != nil {
					errs = append(errs, err)
				}
			}
			if err := multierr.Combine(errs...); err != nil {
				l.Error(err, "failed to add user to target groups (dry-run)")
				responder.Error(fmt.Errorf("failed to add user to target groups (dry-run): %w", err))
				return
			}
		}

		{
			var errs []error
			for _, target := range inv.Spec.TargetRefs {
				err := addUserToTargetGroup(ctx, s.client, user.GetName(), target, false)
				if err != nil {
					errs = append(errs, err)
				}
			}
			if err := multierr.Combine(errs...); err != nil {
				l.Error(err, "failed to add user to target groups")
				responder.Error(fmt.Errorf("failed to add user to target groups: %w", err))
				return
			}
		}

		apimeta.SetStatusCondition(&inv.Status.Conditions, metav1.Condition{
			Type:    userv1.ConditionRedeemed,
			Status:  metav1.ConditionTrue,
			Reason:  userv1.ConditionRedeemed,
			Message: fmt.Sprintf("Redeemed by %q", user.GetName()),
		})

		if err := s.client.Status().Update(ctx, inv); err != nil {
			responder.Error(err)
			return
		}

		responder.Object(http.StatusOK, &metav1.Status{
			Status: metav1.StatusSuccess,
		})
	}), nil
}

// userFrom returns the user from the context if it is a non-serviceaccount user and has the usernamePrefix.
func userFrom(ctx context.Context, usernamePrefix string) (u user.Info, ok bool) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return nil, false
	}
	if !strings.HasPrefix(user.GetName(), usernamePrefix) {
		return nil, false
	}

	for _, u := range user.GetGroups() {
		if u == "system:serviceaccounts" {
			return nil, false
		}
	}

	return user, true
}

func unauthorized(resp rest.Responder) {
	resp.Object(http.StatusUnauthorized, &metav1.Status{
		Status: metav1.StatusFailure,
		Code:   http.StatusUnauthorized,
		Reason: metav1.StatusReasonUnauthorized,
	})
}

func addUserToTargetGroup(ctx context.Context, c client.Client, user string, target userv1.TargetRef, dryrun bool) error {
	var updateOpts []client.UpdateOption
	if dryrun {
		updateOpts = append(updateOpts, client.DryRunAll)
	}

	switch {
	case target.APIGroup == "appuio.io" && target.Kind == "OrganizationMembers":
		om := controlv1.OrganizationMembers{}
		if err := c.Get(ctx, client.ObjectKey{Name: target.Name, Namespace: target.Namespace}, &om); err != nil {
			return err
		}
		om.Spec.UserRefs, _ = ensure(om.Spec.UserRefs, controlv1.UserRef{Name: user})
		return c.Update(ctx, &om, updateOpts...)
	case target.APIGroup == "appuio.io" && target.Kind == "Team":
		te := controlv1.Team{}
		if err := c.Get(ctx, client.ObjectKey{Name: target.Name, Namespace: target.Namespace}, &te); err != nil {
			return err
		}
		te.Spec.UserRefs, _ = ensure(te.Spec.UserRefs, controlv1.UserRef{Name: user})
		return c.Update(ctx, &te, updateOpts...)
	case target.APIGroup == "rbac.authorization.k8s.io" && target.Kind == "ClusterRoleBinding":
		crb := rbacv1.ClusterRoleBinding{}
		if err := c.Get(ctx, client.ObjectKey{Name: target.Name, Namespace: target.Namespace}, &crb); err != nil {
			return err
		}
		crb.Subjects, _ = ensure(crb.Subjects, newSubject(user))
		return c.Update(ctx, &crb, updateOpts...)
	case target.APIGroup == "rbac.authorization.k8s.io" && target.Kind == "RoleBinding":
		rb := rbacv1.RoleBinding{}
		if err := c.Get(ctx, client.ObjectKey{Name: target.Name, Namespace: target.Namespace}, &rb); err != nil {
			return err
		}
		rb.Subjects, _ = ensure(rb.Subjects, newSubject(user))
		return c.Update(ctx, &rb, updateOpts...)
	}

	return fmt.Errorf("unsupported target %s.%s", target.APIGroup, target.Kind)
}

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
