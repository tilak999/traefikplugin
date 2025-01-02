// Package traefikplugin a demo plugin.
package traefikplugin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Config the plugin configuration.
type Config struct {
	CloudflareZone  string   `json:"cloudflarezone,omitempty"`
	CloudflareToken string   `json:"cloudflaretoken,omitempty"`
	Headers         []string `json:"headers,omitempty"`
	DryRun          bool     `json:"dryrun,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Headers:         make([]string, 0),
		DryRun:          false,
		CloudflareZone:  "",
		CloudflareToken: "",
	}
}

// HeaderDetectionPlugin a HeaderDetectionPlugin plugin.
type HeaderDetectionPlugin struct {
	config *Config
	next   http.Handler
	logger *log.Logger
	name   string
}

type CustomResponseWriter struct {
	http.ResponseWriter
	*HeaderDetectionPlugin
	Url             string            `json:"url,omitempty"`
	DetectedHeaders map[string]string `json:"detectedHeaders,omitempty"`
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	logger := log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime)

	if len(config.Headers) == 0 {
		return nil, fmt.Errorf("headers cannot be empty")
	}

	if config.CloudflareToken == "" || config.CloudflareZone == "" {
		return nil, fmt.Errorf("cloudflare zone or token is not defined, zone=%s | token=%s",
			config.CloudflareZone, config.CloudflareToken)
	}

	logger.Println("Plugin initialised, ready to accept connections.")

	return &HeaderDetectionPlugin{
		config: config,
		logger: logger,
		next:   next,
		name:   name,
	}, nil
}

// WriteHeader captures headers before they're written
func (crw *CustomResponseWriter) WriteHeader(code int) {
	for _, header := range crw.config.Headers {
		value := crw.ResponseWriter.Header().Get(header)
		if value != "" {
			crw.DetectedHeaders[header] = value
		}
	}
	crw.ResponseWriter.WriteHeader(code)
}

func (a *HeaderDetectionPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	customRW := &CustomResponseWriter{
		ResponseWriter:        rw,
		HeaderDetectionPlugin: a,
		DetectedHeaders:       make(map[string]string),
	}
	a.next.ServeHTTP(customRW, req)
	if len(customRW.DetectedHeaders) > 0 {
		a.logger.Printf("req [host:%s][path:%s]", req.Host, req.URL.Path)
		if a.config.DryRun {
			for k, v := range customRW.DetectedHeaders {
				a.logger.Printf("%s=%s", k, v)
			}
		}
		go PurgeCache(a.config, customRW.DetectedHeaders, a.logger)
	}
}
