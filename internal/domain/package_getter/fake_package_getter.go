package packagegetter

import (
	"errors"
)

type FakePackageGetter struct {
	OverrideVersions      map[string]NpmPackageResponse
	FetchPackageError     bool
	FetchPackageMetaError bool
}

func NewFakePackageGetter() FakePackageGetter {
	return FakePackageGetter{
		OverrideVersions:      map[string]NpmPackageResponse{},
		FetchPackageError:     false,
		FetchPackageMetaError: false,
	}
}

func (fpg FakePackageGetter) FetchPackage(name, version string) (*NpmPackageResponse, error) {
	if fpg.FetchPackageError {
		return nil, errors.New("FetchPackage failed")
	}
	wantedPackage := NpmPackageResponse{
		Name:         name,
		Version:      version,
		Dependencies: map[string]string{},
	}
	return &wantedPackage, nil
}

func (fpg FakePackageGetter) FetchPackageMeta(name string) (*NpmPackageMetaResponse, error) {
	if fpg.FetchPackageMetaError {
		return nil, errors.New("FetchPackageMeta fail")
	}

	packageMeta := NpmPackageMetaResponse{
		Versions: fpg.OverrideVersions,
	}

	return &packageMeta, nil
}

func (fpg *FakePackageGetter) AddPackagesToMetaResponse(OverrideVersions map[string]NpmPackageResponse) {
	fpg.OverrideVersions = OverrideVersions
}

func (fpg *FakePackageGetter) ResetPackagesToMetaResponse() {
	fpg.OverrideVersions = map[string]NpmPackageResponse{}
}
