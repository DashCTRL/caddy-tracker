package main

import (
	"net/http"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/caddyserver/caddy/v2/modules/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/oschwald/maxminddb-golang"
	"net"
)

// CaddyTracking Module
type CaddyTracking struct {
	Logger *logging.Logger
	dbInst *maxminddb.Reader
}

func init() {
	caddy.RegisterModule(CaddyTracking{})
}

func (CaddyTracking) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.caddy_tracking",
		New: func() caddy.Module { return new(CaddyTracking) },
	}
}

func (m *CaddyTracking) Provision(ctx caddy.Context) error {
	m.Logger = ctx.Logger(m)
	var err error
	m.dbInst, err = maxminddb.Open("path/to/GeoLite2-Country.mmdb")
	if err != nil {
		return err
	}
	return nil
}

func (m *CaddyTracking) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {

    // Start a timer to track request duration
	startTime := time.Now()
    // Call the next handler in the chain and capture the response
	rec := httptest.NewRecorder()
	err := next.ServeHTTP(rec, r)
	if err != nil {
		return err
	}
	duration := time.Since(startTime)
    
	// Existing code for IP and GeoIP
	remoteIp, _, _ := net.SplitHostPort(r.RemoteAddr)
	addr := net.ParseIP(remoteIp)
	if addr == nil {
		return next.ServeHTTP(w, r)
	}

	var record Record // Record is assumed to be in the same package or imported
	err := m.dbInst.Lookup(addr, &record)
	if err != nil {
		return next.ServeHTTP(w, r)
	}

	// Additional tracking details
	httpVersion := r.Proto
	referrer := r.Referer()
	contentType := r.Header.Get("Content-Type")
	queryParams := r.URL.Query().Encode()
	cookieNames := []string{}
	for _, cookie := range r.Cookies() {
		cookieNames = append(cookieNames, cookie.Name)
	}
	cookies := "Cookies: " + strings.Join(cookieNames, ", ")
    host := r.Host
	contentLength := r.ContentLength
	authorization := r.Header.Get("Authorization")
	acceptEncoding := r.Header.Get("Accept-Encoding")
	acceptLanguage := r.Header.Get("Accept-Language")
	customHeader := r.Header.Get("X-Custom-Header")  // Replace with your actual custom header name
	statusCode := rec.Code

	// Log the tracked details
	m.Logger.Info("Tracking request",
		"method", r.Method,
		"uri", r.RequestURI,
		"remote_ip", remoteIp,
		"country", record.Country.ISOCode,
		"subdivisions", record.Subdivisions.CommaSeparatedISOCodes(),
		"user_agent", r.Header.Get("User-Agent"),
		"http_version", httpVersion,
		"referrer", referrer,
		"content_type", contentType,
		"query_params", queryParams,
		"cookies", cookies,
		"host", host,
		"content_length", contentLength,
		"authorization", authorization,  // Caution: This could be sensitive data
		"accept_encoding", acceptEncoding,
		"accept_language", acceptLanguage,
		"custom_header", customHeader,
		"time_taken", duration.String(),
		"status_code", statusCode,
	)

    	// Copy the recorded response to the original response writer
	for k, v := range rec.Header() {
		w.Header()[k] = v
	}
	w.WriteHeader(rec.Code)
	_, _ = io.Copy(w, rec.Body)

	return nil

	return next.ServeHTTP(w, r)
}
