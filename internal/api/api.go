package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	packagemanager "github.com/snyk/npmjs-deps-fetcher/internal/domain/package_manager"
)

const (
	responseOK    = 200
	responseError = 500
)

func New(mngr packagemanager.PackageManagerService) http.Handler {
	router := mux.NewRouter()
	router.Handle("/", http.HandlerFunc(basicHandler))
	router.Handle("/package/{package}/{version}", packageHandler(mngr))
	return router
}

func basicHandler(w http.ResponseWriter, _ *http.Request) {
	response := []byte("Hello World")
	w.WriteHeader(responseOK)
	w.Write(response)
}

func packageHandler(mngr packagemanager.PackageManagerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pkgName := vars["package"]
		pkgVersion := vars["version"]

		pkgResult, err := mngr.GetPackageDependencies(pkgName, pkgVersion)
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
}
