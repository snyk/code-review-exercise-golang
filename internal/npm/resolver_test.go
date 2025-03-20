package npm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/snyk/npmjs-deps-fetcher/internal/npm"
	mocksnpm "github.com/snyk/npmjs-deps-fetcher/internal/npm/mocks"
)

func TestResolver_ResolvePackage(t *testing.T) {
	constraint, err := semver.NewConstraint("^1.0.5")
	require.NoError(t, err)
	pkgName := "foo"

	testCases := []struct {
		name           string
		setup          func(testing.TB) npm.PackageFetcher
		expectedNpmPkg *npm.NpmPackageVersion
		expectedErr    string
	}{
		{
			name: "fetch meta failure for root package",
			setup: func(tb testing.TB) npm.PackageFetcher {
				tb.Helper()
				fetcher := mocksnpm.NewMockPackageFetcher(gomock.NewController(t))
				fetcher.EXPECT().FetchPackageMeta(gomock.Any(), pkgName).Return(nil, errors.New("something bad happened"))
				return fetcher
			},
			expectedErr: "fetch package meta foo: something bad happened",
		},
		{
			name: "no compatible version for root package",
			setup: func(tb testing.TB) npm.PackageFetcher {
				tb.Helper()
				fetcher := mocksnpm.NewMockPackageFetcher(gomock.NewController(t))
				fetcher.EXPECT().FetchPackageMeta(gomock.Any(), pkgName).Return(&npm.PackageMeta{
					Name: pkgName,
					Versions: map[string]npm.Package{
						"1.0.1": {Name: pkgName, Version: "1.0.1"},
						"1.0.2": {Name: pkgName, Version: "1.0.2"},
					},
				}, nil)
				return fetcher
			},
			expectedErr: "resolve highest version: no compatible versions found",
		},
		{
			name: "fetch package failure for highest root package version",
			setup: func(tb testing.TB) npm.PackageFetcher {
				tb.Helper()
				fetcher := mocksnpm.NewMockPackageFetcher(gomock.NewController(t))
				fetcher.EXPECT().FetchPackageMeta(gomock.Any(), pkgName).Return(&npm.PackageMeta{
					Name: pkgName,
					Versions: map[string]npm.Package{
						"1.0.4": {Name: pkgName, Version: "1.0.1"},
						"1.0.5": {Name: pkgName, Version: "1.0.5"},
						"1.0.6": {Name: pkgName, Version: "1.0.6"},
					},
				}, nil)
				fetcher.EXPECT().FetchPackage(gomock.Any(), pkgName, "1.0.6").Return(nil, errors.New("something bad happened"))
				return fetcher
			},
			expectedErr: "fetch package foo/1.0.6: something bad happened",
		},
		{
			name: "invalid package dependency version constraint",
			setup: func(tb testing.TB) npm.PackageFetcher {
				tb.Helper()
				fetcher := mocksnpm.NewMockPackageFetcher(gomock.NewController(t))
				fetcher.EXPECT().FetchPackageMeta(gomock.Any(), pkgName).Return(&npm.PackageMeta{
					Name: pkgName,
					Versions: map[string]npm.Package{
						"1.0.5": {Name: pkgName, Version: "1.0.5"},
					},
				}, nil)
				fetcher.EXPECT().FetchPackage(gomock.Any(), pkgName, "1.0.5").Return(&npm.Package{
					Name:         pkgName,
					Version:      "1.0.5",
					Dependencies: map[string]string{"bar": "latest"},
				}, nil)
				return fetcher
			},
			expectedErr: "invalid version constraint: improper constraint: latest",
		},
		{
			name: "successful resolved package",
			setup: func(tb testing.TB) npm.PackageFetcher {
				tb.Helper()
				fetcher := mocksnpm.NewMockPackageFetcher(gomock.NewController(t))
				fetcher.EXPECT().FetchPackageMeta(gomock.Any(), pkgName).Return(&npm.PackageMeta{
					Name: pkgName,
					Versions: map[string]npm.Package{
						"1.0.4": {Name: pkgName, Version: "1.0.1"},
						"1.0.5": {Name: pkgName, Version: "1.0.5"},
						"1.0.8": {Name: pkgName, Version: "1.0.8"},
						"2.0.0": {Name: pkgName, Version: "1.0.8"},
					},
				}, nil)
				fetcher.EXPECT().FetchPackage(gomock.Any(), pkgName, "1.0.8").Return(&npm.Package{
					Name:         pkgName,
					Version:      "1.0.8",
					Dependencies: map[string]string{"bar": "^2.0.1", "baz": "1.x"},
				}, nil)
				fetcher.EXPECT().FetchPackageMeta(gomock.Any(), "bar").Return(&npm.PackageMeta{
					Name: pkgName,
					Versions: map[string]npm.Package{
						"1.0.0": {Name: "bar", Version: "1.0.0"},
						"2.0.0": {Name: "bar", Version: "2.0.0"},
						"2.0.1": {Name: "bar", Version: "2.0.1"},
						"3.0.0": {Name: "bar", Version: "3.0.0"},
					},
				}, nil)
				fetcher.EXPECT().FetchPackage(gomock.Any(), "bar", "2.0.1").Return(&npm.Package{
					Name:    "bar",
					Version: "2.0.1",
				}, nil)
				fetcher.EXPECT().FetchPackageMeta(gomock.Any(), "baz").Return(&npm.PackageMeta{
					Name: pkgName,
					Versions: map[string]npm.Package{
						"1.0.0": {Name: "baz", Version: "1.0.0"},
						"1.0.1": {Name: "baz", Version: "1.0.1"},
						"1.0.2": {Name: "baz", Version: "1.0.2"},
						"1.1.0": {Name: "baz", Version: "1.1.0"},
					},
				}, nil)
				fetcher.EXPECT().FetchPackage(gomock.Any(), "baz", "1.1.0").Return(&npm.Package{
					Name:    "baz",
					Version: "1.1.0",
				}, nil)
				return fetcher
			},
			expectedNpmPkg: &npm.NpmPackageVersion{
				Name:    pkgName,
				Version: "1.0.8",
				Dependencies: map[string]*npm.NpmPackageVersion{
					"bar": {
						Name:         "bar",
						Version:      "2.0.1",
						Dependencies: map[string]*npm.NpmPackageVersion{},
					},
					"baz": {
						Name:         "baz",
						Version:      "1.1.0",
						Dependencies: map[string]*npm.NpmPackageVersion{},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver := npm.NewResolver(tc.setup(t))
			npmPkg := &npm.NpmPackageVersion{
				Name:         pkgName,
				Dependencies: map[string]*npm.NpmPackageVersion{},
			}

			err := resolver.ResolvePackage(context.Background(), constraint, npmPkg)

			if tc.expectedErr == "" {
				assert.Equal(t, tc.expectedNpmPkg, npmPkg)
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
