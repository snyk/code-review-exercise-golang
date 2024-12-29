package packagegetter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type NpmPackageGetter struct{}

func NewNpmPackageGetter() NpmPackageGetter {
	return NpmPackageGetter{}
}

func (npg NpmPackageGetter) FetchPackage(name, version string) (*NpmPackageResponse, error) {
	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s/%s", name, version))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed NpmPackageResponse
	_ = json.Unmarshal(body, &parsed)
	return &parsed, nil
}

func (npg NpmPackageGetter) FetchPackageMeta(name string) (*NpmPackageMetaResponse, error) {
	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", name))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed NpmPackageMetaResponse
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return nil, err
	}

	return &parsed, nil
}
