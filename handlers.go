package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func handleGetRates(svc *ExchangeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base := r.URL.Query().Get("base")
		if base == "" {
			base = "USD"
		}

		data, err := svc.getLatestRates(base)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, data)
	}
}

func handleGetRatesByBase(svc *ExchangeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base := chi.URLParam(r, "base")

		data, err := svc.getLatestRates(base)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, data)
	}
}

func handleConvert(svc *ExchangeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ConversionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		err = validateConversionRequest(&req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		rate, err := svc.getConversionRate(req.From, req.To)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		result := req.Amount * rate

		resp := ConversionResponse{
			From:      req.From,
			To:        req.To,
			Amount:    req.Amount,
			Result:    result,
			Rate:      rate,
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if db != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			_, dbErr := db.ExecContext(ctx,
				"EXEC sp_InsertConversion @from_currency=@p1, @to_currency=@p2, @amount=@p3, @result=@p4, @rate=@p5",
				req.From, req.To, req.Amount, result, rate,
			)
			if dbErr != nil {
				log.Error().Err(dbErr).Msg("failed to save conversion to db")
			}
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

func handleGetHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeError(w, http.StatusServiceUnavailable, "database not available")
			return
		}

		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		limitStr := r.URL.Query().Get("limit")

		limit, err := validateLimit(limitStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var fromParam, toParam interface{}
		if from != "" {
			fromParam = from
		}
		if to != "" {
			toParam = to
		}

		rows, err := db.QueryContext(ctx,
			"EXEC sp_GetConversionHistory @from_currency=@p1, @to_currency=@p2, @limit=@p3",
			fromParam, toParam, limit,
		)
		if err != nil {
			log.Error().Err(err).Msg("failed to query history")
			writeError(w, http.StatusInternalServerError, "failed to fetch history")
			return
		}
		defer rows.Close()

		var history []ConversionHistory
		for rows.Next() {
			var row ConversionHistory
			err = rows.Scan(&row.ID, &row.From, &row.To, &row.Amount, &row.Result, &row.Rate, &row.CreatedAt)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to read data")
				return
			}
			history = append(history, row)
		}

		if history == nil {
			history = []ConversionHistory{}
		}

		writeJSON(w, http.StatusOK, history)
	}
}

func handleHealth(cache *CacheService, apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := map[string]string{
			"status": "healthy",
		}
		dbStatus := "down"
		if db != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			if db.PingContext(ctx) == nil {
				dbStatus = "up"
			}
		}
		status["database"] = dbStatus
		redisStatus := "down"
		if cache != nil && cache.Ping() == nil {
			redisStatus = "up"
		}
		status["redis"] = redisStatus
		apiStatus := "down"
		resp, err := http.Get(apiURL + "/latest")
		if err == nil && resp.StatusCode == http.StatusOK {
			apiStatus = "up"
		}
		if resp != nil {
			resp.Body.Close()
		}
		status["api"] = apiStatus

		if dbStatus == "down" || redisStatus == "down" || apiStatus == "down" {
			status["status"] = "degraded"
		}

		writeJSON(w, http.StatusOK, status)
	}
}
