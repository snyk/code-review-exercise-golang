package npm_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snyk/npmjs-deps-fetcher/internal/npm"
)

func TestNewClient(t *testing.T) {
	testCases := []struct {
		name        string
		cfg         npm.ClientConfig
		expectedErr string
	}{
		{
			name: "invalid registry url configuration",
			cfg: npm.ClientConfig{
				RegistryURL: "\x7f",
				Timeout:     15 * time.Second,
			},
			expectedErr: "registry URL configuration: parse \"\\x7f\": net/url: invalid control character in URL",
		},
		{
			name: "valid configuration",
			cfg: npm.ClientConfig{
				RegistryURL: "https://a-valid-url",
				Timeout:     15 * time.Second,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := npm.NewClient(tc.cfg)

			if tc.expectedErr == "" {
				assert.NotNil(t, c)
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

type fakeTransport struct {
	resp *http.Response
	err  error
}

func (rt fakeTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return rt.resp, rt.err
}

func TestClient_FetchPackage(t *testing.T) {
	testCases := []struct {
		name        string
		transport   http.RoundTripper
		expectedErr string
		expectedPkg *npm.Package
	}{
		{
			name: "round trip error",
			transport: fakeTransport{
				err: errors.New("round trip error"),
			},
			expectedErr: "http request roundtrip for: Get \"http://localhost:8080/fake-name/fake-version\": round trip error",
		},
		{
			name: "http error response decoding error",
			transport: fakeTransport{
				resp: &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(strings.NewReader(`{"error":"json error"}`)),
				},
			},
			expectedErr: "error response decoding of \"http://localhost:8080/fake-name/fake-version\": json: cannot unmarshal object into Go value of type string",
		},
		{
			name: "http error response decoding success",
			transport: fakeTransport{
				resp: &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader(`"internal server error"`)),
				},
			},
			expectedErr: "http response for \"http://localhost:8080/fake-name/fake-version\": internal server error",
		},
		{
			name: "http response decoding error",
			transport: fakeTransport{
				resp: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`"invalid payload"`)),
				},
			},
			expectedErr: "response decoding of \"http://localhost:8080/fake-name/fake-version\": json: cannot unmarshal string into Go value of type npm.Package",
		},
		{
			name: "http response package not found",
			transport: fakeTransport{
				resp: &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader(`"not found"`)),
				},
			},
			expectedErr: npm.ErrPackageNotFound.Error(),
		},
		{
			name: "http response decoding",
			transport: fakeTransport{
				resp: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"name":"awesome","version":"1.1.1","dependencies":{"great":"^1.0.1","so-so":"^2.0.1"}}`)),
				},
			},
			expectedPkg: &npm.Package{
				Name:    "awesome",
				Version: "1.1.1",
				Dependencies: map[string]string{
					"great": "^1.0.1",
					"so-so": "^2.0.1",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := npm.NewClient(npm.ClientConfig{
				RegistryURL: "http://localhost:8080",
				Timeout:     15 * time.Second,
			}, npm.ClientOptionHTTPTransport(tc.transport))
			require.NoError(t, err)

			pkg, err := client.FetchPackage(context.Background(), "fake-name", "fake-version")

			assert.Equal(t, tc.expectedPkg, pkg)
			if tc.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestClient_FetchPackageMeta(t *testing.T) {
	testCases := []struct {
		name            string
		transport       http.RoundTripper
		expectedErr     string
		expectedPkgMeta *npm.PackageMeta
	}{
		{
			name: "round trip error",
			transport: fakeTransport{
				err: errors.New("round trip error"),
			},
			expectedErr: "http request roundtrip for: Get \"http://localhost:8080/fake-name\": round trip error",
		},
		{
			name: "http error response decoding error",
			transport: fakeTransport{
				resp: &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(strings.NewReader(`{"error":"json error"}`)),
				},
			},
			expectedErr: "error response decoding of \"http://localhost:8080/fake-name\": json: cannot unmarshal object into Go value of type string",
		},
		{
			name: "http error response decoding success",
			transport: fakeTransport{
				resp: &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader(`"internal server error"`)),
				},
			},
			expectedErr: "http response for \"http://localhost:8080/fake-name\": internal server error",
		},
		{
			name: "http response decoding error",
			transport: fakeTransport{
				resp: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`"invalid payload"`)),
				},
			},
			expectedErr: "response decoding of \"http://localhost:8080/fake-name\": json: cannot unmarshal string into Go value of type npm.PackageMeta",
		},
		{
			name: "http response package not found",
			transport: fakeTransport{
				resp: &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader(`"not found"`)),
				},
			},
			expectedErr: npm.ErrPackageNotFound.Error(),
		},
		{
			name: "http response decoding",
			transport: fakeTransport{
				resp: &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
"name":"awesome",
"versions":{
"1.0.1":{"name":"awesome","version":"1.0.1","dependencies":{"great":"^1.0.1"}},
"1.0.2":{"name":"awesome","version":"1.0.2","dependencies":{"so-so":"^2.0.1"}}}}`)),
				},
			},
			expectedPkgMeta: &npm.PackageMeta{
				Name: "awesome",
				Versions: map[string]npm.Package{
					"1.0.1": {
						Name:    "awesome",
						Version: "1.0.1",
						Dependencies: map[string]string{
							"great": "^1.0.1",
						},
					},
					"1.0.2": {
						Name:    "awesome",
						Version: "1.0.2",
						Dependencies: map[string]string{
							"so-so": "^2.0.1",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := npm.NewClient(npm.ClientConfig{
				RegistryURL: "http://localhost:8080",
				Timeout:     15 * time.Second,
			}, npm.ClientOptionHTTPTransport(tc.transport))
			require.NoError(t, err)

			pkgMeta, err := client.FetchPackageMeta(context.Background(), "fake-name")

			assert.Equal(t, tc.expectedPkgMeta, pkgMeta)
			if tc.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
