package fmp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ExchangeMarketHoursInput struct {
	Exchange string
}

type ExchangeMarketHoursOutput struct {
	Exchange     string
	Name         string
	OpeningHour  string
	ClosingHour  string
	Timezone     string
	IsMarketOpen bool
}

func (client Client) ExchangeMarketHours(ctx context.Context, input ExchangeMarketHoursInput) (*ExchangeMarketHoursOutput, error) {
	baseURL, err := url.Parse("https://financialmodelingprep.com/stable/exchange-market-hours")
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	query := baseURL.Query()
	query.Set("exchange", input.Exchange)
	query.Set("apikey", client.apiKey)
	baseURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	if err := client.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("wait for FMP rate limiter: %w", err)
	}

	resp, err := client.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slurp, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return nil, fmt.Errorf("unexpected status %s: %s", resp.Status, string(slurp))
	}

	var items []struct {
		Exchange     string `json:"exchange"`
		Name         string `json:"name"`
		OpeningHour  string `json:"openingHour"`
		ClosingHour  string `json:"closingHour"`
		Timezone     string `json:"timezone"`
		IsMarketOpen bool   `json:"isMarketOpen"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil && err != io.EOF {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(items) < 1 {
		return nil, fmt.Errorf("no quote found for symbol %q", input.Exchange)
	}

	return &ExchangeMarketHoursOutput{
		Exchange:     items[0].Exchange,
		Name:         items[0].Name,
		OpeningHour:  items[0].OpeningHour,
		ClosingHour:  items[0].ClosingHour,
		Timezone:     items[0].Timezone,
		IsMarketOpen: items[0].IsMarketOpen,
	}, nil
}
