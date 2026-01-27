package handler_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
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
		setup              func(testing.TB) (*http.Request, handler.PackageResolver)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "invalid version constraint",
			setup: func(tb testing.TB) (*http.Request, handler.PackageResolver) {
				tb.Helper()

				req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/package/foo/latest", http.NoBody)
				req.SetPathValue("packageName", "foo")
				req.SetPathValue("packageVersion", "latest")

				return req, mockshandler.NewMockPackageResolver(gomock.NewController(t))
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "{\"error\":\"invalid version constraint\"}\n",
		},
		{
			name: "missing version",
			setup: func(tb testing.TB) (*http.Request, handler.PackageResolver) {
				tb.Helper()

				req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/package/", http.NoBody)
				req.SetPathValue("packageName", "package")

				return req, mockshandler.NewMockPackageResolver(gomock.NewController(t))
			},
			expectedStatusCode: http.StatusBadRequest,
			// TODO : incorrect error message reference to invalid version constraint but should be "missing version"
			expectedBody: "{\"error\":\"missing version\"}\n",
		},
		{
			name: "missing package name",
			setup: func(tb testing.TB) (*http.Request, handler.PackageResolver) {
				tb.Helper()

				req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/package//latest", http.NoBody)
				req.SetPathValue("packageVersion", "latest")

				return req, mockshandler.NewMockPackageResolver(gomock.NewController(t))
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "{\"error\":\"missing package name\"}\n",
		},
		{
			name: "malformed url with trailing slash",
			setup: func(tb testing.TB) (*http.Request, handler.PackageResolver) {
				tb.Helper()

				req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/package/foo/1.0.1/", http.NoBody)
				req.SetPathValue("packageName", "foo")
				req.SetPathValue("packageVersion", "latest")

				return req, mockshandler.NewMockPackageResolver(gomock.NewController(t))
			},
			expectedStatusCode: http.StatusBadRequest,
			// suggestion: should return "invalid request path" error instead of "invalid version constraint"
			expectedBody: "{\"error\":\"invalid version constraint\"}\n",
		},
		{
			name: "package not found",
			setup: func(tb testing.TB) (*http.Request, handler.PackageResolver) {
				tb.Helper()

				req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/package/foo/1.0.1", http.NoBody)
				req.SetPathValue("packageName", "foo")
				req.SetPathValue("packageVersion", "1.0.1")

				resolver := mockshandler.NewMockPackageResolver(gomock.NewController(t))
				resolver.EXPECT().ResolvePackage(gomock.Any(), gomock.Any(), gomock.Any()).Return(npm.ErrPackageNotFound)

				return req, resolver
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "{\"error\":\"package not found\"}\n",
		},
		{
			name: "resolve deps failed",
			setup: func(tb testing.TB) (*http.Request, handler.PackageResolver) {
				tb.Helper()

				req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/package/foo/1.0.1", http.NoBody)
				req.SetPathValue("packageName", "foo")
				req.SetPathValue("packageVersion", "1.0.1")

				resolver := mockshandler.NewMockPackageResolver(gomock.NewController(t))
				resolver.EXPECT().ResolvePackage(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("something bad happened"))

				return req, resolver
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "{\"error\":\"internal server error\"}\n",
		},
		{
			name: "context canceled",
			setup: func(tb testing.TB) (*http.Request, handler.PackageResolver) {
				tb.Helper()

				req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/package/foo/1.0.1", http.NoBody)
				req.SetPathValue("packageName", "foo")
				req.SetPathValue("packageVersion", "1.0.1")

				ctx, _ := context.WithTimeout(req.Context(), 1*time.Millisecond)
				req = req.WithContext(ctx)

				resolver := mockshandler.NewMockPackageResolver(gomock.NewController(t))
				resolver.EXPECT().ResolvePackage(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, _ *semver.Constraints, npmPkg *npm.NpmPackageVersion) error {
						<-ctx.Done()
						return ctx.Err()
					})

				return req, resolver
			},
			// TODO currently returns 500, but could return something more specific for the user as this is a client-side issue
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "{\"error\":\"internal server error\"}\n",
		},
		{
			name: "resolve deps succeeded",
			setup: func(tb testing.TB) (*http.Request, handler.PackageResolver) {
				tb.Helper()

				req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/package/foo/1.0.1", http.NoBody)
				req.SetPathValue("packageName", "foo")
				req.SetPathValue("packageVersion", "1.0.1")

				resolver := mockshandler.NewMockPackageResolver(gomock.NewController(t))
				resolver.EXPECT().ResolvePackage(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ *semver.Constraints, npmPkg *npm.NpmPackageVersion) error {
						npmPkg.Version = "1.0.1"
						npmPkg.Status = npm.StatusResolved
						npmPkg.Dependencies = map[string]*npm.NpmPackageVersion{
							"bar": {Name: "bar", Version: "0.1.0", Status: npm.StatusResolved},
							"baz": {Name: "baz", Version: "2.0.1", Status: npm.StatusResolved},
						}
						return nil
					})

				return req, resolver
			},
			expectedStatusCode: http.StatusOK,
			expectedBody: "{\"name\":\"foo\",\"version\":\"1.0.1\",\"dependencies\"" +
				":{\"bar\":{\"name\":\"bar\",\"version\":\"0.1.0\",\"dependencies\":null,\"status\":\"resolved\"},\"baz\":{\"name\":\"baz\",\"version\":\"2.0.1\",\"dependencies\":null,\"status\":\"resolved\"}},\"status\":\"resolved\"}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, resolver := tc.setup(t)

			h := handler.PackageVersion(slog.DiscardHandler, resolver)
			w := httptest.NewRecorder()

			h.ServeHTTP(w, req)

			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, tc.expectedBody, string(body))
		})
	}
}
