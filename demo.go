// Package traefikplugin a demo plugin.
package traefikplugin

import (
	"context"
	"fmt"
	"net/http"
)

// Config holds the plugin configuration.
type Config struct {
	HeaderToRead string `json:"headerToRead,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		HeaderToRead: "X-Response-Header",
	}
}

// Plugin holds the plugin configuration
type Plugin struct {
	next         http.Handler
	headerToRead string
}

// New creates a new plugin instance
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &Plugin{
		next:         next,
		headerToRead: config.HeaderToRead,
	}, nil
}

// CustomResponseWriter wraps the standard http.ResponseWriter
type CustomResponseWriter struct {
	http.ResponseWriter
	headerValue  string
	headerToRead string
}

// WriteHeader captures headers before they're written
func (crw *CustomResponseWriter) WriteHeader(code int) {
	// Read the header we're interested in
	crw.headerValue = crw.ResponseWriter.Header().Get(crw.headerToRead)

	// You can do something with the header value here
	// For example, log it or modify it

	crw.ResponseWriter.WriteHeader(code)
}

// ServeHTTP implements the middleware interface
func (p *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("ServeHTTP: %s", req.URL)
	// Create custom response writer
	customRW := &CustomResponseWriter{
		ResponseWriter: rw,
		headerToRead:   p.headerToRead,
	}

	// Call the next handler with our custom response writer
	p.next.ServeHTTP(customRW, req)

	// After the response has been written, you can access the header value
	// For example, you could add it to a different header
	if customRW.headerValue != "" {
		//fmt.Printf("X-Captured-Header: %s", customRW.headerValue)
		rw.Header().Add("X-Captured-Header", customRW.headerValue)
	}
}
