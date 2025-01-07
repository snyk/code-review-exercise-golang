package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/snyk/npmjs-deps-fetcher/internal/api"
)

func main() {
	fmt.Println("Ciao!")
	handler := api.New()

	server := &http.Server{
		Addr:              "localhost:3000",
		Handler:           handler,
		ReadHeaderTimeout: time.Second * 10,
		WriteTimeout:      time.Second * 30,
	}
	fmt.Println("Server running on http://localhost:3000/")
	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
