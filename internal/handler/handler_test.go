package handler_test

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/snyk/npmjs-deps-fetcher/internal/handler"
	mockshandler "github.com/snyk/npmjs-deps-fetcher/internal/handler/mocks"
	"github.com/snyk/npmjs-deps-fetcher/internal/npm"
)

func TestPackageVersion(t *testing.T) {
	testCases := []struct {
		name               string
		setup              func(testing.TB) handler.PackageResolver
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "resolve deps failed",
			setup: func(tb testing.TB) handler.PackageResolver {
				tb.Helper()
				resolver := mockshandler.NewMockPackageResolver(gomock.NewController(t))
				resolver.EXPECT().ResolvePackage("foo", "1.0.1").Return(nil, errors.New("something bad happened"))
				return resolver
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "{\"error\":\"internal server error\"}\n",
		}, {
			name: "resolve deps succeeded",
			setup: func(tb testing.TB) handler.PackageResolver {
				tb.Helper()
				resolver := mockshandler.NewMockPackageResolver(gomock.NewController(t))
				resolver.EXPECT().ResolvePackage("foo", "1.0.1").Return(&npm.Package{
					Name:    "foo",
					Version: "1.0.1",
					Dependencies: map[string]string{
						"bar": "0.1.0",
						"baz": "2.0.1",
						"qux": "1.2.1",
					},
				}, nil)
				return resolver
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "{\"name\":\"foo\",\"version\":\"1.0.1\",\"dependencies\":{\"bar\":\"0.1.0\",\"baz\":\"2.0.1\",\"qux\":\"1.2.1\"}}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.PackageVersion(slog.DiscardHandler, tc.setup(t))
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/package/foo/1.0.1", http.NoBody)
			req.SetPathValue("package", "foo")
			req.SetPathValue("version", "1.0.1")

			h.ServeHTTP(w, req)

			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, tc.expectedBody, string(body))
		})
	}
}
