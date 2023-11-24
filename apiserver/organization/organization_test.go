package organization

import (
	"fmt"
	"testing"

	"github.com/appuio/control-api/apiserver/authwrapper"
	authmock "github.com/appuio/control-api/apiserver/authwrapper/mock"
	mock "github.com/appuio/control-api/apiserver/organization/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

func newMockedOrganizationStorage(t *testing.T, ctrl *gomock.Controller) (authwrapper.StandardStorage, *mock.MocknamespaceProvider, *authmock.MockAuthorizer) {
	t.Helper()

	mnp := mock.NewMocknamespaceProvider(ctrl)
	mauth := authmock.NewMockAuthorizer(ctrl)
	os, err := authwrapper.NewAuthorizedStorage(&organizationStorage{
		namepaces:               mnp,
		allowEmptyBillingEntity: true,
		impersonator:            &fakeImpersonator{t},
	}, metav1.GroupVersionResource{
		Group:    orgv1.GroupVersion.Group,
		Version:  orgv1.GroupVersion.Version,
		Resource: "organizations",
	}, mauth)
	require.NoError(t, err)
	return os.(authwrapper.StandardStorage), mnp, mauth
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

type fakeImpersonator struct{ t *testing.T }

func (c fakeImpersonator) Impersonate(u user.Info) (client.Client, error) {
	c.t.Helper()

	scheme := runtime.NewScheme()
	require.NoError(c.t, billingv1.AddToScheme(scheme))

	objs := []runtime.Object{
		&billingv1.BillingEntity{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
		},
		&billingv1.BillingEntity{
			ObjectMeta: metav1.ObjectMeta{
				Name: "bar",
			},
		},
	}

	return fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...).Build(), nil
}
