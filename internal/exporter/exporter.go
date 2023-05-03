package exporter

import (
	"github.com/chickenzord/goksei"
	"github.com/chickenzord/ksei-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	accounts  []config.Account
	authStore goksei.AuthStore
}

func New(ksei config.KSEI) (*Exporter, error) {
	authStore, err := goksei.NewFileAuthStore(ksei.AuthDir)
	if err != nil {
		return nil, err
	}

	return &Exporter{
		accounts:  ksei.Accounts,
		authStore: authStore,
	}, nil
}

func (e *Exporter) Collectors() []prometheus.Collector {
	portfolioTypes := []goksei.PortfolioType{
		goksei.EquityType,
		goksei.MutualFundType,
		goksei.BondType,
	}

	collectors := []prometheus.Collector{}

	for _, account := range e.accounts {
		client := goksei.NewClient(goksei.ClientOpts{
			AuthStore: e.authStore,
			Username:  account.Username,
			Password:  account.Password,
		})

		for _, portfolioType := range portfolioTypes {
			collectors = append(collectors, NewShareCollector(client, account.Username, portfolioType))
		}

		collectors = append(collectors, NewCashCollector(client, account.Username))
	}

	return collectors
}
