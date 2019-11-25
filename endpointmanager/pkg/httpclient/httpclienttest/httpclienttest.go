package httpclienttest

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/httpclient"
)

type TestClient struct {
	teardown func()
	httpclient.Client
}

func NewTestClient(handler http.Handler) *TestClient {
	httpcli, teardown := testingHTTPClient(handler)
	cli := httpclient.NewClient(httpclient.SetHTTPClient(httpcli))

	tc := TestClient{
		teardown,
		*cli,
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
