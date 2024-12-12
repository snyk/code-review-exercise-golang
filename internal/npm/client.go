package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var _ PackageFetcher = (*Client)(nil)

type (
	// Client represents the NPM HTTP client.
	Client struct {
		client      *http.Client
		registryURL string
	}

	// ClientConfig provides the configuration of the NPM HTTP client.
	ClientConfig struct {
		// RegistryURL is the HTTP URL of the NPM registry.
		RegistryURL string `json:"registryUrl"`
		// Timeout configures the timeout of the HTTP client.
		Timeout time.Duration `json:"timeout"`
	}

	// ClientOption represent optional configuration for the NPM client.
	ClientOption func(*Client)
)

// ClientOptionHTTPTransport is a client option to customize the [http.RoundTripper] of
// the NPM HTTP client. The client uses the [http.DefaultTransport] by default.
func ClientOptionHTTPTransport(rt http.RoundTripper) ClientOption {
	return func(c *Client) {
		c.client.Transport = rt
	}
}

// NewClient creates an HTTP client to communicate with the NPM registry provided in the configuration.
func NewClient(cfg ClientConfig, opts ...ClientOption) (c *Client, err error) {
	if _, err := url.Parse(cfg.RegistryURL); err != nil {
		return nil, fmt.Errorf("registry URL configuration: %w", err)
	}

	c = &Client{
		client: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: http.DefaultTransport,
		},
		registryURL: cfg.RegistryURL,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// FetchPackage fetches the information of the NPM package identified by the provided name and version.
func (c *Client) FetchPackage(ctx context.Context, name, version string) (*Package, error) {
	u := c.registryURL + "/" + name + "/" + version
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("http request creation for %q: %w", u, err)
	}

	var pkg Package
	if err := c.fetch(req, &pkg); err != nil {
		return nil, err
	}

	return &pkg, nil
}

// FetchPackage fetches the metadata of the NPM package identified by the provided name.
func (c *Client) FetchPackageMeta(ctx context.Context, name string) (*PackageMeta, error) {
	u := c.registryURL + "/" + name
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("http request creation for %q: %w", u, err)
	}

	var pkgMeta PackageMeta
	if err := c.fetch(req, &pkgMeta); err != nil {
		return nil, err
	}

	return &pkgMeta, nil
}

func (c *Client) fetch(req *http.Request, obj any) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("http request roundtrip for: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return ErrPackageNotFound
	default:
		var body string
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return fmt.Errorf("error response decoding of %q: %w", req.URL.String(), err)
		}
		return fmt.Errorf("http response for %q: %s", req.URL.String(), body)
	}

	if err := json.NewDecoder(resp.Body).Decode(obj); err != nil {
		return fmt.Errorf("response decoding of %q: %w", req.URL.String(), err)
	}

	return nil
}
