package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/snyk/npmjs-deps-fetcher/internal/npm"
)

//go:generate go tool mockgen -destination=mocks/handler.go -source=handler.go -package mockshandler

// PackageResolver resolves the metadata and dependencies of a given package,
// identified by its name and version constraint.
type PackageResolver interface {
	ResolvePackage(name, constraint string) (*npm.Package, error)
}

// PackageVersion is the [http.HandlerFunc] for GET /package/{package}/{version}.
func PackageVersion(logHandler slog.Handler, resolver PackageResolver) http.HandlerFunc {
	log := slog.New(logHandler)

	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		deps, err := resolver.ResolvePackage(req.PathValue("package"), req.PathValue("version"))
		if err != nil {
			log.Error("deps resolution error", slog.String("error", err.Error()))
			writeError(w, log)
			return
		}

		if err := json.NewEncoder(w).Encode(deps); err != nil {
			log.Error("deps encoding error", slog.Any("error", err))
			writeError(w, log)
			return
		}
	}
}

func writeError(w http.ResponseWriter, log *slog.Logger) {
	w.WriteHeader(http.StatusInternalServerError)
	if _, err := fmt.Fprintln(w, `{"error":"internal server error"}`); err != nil {
		log.Error("failed to write error", slog.String("error", err.Error()))
	}
}
