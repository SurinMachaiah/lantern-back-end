package httpclient

import (
	"net/http"
	"time"
)

type Option func(*Client)

type Client struct {
	httpClient *http.Client
}

func SetHTTPClient(httpClient *http.Client) Option {
	return func(cli *Client) {
		cli.httpClient = httpClient
	}
}

// NewClient creates a wrapper around an httpclient.
// The client is created with a default timeout of 35 seconds. This can be set using an option argument.
func NewClient(options ...Option) *Client {
	cli := Client{
		httpClient: &http.Client{
			Timeout: 35 * time.Second,
		},
	}

	for i := range options {
		options[i](&cli)
	}

	return &cli
}

func (cli *Client) Do(req *http.Request) (*http.Response, error) {
	return cli.httpClient.Do(req)
}
