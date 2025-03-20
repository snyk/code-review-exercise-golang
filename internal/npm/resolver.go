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

// PackageResolver resolves the metadata and dependencies of a given [Package],
// based on its name and a version constraint.
func (r Resolver) ResolvePackage(ctx context.Context, constraint *semver.Constraints, npmPkg *NpmPackageVersion) error {
	meta, err := r.client.FetchPackageMeta(ctx, npmPkg.Name)
	if err != nil {
		return fmt.Errorf("fetch package meta %s: %w", npmPkg.Name, err)
	}

	version, err := semverutil.ResolveHighestVersion(constraint, maps.Keys(meta.Versions))
	if err != nil {
		return fmt.Errorf("resolve highest version: %w", err)
	}
	npmPkg.Version = version

	pkg, err := r.client.FetchPackage(ctx, npmPkg.Name, version)
	if err != nil {
		return fmt.Errorf("fetch package %s/%s: %w", npmPkg.Name, version, err)
	}

	for depName, depConstraintStr := range pkg.Dependencies {
		depConstraint, err := semver.NewConstraint(depConstraintStr)
		if err != nil {
			return fmt.Errorf("invalid version constraint: %w", err)
		}

		npmPkg.Dependencies[depName] = &NpmPackageVersion{
			Name:         depName,
			Dependencies: map[string]*NpmPackageVersion{},
		}

		r.ResolvePackage(ctx, depConstraint, npmPkg.Dependencies[depName]) //nolint:errcheck // best effort
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
