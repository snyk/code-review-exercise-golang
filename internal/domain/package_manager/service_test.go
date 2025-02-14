package packagemanager_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	packagegetter "github.com/snyk/npmjs-deps-fetcher/internal/domain/package_getter"
	packagemanager "github.com/snyk/npmjs-deps-fetcher/internal/domain/package_manager"
	"github.com/snyk/npmjs-deps-fetcher/internal/npm"
)

type PackageManagerSuite struct {
	suite.Suite
	PackageManagerService packagemanager.PackageManagerService
	Getter                *packagegetter.FakePackageGetter
}

func Test_PackageManagerService_TestSuite(t *testing.T) {
	suite.Run(t, new(PackageManagerSuite))
}

func (s *PackageManagerSuite) SetupTest() {
	fakeGetter := packagegetter.NewFakePackageGetter()
	s.Getter = &fakeGetter
	s.PackageManagerService = packagemanager.NewPackageManagerService(s.Getter)
}

func (s *PackageManagerSuite) TearDownTest() {
}

func (s *PackageManagerSuite) Test_ResolvePackage_WithValidInput_ReturnPackageData() {
	name := "react"
	version := "16.3.0"

	versionMap := map[string]npm.Package{
		"16.3.0": {
			Name:         name,
			Version:      version,
			Dependencies: map[string]string{},
		},
	}

	s.Getter.OverrideVersions = versionMap
	result, err := s.PackageManagerService.ResolvePackage(name, version)
	s.Require().NoError(err)

	s.Equal(name, result.Name)
	s.Equal(version, result.Version)
	s.Equal(map[string]string{}, result.Dependencies)
}

func (s *PackageManagerSuite) Test_ResolvePackage_WithUnexistingPackage_ShouldError() {
	name := "react"
	version := "16.3.0"

	s.Getter.FetchPackageMetaError = true
	_, err := s.PackageManagerService.ResolvePackage(name, version)
	s.Require().Error(err)
	// TODO: switch for error type check
	s.Contains(err.Error(), "failed to get package")
}

func (s *PackageManagerSuite) Test_ResolvePackage_WithUnmatchingVersion_ShouldError() {
	name := "react"
	version := "16.3.0"

	versionMap := map[string]npm.Package{
		"0.0.0": {
			Name:         name,
			Version:      "0.0.0",
			Dependencies: map[string]string{},
		},
	}

	s.Getter.OverrideVersions = versionMap

	s.Getter.FetchPackageError = true
	_, err := s.PackageManagerService.ResolvePackage(name, version)
	s.Require().Error(err)
	// TODO: switch for error type check
	s.Contains(err.Error(), "failed to match version")
}

func (s *PackageManagerSuite) Test_ResolvePackage_WithUnexistingVersion_ShouldError() {
	name := "react"
	version := "16.3.0"

	versionMap := map[string]npm.Package{
		"16.3.0": {
			Name:         name,
			Version:      version,
			Dependencies: map[string]string{},
		},
	}

	s.Getter.OverrideVersions = versionMap

	// Can this happen? The previous part checks that the version exists, so why would
	// this not return the right info? Network Error maybe?
	s.Getter.FetchPackageError = true
	_, err := s.PackageManagerService.ResolvePackage(name, version)
	s.Require().Error(err)
	// TODO: switch for error type check
	s.Contains(err.Error(), "failed to get package "+name+" by version "+version)
}

func (s *PackageManagerSuite) Test_HighestCompatibleVersion_WithMatchingVersion_ShouldReturnIt() {
	constraint := "^16.3"
	versions := npm.PackageMeta{
		Versions: map[string]npm.Package{
			"16.3.0": {},
			"16.3.1": {},
		},
	}

	result, err := packagemanager.HighestCompatibleVersion(constraint, &versions)
	s.Require().NoError(err)

	s.Equal("16.3.1", result)
}

func (s *PackageManagerSuite) Test_HighestCompatibleVersion_WithNoMatchingVersion_ShouldError() {
	constraint := "^16.4"
	versions := npm.PackageMeta{
		Versions: map[string]npm.Package{
			"16.3.0": {},
			"16.3.1": {},
		},
	}

	_, err := packagemanager.HighestCompatibleVersion(constraint, &versions)
	s.Require().Error(err)
}
