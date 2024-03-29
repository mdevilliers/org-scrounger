package jaegar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client manages communication with the Jaegar HTTP API.
type Client struct {
	host *url.URL // Base host URL for API requests.

	// HTTP client used to communicate with the API. By default
	// http.DefaultClient will be used.
	httpClient *http.Client
}

// Option can be supplied that override the default Clients properties
type Option func(c *Client)

// WithHTTPClient allows a specific http.Client to be set
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient returns a new Jaegar API client or an error
func NewClient(host string, opts ...Option) (*Client, error) {
	hostURL, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("error parsing url %s: %w", host, err)
	}

	c := &Client{
		host:       hostURL,
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Host returns the API root URL the Client is configured to talk to.
func (c *Client) Host() string {
	return c.host.String()
}

// NewRequest creates an API request. A relative URL can be provided in path,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON-encoded and included as the
// request body.
func (c *Client) NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	url := c.host.ResolveReference(rel)

	var contentType string
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		errEnc := json.NewEncoder(buf).Encode(body)
		if errEnc != nil {
			return nil, errEnc
		}
		contentType = "application/json"
	}

	request, err := http.NewRequestWithContext(ctx, method, url.String(), buf)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "application/json")
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}

	return request, nil
}

// Do sends an API request and returns the API response. The API response is
// JSON-decoded and stored in the value pointed to by v, or returned as an
// error if an API or HTTP error has occurred.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 { //nolint: gomnd
		return response, fmt.Errorf("error calling Jaegar API : %d", response.StatusCode)
	}

	if v != nil {
		err = json.NewDecoder(response.Body).Decode(v)
		if err == io.EOF {
			err = nil // ignore EOF, empty response body
		}
	}

	return response, err
}

func (c *Client) get(ctx context.Context, path string, v interface{}) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil, v)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body, v interface{}) (*http.Response, error) {
	request, err := c.NewRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}

	return c.Do(request, v)
}

type Trace struct {
	Data []struct {
		Processes map[string]struct {
			ServiceName string `json:"serviceName"`
		} `json:"processes"`
	} `json:"data"`
}

func (c *Client) GetTraceByID(ctx context.Context, traceID string) (*Trace, error) {
	ret := &Trace{}
	response, err := c.get(ctx, fmt.Sprintf("/api/traces/%s", traceID), ret)
	if err != nil {
		return nil, fmt.Errorf("error retrieving trace '%s': %w", traceID, err)
	}
	defer response.Body.Close()
	return ret, nil
}
