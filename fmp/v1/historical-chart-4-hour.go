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

func (client Client) HistoricalChart4Hour(ctx context.Context, input HistoricalChartInput) ([]*HistoricalChartOutput, error) {
	baseURL, err := url.Parse("https://financialmodelingprep.com/stable/historical-chart/4hour")
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	query := baseURL.Query()

	if input.From != nil {
		query.Set("from", input.From.Format("2006-01-02"))
	}

	if input.To != nil {
		query.Set("to", input.To.Format("2006-01-02"))
	}

	if input.Nonadjusted != nil {
		query.Set("nonadjusted", strconv.FormatBool(*input.Nonadjusted))
	}

	if input.Extended != nil {
		query.Set("extended", strconv.FormatBool(*input.Extended))
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
		Date   string   `json:"date"`
		Open   *float64 `json:"open"`
		Low    *float64 `json:"low"`
		High   *float64 `json:"high"`
		Close  *float64 `json:"close"`
		Volume *float64 `json:"volume"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil && err != io.EOF {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	response := make([]*HistoricalChartOutput, len(items))
	for i, item := range items {
		date, err := time.Parse("2006-01-02 15:04:05", item.Date)
		if err != nil {
			return nil, fmt.Errorf("parse historical date %q: %w", item.Date, err)
		}
		response[i] = &HistoricalChartOutput{
			Date:   date,
			Open:   item.Open,
			Low:    item.Low,
			High:   item.High,
			Close:  item.Close,
			Volume: item.Volume,
		}
	}

	return response, nil
}
