package api

import (
	"time"

	"currency-exchange/cache"
	"currency-exchange/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRouter(apiBaseURL string, cacheSvc *cache.CacheService) *chi.Mux {
	r := chi.NewRouter()

	r.Use(requestLogger)
	r.Use(corsMiddleware)
	r.Use(rateLimiter(100, time.Minute))
	r.Use(middleware.Recoverer)
	exchangeSvc := services.NewExchangeService(apiBaseURL, cacheSvc)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handleHealth(cacheSvc, apiBaseURL))
		r.Get("/rates", handleGetRates(exchangeSvc))
		r.Get("/rates/{base}", handleGetRatesByBase(exchangeSvc))
		r.Post("/convert", handleConvert(exchangeSvc))
		r.Get("/history", handleGetHistory())
	})

	return r
}
