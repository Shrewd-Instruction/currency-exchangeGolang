package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

func initLogger() {
	w := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	lvl := os.Getenv("LOG_LEVEL")
	if lvl == "" {
		lvl = "info"
	}

	level, err := zerolog.ParseLevel(lvl)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	log = zerolog.New(w).With().Timestamp().Logger()
}

var currencyRegex = regexp.MustCompile(`^[A-Z]{3}$`)

func validateConversionRequest(req *ConversionRequest) error {
	if req.From == "" {
		return fmt.Errorf("from currency is required")
	}
	if !currencyRegex.MatchString(req.From) {
		return fmt.Errorf("from currency must be 3 uppercase letters")
	}
	if req.To == "" {
		return fmt.Errorf("to currency is required")
	}
	if !currencyRegex.MatchString(req.To) {
		return fmt.Errorf("to currency must be 3 uppercase letters")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	return nil
}

func validateCurrencyCode(code string) bool {
	return currencyRegex.MatchString(code)
}

func validateLimit(limitStr string) (int, error) {
	if limitStr == "" {
		return 50, nil
	}
	n, err := strconv.Atoi(limitStr)
	if err != nil {
		return 0, fmt.Errorf("invalid limit value")
	}
	if n < 1 || n > 100 {
		return 0, fmt.Errorf("limit must be between 1 and 100")
	}
	return n, nil
}
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, APIError{Code: status, Message: msg})
}
