package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Masterminds/semver/v3"

	"github.com/snyk/npmjs-deps-fetcher/internal/npm"
)

//go:generate go tool mockgen -destination=mocks/handler.go -source=handler.go -package mockshandler

// PackageResolver resolves the metadata and dependencies of an [npm.Package],
// based on its name and a version constraint.
type PackageResolver interface {
	ResolvePackage(ctx context.Context, name string, constraint *semver.Constraints) (*npm.Package, error)
}

// PackageVersion is the [http.HandlerFunc] for GET /package/{package}/{version}.
func PackageVersion(logHandler slog.Handler, resolver PackageResolver) http.HandlerFunc {
	log := slog.New(logHandler)

	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		w.Header().Set("Content-Type", "application/json")

		constraint, err := semver.NewConstraint(req.PathValue("packageVersion"))
		if err != nil {
			log.Debug("invalid version constraint", slog.String("error", err.Error()))
			writeError(w, log, http.StatusBadRequest, "invalid version constraint")
			return
		}

		deps, err := resolver.ResolvePackage(ctx, req.PathValue("packageName"), constraint)
		if err != nil {
			log.Error("deps resolution error", slog.String("error", err.Error()))
			writeError(w, log, http.StatusInternalServerError, "internal server error")
			return
		}

		if err := json.NewEncoder(w).Encode(deps); err != nil {
			log.Error("deps encoding error", slog.Any("error", err))
			writeError(w, log, http.StatusInternalServerError, "internal server error")
			return
		}
	}
}

func writeError(w http.ResponseWriter, log *slog.Logger, statusCode int, msg string) {
	w.WriteHeader(statusCode)
	if _, err := fmt.Fprintln(w, `{"error":"`+msg+`"}`); err != nil {
		log.Error("failed to write error", slog.String("error", err.Error()))
	}
}
