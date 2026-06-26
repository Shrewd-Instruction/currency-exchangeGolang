package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ExchangeService struct {
	baseURL string
	client  *http.Client
	cache   *CacheService
}

func newExchangeService(baseURL string, cache *CacheService) *ExchangeService {
	return &ExchangeService{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
		cache:   cache,
	}
}
func (s *ExchangeService) getLatestRates(base string) (*ExchangeRateResponse, error) {
	cacheKey := "rates:" + base


	if s.cache != nil {
		ctx := context.Background()
		cached, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var data ExchangeRateResponse
			if json.Unmarshal([]byte(cached), &data) == nil {
				log.Debug().Msgf("cache hit for %s", cacheKey)
				return &data, nil
			}
		}
	}

	url := s.baseURL + "/latest?from=" + base
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var data ExchangeRateResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if s.cache != nil {
		ctx := context.Background()
		s.cache.Set(ctx, cacheKey, string(body), 5*time.Minute)
	}

	log.Info().Msgf("fetched rates for %s from api", base)
	return &data, nil
}

func (s *ExchangeService) getConversionRate(from, to string) (float64, error) {
	url := s.baseURL + "/latest?from=" + from + "&to=" + to
	resp, err := s.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch rate: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %v", err)
	}

	var data ExchangeRateResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, fmt.Errorf("failed to parse response: %v", err)
	}

	rate, ok := data.Rates[to]
	if !ok {
		return 0, fmt.Errorf("rate not found for %s", to)
	}

	return rate, nil
}
