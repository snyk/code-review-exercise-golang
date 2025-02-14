package packagemanager

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"

	"github.com/snyk/npmjs-deps-fetcher/internal/npm"
)

type PackageManagerService struct {
	fetcher PackageFetcher
}

type PackageFetcher interface {
	FetchPackage(ctx context.Context, name, version string) (*npm.Package, error)
	FetchPackageMeta(ctx context.Context, name string) (*npm.PackageMeta, error)
}

func NewPackageManagerService(pkgFetcher PackageFetcher) PackageManagerService {
	return PackageManagerService{
		fetcher: pkgFetcher,
	}
}

func (pms PackageManagerService) ResolvePackage(pkgName, pkgVersion string) (*npm.Package, error) {
	pkgMeta, err := pms.fetcher.FetchPackageMeta(context.TODO(), pkgName)
	if err != nil {
		return nil, fmt.Errorf("failed to get package: %w", err)
	}

	concreteVersion, err := HighestCompatibleVersion(pkgVersion, pkgMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to match version: %w", err)
	}

	rootPkg := &npm.Package{Name: pkgName, Version: concreteVersion, Dependencies: map[string]string{}}
	npmPkg, err := pms.fetcher.FetchPackage(context.TODO(), rootPkg.Name, rootPkg.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get package %v by version %v : %w", rootPkg.Name, rootPkg.Version, err)
	}

	for dependencyName, dependencyVersionConstraint := range npmPkg.Dependencies {
		pkgMeta, err := pms.fetcher.FetchPackageMeta(context.TODO(), dependencyName)
		if err != nil {
			return nil, fmt.Errorf("failed to get sub-package: %w", err)
		}
		concreteVersion, err := HighestCompatibleVersion(dependencyVersionConstraint, pkgMeta)
		if err != nil {
			return nil, fmt.Errorf("failed to match version for sub-package: %w", err)
		}
		rootPkg.Dependencies[dependencyName] = concreteVersion
	}

	return rootPkg, nil
}

func filterCompatibleVersions(constraint *semver.Constraints, pkgMeta *npm.PackageMeta) semver.Collection {
	var compatible semver.Collection
	for version := range pkgMeta.Versions {
		semVer, err := semver.NewVersion(version)
		if err != nil {
			continue
		}
		if constraint.Check(semVer) {
			compatible = append(compatible, semVer)
		}
	}
	return compatible
}

func HighestCompatibleVersion(constraintStr string, versions *npm.PackageMeta) (string, error) {
	constraint, err := semver.NewConstraint(constraintStr)
	if err != nil {
		return "", err
	}
	filtered := filterCompatibleVersions(constraint, versions)
	sort.Sort(filtered)
	if len(filtered) == 0 {
		return "", errors.New("no compatible versions found")
	}
	return filtered[len(filtered)-1].String(), nil
}
