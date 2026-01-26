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
	}
)

// NewResolver constructs a [Resolver] with the provider [PackageFetcher] client.
func NewResolver(client PackageFetcher) Resolver {
	return Resolver{client: client}
}

// suggestion: update the comment to document its recursive nature and any assumptions and
// edge cases like no cycle detection and unordered dependencies. This helps future maintainers
// understand the function's behavior and limitations.

// suggestion: consider a more descriptive name for ResolvePackage
// The function performs a recursive search and builds a dependency tree
// the name could reflect that behavior more clearly.

// PackageResolver resolves the metadata and dependencies of a given [Package],
// based on its name and a version constraint.
func (r Resolver) ResolvePackage(ctx context.Context, constraint *semver.Constraints, npmPkg *NpmPackageVersion) error {
	visited := make(map[string]bool)
	return r.resolvePackageRecursive(ctx, constraint, npmPkg, visited)
}

func (r Resolver) resolvePackageRecursive(ctx context.Context, constraint *semver.Constraints, npmPkg *NpmPackageVersion, visited map[string]bool) error {
	meta, err := r.client.FetchPackageMeta(ctx, npmPkg.Name)
	if err != nil {
		return fmt.Errorf("fetch package meta %s: %w", npmPkg.Name, err)
	}

	version, err := semverutil.ResolveHighestVersion(constraint, maps.Keys(meta.Versions))
	if err != nil {
		return fmt.Errorf("resolve highest version: %w", err)
	}
	npmPkg.Version = version

	key := fmt.Sprintf("%s@%s", npmPkg.Name, version)
	if _, ok := visited[key]; ok {
		return nil
	}
	visited[key] = true
	defer delete(visited, key)

	pkg, err := r.client.FetchPackage(ctx, npmPkg.Name, version)
	if err != nil {
		return fmt.Errorf("fetch package %s/%s: %w", npmPkg.Name, version, err)
	}

	// issue: npm dependencies can contain cycles and without detection this recursive
	// approach could lead to infinite recursion or stack overflow crash
	for depName, depConstraintStr := range pkg.Dependencies {
		depConstraint, err := semver.NewConstraint(depConstraintStr)
		if err != nil {
			return fmt.Errorf("invalid version constraint: %w", err)
		}

		npmPkg.Dependencies[depName] = &NpmPackageVersion{
			Name:         depName,
			Dependencies: map[string]*NpmPackageVersion{},
		}

		// suggestion: For large problems DFS may lead to stack overflows, to improve performance
		// and reliability consider using an approach with a cache of resolved dependencies to avoid
		// repeated work
		r.resolvePackageRecursive(ctx, depConstraint, npmPkg.Dependencies[depName], visited) //nolint:errcheck // best effort
	}

	return nil
}

func (r Resolver) resolvePackageHighestVersion(ctx context.Context, name string, constraint *semver.Constraints) (string, error) {
	meta, err := r.client.FetchPackageMeta(ctx, name)
	if err != nil {
		return "", fmt.Errorf("fetch package meta %s: %w", name, err)
	}

	version, err := semverutil.ResolveHighestVersion(constraint, maps.Keys(meta.Versions))
	if err != nil {
		return "", fmt.Errorf("resolve highest version: %w", err)
	}

	return version, nil
}
