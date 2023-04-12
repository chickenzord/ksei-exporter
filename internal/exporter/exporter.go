package exporter

import (
	"fmt"
	"math/rand"
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

	refreshInterval time.Duration
	refreshJitter   float32
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
			"asset_subtype",
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
		refreshInterval:  ksei.RefreshInterval,
		refreshJitter:    ksei.RefreshJitter,
	}, nil
}

func (e *Exporter) updateMetrics(a config.Account) error {
	c := goksei.NewClient(goksei.ClientOpts{
		AuthStore: e.authStore,
		Username:  a.Username,
		Password:  a.Password,
	})

	errs := errgroup.Group{}

	errs.Go(func() error {
		cashBalances, err := c.GetCashBalances()
		if err != nil {
			return err
		}

		for _, b := range cashBalances.Data {
			bankName := "Unknown Bank"

			if name, ok := goksei.CustodianBankNameByID(b.BankID); ok {
				bankName = name
			}

			e.metricAssetValue.With(prometheus.Labels{
				"ksei_account":     a.Username,
				"security_account": b.AccountNumber,
				"security_name":    bankName,
				"currency":         b.Currency,
				"asset_type":       goksei.CashType.Name(),
				"asset_subtype":    "rdn",
				"asset_symbol":     b.BankID,
				"asset_name":       fmt.Sprintf("Cash %s", bankName),
			}).Set(b.CurrentBalance())
		}
		return nil
	})

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

			shareBalances, err := c.GetShareBalances(t)
			if err != nil {
				return err
			}

			for _, b := range shareBalances.Data {
				subtype := ""

				if t == goksei.MutualFundType {
					if mutualFund, ok := goksei.MutualFundByCode(b.Symbol()); ok {
						subtype = mutualFund.FundType
					} else {
						subtype = "unknown"
					}
				}

				e.metricAssetValue.With(prometheus.Labels{
					"ksei_account":     a.Username,
					"security_account": b.Account,
					"security_name":    b.Participant,
					"currency":         b.Currency,
					"asset_type":       t.Name(),
					"asset_subtype":    subtype,
					"asset_symbol":     b.Symbol(),
					"asset_name":       b.Name(),
				}).Set(b.CurrentValue())
			}

			return nil
		})
	}

	return errs.Wait()
}

func (e *Exporter) UpdateMetrics() error {
	errs := errgroup.Group{}

	for _, account := range e.accounts {
		account := account

		errs.Go(func() error {
			return e.updateMetrics(account)
		})
	}

	return errs.Wait()
}

func (e *Exporter) WatchMetrics() {
	for {
		if err := e.UpdateMetrics(); err != nil {
			log.Err(err).Msg("error updating metrics")
		}

		delay := e.refreshInterval + time.Duration(rand.Float32()*e.refreshJitter*float32(e.refreshInterval.Nanoseconds()))

		log.Debug().Msgf("sleeping %s", delay)
		time.Sleep(delay)
	}
}

func (e *Exporter) HTTPHandler() http.Handler {
	return promhttp.HandlerFor(e.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}
