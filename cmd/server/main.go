package main

import (
	"fmt"
	"log"
	"net/http"

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

	mux := http.NewServeMux()
	mux.Handle("/", h)

	log.Printf("Server is listerning on port :%d", cfg.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), mux)
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
