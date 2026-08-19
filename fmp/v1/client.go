package fmp

import (
	"net/http"
)

func New(apiKey string, client *http.Client) *Client {
	handler := Client{
		apiKey: apiKey,
		client: client,
	}

	return &handler
}

type Client struct {
	apiKey string
	client *http.Client
}
