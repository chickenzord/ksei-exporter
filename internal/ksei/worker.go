package ksei

import (
	"fmt"
	"strings"
	"time"

	"github.com/chickenzord/goksei/pkg/goksei"
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
	UpdateInterval time.Duration
	Accounts       []Account

	balance *prometheus.GaugeVec
}

func (w *Worker) Register(p *prometheus.Registry) error {
	if p == nil {
		return fmt.Errorf("cannot use nil registry")
	}

	w.balance = prometheus.NewGaugeVec(
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

	if err := p.Register(w.balance); err != nil {
		return err
	}

	return nil
}

func (w *Worker) UpdateMetrics(a Account) error {
	c := goksei.NewClient(a.Username, a.Password)

	e := errgroup.Group{}

	for _, t := range portfolioTypes {
		t := t

		e.Go(func() error {
			res, err := c.GetShareBalances(t)
			if err != nil {
				return err
			}

			for _, b := range res.Data {
				w.balance.With(prometheus.Labels{
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
