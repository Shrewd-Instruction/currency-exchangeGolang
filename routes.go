package main

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func setupRouter(apiBaseURL string, cache *CacheService) *chi.Mux {
	r := chi.NewRouter()

	r.Use(requestLogger)
	r.Use(corsMiddleware)
	r.Use(rateLimiter(100, time.Minute))
	r.Use(middleware.Recoverer)
	exchangeSvc := newExchangeService(apiBaseURL, cache)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handleHealth(cache, apiBaseURL))
		r.Get("/rates", handleGetRates(exchangeSvc))
		r.Get("/rates/{base}", handleGetRatesByBase(exchangeSvc))
		r.Post("/convert", handleConvert(exchangeSvc))
		r.Get("/history", handleGetHistory())
	})

	return r
}
