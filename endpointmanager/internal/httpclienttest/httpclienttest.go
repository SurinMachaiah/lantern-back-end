package httpclienttest

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
)

type TestClient struct {
	teardown func()
	http.Client
}

func NewTestClient(handler http.Handler) *TestClient {
	httpcli, teardown := testingHTTPClient(handler)

	tc := TestClient{
		teardown,
		*httpcli,
	}

	return &tc
}

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
