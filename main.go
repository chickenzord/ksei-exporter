package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/chickenzord/ksei-exporter/internal/config"
	"github.com/chickenzord/ksei-exporter/internal/exporter"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	_ = godotenv.Overload(".env")

	cfg, err := config.FromEnv()
	if err != nil {
		panic(err)
	}

	log.Level(zerolog.Level(zerolog.DebugLevel))
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Int("count", len(cfg.KSEI.Accounts)).Msg("KSEI accounts loaded")

	log.Info().Msg("initializing metrics")
	exp, err := exporter.New(cfg.KSEI)
	if err != nil {
		panic(err)
	}

	log.Info().Msg("starting background metrics updater")
	go func() {
		exp.WatchMetrics()
	}()

	r := chi.NewRouter()
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "pong")
	})
	r.Get("/metrics", exp.HTTPHandler().ServeHTTP)

	log.Info().Msgf("server listening on %s", cfg.Server.BindAddress())
	if err := http.ListenAndServe(cfg.Server.BindAddress(), r); err != nil {
		panic(err)
	}
}
