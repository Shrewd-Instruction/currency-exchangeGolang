package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockAPI() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ExchangeRateResponse{
			Amount: 1,
			Base:   "USD",
			Date:   "2024-01-15",
			Rates: map[string]float64{
				"EUR": 0.92,
				"GBP": 0.79,
				"JPY": 148.5,
				"INR": 83.12,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestHandleGetRates(t *testing.T) {
	initLogger()
	ts := mockAPI()
	defer ts.Close()

	svc := newExchangeService(ts.URL, nil)
	handler := handleGetRates(svc)

	req := httptest.NewRequest("GET", "/api/v1/rates?base=USD", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp ExchangeRateResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.Base != "USD" {
		t.Errorf("expected base USD, got %s", resp.Base)
	}
	if len(resp.Rates) == 0 {
		t.Error("expected rates to not be empty")
	}
}

func TestHandleGetRates_DefaultBase(t *testing.T) {
	initLogger()
	ts := mockAPI()
	defer ts.Close()

	svc := newExchangeService(ts.URL, nil)
	handler := handleGetRates(svc)

	req := httptest.NewRequest("GET", "/api/v1/rates", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestHandleConvert(t *testing.T) {
	initLogger()
	ts := mockAPI()
	defer ts.Close()

	svc := newExchangeService(ts.URL, nil)
	handler := handleConvert(svc)

	body := `{"from": "USD", "to": "EUR", "amount": 100}`
	req := httptest.NewRequest("POST", "/api/v1/convert", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp ConversionResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.From != "USD" || resp.To != "EUR" {
		t.Errorf("expected USD->EUR, got %s->%s", resp.From, resp.To)
	}

	expected := 92.0
	if resp.Result != expected {
		t.Errorf("expected %.2f, got %.2f", expected, resp.Result)
	}
}

func TestHandleConvert_BadJSON(t *testing.T) {
	initLogger()
	ts := mockAPI()
	defer ts.Close()

	svc := newExchangeService(ts.URL, nil)
	handler := handleConvert(svc)

	req := httptest.NewRequest("POST", "/api/v1/convert", strings.NewReader("{bad json}"))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleConvert_InvalidAmount(t *testing.T) {
	initLogger()
	ts := mockAPI()
	defer ts.Close()

	svc := newExchangeService(ts.URL, nil)
	handler := handleConvert(svc)

	body := `{"from": "USD", "to": "EUR", "amount": 0}`
	req := httptest.NewRequest("POST", "/api/v1/convert", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleHistory_NoDB(t *testing.T) {
	initLogger()
	db = nil
	handler := handleGetHistory()

	req := httptest.NewRequest("GET", "/api/v1/history", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestValidateConversionRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     ConversionRequest
		wantErr bool
	}{
		{"valid", ConversionRequest{From: "USD", To: "EUR", Amount: 100}, false},
		{"empty from", ConversionRequest{From: "", To: "EUR", Amount: 100}, true},
		{"empty to", ConversionRequest{From: "USD", To: "", Amount: 100}, true},
		{"zero amount", ConversionRequest{From: "USD", To: "EUR", Amount: 0}, true},
		{"negative amount", ConversionRequest{From: "USD", To: "EUR", Amount: -50}, true},
		{"bad from code", ConversionRequest{From: "US", To: "EUR", Amount: 100}, true},
		{"lowercase", ConversionRequest{From: "usd", To: "EUR", Amount: 100}, true},
		{"numbers in code", ConversionRequest{From: "USD", To: "E2R", Amount: 100}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConversionRequest(&tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("got err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCurrencyCode(t *testing.T) {
	good := []string{"USD", "EUR", "GBP", "JPY", "INR"}
	for _, code := range good {
		if !validateCurrencyCode(code) {
			t.Errorf("%s should be valid", code)
		}
	}

	bad := []string{"usd", "US", "USDD", "U1D", "", "123"}
	for _, code := range bad {
		if validateCurrencyCode(code) {
			t.Errorf("%s should be invalid", code)
		}
	}
}
