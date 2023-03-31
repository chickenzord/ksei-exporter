package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/chickenzord/ksei-exporter/internal/ksei"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	_ = godotenv.Overload(".env")

	accounts := ksei.AccountsFromEnv()
	fmt.Printf("Loaded %d KSEI accounts\n", len(accounts))
	for _, acc := range accounts {
		fmt.Printf("%s\n", acc.Username)
	}

	registry := prometheus.NewRegistry()

	worker := &ksei.Worker{
		UpdateInterval: 5 * time.Minute,
		Accounts:       accounts,
	}

	if err := worker.Register(registry); err != nil {
		panic(err)
	}

	for _, account := range accounts {
		if err := worker.UpdateMetrics(account); err != nil {
			panic(err)
		}
	}

	r := chi.NewRouter()
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "pong")
	})
	r.Get("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP)

	fmt.Println("Starting server")
	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
