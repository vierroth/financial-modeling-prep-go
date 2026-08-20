package fmp

import (
	"net/http"

	"golang.org/x/time/rate"
)

func New(apiKey string, client *http.Client) *Client {
	handler := Client{
		apiKey:  apiKey,
		client:  http.DefaultClient,
		limiter: rate.NewLimiter(rate.Limit(49), 1),
	}

	if client != nil {
		handler.client = client
	}

	return &handler
}

type Client struct {
	apiKey  string
	client  *http.Client
	limiter *rate.Limiter
}
