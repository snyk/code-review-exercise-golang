//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var appAddr string

func TestMain(m *testing.M) {
	srv := httptest.NewServer(registryHandler())

	app := application{registryURL: srv.URL}
	if err := app.start(); err != nil {
		log.Printf("app run error: %s", err)
		os.Exit(1)
	}

	code := m.Run()

	app.close()
	srv.Close()
	os.Exit(code)
}

func TestHealthcheckEndpoint(t *testing.T) {
	ctx := context.Background()
	url := fmt.Sprintf("http://%s/healthcheck", appAddr)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestPackageNameVersionEndpoint(t *testing.T) {
	ctx := context.Background()
	url := fmt.Sprintf("http://%s/package/react/16.13.0", appAddr)

	expectedBody, err := os.ReadFile("testdata/expect_react_16.13.0.json")
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, expectedBody, body)
}
