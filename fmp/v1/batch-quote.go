package fmp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type BatchQuoteInput struct {
	Symbols []string
}

func (client Client) BatchQuote(ctx context.Context, input BatchQuoteInput) ([]*QuoteOutput, error) {
	baseURL, err := url.Parse("https://financialmodelingprep.com/stable/batch-quote")
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	query := baseURL.Query()
	query.Set("symbol", strings.Join(input.Symbols, ","))
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

	response := make([]*QuoteOutput, len(items))
	for _, item := range items {
		response = append(response, &QuoteOutput{
			Symbol:           item.Symbol,
			Name:             item.Name,
			Price:            item.Price,
			ChangePercentage: item.ChangePercentage,
			Change:           item.Change,
			Volume:           item.Volume,
			DayLow:           item.DayLow,
			DayHigh:          item.DayHigh,
			YearHigh:         item.YearHigh,
			YearLow:          item.YearLow,
			MarketCap:        item.MarketCap,
			PriceAvg50:       item.PriceAvg50,
			PriceAvg200:      item.PriceAvg200,
			Exchange:         item.Exchange,
			Open:             item.Open,
			PreviousClose:    item.PreviousClose,
			Timestamp:        time.Unix(item.Timestamp, 0),
		})
	}

	return response, nil
}
