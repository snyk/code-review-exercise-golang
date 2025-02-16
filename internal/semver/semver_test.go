package semver_test

import (
	"slices"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	semverutil "github.com/snyk/npmjs-deps-fetcher/internal/semver"
)

func TestResolveHighestVersion(t *testing.T) {
	constraint, err := semver.NewConstraint("^1.0.5")
	require.NoError(t, err)

	testCases := []struct {
		name            string
		versions        []string
		expectedVersion string
		expectedErr     string
	}{
		{
			name:        "strict parsing error",
			versions:    []string{"^1.0.2", "1.x"},
			expectedErr: "version ^1.0.2: Invalid characters in version\nversion 1.x: Invalid Semantic Version",
		},
		{
			name:        "empty version list",
			versions:    []string{},
			expectedErr: "no compatible versions found",
		},
		{
			name:        "no compatible versions",
			versions:    []string{"0.0.1", "0.0.2", "1.0.0", "1.0.1"},
			expectedErr: "no compatible versions found",
		},
		{
			name:            "compatible version",
			versions:        []string{"0.0.1", "0.0.2", "1.0.0", "1.0.1", "1.0.5", "1.0.6", "2.0.7"},
			expectedVersion: "1.0.6",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
