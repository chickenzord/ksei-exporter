package main

import (
	"net/http"
	"os"

	"github.com/chickenzord/ksei-exporter/internal/config"
	"github.com/chickenzord/ksei-exporter/internal/exporter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	kseiExporter, err := exporter.New(cfg.KSEI)
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.StripSlashes)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: &log.Logger}))

	if creds := cfg.Server.BasicAuthCredentials(); len(creds) > 0 {
		r.Use(middleware.BasicAuth("ksei-exporter", cfg.Server.BasicAuthCredentials()))
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(kseiExporter)
	metricsHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics:   true,
		MaxRequestsInFlight: 1,
	})

	r.Get("/metrics", metricsHandler.ServeHTTP)

	log.Info().Msgf("server listening on %s", cfg.Server.BindAddress())

	if err := http.ListenAndServe(cfg.Server.BindAddress(), r); err != nil {
		panic(err)
	}
}
