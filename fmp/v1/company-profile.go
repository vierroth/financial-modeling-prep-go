package fmp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type CompanyProfileInput struct {
	Symbol string
}

type CompanyProfileOutput struct {
	Symbol            string
	Price             float64
	MarketCap         *float64
	Beta              float64
	LastDividend      float64
	Range             string
	Change            float64
	ChangePercentage  float64
	Volume            int64
	AverageVolume     int64
	CompanyName       string
	Currency          string
	Cik               *string
	Isin              string
	Cusip             string
	ExchangeFullName  string
	Exchange          string
	Industry          string
	Website           string
	Description       string
	Ceo               string
	Sector            string
	Country           string
	FullTimeEmployees string
	Phone             string
	Address           string
	City              string
	State             string
	Zip               string
	Image             *string
	IpoDate           string
	DefaultImage      bool
	IsEtf             bool
	IsActivelyTrading bool
	IsAdr             bool
	IsFund            bool
}

func (client Client) CompanyInformation(ctx context.Context, input CompanyProfileInput) (*CompanyProfileOutput, error) {
	baseURL, err := url.Parse("https://financialmodelingprep.com/stable/profile")
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
		Symbol            string   `json:"symbol"`
		Price             float64  `json:"price"`
		MarketCap         *float64 `json:"marketCap"`
		Beta              float64  `json:"beta"`
		LastDividend      float64  `json:"lastDividend"`
		Range             string   `json:"range"`
		Change            float64  `json:"change"`
		ChangePercentage  float64  `json:"changePercentage"`
		Volume            int64    `json:"volume"`
		AverageVolume     int64    `json:"averageVolume"`
		CompanyName       string   `json:"companyName"`
		Currency          string   `json:"currency"`
		Cik               *string  `json:"cik"`
		Isin              string   `json:"isin"`
		Cusip             string   `json:"cusip"`
		ExchangeFullName  string   `json:"exchangeFullName"`
		Exchange          string   `json:"exchange"`
		Industry          string   `json:"industry"`
		Website           string   `json:"website"`
		Description       string   `json:"description"`
		Ceo               string   `json:"ceo"`
		Sector            string   `json:"sector"`
		Country           string   `json:"country"`
		FullTimeEmployees string   `json:"fullTimeEmployees"`
		Phone             string   `json:"phone"`
		Address           string   `json:"address"`
		City              string   `json:"city"`
		State             string   `json:"state"`
		Zip               string   `json:"zip"`
		Image             *string  `json:"image"`
		IpoDate           string   `json:"ipoDate"`
		DefaultImage      bool     `json:"defaultImage"`
		IsEtf             bool     `json:"isEtf"`
		IsActivelyTrading bool     `json:"isActivelyTrading"`
		IsAdr             bool     `json:"isAdr"`
		IsFund            bool     `json:"isFund"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil && err != io.EOF {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(items) < 1 {
		return nil, fmt.Errorf("no quote found for symbol %q", input.Symbol)
	}

	return &CompanyProfileOutput{
		Symbol:            items[0].Symbol,
		Price:             items[0].Price,
		MarketCap:         items[0].MarketCap,
		Beta:              items[0].Beta,
		LastDividend:      items[0].LastDividend,
		Range:             items[0].Range,
		Change:            items[0].Change,
		ChangePercentage:  items[0].ChangePercentage,
		Volume:            items[0].Volume,
		AverageVolume:     items[0].AverageVolume,
		CompanyName:       items[0].CompanyName,
		Currency:          items[0].Currency,
		Cik:               items[0].Cik,
		Isin:              items[0].Isin,
		Cusip:             items[0].Cusip,
		ExchangeFullName:  items[0].ExchangeFullName,
		Exchange:          items[0].Exchange,
		Industry:          items[0].Industry,
		Website:           items[0].Website,
		Description:       items[0].Description,
		Ceo:               items[0].Ceo,
		Sector:            items[0].Sector,
		Country:           items[0].Country,
		FullTimeEmployees: items[0].FullTimeEmployees,
		Phone:             items[0].Phone,
		Address:           items[0].Address,
		City:              items[0].City,
		State:             items[0].State,
		Zip:               items[0].Zip,
		Image:             items[0].Image,
		IpoDate:           items[0].IpoDate,
		DefaultImage:      items[0].DefaultImage,
		IsEtf:             items[0].IsEtf,
		IsActivelyTrading: items[0].IsActivelyTrading,
		IsAdr:             items[0].IsAdr,
		IsFund:            items[0].IsFund,
	}, nil
}
