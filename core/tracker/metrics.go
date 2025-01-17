// Copyright © 2022-2023 Obol Labs Inc. Licensed under the terms of a Business Source License 1.1

package tracker

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/obolnetwork/charon/app/promauto"
)

var (
	participationGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "core",
		Subsystem: "tracker",
		Name:      "participation",
		Help:      "Set to 1 if peer participated successfully for the given duty or else 0",
	}, []string{"duty", "peer"})

	participationCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "core",
		Subsystem: "tracker",
		Name:      "participation_total",
		Help:      "Total number of successful participations by peer and duty type",
	}, []string{"duty", "peer"})

	failedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "core",
		Subsystem: "tracker",
		Name:      "failed_duties_total",
		Help:      "Total number of failed duties by type",
	}, []string{"duty"})

	unexpectedEventsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "core",
		Subsystem: "tracker",
		Name:      "unexpected_events_total",
		Help:      "Total number of unexpected events by peer",
	}, []string{"peer"})

	inconsistentCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "core",
		Subsystem: "tracker",
		Name:      "inconsistent_parsigs_total",
		Help:      "Total number of duties that contained inconsistent partial signed data by duty type",
	}, []string{"duty"})

	inclusionDelay = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "core",
		Subsystem: "tracker",
		Name:      "inclusion_delay",
		Help:      "Cluster's average attestation inclusion delay in slots",
	})
)
