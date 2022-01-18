package organization

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/endpoints/request"
)

func TestOrganizationStorage_Delete(t *testing.T) {
	tests := map[string]struct {
		name string

		namespace          *corev1.Namespace
		namespaceGetErr    error
		namespaceDeleteErr error

		authDecision authResponse

		organization *orgv1.Organization
		err          error
	}{
		"GivenDeleteOrg_ThenSuccess": {
			name:               "foo",
			namespaceDeleteErr: nil,
			namespace:          fooNs,
			authDecision: authResponse{
				decision: authorizer.DecisionAllow,
			},
			organization: &orgv1.Organization{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			},
		},
		"GivenDeleteNonOrg_ThenFail": {
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
		"GivenDeleteFails_ThenFail": {
			name:      "foo",
			namespace: fooNs,
			authDecision: authResponse{
				decision: authorizer.DecisionAllow,
			},
			namespaceDeleteErr: apierrors.NewNotFound(schema.GroupResource{
				Resource: "namepaces",
			}, "foo"),
			err: apierrors.NewNotFound(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "foo"),
		},
		"GivenAuthFails_ThenFail": {
			name: "foo",
			authDecision: authResponse{
				err: errors.New("failed"),
			},
			err: errors.New("failed"),
		},
		"GivenForbidden_ThenForbidden": {
			name: "foo",
			authDecision: authResponse{
				decision: authorizer.DecisionDeny,
				reason:   "confidential",
			},
			err: apierrors.NewForbidden(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "foo", errors.New("confidential")),
		},
	}

	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			os, mnp, mauth := newMockedOrganizationStorage(ctrl)

			mauth.EXPECT().
				Authorize(gomock.Any(), isAuthRequest("get")).
				Return(authorizer.DecisionAllow, "", nil).
				AnyTimes()
			mauth.EXPECT().
				Authorize(gomock.Any(), isAuthRequest("delete")).
				Return(tc.authDecision.decision, tc.authDecision.reason, tc.authDecision.err).
				Times(1)
			mnp.EXPECT().
				GetNamespace(gomock.Any(), tc.name, gomock.Any()).
				Return(tc.namespace, tc.namespaceGetErr).
				AnyTimes()
			mnp.EXPECT().
				DeleteNamespace(gomock.Any(), tc.name, gomock.Any()).
				Return(tc.namespace, tc.namespaceDeleteErr).
				AnyTimes()

			nopValidate := func(ctx context.Context, obj runtime.Object) error {
				return nil
			}
			org, _, err := os.Delete(request.WithRequestInfo(request.NewContext(),
				&request.RequestInfo{
					Verb:     "delete",
					APIGroup: orgv1.GroupVersion.Group,
					Resource: "organizations",
					Name:     tc.name,
				}),
				tc.name, nopValidate, nil)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.organization, org)
		})
	}
}
