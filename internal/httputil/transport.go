package httputil

import (
	"crypto/tls"
	"net/http"
	"time"
)

// NewHTTPClient creates an HTTP client with common configuration.
func NewHTTPClient(timeout time.Duration, skipTLSVerify bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if skipTLSVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// HeaderRoundTripper wraps a transport to inject headers into all requests.
type HeaderRoundTripper struct {
	Headers   map[string]string
	Transport http.RoundTripper
}

// RoundTrip implements http.RoundTripper.
func (rt *HeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range rt.Headers {
		req.Header.Set(k, v)
	}
	return rt.Transport.RoundTrip(req)
}

// NewTransport creates an http.RoundTripper with optional TLS skip verification.
func NewTransport(skipTLSVerify bool) http.RoundTripper {
	if skipTLSVerify {
		return &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	return http.DefaultTransport
}
