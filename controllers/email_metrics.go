package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
)

func newSuccessCounter(subsystem string) prometheus.Counter {
	return prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: subsystem,
		Name:      "sent_success_total",
		Help:      "Total number of successfully sent invitation e-mails",
	})
}

func newFailureCounter(subsystem string) prometheus.Counter {
	return prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: subsystem,
		Name:      "sent_failed_total",
		Help:      "Total number of invitation e-mails which failed to send",
	})
}
