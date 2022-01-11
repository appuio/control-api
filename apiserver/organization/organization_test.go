package organization

import (
	"context"
	"errors"
	"fmt"
	"testing"

	mock "github.com/appuio/control-api/apiserver/organization/mock"
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

func newMockedOrganizationStorage(ctrl *gomock.Controller) (organizationStorage, *mock.MocknamespaceProvider, *mock.MockAuthorizer) {
	mnp := mock.NewMocknamespaceProvider(ctrl)
	mauth := mock.NewMockAuthorizer(ctrl)
	os := organizationStorage{
		namepaces:  mnp,
		authorizer: rbacAuthorizer{Authorizer: mauth},
	}
	return os, mnp, mauth
}

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
			os, mnp, mauth := newMockedOrganizationStorage(ctrl)

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
			mauth.EXPECT().
				Authorize(gomock.Any(), isAuthRequest("create")).
				Return(tc.authDecision.decision, tc.authDecision.reason, tc.authDecision.err).
				Times(1)
			mnp.EXPECT().
				CreateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tc.namespaceErr).
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
			name: "foo",
			authDecision: authResponse{
				decision: authorizer.DecisionAllow,
			},
			organization: fooOrg,
			namespace:    fooNs,
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

type testUpdateInfo func(obj runtime.Object) runtime.Object

func (_ testUpdateInfo) Preconditions() *metav1.Preconditions {
	return nil
}
func (ui testUpdateInfo) UpdatedObject(ctx context.Context, oldObj runtime.Object) (newObj runtime.Object, err error) {
	return ui(oldObj), nil
}

func TestOrganizationStorage_Update(t *testing.T) {
	tests := map[string]struct {
		name       string
		updateFunc func(obj runtime.Object) runtime.Object

		namespace          *corev1.Namespace
		namespaceGetErr    error
		namespaceUpdateErr error

		authDecision authResponse

		organization *orgv1.Organization
		err          error
	}{
		"GivenUpdateOrg_ThenSuccess": {
			name: "foo",
			updateFunc: func(obj runtime.Object) runtime.Object {
				org, ok := obj.(*orgv1.Organization)
				if !ok {
					return nil
				}
				org.Spec.DisplayName = "New Foo Inc."
				return org
			},

			namespace: fooNs,
			authDecision: authResponse{
				decision: authorizer.DecisionAllow,
			},

			organization: &orgv1.Organization{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "foo",
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				},
				Spec: orgv1.OrganizationSpec{
					DisplayName: "New Foo Inc.",
				},
			},
		},
		"GivenUpdateNonOrg_ThenFail": {
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
		"GivenUpdateFails_ThenFail": {
			name:      "foo",
			namespace: fooNs,
			updateFunc: func(obj runtime.Object) runtime.Object {
				org, ok := obj.(*orgv1.Organization)
				if !ok {
					return nil
				}
				org.Spec.DisplayName = "New Foo Inc."
				return org
			},
			namespaceUpdateErr: apierrors.NewNotFound(schema.GroupResource{
				Resource: "namepaces",
			}, "foo"),
			authDecision: authResponse{
				decision: authorizer.DecisionAllow,
			},
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
				Authorize(gomock.Any(), isAuthRequest("update")).
				Return(tc.authDecision.decision, tc.authDecision.reason, tc.authDecision.err).
				Times(1)

			mnp.EXPECT().
				GetNamespace(gomock.Any(), tc.name, gomock.Any()).
				Return(tc.namespace, tc.namespaceGetErr).
				AnyTimes()

			mnp.EXPECT().
				UpdateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tc.namespaceUpdateErr).
				AnyTimes()

			org, _, err := os.Update(request.WithRequestInfo(request.NewContext(),
				&request.RequestInfo{
					Verb:     "update",
					APIGroup: orgv1.GroupVersion.Group,
					Resource: "organizations",
					Name:     tc.name,
				}),
				tc.name, testUpdateInfo(tc.updateFunc), nil, nil, false, nil)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.organization, org)
		})
	}
}

// Some common test organizations
var (
	fooOrg = &orgv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "foo",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Spec: orgv1.OrganizationSpec{
			DisplayName: "Foo Inc.",
		},
	}
	fooNs = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
			Labels: map[string]string{
				orgv1.TypeKey: orgv1.OrgType,
			},
			Annotations: map[string]string{
				orgv1.DisplayNameKey: "Foo Inc.",
			},
		},
	}
	barOrg = &orgv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "bar",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Spec: orgv1.OrganizationSpec{
			DisplayName: "Bar Gmbh.",
		},
	}
	barNs = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
			Labels: map[string]string{
				orgv1.TypeKey: orgv1.OrgType,
			},
			Annotations: map[string]string{
				orgv1.DisplayNameKey: "Bar Gmbh.",
			},
		},
	}
	defaultNs = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
		},
	}
)

type authResponse struct {
	decision authorizer.Decision
	reason   string
	err      error
}

type authRequestMatcher struct {
	verb string
}

func (m authRequestMatcher) Matches(x interface{}) bool {
	attr, ok := x.(authorizer.Attributes)
	if !ok {
		return ok
	}
	return attr.GetVerb() == m.verb
}

func (m authRequestMatcher) String() string {
	return fmt.Sprintf("authenticates %s", m.verb)
}

func isAuthRequest(verb string) authRequestMatcher {
	return authRequestMatcher{verb: verb}
}
