package semver_test

import (
	"slices"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	semverutil "github.com/snyk/npmjs-deps-fetcher/internal/semver"
)

func TestResolveVersion(t *testing.T) {
	testCases := []struct {
		name            string
		constraint      string
		versions        []string
		expectedVersion string
		expectedErr     string
	}{
		{
			name:        "strict parsing error",
			constraint:  "^1.0.5",
			versions:    []string{"^1.0.2", "1.x"},
			expectedErr: "version ^1.0.2: Invalid characters in version\nversion 1.x: Invalid Semantic Version",
		},
		{
			name:        "empty version list",
			constraint:  "^1.0.5",
			versions:    []string{},
			expectedErr: "no compatible versions found",
		},
		{
			name:        "no compatible versions",
			constraint:  "^1.0.5",
			versions:    []string{"0.0.1", "0.0.2", "1.0.0", "1.0.1"},
			expectedErr: "no compatible versions found",
		},
		{
			name:            "duplicate versions",
			constraint:      "^1.0.5",
			versions:        []string{"1.0.5", "1.0.5", "1.0.6"},
			expectedVersion: "1.0.6",
		},
		{
			name:            "unordered versions",
			constraint:      "^1.0.5",
			versions:        []string{"1.0.6", "0.9.9", "1.0.5", "1.0.1"},
			expectedVersion: "1.0.6",
		},
		{
			name:            "compatible version",
			constraint:      "^1.0.5",
			versions:        []string{"0.0.1", "0.0.2", "1.0.0", "1.0.1", "1.0.5", "1.0.6", "2.0.7"},
			expectedVersion: "1.0.6",
		},
		{
			name:            "exact version match",
			constraint:      "1.0.5",
			versions:        []string{"1.0.4", "1.0.5", "1.0.6"},
			expectedVersion: "1.0.5",
		},
		{
			name:            "tilde constraint locks minor version",
			constraint:      "~1.2.3",
			versions:        []string{"1.2.3", "1.2.9", "1.3.0"},
			expectedVersion: "1.2.9",
		},
		{
			name:            "pre-release stable version preferred",
			constraint:      "^1.0.0-alpha1",
			versions:        []string{"1.0.0-alpha1", "1.0.0-alpha2", "1.0.0"},
			expectedVersion: "1.0.0",
		},
		{
			name:            "pre-release highest preferred",
			constraint:      "^1.0.1-alpha1",
			versions:        []string{"1.0.1-alpha1", "1.0.1-alpha2", "1.0.1-alpha3"},
			expectedVersion: "1.0.1-alpha3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			constraint, err := semver.NewConstraint(tc.constraint)
			require.NoError(t, err)

			version, err := semverutil.ResolveHighestVersion(constraint, slices.Values(tc.versions))

			assert.Equal(t, tc.expectedVersion, version)
			if tc.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
