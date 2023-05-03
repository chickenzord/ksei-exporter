package exporter

import (
	"fmt"
	"time"

	"github.com/chickenzord/goksei"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

type CashCollector struct {
	client   *goksei.Client
	username string

	assetValue *prometheus.GaugeVec
}

func NewCashCollector(client *goksei.Client, username string) *CashCollector {
	return &CashCollector{
		client:   client,
		username: username,

		assetValue: assetValue.MustCurryWith(prometheus.Labels{
			"ksei_account": username,
			"asset_type":   goksei.CashType.Name(),
		}),
	}
}

func (c *CashCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c.assetValue, ch)
}

func (c *CashCollector) Collect(ch chan<- prometheus.Metric) {
	var err error

	start := time.Now()

	defer func() {
		log.Debug().
			Str("account", c.username).
			Str("type", goksei.CashType.Name()).
			TimeDiff("elapsed", time.Now(), start).
			Err(err).
			Msg("asset value collected")

		labels := prometheus.Labels{
			"ksei_account": c.username,
			"endpoint":     "GetCashBalances",
		}

		requestDuration.With(labels).Observe(time.Since(start).Seconds())

		m, err := requestDuration.MetricVec.GetMetricWith(labels)
		if err != nil {
			panic(err)
		}

		ch <- m
	}()

	cashBalances, err := c.client.GetCashBalances()
	if err != nil {
		log.Err(err).Msg("error GetCashBalances")
		return
	}

	for _, b := range cashBalances.Data {
		bankName := "Unknown Bank"

		if name, ok := goksei.CustodianBankNameByID(b.BankID); ok {
			bankName = name
		}

		assetValue := c.assetValue.With(prometheus.Labels{
			"security_account": b.AccountNumber,
			"security_name":    bankName,
			"currency":         b.Currency,
			"asset_subtype":    "rdn",
			"asset_symbol":     b.BankID,
			"asset_name":       fmt.Sprintf("Cash %s", bankName),
		})
		assetValue.Set(b.CurrentBalance())

		ch <- assetValue
	}
}
