package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/snyk/npmjs-deps-fetcher/internal/api"
	packagemanager "github.com/snyk/npmjs-deps-fetcher/internal/domain/package_manager"
	"github.com/snyk/npmjs-deps-fetcher/internal/npm"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := parseConfig()
	if err != nil {
		log.Error("failed to parse configuration", slog.Any("error", err))
		os.Exit(1)
	}

	npmClient, err := npm.NewClient(cfg.NPM)
	if err != nil {
		log.Error("failed to create NPM client", slog.Any("error", err))
	}

	handler := api.New(packagemanager.NewPackageManagerService(npmClient))

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           handler,
		ReadHeaderTimeout: time.Second * 10,
		WriteTimeout:      time.Second * 30,
	}

	log.Info("HTTP server running", slog.String("addr", cfg.ListenAddr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("HTTP server exited ungracefully", slog.Any("error", err))
		os.Exit(1)
	}
}
