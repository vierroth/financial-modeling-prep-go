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

type HistoricalPriceEodInput struct {
	Symbol string
	From   time.Time
	To     time.Time
}

type HistoricalPriceEodOutput struct {
	Symbol        string
	Date          time.Time
	Open          float64
	Close         float64
	High          float64
	Low           float64
	Volume        uint64
	Change        float64
	ChangePercent float64
	Vwap          float64
}

func (client Client) HistoricalPriceEod(ctx context.Context, input HistoricalPriceEodInput) ([]*HistoricalPriceEodOutput, error) {
	baseURL, err := url.Parse("https://financialmodelingprep.com/stable/historical-price-eod")
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	query := baseURL.Query()
	query.Set("symbol", input.Symbol)
	query.Set("from", input.From.Format("2006-01-02"))
	query.Set("to", input.To.Format("2006-01-02"))
	query.Set("apikey", client.apiKey)
	baseURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
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
		Symbol        string    `json:"symbol"`
		Date          time.Time `json:"date"`
		Open          float64   `json:"open"`
		Close         float64   `json:"close"`
		High          float64   `json:"high"`
		Low           float64   `json:"low"`
		Volume        uint64    `json:"volume"`
		Change        float64   `json:"change"`
		ChangePercent float64   `json:"changePercent"`
		Vwap          float64   `json:"vwap"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil && err != io.EOF {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	response := make([]*HistoricalPriceEodOutput, len(items))
	for _, item := range items {
		response = append(response, &HistoricalPriceEodOutput{
			Symbol:        item.Symbol,
			Date:          item.Date,
			Open:          item.Open,
			Close:         item.Close,
			High:          item.High,
			Low:           item.Low,
			Volume:        item.Volume,
			Change:        item.Change,
			ChangePercent: item.ChangePercent,
			Vwap:          item.Vwap,
		})
	}

	return response, nil
}
