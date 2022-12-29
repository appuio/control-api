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
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/endpoints/request"
)

func TestOrganizationStorage_List(t *testing.T) {
	tests := map[string]struct {
		namespaces   *corev1.NamespaceList
		namespaceErr error

		authList authResponse
		authGet  []authResponse

		organizations *orgv1.OrganizationList
		err           error
	}{
		"GivenList_ThenSucceed": {
			namespaces: &corev1.NamespaceList{
				Items: []corev1.Namespace{
					*fooNs,
					*barNs,
				},
			},
			authList: authResponse{
				decision: authorizer.DecisionAllow,
			},
			authGet: []authResponse{
				{decision: authorizer.DecisionAllow},
				{decision: authorizer.DecisionAllow},
			},
			organizations: &orgv1.OrganizationList{
				Items: []orgv1.Organization{
					*fooOrg,
					*barOrg,
				},
			},
		},
		"GivenErrNotFound_ThenErrNotFound": {
			authList: authResponse{
				decision: authorizer.DecisionAllow,
			},
			namespaceErr: apierrors.NewNotFound(schema.GroupResource{
				Resource: "namepaces",
			}, "not-found"),
			err: apierrors.NewNotFound(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "not-found"),
		},
		"GivenForbidden_ThenForbidden": {
			authList: authResponse{
				decision: authorizer.DecisionDeny,
				reason:   "confidential",
			},
			err: apierrors.NewForbidden(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "", errors.New("confidential")),
		},
		"GivenList_ThenFilter": {
			namespaces: &corev1.NamespaceList{
				Items: []corev1.Namespace{
					*fooNs,
					*barNs,
				},
			},
			authList: authResponse{
				decision: authorizer.DecisionAllow,
			},
			authGet: []authResponse{
				{decision: authorizer.DecisionAllow},
				{decision: authorizer.DecisionDeny},
			},
			organizations: &orgv1.OrganizationList{
				Items: []orgv1.Organization{
					*fooOrg,
				},
			},
		},
	}
	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			os, mnp, mauth := newMockedOrganizationStorage(t, ctrl)

			mauth.EXPECT().
				Authorize(gomock.Any(), isAuthRequest("list")).
				Return(tc.authList.decision, tc.authList.reason, tc.authList.err).
				Times(1)
			mnp.EXPECT().
				ListNamespaces(gomock.Any(), gomock.Any()).
				Return(tc.namespaces, tc.namespaceErr).
				AnyTimes()

			for _, getAuth := range tc.authGet {
				mauth.EXPECT().
					Authorize(gomock.Any(), isAuthRequest("get")).
					Return(getAuth.decision, getAuth.reason, getAuth.err).
					Times(1)
			}

			org, err := os.List(request.WithRequestInfo(request.NewContext(),
				&request.RequestInfo{
					Verb:     "list",
					APIGroup: orgv1.GroupVersion.Group,
					Resource: "organizations",
				}), nil)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.organizations, org)
		})
	}
}

type testWatcher struct {
	events chan watch.Event
}

func (w testWatcher) Stop() {}

func (w testWatcher) ResultChan() <-chan watch.Event {
	return w.events
}

func TestOrganizationStorage_Watch(t *testing.T) {
	tests := map[string]struct {
		namespacesEvents []watch.Event
		namespaceErr     error

		authWatch authResponse
		authGet   []authResponse

		organizationEvents []watch.Event
		err                error
	}{
		"GivenNsEvents_ThenOrgEvents": {
			namespacesEvents: []watch.Event{
				{
					Type:   watch.Added,
					Object: fooNs,
				},
				{
					Type:   watch.Modified,
					Object: barNs,
				},
			},
			authWatch: authResponse{
				decision: authorizer.DecisionAllow,
			},
			authGet: []authResponse{
				{decision: authorizer.DecisionAllow},
				{decision: authorizer.DecisionAllow},
			},
			organizationEvents: []watch.Event{
				{
					Type:   watch.Added,
					Object: fooOrg,
				},
				{
					Type:   watch.Modified,
					Object: barOrg,
				},
			},
		},
		"GivenErrNotFound_ThenErrNotFound": {
			namespaceErr: apierrors.NewNotFound(schema.GroupResource{
				Resource: "namepaces",
			}, "not-found"),
			authWatch: authResponse{
				decision: authorizer.DecisionAllow,
			},
			err: apierrors.NewNotFound(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "not-found"),
		},
		"GivenForbidden_ThenForbidden": {
			authWatch: authResponse{
				decision: authorizer.DecisionDeny,
				reason:   "confidential",
			},
			err: apierrors.NewForbidden(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "", errors.New("confidential")),
		},
		"GivenNsEvents_ThenFilter": {
			namespacesEvents: []watch.Event{
				{
					Type:   watch.Added,
					Object: fooNs,
				},
				{
					Type:   watch.Modified,
					Object: barNs,
				},
				{
					Type:   watch.Modified,
					Object: fooNs,
				},
				{
					Type:   watch.Modified,
					Object: barNs,
				},
			},
			authWatch: authResponse{
				decision: authorizer.DecisionAllow,
			},
			authGet: []authResponse{
				{decision: authorizer.DecisionAllow},
				{decision: authorizer.DecisionDeny},
				{decision: authorizer.DecisionAllow},
				{decision: authorizer.DecisionDeny},
			},
			organizationEvents: []watch.Event{
				{
					Type:   watch.Added,
					Object: fooOrg,
				},
				{
					Type:   watch.Modified,
					Object: fooOrg,
				},
			},
		},
	}

	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			os, mnp, mauth := newMockedOrganizationStorage(t, ctrl)

			nsWatcher := testWatcher{
				events: make(chan watch.Event, len(tc.namespacesEvents)),
			}
			for _, e := range tc.namespacesEvents {
				nsWatcher.events <- e
			}
			close(nsWatcher.events)

			mauth.EXPECT().
				Authorize(gomock.Any(), isAuthRequest("watch")).
				Return(tc.authWatch.decision, tc.authWatch.reason, tc.authWatch.err).
				Times(1)

			mnp.EXPECT().
				WatchNamespaces(gomock.Any(), gomock.Any()).
				Return(nsWatcher, tc.namespaceErr).
				AnyTimes()

			for _, getAuth := range tc.authGet {
				mauth.EXPECT().
					Authorize(gomock.Any(), isAuthRequest("get")).
					Return(getAuth.decision, getAuth.reason, getAuth.err).
					Times(1)
			}
			orgWatch, err := os.Watch(request.WithRequestInfo(request.NewContext(),
				&request.RequestInfo{
					Verb:     "watch",
					APIGroup: orgv1.GroupVersion.Group,
					Resource: "organizations",
				}), nil)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				return
			}
			require.NoError(t, err)
			orgEvents := []watch.Event{}
			for e := range orgWatch.ResultChan() {
				orgEvents = append(orgEvents, e)
			}
			assert.Equal(t, tc.organizationEvents, orgEvents)
		})
	}

}
