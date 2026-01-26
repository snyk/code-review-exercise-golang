package npm_test

import (
	"context"
	"errors"
	"strconv"
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
			name: "empty dependency versions for root package",
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
					Dependencies: map[string]string{},
				}, nil)
				return fetcher
			},
			expectedNpmPkg: &npm.NpmPackageVersion{
				Name:         pkgName,
				Version:      "1.0.5",
				Dependencies: map[string]*npm.NpmPackageVersion{},
			},
		},
		{
			name: "dependency listed in a package but meta is not found",
			setup: func(tb testing.TB) npm.PackageFetcher {
				tb.Helper()
				fetcher := mocksnpm.NewMockPackageFetcher(gomock.NewController(t))
				fetcher.EXPECT().FetchPackageMeta(gomock.Any(), "foo").Return(&npm.PackageMeta{
					Name: "foo",
					Versions: map[string]npm.Package{
						"1.0.5": {Name: "foo", Version: "1.0.5"},
					},
				}, nil)
				fetcher.EXPECT().FetchPackage(gomock.Any(), "foo", "1.0.5").Return(&npm.Package{
					Name:         "foo",
					Version:      "1.0.5",
					Dependencies: map[string]string{"bar": "1.0.0"},
				}, nil)
				fetcher.EXPECT().FetchPackageMeta(gomock.Any(), "bar").Return(nil, npm.ErrPackageNotFound).AnyTimes()
				return fetcher
			},
			// question: should the resolver fail if a dependency is missing?
			// Currently it continues resolving other dependencies and returns the ones found
			expectedNpmPkg: &npm.NpmPackageVersion{
				Name:    "foo",
				Version: "1.0.5",
				Dependencies: map[string]*npm.NpmPackageVersion{
					"bar": {
						Name:         "bar",
						Version:      "",
						Dependencies: map[string]*npm.NpmPackageVersion{},
					},
				},
			},
			expectedErr: "",
		},
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
					Name: "bar",
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
					Name: "baz",
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
		{
			name: "Dependency listed but no versions satisfy constraint",
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
					Dependencies: map[string]string{"bar": "^2.0.0"},
				}, nil)
				fetcher.EXPECT().FetchPackageMeta(gomock.Any(), "bar").Return(&npm.PackageMeta{
					Name: "bar",
					Versions: map[string]npm.Package{
						"1.0.0": {Name: "bar", Version: "1.0.0"},
						"1.1.0": {Name: "bar", Version: "1.1.0"},
						"1.5.0": {Name: "bar", Version: "1.5.0"},
					},
				}, nil)
				return fetcher
			},
			expectedNpmPkg: &npm.NpmPackageVersion{
				Name:    pkgName,
				Version: "1.0.5",
				Dependencies: map[string]*npm.NpmPackageVersion{
					"bar": {
						Name:         "bar",
						Version:      "",
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

// This test demonstrates that ResolvePackage correctly handles cyclic dependencies
// by tracking visited packages
func TestResolver_ResolvePackage_Cyclic(t *testing.T) {
	constraint, err := semver.NewConstraint("^1.0.5")
	require.NoError(t, err)

	fetcher := mocksnpm.NewMockPackageFetcher(gomock.NewController(t))
	fetcher.EXPECT().FetchPackageMeta(gomock.Any(), "foo").Return(&npm.PackageMeta{
		Name: "foo",
		Versions: map[string]npm.Package{
			"1.0.5": {Name: "foo", Version: "1.0.5"},
		},
	}, nil).AnyTimes()
	fetcher.EXPECT().FetchPackage(gomock.Any(), "foo", "1.0.5").Return(&npm.Package{
		Name:         "foo",
		Version:      "1.0.5",
		Dependencies: map[string]string{"bar": "1.0.0"},
	}, nil).AnyTimes()
	fetcher.EXPECT().FetchPackageMeta(gomock.Any(), "bar").Return(&npm.PackageMeta{
		Name: "bar",
		Versions: map[string]npm.Package{
			"1.0.0": {Name: "bar", Version: "1.0.0"},
		},
	}, nil).AnyTimes()
	fetcher.EXPECT().FetchPackage(gomock.Any(), "bar", "1.0.0").Return(&npm.Package{
		Name:         "bar",
		Version:      "1.0.0",
		Dependencies: map[string]string{"foo": "1.0.5"},
	}, nil).AnyTimes()

	resolver := npm.NewResolver(fetcher)
	npmPkg := &npm.NpmPackageVersion{
		Name:         "foo",
		Dependencies: map[string]*npm.NpmPackageVersion{},
	}

	err = resolver.ResolvePackage(context.Background(), constraint, npmPkg)
	require.NoError(t, err)

	assert.Equal(t, "foo", npmPkg.Name)
	assert.Equal(t, "1.0.5", npmPkg.Version)

	assert.Contains(t, npmPkg.Dependencies, "bar")
	assert.Equal(t, "1.0.0", npmPkg.Dependencies["bar"].Version)
}

// Test that ResolvePackage handles a long chain of dependencies.
func TestResolver_ResolvePackage_LargeDependencyChain(t *testing.T) {
	const depth = 50

	constraint, err := semver.NewConstraint("1.0.5")
	require.NoError(t, err)

	fetcher := mocksnpm.NewMockPackageFetcher(gomock.NewController(t))
	fetcher.EXPECT().FetchPackageMeta(gomock.Any(), "foo").Return(&npm.PackageMeta{
		Name: "foo",
		Versions: map[string]npm.Package{
			"1.0.5": {Name: "foo", Version: "1.0.5"},
		},
	}, nil).AnyTimes()
	fetcher.EXPECT().FetchPackage(gomock.Any(), "foo", "1.0.5").Return(&npm.Package{
		Name:         "foo",
		Version:      "1.0.5",
		Dependencies: map[string]string{"1": "1.0.0"},
	}, nil).AnyTimes()

	for i := 1; i < depth; i++ {
		name := strconv.Itoa(i)

		fetcher.EXPECT().FetchPackageMeta(gomock.Any(), name).Return(&npm.PackageMeta{
			Name: name,
			Versions: map[string]npm.Package{
				"1.0.0": {Name: name, Version: "1.0.0"},
			},
		}, nil).AnyTimes()
		fetcher.EXPECT().FetchPackage(gomock.Any(), name, "1.0.0").Return(&npm.Package{
			Name:         name,
			Version:      "1.0.0",
			Dependencies: map[string]string{strconv.Itoa(i + 1): "1.0.0"},
		}, nil).AnyTimes()
	}

	//Add last depency that has no other dependencies
	name := strconv.Itoa(depth)

	fetcher.EXPECT().FetchPackageMeta(gomock.Any(), name).Return(&npm.PackageMeta{
		Name: name,
		Versions: map[string]npm.Package{
			"1.0.0": {Name: name, Version: "1.0.0"},
		},
	}, nil).AnyTimes()
	fetcher.EXPECT().FetchPackage(gomock.Any(), name, "1.0.0").Return(&npm.Package{
		Name:         name,
		Version:      "1.0.0",
		Dependencies: map[string]string{},
	}, nil).AnyTimes()

	resolver := npm.NewResolver(fetcher)
	npmPkg := &npm.NpmPackageVersion{
		Name:         "foo",
		Dependencies: map[string]*npm.NpmPackageVersion{},
	}

	err = resolver.ResolvePackage(context.Background(), constraint, npmPkg)

	assert.Equal(t, "foo", npmPkg.Name)
	assert.NoError(t, err)

	// Check the dependency chain
	current := npmPkg
	for i := 1; i <= depth && current != nil; i++ {
		dep := strconv.Itoa(i)
		require.Contains(t, current.Dependencies, dep)
		current = current.Dependencies[dep]
	}
}

// Test that ResolvePackage handles a wide chain of dependencies.
func TestResolver_ResolvePackage_WideDependencyChain(t *testing.T) {

	constraint, err := semver.NewConstraint("1.0.5")
	require.NoError(t, err)

	fetcher := mocksnpm.NewMockPackageFetcher(gomock.NewController(t))
	fetcher.EXPECT().FetchPackageMeta(gomock.Any(), "foo").Return(&npm.PackageMeta{
		Name: "foo",
		Versions: map[string]npm.Package{
			"1.0.5": {Name: "foo", Version: "1.0.5"},
		},
	}, nil).AnyTimes()
	fetcher.EXPECT().FetchPackage(gomock.Any(), "foo", "1.0.5").Return(&npm.Package{
		Name:         "foo",
		Version:      "1.0.5",
		Dependencies: map[string]string{"1": "1.0.0", "2": "1.0.0", "3": "1.0.0", "4": "1.0.0"},
	}, nil).AnyTimes()

	// Create 4 main children each with 4 dependencies
	children := []string{"1", "2", "3", "4"}
	for _, name := range children {
		fetcher.EXPECT().FetchPackageMeta(gomock.Any(), name).Return(&npm.PackageMeta{
			Name: name,
			Versions: map[string]npm.Package{
				"1.0.0": {Name: name, Version: "1.0.0"},
			},
		}, nil).AnyTimes()
		fetcher.EXPECT().FetchPackage(gomock.Any(), name, "1.0.0").Return(&npm.Package{
			Name:    name,
			Version: "1.0.0",
			Dependencies: map[string]string{
				name + "a": "1.0.0",
				name + "b": "1.0.0",
				name + "c": "1.0.0",
				name + "d": "1.0.0",
			},
		}, nil).AnyTimes()
	}

	// Each child has 4 dependencies with no further dependencies
	edges := []string{"1a", "1b", "1c", "1d", "2a", "2b", "2c", "2d", "3a", "3b", "3c", "3d",
		"4a", "4b", "4c", "4d"}
	for _, name := range edges {
		fetcher.EXPECT().FetchPackageMeta(gomock.Any(), name).Return(&npm.PackageMeta{
			Name: name,
			Versions: map[string]npm.Package{
				"1.0.0": {Name: name, Version: "1.0.0"},
			},
		}, nil).AnyTimes()
		fetcher.EXPECT().FetchPackage(gomock.Any(), name, "1.0.0").Return(&npm.Package{
			Name:         name,
			Version:      "1.0.0",
			Dependencies: map[string]string{},
		}, nil).AnyTimes()
	}

	resolver := npm.NewResolver(fetcher)
	npmPkg := &npm.NpmPackageVersion{
		Name:         "foo",
		Dependencies: map[string]*npm.NpmPackageVersion{},
	}

	err = resolver.ResolvePackage(context.Background(), constraint, npmPkg)

	assert.Equal(t, "foo", npmPkg.Name)
	assert.NoError(t, err)

	// Check that there are 4 main dependencies
	assert.Len(t, npmPkg.Dependencies, 4)

	// Check that each main dependency has 4 dependencies
	for _, v := range npmPkg.Dependencies {
		assert.Len(t, v.Dependencies, 4)
	}
}
