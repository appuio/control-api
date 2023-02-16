package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	"github.com/appuio/control-api/apiserver/user/mock"
)

//go:generate go run github.com/golang/mock/mockgen -destination=./mock/responder.go -package mock k8s.io/apiserver/pkg/registry/rest Responder

func TestConnect_Redeem_Success(t *testing.T) {
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

	executeRequest(t, subject, "redeeming-user", "/"+inv.Status.Token, http.StatusOK)

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

	executeRequest(t, subject, "redeeming-user", "/invalid", http.StatusForbidden)
}

func TestConnect_Redeem_Fail_InvalidUser(t *testing.T) {
	c := prepareTest(t, redeemableInvitation())
	subject := invitationRedeemer{
		client:         c,
		usernamePrefix: "appuio#",
	}

	executeRequest(t, subject, "redeeming-user", "/token", http.StatusForbidden)
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

	ctrl := gomock.NewController(t)
	r := mock.NewMockResponder(ctrl)
	defer ctrl.Finish()
	r.EXPECT().Object(expectedHTTPStatus, gomock.Any())

	reqCtx := request.WithUser(context.Background(), &user.DefaultInfo{
		Name: username,
	})
	h, err := subject.Connect(reqCtx, "subject", &userv1.RedeemOptions{
		Token: "/" + token,
	}, r)
	require.NoError(t, err)
	require.NotNil(t, h)
	req, err := http.NewRequestWithContext(reqCtx, "REDEEM", "", nil)
	require.NoError(t, err)
	h.ServeHTTP(httptest.NewRecorder(), req)
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
