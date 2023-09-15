package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestServeHTTP(t *testing.T) {
	// Create an instance of PrometheusGeoIP
	promGeoIP := &PrometheusGeoIP{}

	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Could not create HTTP request: %v", err)
	}

	// Create a new HTTP response recorder
	rec := httptest.NewRecorder()

	// Create a next handler
	next := caddyhttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("next handler"))
		return nil
	})

	// Call ServeHTTP method
	err = promGeoIP.ServeHTTP(rec, req, next)
	if err != nil {
		t.Fatalf("Could not call ServeHTTP: %v", err)
	}

	// Check if next handler is called
	assert.Equal(t, "next handler", rec.Body.String())

	// Check if metrics are incremented
	assert.Equal(t, 1, testutil.CollectAndCount(requestMethod, "method=\"GET\""))
	assert.Equal(t, 1, testutil.CollectAndCount(requestURI, "uri=\"/test\""))
	assert.Equal(t, 1, testutil.CollectAndCount(userAgent, "user_agent=\"\"")) // Empty user agent for this test
	assert.Equal(t, 1, testutil.CollectAndCount(remoteIP, "remote_ip=\"\""))  // Empty remote IP for this test
}
