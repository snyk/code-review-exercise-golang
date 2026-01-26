package npm

type (

	// suggestion: NpmPackageVersion sounds like a raw data type from the npm registry
	// consider renaming NpmPackageVersion to make it known this type is constructed
	// by the resolver rather than returned by the npm client e.g. ResolvedPackageVersion
	NpmPackageVersion struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		// question: dependency is stored in a map that has random access order
		// is dependency ordering important for the users of this API?
		Dependencies map[string]*NpmPackageVersion `json:"dependencies"`
		Status       ResolutionStatus              `json:"status"`
		Error        string                        `json:"error,omitempty"`
	}

	// Package contains the info of an NPM package version.
	Package struct {
		// Name is the name of the NPM package.
		Name string `json:"name,omitempty"`
		// Version is the version of the NPM package.
		Version string `json:"version,omitempty"`
		// Dependencies contains the direct dependencies of an NPM package,
		// mapping the package name to its version constraint.
		Dependencies map[string]string `json:"dependencies,omitempty"`
	}

	// PackageMeta contains the metadata of an NPM package.
	PackageMeta struct {
		// Name is the name of the NPM package.
		Name string `json:"name,omitempty"`
		// Versions contains all the versions of the given NPM package.
		Versions map[string]Package `json:"versions,omitempty"`
	}
)
