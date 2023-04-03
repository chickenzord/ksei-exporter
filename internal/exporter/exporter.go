package exporter

import (
	"fmt"
	"net/http"
	"time"

	"github.com/chickenzord/goksei"
	"github.com/chickenzord/ksei-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

var (
	portfolioTypes = []goksei.PortfolioType{
		goksei.EquityType,
		goksei.MutualFundType,
		goksei.BondType,
	}
)

type Exporter struct {
	accounts         []config.Account
	authStore        goksei.AuthStore
	registry         *prometheus.Registry
	metricAssetValue *prometheus.GaugeVec
}

func New(ksei config.KSEI) (*Exporter, error) {
	authStore, err := goksei.NewFileAuthStore(ksei.AuthDir)
	if err != nil {
		return nil, err
	}

	registry := prometheus.NewRegistry()
	metricAssetValue := prometheus.NewGaugeVec(
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
			"asset_symbol",
			"asset_name",
		},
	)

	if err := registry.Register(metricAssetValue); err != nil {
		return nil, err
	}

	return &Exporter{
		accounts:         ksei.Accounts,
		authStore:        authStore,
		registry:         registry,
		metricAssetValue: metricAssetValue,
	}, nil
}

func (e *Exporter) updateMetrics(a config.Account) error {
	c := goksei.NewClient(goksei.ClientOpts{
		AuthStore: e.authStore,
		Username:  a.Username,
		Password:  a.Password,
	})

	errs := errgroup.Group{}

	for _, t := range portfolioTypes {
		t := t

		errs.Go(func() error {
			var err error
			start := time.Now()

			defer func() {
				log.Debug().
					Str("account", a.Username).
					Str("type", t.Name()).
					TimeDiff("elapsed", time.Now(), start).
					Err(err).
					Msg("metrics updated")
			}()

			res, err := c.GetShareBalances(t)
			if err != nil {
				return err
			}

			for _, b := range res.Data {
				e.metricAssetValue.With(prometheus.Labels{
					"ksei_account":     a.Username,
					"security_account": b.Account,
					"security_name":    b.Participant,
					"currency":         b.Currency,
					"asset_type":       t.Name(),
					"asset_symbol":     b.Symbol(),
					"asset_name":       b.Name(),
				}).Set(b.CurrentValue())
			}

			return nil
		})
	}

	if err := errs.Wait(); err != nil {
		return err
	}

	return nil
}

func (e *Exporter) UpdateMetrics() error {
	errs := errgroup.Group{}

	for _, account := range e.accounts {
		account := account

		errs.Go(func() error {
			return e.updateMetrics(account)
		})
	}

	if err := errs.Wait(); err != nil {
		return err
	}

	return nil
}

func (e *Exporter) WatchMetrics() {
	if err := e.UpdateMetrics(); err != nil {
		fmt.Println(err)
	}

	for range time.Tick(5 * time.Minute) {
		if err := e.UpdateMetrics(); err != nil {
			fmt.Println(err)
		}
	}
}

func (e *Exporter) HTTPHandler() http.Handler {
	return promhttp.HandlerFor(e.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}