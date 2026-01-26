package npm

import (
	"context"
	"fmt"
	"maps"

	"github.com/Masterminds/semver/v3"

	semverutil "github.com/snyk/npmjs-deps-fetcher/internal/semver"
)

//go:generate go tool mockgen -destination=mocks/resolver.go -source=resolver.go -package mocksnpm

type (
	// PackageFetcher fetches data of NPM packages.
	PackageFetcher interface {
		// FetchPackage fetches the [Package] information of an NPM package at a given version.
		FetchPackage(ctx context.Context, name, version string) (*Package, error)
		// FetchPackageMeta fetches the [PackageMeta] metadata of an NPM package.
		FetchPackageMeta(ctx context.Context, name string) (*PackageMeta, error)
	}

	// Resolver resolves an NPM package, as well as its dependencies.
	Resolver struct {
		client PackageFetcher
		cache  map[string]*NpmPackageVersion
	}
)

// NewResolver constructs a [Resolver] with the provider [PackageFetcher] client.
func NewResolver(client PackageFetcher) Resolver {
	return Resolver{client: client, cache: map[string]*NpmPackageVersion{}}
}

// suggestion: consider a more descriptive name for ResolvePackage
// The function performs a recursive search and builds a dependency tree
// the name could reflect that behavior more clearly.

// ResolvePackage resolves the highest matching version of an NPM package
// for the given version constraint and recursively builds its dependency tree.
// An internal cache is used so we can reuse already resolved packages and detect
// cycles to prevent infinite recursion.
//
// This is best effort, failures in resolving individual dependencies do not stop
// resolution of the parent package.
func (r Resolver) ResolvePackage(ctx context.Context, constraint *semver.Constraints, npmPkg *NpmPackageVersion) error {
	meta, err := r.client.FetchPackageMeta(ctx, npmPkg.Name)
	if err != nil {
		npmPkg.Status = StatusFailed
		npmPkg.Error = err.Error()
		return fmt.Errorf("fetch package meta %s: %w", npmPkg.Name, err)
	}

	version, err := semverutil.ResolveHighestVersion(constraint, maps.Keys(meta.Versions))
	if err != nil {
		npmPkg.Status = StatusFailed
		npmPkg.Error = err.Error()
		return fmt.Errorf("resolve highest version: %w", err)
	}
	npmPkg.Version = version

	// Cyclic dependencies are detected via the cache.
	// We reuse the already resolved package version in the cache
	// without expanding its dependencies to avoid infinite recursion
	key := fmt.Sprintf("%s@%s", npmPkg.Name, version)
	if cached, ok := r.cache[key]; ok {
		*npmPkg = *cached
		npmPkg.Status = StatusCyclic
		return nil
	}

	pkg, err := r.client.FetchPackage(ctx, npmPkg.Name, version)
	if err != nil {
		npmPkg.Status = StatusFailed
		npmPkg.Error = err.Error()
		return fmt.Errorf("fetch package %s/%s: %w", npmPkg.Name, version, err)
	}

	npmPkg.Status = StatusResolved
	r.cache[key] = npmPkg

	for depName, depConstraintStr := range pkg.Dependencies {
		// Check if context was cancelled before processing each dependency
		if err := ctx.Err(); err != nil {
			npmPkg.Dependencies[depName] = &NpmPackageVersion{
				Name:         depName,
				Dependencies: map[string]*NpmPackageVersion{},
				Status:       StatusFailed,
				Error:        err.Error(),
			}
			continue
		}

		depConstraint, err := semver.NewConstraint(depConstraintStr)
		if err != nil {
			return fmt.Errorf("invalid version constraint: %w", err)
		}

		npmPkg.Dependencies[depName] = &NpmPackageVersion{
			Name:         depName,
			Dependencies: map[string]*NpmPackageVersion{},
		}

		// Any errors from resolving are swallowed as this is best effort and can be found
		// in the NpmPackageVersions Status/Error fields.
		r.ResolvePackage(ctx, depConstraint, npmPkg.Dependencies[depName]) //nolint:errcheck  // best effort
	}

	return nil
}
