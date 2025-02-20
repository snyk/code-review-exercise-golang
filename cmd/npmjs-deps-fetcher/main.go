package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/snyk/npmjs-deps-fetcher/internal/handler"
	"github.com/snyk/npmjs-deps-fetcher/internal/npm"
)

func main() {
	if err := run(); err != nil {
		slog.Error("runtime error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := parseConfig()
	if err != nil {
		return fmt.Errorf("parse configuration: %w", err)
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: &cfg.Logger.Level,
	}))
	slog.SetDefault(log)

	client, err := npm.NewClient(cfg.NPM)
	if err != nil {
		return fmt.Errorf("create NPM client: %w", err)
	}
	resolver := npm.NewResolver(client)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthcheck", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	mux.HandleFunc("GET /package/{packageName}/{packageVersion}", handler.PackageVersion(log.Handler(), resolver))

	srv := http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           mux,
		ReadHeaderTimeout: cfg.Server.readHeaderTimeout,
		WriteTimeout:      cfg.Server.writeTimeout,
	}

	log.Info("HTTP server running", slog.String("addr", cfg.Server.Addr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server exited ungracefully: %w", err)
	}

	return nil
}
