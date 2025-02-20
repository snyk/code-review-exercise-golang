//go:build integration

package test

import (
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthcheckEndpoint(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/healthcheck")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestPackageNameVersionEndpoint(t *testing.T) {
	expectedBody, err := os.ReadFile("testdata/react_16.13.0.json")
	require.NoError(t, err)

	resp, err := http.Get("http://localhost:8080/package/react/16.13.0")

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, expectedBody, body)
}
