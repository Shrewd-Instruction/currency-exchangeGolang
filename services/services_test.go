package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"currency-exchange/logger"
	"currency-exchange/models"
)

func TestGetLatestRates_Service(t *testing.T) {
	logger.InitLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.ExchangeRateResponse{
			Amount: 1, Base: "EUR", Date: "2024-01-15",
			Rates: map[string]float64{"USD": 1.09, "GBP": 0.86},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	svc := NewExchangeService(ts.URL, nil)
	resp, err := svc.GetLatestRates("EUR")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Base != "EUR" {
		t.Errorf("expected EUR, got %s", resp.Base)
	}
	if resp.Rates["USD"] != 1.09 {
		t.Errorf("expected 1.09, got %f", resp.Rates["USD"])
	}
}

func TestGetLatestRates_ServerError(t *testing.T) {
	logger.InitLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()

	svc := NewExchangeService(ts.URL, nil)
	_, err := svc.GetLatestRates("USD")
	if err == nil {
		t.Error("expected error for 500, got nil")
	}
}

func TestGetConversionRate_Service(t *testing.T) {
	logger.InitLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.ExchangeRateResponse{
			Amount: 1, Base: "USD", Date: "2024-01-15",
			Rates: map[string]float64{"EUR": 0.92},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	svc := NewExchangeService(ts.URL, nil)
	rate, err := svc.GetConversionRate("USD", "EUR")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rate != 0.92 {
		t.Errorf("expected 0.92, got %f", rate)
	}
}

func TestGetConversionRate_NotFound(t *testing.T) {
	logger.InitLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.ExchangeRateResponse{
			Amount: 1, Base: "USD", Date: "2024-01-15",
			Rates: map[string]float64{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	svc := NewExchangeService(ts.URL, nil)
	_, err := svc.GetConversionRate("USD", "XYZ")
	if err == nil {
		t.Error("expected error for missing rate")
	}
}
