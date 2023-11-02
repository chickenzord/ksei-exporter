package exporter

import "github.com/prometheus/client_golang/prometheus"

var (
	assetValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "ksei",
			Name:      "asset_value",
			Help:      "Financial account balance current value",
		},
		[]string{
			"ksei_account",
			"security_account",
			"security_name",
			"currency",
			"asset_type",
			"asset_subtype",
			"asset_symbol",
			"asset_name",
		},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ksei",
			Name:      "client_request_duration_seconds",
			Help:      "KSEI client request duration",
			Buckets: []float64{
				0.5, 1, 2.5, 5, 10,
			},
		},
		[]string{
			"ksei_account",
			"endpoint",
		},
	)
)
