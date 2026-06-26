package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"currency-exchange/api"
	"currency-exchange/cache"
	"currency-exchange/database"
	"currency-exchange/logger"
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
	logger.InitLogger()
	logger.Log.Info().Msg("starting currency exchange service")

	err := database.ConnectDB(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to connect to database")
		logger.Log.Info().Msg("running without database - history features wont work")
	} else {
		defer database.CloseDB()
		logger.Log.Info().Msg("connected to MSSQL database")
	}
	cacheSvc := cache.NewCacheService(redisAddr, redisPwd, 0)
	if cacheSvc != nil {
		defer cacheSvc.Close()
	}
	r := api.SetupRouter(apiBaseURL, cacheSvc)
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
		logger.Log.Info().Msgf("server listening on :%s", serverPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error().Err(err).Msg("server error")
			os.Exit(1)
		}
	}()

	<-quit
	fmt.Println()
	logger.Log.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error().Err(err).Msg("server forced to shutdown")
	}

	logger.Log.Info().Msg("server stopped")
}
