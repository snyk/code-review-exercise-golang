package packagegetter

import (
	"context"
	"errors"

	"github.com/snyk/npmjs-deps-fetcher/internal/npm"
)

type FakePackageGetter struct {
	OverrideVersions      map[string]npm.Package
	FetchPackageError     bool
	FetchPackageMetaError bool
}

func NewFakePackageGetter() FakePackageGetter {
	return FakePackageGetter{
		OverrideVersions:      map[string]npm.Package{},
		FetchPackageError:     false,
		FetchPackageMetaError: false,
	}
}

func (fpg FakePackageGetter) FetchPackage(_ context.Context, name, version string) (*npm.Package, error) {
	if fpg.FetchPackageError {
		return nil, errors.New("FetchPackage failed")
	}
	wantedPackage := npm.Package{
		Name:         name,
		Version:      version,
		Dependencies: map[string]string{},
	}
	return &wantedPackage, nil
}

func (fpg FakePackageGetter) FetchPackageMeta(_ context.Context, name string) (*npm.PackageMeta, error) {
	if fpg.FetchPackageMetaError {
		return nil, errors.New("FetchPackageMeta fail")
	}

	packageMeta := npm.PackageMeta{
		Versions: fpg.OverrideVersions,
	}

	return &packageMeta, nil
}

func (fpg *FakePackageGetter) AddPackagesToMetaResponse(OverrideVersions map[string]npm.Package) {
	fpg.OverrideVersions = OverrideVersions
}

func (fpg *FakePackageGetter) ResetPackagesToMetaResponse() {
	fpg.OverrideVersions = map[string]npm.Package{}
}
