package models

import "time"

type ExchangeRateResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
}

type ConversionRequest struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}

type ConversionResponse struct {
	From      string  `json:"from"`
	To        string  `json:"to"`
	Amount    float64 `json:"amount"`
	Result    float64 `json:"result"`
	Rate      float64 `json:"rate"`
	Timestamp string  `json:"timestamp"`
}

type ConversionHistory struct {
	ID        int       `json:"id"`
	From      string    `json:"from_currency"`
	To        string    `json:"to_currency"`
	Amount    float64   `json:"amount"`
	Result    float64   `json:"result"`
	Rate      float64   `json:"rate"`
	CreatedAt time.Time `json:"created_at"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
