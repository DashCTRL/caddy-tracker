package main

import (
    "net/http"
    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define custom metrics
var (
    requestMethod = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_method_total",
            Help: "Total number of requests by HTTP method.",
        },
        []string{"method"},
    )
    requestURI = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_uri_total",
            Help: "Total number of requests by URI.",
        },
        []string{"uri"},
    )
    userAgent = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_user_agent_total",
            Help: "Total number of requests by User-Agent.",
        },
        []string{"user_agent"},
    )
    remoteIP = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_remote_ip_total",
            Help: "Total number of requests by remote IP.",
        },
        []string{"remote_ip"},
    )
)

func init() {
    // Register the custom metrics with Prometheus
    prometheus.MustRegister(requestMethod, requestURI, userAgent, remoteIP)

    // Register the module with Caddy
    caddy.RegisterModule(PrometheusGeoIP{})
}

type PrometheusGeoIP struct{}

func (PrometheusGeoIP) CaddyModule() caddy.ModuleInfo {
    return caddy.ModuleInfo{
        ID:  "http.handlers.prometheus_geoip",
        New: func() caddy.Module { return new(PrometheusGeoIP) },
    }
}

// ServeHTTP implements the Caddy HTTP handler interface
func (m *PrometheusGeoIP) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    // Increment the custom metrics based on the incoming request
    requestMethod.WithLabelValues(r.Method).Inc()
    requestURI.WithLabelValues(r.RequestURI).Inc()
    userAgent.WithLabelValues(r.Header.Get("User-Agent")).Inc()
    remoteIP.WithLabelValues(r.RemoteAddr).Inc()

    // Serve Prometheus metrics
    promhttp.Handler().ServeHTTP(w, r)

    // Call the next handler in the chain
    return next.ServeHTTP(w, r)
}
