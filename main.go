package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	serverPort = getEnv("SERVER_PORT", "8080")
	dbHost     = getEnv("DB_HOST", "localhost")
	dbPort     = getEnv("DB_PORT", "1433")
	dbUser     = getEnv("DB_USER", "sa")
	dbPassword = getEnv("DB_PASSWORD", "sqlPass!223!!")
	dbName     = getEnv("DB_NAME", "currency_exchange")
	redisAddr  = getEnv("REDIS_ADDR", "localhost:6379")
	redisPwd   = getEnv("REDIS_PASSWORD", "")
	apiBaseURL = getEnv("API_BASE_URL", "https://api.frankfurter.app")
)

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func main() {
	initLogger()
	log.Info().Msg("starting currency exchange service")

	err := connectDB(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to database")
		log.Info().Msg("running without database - history features wont work")
	} else {
		defer closeDB()
		log.Info().Msg("connected to MSSQL database")
	}
	cache := newCacheService(redisAddr, redisPwd, 0)
	if cache != nil {
		defer cache.Close()
	}
	r := setupRouter(apiBaseURL, cache)
	srv := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Msgf("server listening on :%s", serverPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("server error")
			os.Exit(1)
		}
	}()

	<-quit
	fmt.Println()
	log.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server stopped")
}
