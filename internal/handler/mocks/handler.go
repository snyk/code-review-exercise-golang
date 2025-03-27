// Code generated by MockGen. DO NOT EDIT.
// Source: handler.go
//
// Generated by this command:
//
//	mockgen -destination=mocks/handler.go -source=handler.go -package mockshandler
//

// Package mockshandler is a generated GoMock package.
package mockshandler

import (
	context "context"
	reflect "reflect"

	semver "github.com/Masterminds/semver/v3"
	npm "github.com/snyk/npmjs-deps-fetcher/internal/npm"
	gomock "go.uber.org/mock/gomock"
)

// MockPackageResolver is a mock of PackageResolver interface.
type MockPackageResolver struct {
	ctrl     *gomock.Controller
	recorder *MockPackageResolverMockRecorder
	isgomock struct{}
}

// MockPackageResolverMockRecorder is the mock recorder for MockPackageResolver.
type MockPackageResolverMockRecorder struct {
	mock *MockPackageResolver
}

// NewMockPackageResolver creates a new mock instance.
func NewMockPackageResolver(ctrl *gomock.Controller) *MockPackageResolver {
	mock := &MockPackageResolver{ctrl: ctrl}
	mock.recorder = &MockPackageResolverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPackageResolver) EXPECT() *MockPackageResolverMockRecorder {
	return m.recorder
}

// ResolvePackage mocks base method.
func (m *MockPackageResolver) ResolvePackage(ctx context.Context, constraint *semver.Constraints, npmPkg *npm.NpmPackageVersion) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ResolvePackage", ctx, constraint, npmPkg)
	ret0, _ := ret[0].(error)
	return ret0
}

// ResolvePackage indicates an expected call of ResolvePackage.
func (mr *MockPackageResolverMockRecorder) ResolvePackage(ctx, constraint, npmPkg any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResolvePackage", reflect.TypeOf((*MockPackageResolver)(nil).ResolvePackage), ctx, constraint, npmPkg)
}
