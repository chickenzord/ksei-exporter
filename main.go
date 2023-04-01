package main

import (
	"fmt"
	"net/http"

	"github.com/chickenzord/ksei-exporter/internal/config"
	"github.com/chickenzord/ksei-exporter/internal/exporter"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Overload(".env")

	cfg, err := config.FromEnv()
	if err != nil {
		panic(err)
	}

	exp, err := exporter.New(cfg.KSEI)
	if err != nil {
		panic(err)
	}

	go func() {
		exp.WatchMetrics()
	}()

	r := chi.NewRouter()
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "pong")
	})
	r.Get("/metrics", exp.HTTPHandler().ServeHTTP)

	fmt.Println("Starting server")
	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
