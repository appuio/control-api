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

func TestOrgInfoMetric(t *testing.T) {
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
	}, &orgv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo-org",
		},
		Spec: orgv1.OrganizationSpec{
			BillingEntityRef: "be-234",
		},
		Status: orgv1.OrganizationStatus{
			SaleOrderName: "SO9999",
		},
	})

	require.NoError(t,
		testutil.CollectAndCompare(&controllers.OrgInfoMetric{c}, strings.NewReader(`
# HELP appuio_control_organization_info Information about APPUiO Cloud organizations
# TYPE appuio_control_organization_info gauge
appuio_control_organization_info{billing_entity="be-1734",organization="blub-org",sales_order=""} 1
appuio_control_organization_info{billing_entity="be-234",organization="foo-org",sales_order="SO9999"} 1
appuio_control_organization_info{billing_entity="test-billing-entity",organization="test-org",sales_order=""} 1
`),
			"appuio_control_organization_info"),
	)
}
