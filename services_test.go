package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetLatestRates_Service(t *testing.T) {
	initLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ExchangeRateResponse{
			Amount: 1, Base: "EUR", Date: "2024-01-15",
			Rates: map[string]float64{"USD": 1.09, "GBP": 0.86},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	svc := newExchangeService(ts.URL, nil)
	resp, err := svc.getLatestRates("EUR")
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
	initLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()

	svc := newExchangeService(ts.URL, nil)
	_, err := svc.getLatestRates("USD")
	if err == nil {
		t.Error("expected error for 500, got nil")
	}
}

func TestGetConversionRate_Service(t *testing.T) {
	initLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ExchangeRateResponse{
			Amount: 1, Base: "USD", Date: "2024-01-15",
			Rates: map[string]float64{"EUR": 0.92},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	svc := newExchangeService(ts.URL, nil)
	rate, err := svc.getConversionRate("USD", "EUR")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rate != 0.92 {
		t.Errorf("expected 0.92, got %f", rate)
	}
}

func TestGetConversionRate_NotFound(t *testing.T) {
	initLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ExchangeRateResponse{
			Amount: 1, Base: "USD", Date: "2024-01-15",
			Rates: map[string]float64{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	svc := newExchangeService(ts.URL, nil)
	_, err := svc.getConversionRate("USD", "XYZ")
	if err == nil {
		t.Error("expected error for missing rate")
	}
}

func TestValidateLimit_Service(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"25", 25, false},
		{"1", 1, false},
		{"100", 100, false},
		{"", 50, false},
		{"150", 0, true},
		{"0", 0, true},
		{"-5", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		got, err := validateLimit(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateLimit(%q) err=%v, wantErr=%v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("validateLimit(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
