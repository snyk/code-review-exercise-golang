//go:build integration

package integration_test

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// registryHandler is an HTTP handler that serves requests from files in testdata folder based on url path.
// For instance '/path1/path2' serves `testdata/registry_path1_path2.json`.
func registryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := "registry_" + strings.ReplaceAll(strings.Trim(r.URL.Path, "/"), "/", "_")
		content, err := os.ReadFile("testdata/" + filename + ".json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%q", err.Error())
			return
		}

		if _, err := w.Write(content); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%q", err.Error())
		}
	}
}

// application is a helper struct to manage the process of a Go application
// from a test suite.
type application struct {
	cmd         *exec.Cmd
	registryURL string
}

func (a *application) start() error {
	port, err := freePort()
	if err != nil {
		return err
	}

	appAddr = fmt.Sprintf("localhost:%d", port)

	//nolint:gosec // #nosec G402
	a.cmd = exec.Command("go", "run", "../../cmd/npmjs-deps-fetcher/...",
		"--npm.registryUrl", "https://registry.npmjs.org",
		"--server.addr", appAddr)
	a.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	out, err := a.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	if err = a.cmd.Start(); err != nil {
		return fmt.Errorf("cmd start: %w", err)
	}

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "HTTP server running") {
			break
		}
	}

	return nil
}

func (a *application) close() error {
	pgid, err := syscall.Getpgid(a.cmd.Process.Pid)
	if err != nil {
		return fmt.Errorf("syscall get pgid: %w", err)
	}
	if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("sigkill %d: %w", -pgid, err)
	}
	return nil
}

func freePort() (int, error) {
	ln, err := net.Listen("tcp4", "localhost:0")
	if err != nil {
		return 0, fmt.Errorf("tcp listen: %w", err)
	}
	defer ln.Close()

	if tcpAddr, ok := ln.Addr().(*net.TCPAddr); ok {
		return tcpAddr.Port, nil
	}

	return 0, fmt.Errorf("unexpected add type: %T", ln.Addr())
}
