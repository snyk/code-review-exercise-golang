package npm

import "errors"

// ErrPackageNotFound indicates the package/version is
// not found in the registry.
var ErrPackageNotFound = errors.New("package not found")
