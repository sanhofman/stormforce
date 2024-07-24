package httpclient

import (
	"net/http"
	"time"
)

// NewClient creates a new HTTP client with a specified timeout
func NewClient(timeoutSeconds int) *http.Client {
	return &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

// CustomTransport allows for additional configuration of the HTTP transport
type CustomTransport struct {
	*http.Transport
}

// NewCustomTransport creates a new CustomTransport with optimized settings
func NewCustomTransport() *CustomTransport {
	return &CustomTransport{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  true,
			DisableKeepAlives:   false,
		},
	}
}

// NewClientWithCustomTransport creates a new HTTP client with a custom transport
func NewClientWithCustomTransport(timeoutSeconds int) *http.Client {
	return &http.Client{
		Timeout:   time.Duration(timeoutSeconds) * time.Second,
		Transport: NewCustomTransport(),
	}
}
