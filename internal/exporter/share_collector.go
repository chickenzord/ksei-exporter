package exporter

import (
	"time"

	"github.com/chickenzord/goksei"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

type ShareCollector struct {
	client    *goksei.Client
	username  string
	portfolio goksei.PortfolioType

	assetValue *prometheus.GaugeVec
}

func NewShareCollector(client *goksei.Client, username string, portfolio goksei.PortfolioType) *ShareCollector {
	return &ShareCollector{
		client:    client,
		username:  username,
		portfolio: portfolio,

		assetValue: assetValue.MustCurryWith(prometheus.Labels{
			"ksei_account": username,
			"asset_type":   portfolio.Name(),
		}),
	}
}

func (c *ShareCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c.assetValue, ch)
}

func (c *ShareCollector) Collect(ch chan<- prometheus.Metric) {
	var err error

	start := time.Now()

	defer func() {
		log.Debug().
			Str("account", c.username).
			Str("type", c.portfolio.Name()).
			TimeDiff("elapsed", time.Now(), start).
			Err(err).
			Msg("asset value collected")

		labels := prometheus.Labels{
			"ksei_account": c.username,
			"endpoint":     "GetShareBalances/" + c.portfolio.Name(),
		}

		requestDuration.With(labels).Observe(time.Since(start).Seconds())

		m, err := requestDuration.MetricVec.GetMetricWith(labels)
		if err != nil {
			panic(err)
		}

		ch <- m
	}()

	shareBalances, err := c.client.GetShareBalances(c.portfolio)
	if err != nil {
		log.Err(err).Msg("error GetShareBalances")
		return
	}

	for _, b := range shareBalances.Data {
		subtype := ""

		if c.portfolio == goksei.MutualFundType {
			if mutualFund, ok := goksei.MutualFundByCode(b.Symbol()); ok {
				subtype = mutualFund.FundType
			} else {
				subtype = "unknown"
			}
		}

		assetValue := c.assetValue.With(prometheus.Labels{
			"security_account": b.Account,
			"security_name":    b.Participant,
			"currency":         b.Currency,
			"asset_subtype":    subtype,
			"asset_symbol":     b.Symbol(),
			"asset_name":       b.Name(),
		})
		assetValue.Set(b.CurrentValue())

		ch <- assetValue
	}
}
