package fmp

import (
	"net/http"
)

func New(apiKey string, client *http.Client) *Client {
	handler := Client{
		apiKey: apiKey,
		client: http.DefaultClient,
	}

	if client != nil {
		handler.client = client
	}

	return &handler
}

type Client struct {
	apiKey string
	client *http.Client
}
