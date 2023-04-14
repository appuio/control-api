package controllers

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"

	userv1 "github.com/appuio/control-api/apis/user/v1"
)

//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations,verbs=get;list;watch
//+kubebuilder:rbac:groups="rbac.appuio.io",resources=invitations/status,verbs=get
//+kubebuilder:rbac:groups="user.appuio.io",resources=invitations/status,verbs=get

var emailPendingDesc = prometheus.NewDesc(
	"control_api_invitation_emails_pending_current",
	"Amount of e-mails that have not been sent yet",
	nil,
	nil,
)

// EmailPendingMetric is a Prometheus collector that exposes the number of currently pending invitation e-mails
type EmailPendingMetric struct {
	client.Client
}

var _ prometheus.Collector = &EmailPendingMetric{}

// Describe implements prometheus.Collector.
// Sends the static description of the metric to the provided channel.
func (e *EmailPendingMetric) Describe(ch chan<- *prometheus.Desc) {
	ch <- emailPendingDesc
}

// Collect implements prometheus.Collector.
// Sends a metric to the provided channel.
func (e *EmailPendingMetric) Collect(ch chan<- prometheus.Metric) {
	invs := &userv1.InvitationList{}

	if err := e.List(context.Background(), invs); err != nil {
		ch <- prometheus.NewInvalidMetric(emailPendingDesc, err)
		return
	}

	var count float64 = 0
	for _, inv := range invs.Items {
		if !apimeta.IsStatusConditionTrue(inv.Status.Conditions, userv1.ConditionEmailSent) {
			count++
		}
	}
	ch <- prometheus.MustNewConstMetric(
		emailPendingDesc,
		prometheus.GaugeValue,
		count,
	)
}
