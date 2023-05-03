package exporter

import (
	"fmt"
	"time"

	"github.com/chickenzord/goksei"
	"github.com/chickenzord/ksei-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
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
	accounts  []config.Account
	authStore goksei.AuthStore

	assetValue      *prometheus.GaugeVec
	requestDuration *prometheus.HistogramVec
}

func New(ksei config.KSEI) (*Exporter, error) {
	authStore, err := goksei.NewFileAuthStore(ksei.AuthDir)
	if err != nil {
		return nil, err
	}

	return &Exporter{
		accounts:  ksei.Accounts,
		authStore: authStore,

		assetValue: prometheus.NewGaugeVec(
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
		),

		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "ksei",
				Name:      "client_request_duration",
				Help:      "KSEI client request duration",
				Buckets: []float64{
					0.5, 1, 2.5, 5, 10,
				},
			},
			[]string{
				"ksei_account",
				"endpoint",
			},
		),
	}, nil
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.assetValue.Describe(ch)
	e.requestDuration.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	errs := errgroup.Group{}

	for _, account := range e.accounts {
		account := account

		c := goksei.NewClient(goksei.ClientOpts{
			AuthStore: e.authStore,
			Username:  account.Username,
			Password:  account.Password,
		})

		// GetCashBalances
		errs.Go(func() error {
			var err error
			start := time.Now()

			defer func() {
				log.Debug().
					Str("account", account.Username).
					Str("type", goksei.CashType.Name()).
					TimeDiff("elapsed", time.Now(), start).
					Err(err).
					Msg("asset value collected")

				labels := prometheus.Labels{
					"ksei_account": account.Username,
					"endpoint":     "GetCashBalances",
				}

				e.requestDuration.With(labels).Observe(time.Since(start).Seconds())

				m, err := e.requestDuration.MetricVec.GetMetricWith(labels)
				if err != nil {
					panic(err)
				}

				ch <- m
			}()

			cashBalances, err := c.GetCashBalances()
			if err != nil {
				return &Error{
					Account: account.Username,
					Method:  "GetCashBalances",
					Cause:   err,
				}
			}

			for _, b := range cashBalances.Data {
				bankName := "Unknown Bank"

				if name, ok := goksei.CustodianBankNameByID(b.BankID); ok {
					bankName = name
				}

				assetValue := e.assetValue.With(prometheus.Labels{
					"ksei_account":     account.Username,
					"security_account": b.AccountNumber,
					"security_name":    bankName,
					"currency":         b.Currency,
					"asset_type":       goksei.CashType.Name(),
					"asset_subtype":    "rdn",
					"asset_symbol":     b.BankID,
					"asset_name":       fmt.Sprintf("Cash %s", bankName),
				})
				assetValue.Set(b.CurrentBalance())

				ch <- assetValue
			}

			return nil
		})

		// GetShareBalances
		for _, t := range portfolioTypes {
			t := t

			errs.Go(func() error {
				var err error
				start := time.Now()

				defer func() {
					log.Debug().
						Str("account", account.Username).
						Str("type", t.Name()).
						TimeDiff("elapsed", time.Now(), start).
						Err(err).
						Msg("asset value collected")

					labels := prometheus.Labels{
						"ksei_account": account.Username,
						"endpoint":     "GetShareBalances/" + t.Name(),
					}

					e.requestDuration.With(labels).Observe(time.Since(start).Seconds())

					m, err := e.requestDuration.MetricVec.GetMetricWith(labels)
					if err != nil {
						panic(err)
					}

					ch <- m
				}()

				shareBalances, err := c.GetShareBalances(t)
				if err != nil {
					return &Error{
						Account: account.Username,
						Method:  "GetShareBalances",
						Params:  []string{t.Name()},
						Cause:   err,
					}
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

					assetValue := e.assetValue.With(prometheus.Labels{
						"ksei_account":     account.Username,
						"security_account": b.Account,
						"security_name":    b.Participant,
						"currency":         b.Currency,
						"asset_type":       t.Name(),
						"asset_subtype":    subtype,
						"asset_symbol":     b.Symbol(),
						"asset_name":       b.Name(),
					})
					assetValue.Set(b.CurrentValue())

					ch <- assetValue
				}

				return nil
			})
		}
	}

	if err := errs.Wait(); err != nil {
		log.Err(err).Msg("error collecting metrics")
	}
}
