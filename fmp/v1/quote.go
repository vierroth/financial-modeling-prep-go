package fmp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type QuoteInput struct {
	Symbol string
}

type QuoteOutput struct {
	Symbol           string
	Name             string
	Price            float64
	ChangePercentage float64
	Change           float64
	Volume           float64
	DayLow           float64
	DayHigh          float64
	YearHigh         float64
	YearLow          float64
	MarketCap        float64
	PriceAvg50       float64
	PriceAvg200      float64
	Exchange         string
	Open             float64
	PreviousClose    float64
	Timestamp        time.Time
}

func (client Client) Quote(ctx context.Context, input QuoteInput) (*QuoteOutput, error) {
	baseURL, err := url.Parse("https://financialmodelingprep.com/stable/quote")
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	query := baseURL.Query()
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
		Symbol           string  `json:"symbol"`
		Name             string  `json:"name"`
		Price            float64 `json:"price"`
		ChangePercentage float64 `json:"changePercentage"`
		Change           float64 `json:"change"`
		Volume           float64 `json:"volume"`
		DayLow           float64 `json:"dayLow"`
		DayHigh          float64 `json:"dayHigh"`
		YearHigh         float64 `json:"yearHigh"`
		YearLow          float64 `json:"yearLow"`
		MarketCap        float64 `json:"marketCap"`
		PriceAvg50       float64 `json:"priceAvg50"`
		PriceAvg200      float64 `json:"priceAvg200"`
		Exchange         string  `json:"exchange"`
		Open             float64 `json:"open"`
		PreviousClose    float64 `json:"previousClose"`
		Timestamp        int64   `json:"timestamp"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil && err != io.EOF {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(items) < 1 {
		return nil, fmt.Errorf("no quote found for symbol %q", input.Symbol)
	}

	return &QuoteOutput{
		Symbol:           items[0].Symbol,
		Name:             items[0].Name,
		Price:            items[0].Price,
		ChangePercentage: items[0].ChangePercentage,
		Change:           items[0].Change,
		Volume:           items[0].Volume,
		DayLow:           items[0].DayLow,
		DayHigh:          items[0].DayHigh,
		YearHigh:         items[0].YearHigh,
		YearLow:          items[0].YearLow,
		MarketCap:        items[0].MarketCap,
		PriceAvg50:       items[0].PriceAvg50,
		PriceAvg200:      items[0].PriceAvg200,
		Exchange:         items[0].Exchange,
		Open:             items[0].Open,
		PreviousClose:    items[0].PreviousClose,
		Timestamp:        time.Unix(items[0].Timestamp, 0),
	}, nil
}
