package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/snyk/npmjs-deps-fetcher/internal/api"
)

func main() {
	handler := api.New()

	srv := &http.Server{
		Addr:              "localhost:8080",
		Handler:           handler,
		ReadHeaderTimeout: time.Second * 10,
		WriteTimeout:      time.Second * 30,
	}

	fmt.Println("HTTP server running on localhost:8080")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println(err)
		os.Exit(1)
	}
}
