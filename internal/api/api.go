package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	getter "github.com/snyk/npmjs-deps-fetcher/internal/domain/package_getter"
	packagemanager "github.com/snyk/npmjs-deps-fetcher/internal/domain/package_manager"
)

const (
	responseOK    = 200
	responseError = 500
)

func New() http.Handler {
	router := mux.NewRouter()
	router.Handle("/", http.HandlerFunc(basicHandler))
	router.Handle("/package/{package}/{version}", http.HandlerFunc(packageHandler))
	return router
}

func basicHandler(w http.ResponseWriter, _ *http.Request) {
	response := []byte("Hello World")
	w.WriteHeader(responseOK)
	w.Write(response)
}

func packageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pkgName := vars["package"]
	pkgVersion := vars["version"]

	packageManager := packagemanager.NewPackageManagerService(getter.NewNpmPackageGetter())

	pkgResult, err := packageManager.GetPackageDependencies(pkgName, pkgVersion)
	if err != nil {
		println(err.Error())
		w.WriteHeader(responseError)
		return
	}

	stringified, err := json.MarshalIndent(pkgResult, "", "  ")
	if err != nil {
		println(err.Error())
		w.WriteHeader(responseError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseOK)

	// Ignoring ResponseWriter errors
	_, _ = w.Write(stringified)
}
