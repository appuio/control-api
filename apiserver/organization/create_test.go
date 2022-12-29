package organization

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"
	mock "github.com/appuio/control-api/apiserver/organization/mock"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/endpoints/request"
)

func TestOrganizationStorage_Create(t *testing.T) {
	tests := map[string]struct {
		userID         string
		userGroups     []string
		organizationIn *orgv1.Organization

		namespaceErr error

		authDecision authResponse

		memberName string

		skipRoleBindings bool
		organizationOut  *orgv1.Organization
		err              error
	}{
		"GivenCreateOrg_ThenSuccess": {
			userID:         "appuio#smith",
			organizationIn: fooOrg,
			authDecision: authResponse{
				decision: authorizer.DecisionAllow,
			},
			memberName:      "smith",
			organizationOut: fooOrg,
		},
		"GivenCreateOrgFromNonAPPUiOUser_ThenSuccessButNoSA": {
			userID:         "cluster-admin",
			organizationIn: fooOrg,
			authDecision: authResponse{
				decision: authorizer.DecisionAllow,
			},
			memberName:       "",
			organizationOut:  fooOrg,
			skipRoleBindings: true,
		},
		"GivenCreateOrgFromSA_ThenSuccessButNoSA": {
			userID:         "appuio#serviceacount",
			userGroups:     []string{"system:serviceaccounts"},
			organizationIn: fooOrg,
			authDecision: authResponse{
				decision: authorizer.DecisionAllow,
			},
			memberName:       "",
			organizationOut:  fooOrg,
			skipRoleBindings: true,
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
			skipRoleBindings: true,
		},
		"GivenAuthFails_ThenFail": {
			organizationIn: fooOrg,
			authDecision: authResponse{
				err: errors.New("failed"),
			},
			err:              errors.New("failed"),
			skipRoleBindings: true,
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
			skipRoleBindings: true,
		},
	}

	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			os, mnp, mauth := newMockedOrganizationStorage(t, ctrl)
			mrb := mock.NewMockroleBindingCreator(ctrl)
			ds := os.Storage().(*organizationStorage)
			ds.rbac = mrb
			mmemb := mock.NewMockmemberProvider(ctrl)
			ds.members = mmemb
			ds.usernamePrefix = "appuio#"
			nsOut := &corev1.Namespace{}
			if tc.organizationOut != nil {
				nsOut = tc.organizationOut.ToNamespace()
			}
			mauth.EXPECT().
				Authorize(gomock.Any(), isAuthRequest("create")).
				Return(tc.authDecision.decision, tc.authDecision.reason, tc.authDecision.err).
				Times(1)
			mnp.EXPECT().
				CreateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nsOut, tc.namespaceErr).
				AnyTimes()
			if !tc.skipRoleBindings {
				mrb.EXPECT().
					CreateRoleBindings(gomock.Any(), tc.organizationIn.Name, "appuio#"+tc.memberName).
					Return(nil).
					Times(1)
			}
			mmemb.EXPECT().
				CreateMembers(gomock.Any(), containsMemberAndOwner(tc.organizationIn.Name, tc.memberName)).
				Return(nil).
				AnyTimes()

			nopValidate := func(ctx context.Context, obj runtime.Object) error {
				return nil
			}
			org, err := os.Create(request.WithUser(request.WithRequestInfo(request.NewContext(),
				&request.RequestInfo{
					Verb:     "create",
					APIGroup: orgv1.GroupVersion.Group,
					Resource: "organizations",
					Name:     tc.organizationIn.Name,
				}), &user.DefaultInfo{
				Name:   tc.userID,
				Groups: tc.userGroups,
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

	tests := map[string]struct {
		failRoleBinding bool
		failMembers     bool
	}{
		"GivenRolebindingFails_ThenAbort": {
			failRoleBinding: true,
		},
		"GivenMembersFails_ThenAbort": {
			failMembers: true,
		},
	}
	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			os, mnp, mauth := newMockedOrganizationStorage(t, ctrl)
			mrb := mock.NewMockroleBindingCreator(ctrl)
			ds := os.Storage().(*organizationStorage)
			ds.rbac = mrb
			mmemb := mock.NewMockmemberProvider(ctrl)
			ds.members = mmemb
			ds.usernamePrefix = "appuio#"

			mauth.EXPECT().
				Authorize(gomock.Any(), isAuthRequest("create")).
				Return(authorizer.DecisionAllow, "", nil).
				Times(1)
			mnp.EXPECT().
				CreateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(fooNs, nil).
				Times(1)

			if tc.failRoleBinding {
				mrb.EXPECT().
					CreateRoleBindings(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("")).
					Times(1)
			} else {
				mrb.EXPECT().
					CreateRoleBindings(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				if tc.failMembers {
					mmemb.EXPECT().
						CreateMembers(gomock.Any(), gomock.Any()).
						Return(errors.New("")).
						Times(1)
				}
			}

			mnp.EXPECT().
				DeleteNamespace(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(fooNs, nil).
				Times(1)

			nopValidate := func(ctx context.Context, obj runtime.Object) error {
				return nil
			}
			_, err := os.Create(
				request.WithUser(
					request.WithRequestInfo(request.NewContext(),
						&request.RequestInfo{
							Verb:     "create",
							APIGroup: orgv1.GroupVersion.Group,
							Resource: "organizations",
							Name:     "foo",
						}),
					&user.DefaultInfo{
						Name: "appuio#foo",
					}),
				fooOrg, nopValidate, nil)

			require.Error(t, err)
		})
	}
}

type memberMatcher struct {
	owner string
	user  string
}

func (m memberMatcher) Matches(x interface{}) bool {
	mem, ok := x.(*controlv1.OrganizationMembers)
	if !ok {
		return ok
	}
	correctMembers := (len(mem.Spec.UserRefs) > 0 && mem.Spec.UserRefs[0].Name == m.user) || (m.user == "" && len(mem.Spec.UserRefs) == 0)
	return correctMembers &&
		len(mem.OwnerReferences) > 0 && mem.OwnerReferences[0].Name == m.owner
}

func (m memberMatcher) String() string {
	return fmt.Sprintf("contains %s and owned by %s", m.user, m.owner)
}

func containsMemberAndOwner(owner, user string) memberMatcher {
	return memberMatcher{user: user, owner: owner}
}
