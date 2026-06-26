package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"currency-exchange/database"
	"currency-exchange/logger"
	"currency-exchange/models"
	"currency-exchange/services"
)

func mockAPI() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.ExchangeRateResponse{
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
	logger.InitLogger()
	ts := mockAPI()
	defer ts.Close()

	svc := services.NewExchangeService(ts.URL, nil)
	handler := handleGetRates(svc)

	req := httptest.NewRequest("GET", "/api/v1/rates?base=USD", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp models.ExchangeRateResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.Base != "USD" {
		t.Errorf("expected base USD, got %s", resp.Base)
	}
	if len(resp.Rates) == 0 {
		t.Error("expected rates to not be empty")
	}
}

func TestHandleGetRates_DefaultBase(t *testing.T) {
	logger.InitLogger()
	ts := mockAPI()
	defer ts.Close()

	svc := services.NewExchangeService(ts.URL, nil)
	handler := handleGetRates(svc)

	req := httptest.NewRequest("GET", "/api/v1/rates", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestHandleConvert(t *testing.T) {
	logger.InitLogger()
	ts := mockAPI()
	defer ts.Close()

	svc := services.NewExchangeService(ts.URL, nil)
	handler := handleConvert(svc)

	body := `{"from": "USD", "to": "EUR", "amount": 100}`
	req := httptest.NewRequest("POST", "/api/v1/convert", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp models.ConversionResponse
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
	logger.InitLogger()
	ts := mockAPI()
	defer ts.Close()

	svc := services.NewExchangeService(ts.URL, nil)
	handler := handleConvert(svc)

	req := httptest.NewRequest("POST", "/api/v1/convert", strings.NewReader("{bad json}"))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleConvert_InvalidAmount(t *testing.T) {
	logger.InitLogger()
	ts := mockAPI()
	defer ts.Close()

	svc := services.NewExchangeService(ts.URL, nil)
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
	logger.InitLogger()
	database.DB = nil
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
		req     models.ConversionRequest
		wantErr bool
	}{
		{"valid", models.ConversionRequest{From: "USD", To: "EUR", Amount: 100}, false},
		{"empty from", models.ConversionRequest{From: "", To: "EUR", Amount: 100}, true},
		{"empty to", models.ConversionRequest{From: "USD", To: "", Amount: 100}, true},
		{"zero amount", models.ConversionRequest{From: "USD", To: "EUR", Amount: 0}, true},
		{"negative amount", models.ConversionRequest{From: "USD", To: "EUR", Amount: -50}, true},
		{"bad from code", models.ConversionRequest{From: "US", To: "EUR", Amount: 100}, true},
		{"lowercase", models.ConversionRequest{From: "usd", To: "EUR", Amount: 100}, true},
		{"numbers in code", models.ConversionRequest{From: "USD", To: "E2R", Amount: 100}, true},
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
