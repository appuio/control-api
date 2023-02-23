package controllers

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"
)

//+kubebuilder:rbac:groups="rbac.appuio.io",resources=organizations,verbs=get;list;watch
//+kubebuilder:rbac:groups="organization.appuio.io",resources=organizations,verbs=get;list;watch

var desc = prometheus.NewDesc(
	"control_api_organization_billing_entity_ref",
	"Link between an organization and a billing entity",
	[]string{"organization", "billing_entity"},
	nil,
)

// OrgBillingRefLinkMetric is a Prometheus collector that exposes the link between an organization and a billing entity.
type OrgBillingRefLinkMetric struct {
	client.Client
}

var _ prometheus.Collector = &OrgBillingRefLinkMetric{}

// Describe implements prometheus.Collector.
// Sends the static description of the metric to the provided channel.
func (o *OrgBillingRefLinkMetric) Describe(ch chan<- *prometheus.Desc) {
	ch <- desc
}

// Collect implements prometheus.Collector.
// Sends a metric for each organization and its billing entity to the provided channel.
func (o *OrgBillingRefLinkMetric) Collect(ch chan<- prometheus.Metric) {
	orgs := &orgv1.OrganizationList{}
	if err := o.List(context.Background(), orgs); err != nil {
		ch <- prometheus.NewInvalidMetric(desc, err)
		return
	}

	for _, org := range orgs.Items {
		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			1,
			org.Name,
			org.Spec.BillingEntityRef,
		)
	}
}
