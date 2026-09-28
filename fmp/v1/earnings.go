package fmp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Earnings struct {
	Symbol           string
	Date             *time.Time
	EpsActual        *float64
	EpsEstimated     *float64
	RevenueActual    *float64
	RevenueEstimated *float64
	LastUpdated      *time.Time
}

type EarningsInput struct {
	Symbol             string
	IncludeReportTimes *bool
	Limit              *uint
}

func (client Client) Earnings(ctx context.Context, input EarningsInput) ([]*Earnings, error) {
	baseURL, err := url.Parse("https://financialmodelingprep.com/stable/sp500-constituent")
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	query := baseURL.Query()

	if input.IncludeReportTimes != nil {
		query.Set("includeAllShareClasses", strconv.FormatBool(*input.IncludeReportTimes))
	}
	if input.Limit != nil {
		query.Set("limit", strconv.FormatUint(uint64(*input.Limit), 10))
	}

	query.Set("symbol", input.Symbol)
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
		Symbol           string   `json:"symbol"`
		Date             *string  `json:"date"`
		EpsActual        *float64 `json:"epsActual"`
		EpsEstimated     *float64 `json:"epsEstimated"`
		RevenueActual    *float64 `json:"revenueActual"`
		RevenueEstimated *float64 `json:"revenueEstimated"`
		LastUpdated      *string  `json:"lastUpdated"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil && err != io.EOF {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var date *time.Time
	if items[0].Date != nil && *items[0].Date != "" {
		parsed, err := time.Parse("2006-01-02", *items[0].Date)
		if err != nil {
			return nil, fmt.Errorf("error parsing date: %v", err)
		}
		date = &parsed
	}

	var lastUpdated *time.Time
	if items[0].LastUpdated != nil && *items[0].LastUpdated != "" {
		parsed, err := time.Parse("2006-01-02", *items[0].LastUpdated)
		if err != nil {
			return nil, fmt.Errorf("error parsing lastUpdated: %v", err)
		}
		lastUpdated = &parsed
	}

	response := make([]*Earnings, len(items))
	for i, item := range items {
		response[i] = &Earnings{
			Symbol:           item.Symbol,
			Date:             date,
			EpsActual:        item.EpsActual,
			EpsEstimated:     item.EpsEstimated,
			RevenueActual:    item.RevenueActual,
			RevenueEstimated: item.RevenueEstimated,
			LastUpdated:      lastUpdated,
		}
	}

	return response, nil
}
