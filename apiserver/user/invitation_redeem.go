package user

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	userv1 "github.com/appuio/control-api/apis/user/v1"
)

//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations/status,verbs=get;update;patch

var _ rest.Creater = &invitationRedeemer{}
var _ rest.Storage = &invitationRedeemer{}
var _ rest.Scoper = &invitationRedeemer{}

type invitationRedeemer struct {
	client client.Client

	usernamePrefix string
}

func (ir invitationRedeemer) NamespaceScoped() bool {
	return false
}

func (ir invitationRedeemer) New() runtime.Object {
	return &userv1.InvitationRedeemRequest{}
}

func (ir invitationRedeemer) Destroy() {}

// Create implements redeeming invitations, it accepts `InvitationRedeemRequest`.
// The user is identified by the username in the request context.
// If the invitation is valid, the invitation is marked as redeemed, the user, and a snapshot of the invitations's targets are stored in the status.
// The snapshot is later used in a controller to add the user to the targets in an idempotent and retryable way.
// If user or token are invalid, the request is rejected with a 403.
func (s *invitationRedeemer) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, opts *metav1.CreateOptions) (runtime.Object, error) {
	irr, ok := obj.(*userv1.InvitationRedeemRequest)
	if !ok {
		return nil, fmt.Errorf("not an InvitationRedeemRequest: %#v", obj)
	}

	name := irr.Name
	token := irr.Token

	l := klog.FromContext(ctx).WithName("InvitationRedeemer.Create").WithValues("invitation", name)

	inv := &userv1.Invitation{}
	if err := s.client.Get(ctx, client.ObjectKey{Name: name}, inv); err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	if inv.Status.Token == "" {
		l.Info("token is empty")
		return nil, errForbidden()
	}
	if !inv.Status.ValidUntil.After(time.Now()) {
		l.Info("invitation is expired")
		return nil, errForbidden()
	}
	if inv.IsRedeemed() {
		l.Info("invitation is already redeemed")
		return nil, errForbidden()
	}
	if inv.Status.Token != token {
		l.Info("token does not match")
		return nil, errForbidden()
	}

	user, ok := userFrom(ctx, s.usernamePrefix)
	if !ok {
		l.Info("no allowed user found in request context", "usernamePrefix", s.usernamePrefix)
		return nil, errForbidden()
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
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	return irr, nil
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

func errForbidden() *apierrors.StatusError {
	return &apierrors.StatusError{
		ErrStatus: metav1.Status{
			Status: metav1.StatusFailure,
			Code:   http.StatusForbidden,
			Reason: metav1.StatusReasonForbidden,
		}}
}
