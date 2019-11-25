package testhelper

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
)

// TestClient is test wrapper around an http client, allowing http responses to be mocked.
type TestClient struct {
	teardown func()
	http.Client
}

// NewTestClient creates a new TestClient using an httptest TLS Server. Any http requests using
// this client will be handled by 'handler'.
func NewTestClient(handler http.Handler) *TestClient {
	httpcli, teardown := testingHTTPClient(handler)

	tc := TestClient{
		teardown,
		*httpcli,
	}

	return &tc
}

// Close closes resources associated with the test client and should be called after every
// instantiation of the client.
func (tc *TestClient) Close() {
	tc.teardown()
}

func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewTLSServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return cli, s.Close
}
