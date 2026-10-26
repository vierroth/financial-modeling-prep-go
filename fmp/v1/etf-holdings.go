package fmp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type EtfHoldingsInput struct {
	Symbol string
}

type Asset struct {
	Symbol string
	Asset  string
	Name   string
}

func (client Client) EtfHoldings(ctx context.Context, input EtfHoldingsInput) ([]*Asset, error) {
	baseURL, err := url.Parse("https://financialmodelingprep.com/stable/etf/holdings")
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
		Symbol string `json:"symbol"`
		Asset  string `json:"asset"`
		Name   string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil && err != io.EOF {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	response := make([]*Asset, len(items))
	for i, item := range items {
		response[i] = &Asset{
			Symbol: item.Symbol,
			Asset:  item.Asset,
			Name:   item.Name,
		}
	}

	return response, nil
}
