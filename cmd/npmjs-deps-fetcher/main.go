package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	packagemanager "github.com/snyk/npmjs-deps-fetcher/internal/domain/package_manager"
	"github.com/snyk/npmjs-deps-fetcher/internal/handler"
	"github.com/snyk/npmjs-deps-fetcher/internal/npm"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	if err := run(log); err != nil {
		log.Error("runtime error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := parseConfig()
	if err != nil {
		return fmt.Errorf("parse configuration: %w", err)
	}

	npmClient, err := npm.NewClient(cfg.NPM)
	if err != nil {
		return fmt.Errorf("create NPM client: %w", err)
	}
	mngr := packagemanager.NewPackageManagerService(npmClient)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /package/{package}/{version}", handler.PackageVersion(log.Handler(), mngr))

	srv := http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: time.Second * 10,
		WriteTimeout:      time.Second * 30,
	}

	log.Info("HTTP server running", slog.String("addr", cfg.ListenAddr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server exited ungracefully: %w", err)
	}

	return nil
}
