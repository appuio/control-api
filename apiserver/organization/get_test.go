package organization

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/endpoints/request"
)

func TestOrganizationStorage_Get(t *testing.T) {
	tests := map[string]struct {
		name string

		namespace    *corev1.Namespace
		namespaceErr error

		authDecision authResponse

		organization *orgv1.Organization
		err          error
	}{
		"GivenOrgNS_ThenOrg": {
			name:         "foo",
			namespace:    fooNs,
			organization: fooOrg,
			authDecision: authResponse{
				decision: authorizer.DecisionAllow,
			},
		},
		"GivenErrNotFound_ThenErrNotFound": {
			name: "not-found",
			namespaceErr: apierrors.NewNotFound(schema.GroupResource{
				Resource: "namepaces",
			}, "not-found"),
			authDecision: authResponse{
				decision: authorizer.DecisionAllow,
			},
			err: apierrors.NewNotFound(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "not-found"),
		},
		"GivenNonOrgNs_ThenErrNotFound": {
			name:      "default",
			namespace: defaultNs,
			authDecision: authResponse{
				decision: authorizer.DecisionAllow,
			},
			err: apierrors.NewNotFound(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "default"),
		},
		"GivenAuthFails_ThenFail": {
			name: "auth-offline",
			authDecision: authResponse{
				err: errors.New("failed"),
			},
			err: errors.New("failed"),
		},
		"GivenForbidden_ThenForbidden": {
			name: "secret-org",
			authDecision: authResponse{
				decision: authorizer.DecisionDeny,
				reason:   "confidential",
			},
			err: apierrors.NewForbidden(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "secret-org", errors.New("confidential")),
		},
	}

	for n, tc := range tests {

		t.Run(n, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			os, mnp, mauth := newMockedOrganizationStorage(t, ctrl)

			mauth.EXPECT().
				Authorize(gomock.Any(), isAuthRequest("get")).
				Return(tc.authDecision.decision, tc.authDecision.reason, tc.authDecision.err).
				Times(1)
			mnp.EXPECT().
				GetNamespace(gomock.Any(), tc.name, gomock.Any()).
				Return(tc.namespace, tc.namespaceErr).
				AnyTimes()

			org, err := os.Get(request.WithRequestInfo(request.NewContext(),
				&request.RequestInfo{
					Verb:     "get",
					APIGroup: orgv1.GroupVersion.Group,
					Resource: "organizations",
					Name:     tc.name,
				}),
				tc.name, nil)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.organization, org)
		})
	}
}
