package ksei

import (
	"fmt"
	"strings"
	"time"

	"github.com/chickenzord/goksei"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

var (
	portfolioTypes = []goksei.PortfolioType{
		goksei.EquityType,
		goksei.MutualFundType,
		goksei.BondType,
	}
)

type Worker struct {
	Accounts []Account

	AuthDir         string        `envconfig:"auth_dir"`
	RefreshInterval time.Duration `envconfig:"refresh_interval" default:"1h"`
	RefreshJitter   float32       `envconfig:"refresh_jitter" default:"0.2"`

	authStore        goksei.AuthStore
	metricAssetValue *prometheus.GaugeVec
}

func NewWorkerFromEnv() (*Worker, error) {
	var worker Worker

	if err := envconfig.Process("ksei", &worker); err != nil {
		return nil, err
	}

	authStore, err := goksei.NewFileAuthStore(worker.AuthDir)
	if err != nil {
		return nil, err
	}

	worker.authStore = authStore
	worker.Accounts = accountsFromEnv()

	return &worker, nil
}

func (w *Worker) Register(p *prometheus.Registry) error {
	if p == nil {
		return fmt.Errorf("cannot use nil registry")
	}

	w.metricAssetValue = prometheus.NewGaugeVec(
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

	if err := p.Register(w.metricAssetValue); err != nil {
		return err
	}

	return nil
}

func (w *Worker) updateMetrics(a Account) error {
	c := goksei.NewClient(goksei.ClientOpts{
		AuthStore: w.authStore,
		Username:  a.Username,
		Password:  a.Password,
	})

	e := errgroup.Group{}

	for _, t := range portfolioTypes {
		t := t

		e.Go(func() error {
			res, err := c.GetShareBalances(t)
			if err != nil {
				return err
			}

			for _, b := range res.Data {
				w.metricAssetValue.With(prometheus.Labels{
					"ksei_account":     a.Username,
					"security_account": b.Account,
					"security_name":    b.Participant,
					"currency":         b.Currency,
					"asset_type":       t.Name(),
					"asset_symbol":     b.Symbol(),
					"asset_name":       strings.Split(b.FullName, " - ")[1], // TODO create helper func in goksei lib
				}).Set(b.CurrentValue())
			}

			return nil
		})
	}

	return e.Wait()
}

func (w *Worker) UpdateMetrics() error {
	e := errgroup.Group{}

	for _, account := range w.Accounts {
		account := account

		e.Go(func() error {
			return w.updateMetrics(account)
		})
	}

	return e.Wait()
}

func (w *Worker) WatchMetrics() {
	if err := w.UpdateMetrics(); err != nil {
		fmt.Println(err)
	}

	for range time.Tick(w.RefreshInterval) {
		if err := w.UpdateMetrics(); err != nil {
			fmt.Println(err)
		}
	}
}
