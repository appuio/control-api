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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestOrganizationStorage_Get(t *testing.T) {
	tests := map[string]struct {
		name string

		namespace    *corev1.Namespace
		namespaceErr error

		organization *orgv1.Organization
		err          error
	}{
		"GivenOrgNS_ThenOrg": {
			name:         "foo",
			namespace:    fooNs,
			organization: fooOrg,
		},
		"GivenErrNotFound_ThenErrNotFound": {
			name: "not-found",
			namespaceErr: apierrors.NewNotFound(schema.GroupResource{
				Resource: "namepaces",
			}, "not-found"),
			err: apierrors.NewNotFound(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "not-found"),
		},
		"GivenNonOrgNs_ThenErrNotFound": {
			name:      "default",
			namespace: defaultNs,
			err: apierrors.NewNotFound(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "default"),
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
				GetNamespace(gomock.Any(), tc.name, gomock.Any()).
				Return(tc.namespace, tc.namespaceErr).
				Times(1)
			org, err := os.Get(context.TODO(), tc.name, nil)
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

		organizationOut *orgv1.Organization
		err             error
	}{
		"GivenCreateOrg_ThenSuccess": {
			organizationIn:  fooOrg,
			organizationOut: fooOrg,
		},
		"GivenNsExists_ThenFail": {
			organizationIn: fooOrg,
			namespaceErr: apierrors.NewAlreadyExists(schema.GroupResource{
				Resource: "namepaces",
			}, "foo"),
			err: apierrors.NewAlreadyExists(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "foo"),
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
				CreateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tc.namespaceErr).
				Times(1)

			nopValidate := func(ctx context.Context, obj runtime.Object) error {
				return nil
			}
			org, err := os.Create(context.TODO(), tc.organizationIn, nopValidate, nil)

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
		skipNsDelete       bool
		namespaceDeleteErr error

		organization *orgv1.Organization
		err          error
	}{
		"GivenDeleteOrg_ThenSuccess": {
			name:         "foo",
			organization: fooOrg,
			namespace:    fooNs,
		},
		"GivenDeleteNonOrg_ThenFail": {
			name:         "default",
			namespace:    defaultNs,
			skipNsDelete: true,
			err: apierrors.NewNotFound(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "default"),
		},
		"GivenDeleteFails_ThenFail": {
			name:      "foo",
			namespace: fooNs,
			namespaceDeleteErr: apierrors.NewNotFound(schema.GroupResource{
				Resource: "namepaces",
			}, "foo"),
			err: apierrors.NewNotFound(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "foo"),
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
				GetNamespace(gomock.Any(), tc.name, gomock.Any()).
				Return(tc.namespace, tc.namespaceGetErr).
				Times(1)

			if !tc.skipNsDelete {
				mnp.EXPECT().
					DeleteNamespace(gomock.Any(), tc.name, gomock.Any()).
					Return(tc.namespace, tc.namespaceDeleteErr).
					Times(1)
			}

			nopValidate := func(ctx context.Context, obj runtime.Object) error {
				return nil
			}
			org, _, err := os.Delete(context.TODO(), tc.name, nopValidate, nil)

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
		skipNsUpdate       bool
		namespaceUpdateErr error

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
			name:         "default",
			namespace:    defaultNs,
			skipNsUpdate: true,
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
			err: apierrors.NewNotFound(schema.GroupResource{
				Group:    orgv1.GroupVersion.Group,
				Resource: "organizations",
			}, "foo"),
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
				GetNamespace(gomock.Any(), tc.name, gomock.Any()).
				Return(tc.namespace, tc.namespaceGetErr).
				Times(1)

			if !tc.skipNsUpdate {
				mnp.EXPECT().
					UpdateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(tc.namespaceUpdateErr).
					Times(1)
			}

			org, _, err := os.Update(context.TODO(), tc.name, testUpdateInfo(tc.updateFunc), nil, nil, false, nil)

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
