package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/snyk/npmjs-deps-fetcher/internal/npm"
)

const configFilePath = "config.json"

// config represents the application configuration.
type config struct {
	// NPM configure the client to communicate with the NPM registry.
	NPM npm.ClientConfig `json:"npm"`

	// Server is the HTTP server related configuration.
	Server struct {
		// Addr is the bind address that the server will listen on.
		// For example, "localhost:8080"
		Addr string `json:"addr"`
		// ReadHeaderTimeout is the amount of time allowed to read
		// request headers.
		ReadHeaderTimeout string `json:"readHeaderTimeout"`
		readHeaderTimeout time.Duration
		// WriteTimeout is the maximum duration before timing out
		// writes of the response.
		WriteTimeout string `json:"writeTimeout"`
		writeTimeout time.Duration
	} `json:"server"`
}

// parseConfig parses the app configuration.
func parseConfig() (*config, error) {
	f, err := os.Open(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	var cfg config
	if err = json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decoding file %q: %w", configFilePath, err)
	}

	cfg.Server.readHeaderTimeout, err = time.ParseDuration(cfg.Server.WriteTimeout)
	if err != nil {
		return nil, fmt.Errorf("parsing readHeaderTimeout duration: %w", err)
	}

	cfg.Server.writeTimeout, err = time.ParseDuration(cfg.Server.WriteTimeout)
	if err != nil {
		return nil, fmt.Errorf("parsing writeTimeout duration: %w", err)
	}

	return &cfg, nil
}
