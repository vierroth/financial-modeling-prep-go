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

type SymbolChangeInput struct {
	Invalid *bool
	Limit   *uint
}

type SymbolChangeOutput struct {
	Date        time.Time
	CompanyName string
	OldSymbol   string
	NewSymbol   string
}

func (client Client) SymbolChangeQuote(ctx context.Context, input SymbolChangeInput) ([]*SymbolChangeOutput, error) {
	baseURL, err := url.Parse("https://financialmodelingprep.com/stable/batch-exchange-quote")
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	query := baseURL.Query()

	if input.Invalid != nil {
		query.Set("invalid", strconv.FormatBool(*input.Invalid))
	}

	if input.Limit != nil {
		query.Set("limit", strconv.FormatUint(uint64(*input.Limit), 10))
	}

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
		Date        string `json:"date"`
		CompanyName string `json:"companyName"`
		OldSymbol   string `json:"oldSymbol"`
		NewSymbol   string `json:"newSymbol"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil && err != io.EOF {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	response := make([]*SymbolChangeOutput, len(items))
	for i, item := range items {
		date, err := time.Parse("2006-01-02", item.Date)
		if err != nil {
			return nil, fmt.Errorf("error parsing date: %v", err)
		}

		response[i] = &SymbolChangeOutput{
			Date:        date,
			CompanyName: item.CompanyName,
			OldSymbol:   item.OldSymbol,
			NewSymbol:   item.NewSymbol,
		}
	}

	return response, nil
}
