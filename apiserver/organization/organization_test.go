package organization

import (
	"fmt"

	mock "github.com/appuio/control-api/apiserver/organization/mock"
	"github.com/golang/mock/gomock"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authorization/authorizer"
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
