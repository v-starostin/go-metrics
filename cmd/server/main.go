package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/v-starostin/go-metrics/internal/config"
	"github.com/v-starostin/go-metrics/internal/handler"
	"github.com/v-starostin/go-metrics/internal/repository"
	"github.com/v-starostin/go-metrics/internal/service"
)

func main() {
	cfg := config.New()
	repo := repository.New()
	srv := service.New(repo)
	h := handler.New(srv)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Method(http.MethodPost, "/update/{type}/{name}/{value}", h)
		r.Method(http.MethodGet, "/update/{type}/{name}", h)
		r.Method(http.MethodGet, "/value/{type}/{name}", h)
		r.Method(http.MethodGet, "/", h)
	})

	log.Printf("Server is listerning on port :%d", cfg.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r)
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
