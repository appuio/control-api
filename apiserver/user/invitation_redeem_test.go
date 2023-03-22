package user

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	userv1 "github.com/appuio/control-api/apis/user/v1"
)

func TestCreate_Redeem_Success(t *testing.T) {
	target := userv1.TargetRef{
		APIGroup:  "appuio.io",
		Kind:      "Team",
		Name:      "team",
		Namespace: "team-namespace",
	}

	inv := redeemableInvitation()
	inv.Spec.TargetRefs = []userv1.TargetRef{target}

	c := prepareTest(t, inv)
	subject := invitationRedeemer{client: c}

	executeRequest(t, subject, "redeeming-user", inv.Status.Token, http.StatusOK)

	require.NoError(t, c.Get(context.Background(), client.ObjectKeyFromObject(inv), inv))
	assert.True(t, inv.IsRedeemed())
	assert.Equal(t, "redeeming-user", inv.Status.RedeemedBy)
	assert.Equal(t, []userv1.TargetStatus{
		{
			TargetRef: target,
			Condition: metav1.Condition{
				Type:   userv1.ConditionRedeemed,
				Status: metav1.ConditionUnknown,
			},
		},
	}, inv.Status.TargetStatuses)
}

func TestConnect_Redeem_Fail_InvalidToken(t *testing.T) {
	c := prepareTest(t, redeemableInvitation())
	subject := invitationRedeemer{client: c}

	executeRequest(t, subject, "redeeming-user", "invalid", http.StatusForbidden)
}

func TestConnect_Redeem_Fail_InvalidUser(t *testing.T) {
	c := prepareTest(t, redeemableInvitation())
	subject := invitationRedeemer{
		client:         c,
		usernamePrefix: "appuio#",
	}

	executeRequest(t, subject, "redeeming-user", "token", http.StatusForbidden)
}

func TestConnect_Redeem_Fail_Expired(t *testing.T) {
	inv := redeemableInvitation()
	inv.Status.ValidUntil = metav1.NewTime(metav1.Now().Add(-time.Hour))
	c := prepareTest(t, inv)
	subject := invitationRedeemer{client: c}

	executeRequest(t, subject, "redeeming-user", "token", http.StatusForbidden)
}

// The token is added asynchronously, so it might not be available yet.
func TestConnect_Redeem_Fail_NoTokenYet(t *testing.T) {
	inv := redeemableInvitation()
	inv.Status.Token = ""
	inv.Status.ValidUntil = metav1.Time{}
	c := prepareTest(t, inv)
	subject := invitationRedeemer{client: c}

	executeRequest(t, subject, "redeeming-user", "token", http.StatusForbidden)
}

func TestConnect_Redeem_Fail_AlreadyRedeemed(t *testing.T) {
	inv := redeemableInvitation()
	apimeta.SetStatusCondition(&inv.Status.Conditions, metav1.Condition{
		Type:   userv1.ConditionRedeemed,
		Status: metav1.ConditionTrue,
	})
	c := prepareTest(t, inv)
	subject := invitationRedeemer{client: c}

	executeRequest(t, subject, "redeeming-user", "token", http.StatusForbidden)
}

func redeemableInvitation() *userv1.Invitation {
	return &userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
		Status: userv1.InvitationStatus{
			Token:      "token",
			ValidUntil: metav1.NewTime(metav1.Now().Add(time.Hour)),
		},
	}
}

func executeRequest(t *testing.T, subject invitationRedeemer, username, token string, expectedHTTPStatus int) {
	t.Helper()

	reqCtx := request.WithUser(context.Background(), &user.DefaultInfo{
		Name: username,
	})
	h, err := subject.Create(reqCtx, &userv1.InvitationRedeemRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
		Token: token,
	}, nil, &metav1.CreateOptions{})

	if expectedHTTPStatus == http.StatusOK {
		require.NotNil(t, h)
		require.NoError(t, err)
		return
	}

	require.Error(t, err)
	if assert.IsType(t, &apierrors.StatusError{}, err) {
		status := err.(*apierrors.StatusError)
		assert.Equal(t, expectedHTTPStatus, int(status.Status().Code))
	}
}

func prepareTest(t *testing.T, initObjs ...client.Object) client.WithWatch {
	scheme := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(scheme))
	require.NoError(t, userv1.AddToScheme(scheme))

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjs...).
		Build()
}
