package handler

import (
	"context"
	"encoding/json"
	"errors"
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
	ResolvePackage(ctx context.Context, constraint *semver.Constraints, npmPkg *npm.NpmPackageVersion) error
}

// PackageVersion is the [http.HandlerFunc] for GET /package/{package}/{version}.
func PackageVersion(logHandler slog.Handler, resolver PackageResolver) http.HandlerFunc {
	log := slog.New(logHandler)

	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		// suggestion: consider validating and santising package name and version
		// inputs to prevent potential security issues like injection attacks.

		pkgName, pkgVersion := req.PathValue("packageName"), req.PathValue("packageVersion")

		if pkgName == "" && pkgVersion == "" {
			writeError(w, log, http.StatusBadRequest, "invalid request path")
			return
		}
		if pkgVersion == "" {
			writeError(w, log, http.StatusBadRequest, "missing version")
			return
		}
		if pkgName == "" {
			writeError(w, log, http.StatusBadRequest, "missing package name")
			return
		}

		w.Header().Set("Content-Type", "application/json")

		constraint, err := semver.NewConstraint(pkgVersion)
		if err != nil {
			log.Debug("invalid version constraint", slog.String("error", err.Error()))
			writeError(w, log, http.StatusBadRequest, "invalid version constraint")
			return
		}

		npmPkg := &npm.NpmPackageVersion{Name: pkgName, Dependencies: map[string]*npm.NpmPackageVersion{}}

		err = resolver.ResolvePackage(ctx, constraint, npmPkg)
		if errors.Is(err, npm.ErrPackageNotFound) {
			log.Debug("package not found", slog.String("name", pkgName), slog.String("version", pkgVersion))
			writeError(w, log, http.StatusNotFound, "package not found")
			return
		}

		// question: Should all reolvePackage errors be treated as internal server errors?
		// Some errors could be due to client input or being unable to reach npm registry
		// which could be handled better than returning a generic 500 error
		if err != nil {
			log.Error("deps resolution error", slog.String("error", err.Error()))
			writeError(w, log, http.StatusInternalServerError, "internal server error")
			return
		}

		if err := json.NewEncoder(w).Encode(npmPkg); err != nil {
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
