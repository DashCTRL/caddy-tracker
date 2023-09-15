package main

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/caddyserver/caddy/v2/modules/logging"
	"github.com/luludotdev/caddy-requestid"
	"github.com/oschwald/maxminddb-golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

type CaddyTracking struct {
	Logger         *logging.Logger
	dbInst         *maxminddb.Reader
	RequestIDModule *requestid.RequestID
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
	m.RequestIDModule = new(requestid.RequestID)
	err = m.RequestIDModule.Provision(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (m *CaddyTracking) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	err := m.RequestIDModule.ServeHTTP(w, r, nil)
	if err != nil {
		return err
	}
	startTime := time.Now()
	rec := httptest.NewRecorder()
	err = next.ServeHTTP(rec, r)
	if err != nil {
		return err
	}
	duration := time.Since(startTime)
	remoteIp, _, _ := net.SplitHostPort(r.RemoteAddr)
	addr := net.ParseIP(remoteIp)
	if addr == nil {
		return next.ServeHTTP(w, r)
	}
	var record Record
	err = m.dbInst.Lookup(addr, &record)
	if err != nil {
		return next.ServeHTTP(w, r)
	}
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
	customHeader := r.Header.Get("X-Custom-Header")
	statusCode := rec.Code
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
		"authorization", authorization,
		"accept_encoding", acceptEncoding,
		"accept_language", acceptLanguage,
		"custom_header", customHeader,
		"time_taken", duration.String(),
		"status_code", statusCode,
	)
	for k, v := range rec.Header() {
		w.Header()[k] = v
	}
	w.WriteHeader(rec.Code)
	_, _ = io.Copy(w, rec.Body)
	return nil
}

func main() {}
