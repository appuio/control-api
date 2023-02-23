package controllers_test

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	"github.com/appuio/control-api/controllers"
)

func TestOrgBillingRefLinkMetric(t *testing.T) {
	c := prepareTest(t, &orgv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-org",
		},
		Spec: orgv1.OrganizationSpec{
			BillingEntityRef: "test-billing-entity",
		},
	}, &orgv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "blub-org",
		},
		Spec: orgv1.OrganizationSpec{
			BillingEntityRef: "be-1734",
		},
	})

	require.NoError(t,
		testutil.CollectAndCompare(&controllers.OrgBillingRefLinkMetric{c}, strings.NewReader(`
# HELP control_api_organization_billing_entity_ref Link between an organization and a billing entity
# TYPE control_api_organization_billing_entity_ref gauge
control_api_organization_billing_entity_ref{billing_entity="be-1734",organization="blub-org"} 1
control_api_organization_billing_entity_ref{billing_entity="test-billing-entity",organization="test-org"} 1
`),
			"control_api_organization_billing_entity_ref"),
	)
}
