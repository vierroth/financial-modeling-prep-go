package fmp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type CompanyScreenerInput struct {
	Exchange *string
	Country  *string

	IsEtf                  *bool
	IsFund                 *bool
	IsActivelyTrading      *bool
	IncludeAllShareClasses *bool

	Page  *uint
	Limit *uint
}

type CompanyScreenerOutput struct {
	Symbol             string
	CompanyName        string
	MarketCap          float64
	Sector             string
	Industry           string
	Beta               float64
	Price              float64
	LastAnnualDividend float64
	Volume             float64
	Exchange           string
	ExchangeShortName  string
	Country            string
	IsEtf              bool
	IsFund             bool
	IsActivelyTrading  bool
}

func (client Client) CompanyScreener(ctx context.Context, input CompanyScreenerInput) ([]*CompanyScreenerOutput, error) {
	baseURL, err := url.Parse("https://financialmodelingprep.com/stable/company-screener")
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	query := baseURL.Query()

	if input.Exchange != nil {
		query.Set("exchange", *input.Exchange)
	}

	if input.Country != nil {
		query.Set("country", *input.Country)
	}

	if input.IsEtf != nil {
		query.Set("isEtf", strconv.FormatBool(*input.IsEtf))
	}

	if input.IsFund != nil {
		query.Set("isFund", strconv.FormatBool(*input.IsFund))
	}

	if input.IsActivelyTrading != nil {
		query.Set("isActivelyTrading", strconv.FormatBool(*input.IsActivelyTrading))
	}

	if input.IncludeAllShareClasses != nil {
		query.Set("includeAllShareClasses", strconv.FormatBool(*input.IncludeAllShareClasses))
	}

	if input.Page != nil {
		query.Set("page", strconv.FormatUint(uint64(*input.Page), 10))
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
		Symbol             string  `json:"symbol"`
		CompanyName        string  `json:"companyName"`
		MarketCap          float64 `json:"marketCap"`
		Sector             string  `json:"sector"`
		Industry           string  `json:"industry"`
		Beta               float64 `json:"beta"`
		Price              float64 `json:"price"`
		LastAnnualDividend float64 `json:"lastAnnualDividend"`
		Volume             float64 `json:"volume"`
		Exchange           string  `json:"exchange"`
		ExchangeShortName  string  `json:"exchangeShortName"`
		Country            string  `json:"country"`
		IsEtf              bool    `json:"isEtf"`
		IsFund             bool    `json:"isFund"`
		IsActivelyTrading  bool    `json:"isActivelyTrading"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil && err != io.EOF {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	response := make([]*CompanyScreenerOutput, len(items))
	for i, item := range items {
		response[i] = &CompanyScreenerOutput{
			Symbol:             item.Symbol,
			CompanyName:        item.CompanyName,
			MarketCap:          item.MarketCap,
			Sector:             item.Sector,
			Industry:           item.Industry,
			Beta:               item.Beta,
			Price:              item.Price,
			LastAnnualDividend: item.LastAnnualDividend,
			Volume:             item.Volume,
			Exchange:           item.Exchange,
			ExchangeShortName:  item.ExchangeShortName,
			Country:            item.Country,
			IsEtf:              item.IsEtf,
			IsFund:             item.IsFund,
			IsActivelyTrading:  item.IsActivelyTrading,
		}
	}

	return response, nil
}
