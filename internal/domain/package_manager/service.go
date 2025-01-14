package packagemanager

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"

	getter "github.com/snyk/npmjs-deps-fetcher/internal/domain/package_getter"
)

type PackageManagerService struct {
	packageGetter getter.PackageGetter
}

func NewPackageManagerService(packageGetter getter.PackageGetter) PackageManagerService {
	return PackageManagerService{
		packageGetter: packageGetter,
	}
}

func (pms PackageManagerService) GetPackageDependencies(pkgName, pkgVersion string) (*getter.NpmPackageVersion, error) {
	pkgMeta, err := pms.packageGetter.FetchPackageMeta(pkgName)
	if err != nil {
		return nil, fmt.Errorf("failed to get package: %w", err)
	}

	concreteVersion, err := HighestCompatibleVersion(pkgVersion, pkgMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to match version: %w", err)
	}

	rootPkg := &getter.NpmPackageVersion{Name: pkgName, Version: concreteVersion, Dependencies: map[string]string{}}
	npmPkg, err := pms.packageGetter.FetchPackage(rootPkg.Name, rootPkg.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get package %v by version %v : %w", rootPkg.Name, rootPkg.Version, err)
	}

	for dependencyName, dependencyVersionConstraint := range npmPkg.Dependencies {
		pkgMeta, err := pms.packageGetter.FetchPackageMeta(dependencyName)
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

func filterCompatibleVersions(constraint *semver.Constraints, pkgMeta *getter.NpmPackageMetaResponse) semver.Collection {
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

func HighestCompatibleVersion(constraintStr string, versions *getter.NpmPackageMetaResponse) (string, error) {
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
