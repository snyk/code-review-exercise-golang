package npm

type (
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
