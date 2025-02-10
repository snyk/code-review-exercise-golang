package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/snyk/npmjs-deps-fetcher/internal/npm"
)

const configFilePath = "config.json"

// config represents the application configuration.
type config struct {
	// ListenAddr is the bind address that the server will listen on.
	// For example, "localhost:8080"
	ListenAddr string `json:"listenAddr"`
	// NPM configure the client to communicate with the NPM registry.
	NPM npm.ClientConfig `json:"npm"`
}

// parseConfig parses the app configuration.
func parseConfig() (*config, error) {
	f, err := os.Open(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	var cfg config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decoding file %q: %w", configFilePath, err)
	}

	return &cfg, nil
}
