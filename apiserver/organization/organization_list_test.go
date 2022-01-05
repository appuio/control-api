package organization

import (
	"context"
	"testing"

	mock "github.com/appuio/control-api/apiserver/organization/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

func TestOrganizationStorage_List(t *testing.T) {
	tests := map[string]struct {
		namespaces   *corev1.NamespaceList
		namespaceErr error

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

			organizations: &orgv1.OrganizationList{
				Items: []orgv1.Organization{
					*fooOrg,
					*barOrg,
				},
			},
		},
		"GivenErrNotFound_ThenErrNotFound": {
			namespaceErr: apierrors.NewNotFound(schema.GroupResource{
				Resource: "namepaces",
			}, "not-found"),
			err: apierrors.NewNotFound(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "not-found"),
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mnp := mock.NewMocknamespaceProvider(ctrl)
	os := organizationStorage{
		namepaces: mnp,
	}

	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			mnp.EXPECT().
				ListNamespaces(gomock.Any(), gomock.Any()).
				Return(tc.namespaces, tc.namespaceErr).
				Times(1)
			org, err := os.List(context.TODO(), nil)
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
			err: apierrors.NewNotFound(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "not-found"),
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mnp := mock.NewMocknamespaceProvider(ctrl)
	os := organizationStorage{
		namepaces: mnp,
	}

	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			nsWatcher := testWatcher{
				events: make(chan watch.Event, len(tc.namespacesEvents)),
			}
			for _, e := range tc.namespacesEvents {
				nsWatcher.events <- e
			}
			close(nsWatcher.events)
			mnp.EXPECT().
				WatchNamespaces(gomock.Any(), gomock.Any()).
				Return(nsWatcher, tc.namespaceErr).
				Times(1)
			orgWatch, err := os.Watch(context.TODO(), nil)
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
