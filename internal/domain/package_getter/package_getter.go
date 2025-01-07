package packagegetter

type PackageGetter interface {
	FetchPackage(name, version string) (*NpmPackageResponse, error)
	FetchPackageMeta(name string) (*NpmPackageMetaResponse, error)
}

type NpmPackageMetaResponse struct {
	Versions map[string]NpmPackageResponse `json:"versions"`
}

type NpmPackageResponse struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}

type NpmPackageVersion struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}
