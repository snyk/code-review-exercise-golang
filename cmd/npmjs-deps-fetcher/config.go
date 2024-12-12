package main

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/snyk/npmjs-deps-fetcher/internal/npm"
)

// config represents the application configuration.
type config struct {
	// Logger configures the application logger that prints to stdout.
	Logger struct {
		// Level defines the minimum record level that will be logged.
		Level string `json:"level"`
		level slog.LevelVar
	} `json:"logger"`

	// NPM configures the client to communicate with the NPM registry.
	NPM npm.ClientConfig `json:"npm"`

	// Server is the HTTP server related configuration.
	Server struct {
		// Addr is the bind address that the server will listen on.
		// For example, "localhost:8080"
		Addr string `json:"addr"`
		// ReadHeaderTimeout is the amount of time allowed to read
		// request headers.
		ReadHeaderTimeout time.Duration `json:"readHeaderTimeout"`
		// WriteTimeout is the maximum duration before timing out
		// writes of the response.
		WriteTimeout time.Duration `json:"writeTimeout"`
	} `json:"server"`
}

// parseConfig parses the app configuration.
func parseConfig() (cfg *config, errs error) {
	viper.AddConfigPath(".")
	viper.SetConfigType("json")

	viper.SetDefault("npm.timeout", "15s")
	viper.SetDefault("server.readHeaderTimeout", "10s")
	viper.SetDefault("server.writeTimeout", "30s")

	if err := viper.ReadInConfig(); err != nil {
		var errNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &errNotFound) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var logLevel, registryURL, serverAddr string
	pflag.StringVar(&logLevel, "logger.level", "info", "Log level (debug, info, warn, error)")
	pflag.StringVar(&registryURL, "npm.registryUrl", "", "NPM registry url")
	pflag.StringVar(&serverAddr, "server.addr", "", "Server address")

	pflag.Parse()

	if err := viper.BindPFlag("logger.level", pflag.Lookup("logger.level")); err != nil {
		errs = errors.Join(errs, fmt.Errorf("bind pflag logger.level: %w", err))
	}
	if err := viper.BindPFlag("npm.registryUrl", pflag.Lookup("npm.registryUrl")); err != nil {
		errs = errors.Join(errs, fmt.Errorf("bind pflag npm.registryUrl: %w", err))
	}
	if err := viper.BindPFlag("server.addr", pflag.Lookup("server.addr")); err != nil {
		errs = errors.Join(errs, fmt.Errorf("bind pflag server.addr: %w", err))
	}

	if errs != nil {
		return nil, errs
	}

	cfg = &config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("decoding config: %w", err)
	}

	if err := cfg.Logger.level.UnmarshalText([]byte(cfg.Logger.Level)); err != nil {
		return nil, fmt.Errorf("decoding log level: %w", err)
	}

	return cfg, nil
}
