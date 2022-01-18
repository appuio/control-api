package organization

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	mock "github.com/appuio/control-api/apiserver/organization/mock"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/endpoints/request"
)

func TestOrganizationStorage_Create(t *testing.T) {
	tests := map[string]struct {
		organizationIn *orgv1.Organization

		namespaceErr error

		authDecision authResponse

		organizationOut *orgv1.Organization
		err             error
	}{
		"GivenCreateOrg_ThenSuccess": {
			organizationIn: fooOrg,
			authDecision: authResponse{
				decision: authorizer.DecisionAllow,
			},
			organizationOut: fooOrg,
		},
		"GivenNsExists_ThenFail": {
			organizationIn: fooOrg,
			authDecision: authResponse{
				decision: authorizer.DecisionAllow,
			},
			namespaceErr: apierrors.NewAlreadyExists(schema.GroupResource{
				Resource: "namepaces",
			}, "foo"),
			err: apierrors.NewAlreadyExists(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "foo"),
		},
		"GivenAuthFails_ThenFail": {
			organizationIn: fooOrg,
			authDecision: authResponse{
				err: errors.New("failed"),
			},
			err: errors.New("failed"),
		},
		"GivenForbidden_ThenForbidden": {
			organizationIn: fooOrg,
			authDecision: authResponse{
				decision: authorizer.DecisionDeny,
				reason:   "confidential",
			},
			err: apierrors.NewForbidden(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, fooOrg.Name, errors.New("confidential")),
		},
	}

	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			os, mnp, mauth := newMockedOrganizationStorage(ctrl)
			mrb := mock.NewMockroleBindingCreator(ctrl)
			os.rbac = mrb
			mauth.EXPECT().
				Authorize(gomock.Any(), isAuthRequest("create")).
				Return(tc.authDecision.decision, tc.authDecision.reason, tc.authDecision.err).
				Times(1)
			mnp.EXPECT().
				CreateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tc.namespaceErr).
				AnyTimes()
			mrb.EXPECT().
				CreateRoleBindings(gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()

			nopValidate := func(ctx context.Context, obj runtime.Object) error {
				return nil
			}
			org, err := os.Create(request.WithRequestInfo(request.NewContext(),
				&request.RequestInfo{
					Verb:     "create",
					APIGroup: orgv1.GroupVersion.Group,
					Resource: "organizations",
					Name:     tc.organizationIn.Name,
				}),
				tc.organizationIn, nopValidate, nil)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.organizationOut, org)
		})
	}
}

func TestOrganizationStorage_Create_Abort(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	os, mnp, mauth := newMockedOrganizationStorage(ctrl)
	mrb := mock.NewMockroleBindingCreator(ctrl)
	os.rbac = mrb
	mauth.EXPECT().
		Authorize(gomock.Any(), isAuthRequest("create")).
		Return(authorizer.DecisionAllow, "", nil).
		Times(1)
	mnp.EXPECT().
		CreateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)
	mrb.EXPECT().
		CreateRoleBindings(gomock.Any(), gomock.Any()).
		Return(errors.New("")).
		Times(1)
	mnp.EXPECT().
		DeleteNamespace(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(fooNs, nil).
		Times(1)

	nopValidate := func(ctx context.Context, obj runtime.Object) error {
		return nil
	}
	_, err := os.Create(request.WithRequestInfo(request.NewContext(),
		&request.RequestInfo{
			Verb:     "create",
			APIGroup: orgv1.GroupVersion.Group,
			Resource: "organizations",
			Name:     "foo",
		}),
		fooOrg, nopValidate, nil)

	require.Error(t, err)
}
