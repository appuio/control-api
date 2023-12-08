package controllers

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"
)

//+kubebuilder:rbac:groups="rbac.appuio.io",resources=organizations,verbs=get;list;watch
//+kubebuilder:rbac:groups="organization.appuio.io",resources=organizations,verbs=get;list;watch

var orgInfoMetricDesc = prometheus.NewDesc(
	"appuio_control_organization_info",
	"Information about APPUiO Cloud organizations",
	[]string{"organization", "billing_entity", "sales_order"},
	nil,
)

// OrgInfoMetric is a Prometheus collector that exposes the link between an organization and a billing entity.
type OrgInfoMetric struct {
	client.Client
}

var _ prometheus.Collector = &OrgInfoMetric{}

// Describe implements prometheus.Collector.
// Sends the static description of the metric to the provided channel.
func (o *OrgInfoMetric) Describe(ch chan<- *prometheus.Desc) {
	ch <- orgInfoMetricDesc
}

// Collect implements prometheus.Collector.
// Sends a metric for each organization and its billing entity to the provided channel.
func (o *OrgInfoMetric) Collect(ch chan<- prometheus.Metric) {
	orgs := &orgv1.OrganizationList{}
	if err := o.List(context.Background(), orgs); err != nil {
		ch <- prometheus.NewInvalidMetric(desc, err)
		return
	}

	for _, org := range orgs.Items {
		ch <- prometheus.MustNewConstMetric(
			orgInfoMetricDesc,
			prometheus.GaugeValue,
			1,
			org.Name,
			org.Spec.BillingEntityRef,
			org.Status.SaleOrderName,
		)
	}
}
