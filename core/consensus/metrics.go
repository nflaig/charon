// Copyright © 2022-2023 Obol Labs Inc. Licensed under the terms of a Business Source License 1.1

package consensus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/obolnetwork/charon/app/promauto"
	"github.com/obolnetwork/charon/core"
)

var (
	decidedRoundsGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "core",
		Subsystem: "consensus",
		Name:      "decided_rounds",
		Help:      "Number of rounds it took to decide consensus instances by duty type.",
	}, []string{"duty"}) // Using gauge since the value changes slowly, once per slot.

	consensusDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "core",
		Subsystem: "consensus",
		Name:      "duration_seconds",
		Help:      "Duration of a consensus instance in seconds by duty",
		Buckets:   []float64{.05, .1, .25, .5, 1, 2.5, 5, 10, 20, 30, 60},
	}, []string{"duty"})

	consensusTimeout = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "core",
		Subsystem: "consensus",
		Name:      "timeout_total",
		Help:      "Total count of consensus timeouts by duty",
	}, []string{"duty"})

	consensusError = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "core",
		Subsystem: "consensus",
		Name:      "error_total",
		Help:      "Total count of consensus errors",
	})
)

func instrumentConsensus(duty core.Duty, round int64, startTime time.Time) {
	decidedRoundsGauge.WithLabelValues(duty.Type.String()).Set(float64(round))
	consensusDuration.WithLabelValues(duty.Type.String()).Observe(time.Since(startTime).Seconds())
}
