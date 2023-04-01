package main

import (
	"fmt"
	"net/http"

	"github.com/chickenzord/ksei-exporter/internal/ksei"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	_ = godotenv.Overload(".env")

	registry := prometheus.NewRegistry()

	worker, err := ksei.NewWorkerFromEnv()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loaded %d KSEI accounts\n", len(worker.Accounts))

	if err := worker.Register(registry); err != nil {
		panic(err)
	}

	go func() {
		worker.WatchMetrics()
	}()

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
