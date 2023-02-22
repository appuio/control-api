package user

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

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

//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations/status,verbs=get;update;patch

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

// Connect implements the REDEEM method for invitations.
// It is used to redeem an invitation by a user.
// The user is identified by the username in the request context.
// The token is taken from the path.
// If the invitation is valid, the invitation is marked as redeemed, the user, and a snapshot of the invitations's targets are stored in the status.
// The snapshot is later used in a controller to add the user to the targets in an idempotent and retryable way.
// If user or token are invalid, the request is rejected with a 403.
func (s *invitationRedeemer) Connect(ctx context.Context, name string, options runtime.Object, responder rest.Responder) (http.Handler, error) {
	l := klog.FromContext(ctx).WithName("InvitationRedeemer.Connect").WithValues("invitation", name)
	opts := options.(*userv1.RedeemOptions)
	// Might come from the path, so we need to trim the leading slash
	token := strings.TrimLeft(opts.Token, "/")

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		inv := &userv1.Invitation{}
		if err := s.client.Get(ctx, client.ObjectKey{Name: name}, inv); err != nil {
			responder.Error(err)
			return
		}

		tokenValid := inv.Status.Token != "" && inv.Status.ValidUntil.After(time.Now())
		if inv.IsRedeemed() || !tokenValid || inv.Status.Token != token {
			l.Info("invalid token")
			forbidden(responder)
			return
		}

		user, ok := userFrom(ctx, s.usernamePrefix)
		if !ok {
			l.Info("no allowed user found in request context", "usernamePrefix", s.usernamePrefix)
			forbidden(responder)
			return
		}

		ts := make([]userv1.TargetStatus, len(inv.Spec.TargetRefs))
		for i, target := range inv.Spec.TargetRefs {
			ts[i] = userv1.TargetStatus{
				TargetRef: target,
				Condition: metav1.Condition{
					Type:   userv1.ConditionRedeemed,
					Status: metav1.ConditionUnknown,
				},
			}
		}

		inv.Status.TargetStatuses = ts
		inv.Status.RedeemedBy = user.GetName()
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

func forbidden(resp rest.Responder) {
	resp.Object(http.StatusForbidden, &metav1.Status{
		Status: metav1.StatusFailure,
		Code:   http.StatusForbidden,
		Reason: metav1.StatusReasonUnauthorized,
	})
}
