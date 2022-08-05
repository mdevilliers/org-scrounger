package sonarcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pkg/errors"
)

// Client manages communication with the Sonarcloud HTTP API.
type Client struct {
	host *url.URL // Base host URL for API requests.

	// HTTP client used to communicate with the API. By default
	// http.DefaultClient will be used.
	httpClient *http.Client

	token string
}

// Option can be supplied that override the default Clients properties
type Option func(c *Client)

// WithHTTPClient allows a specific http.Client to be set
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithToken sepcifies the sonarcloud token to use
func WithToken(token string) Option {
	return func(c *Client) {
		c.token = token
	}
}

// NewClientFromEnv returns a Sonercloud API client using
// the env var 'SONARCLOUD_TOKEN' or an error
func NewClientFromEnv(host string, opts ...Option) (bool, *Client, error) {
	token := os.Getenv("SONARCLOUD_TOKEN")
	if token == "" {
		return false, nil, errors.New("sonarcloud token not defined via 'SONARCLOUD_TOKEN'")
	}
	opts = append(opts, WithToken(token))

	client, err := NewClient(host, opts...)
	return true, client, err
}

// NewClient returns a new Sonarcloud API client or an error
func NewClient(host string, opts ...Option) (*Client, error) {
	hostURL, err := url.Parse(host)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing url %s", host)
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
	request.SetBasicAuth(c.token, "")
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

	if response.StatusCode >= 400 { //nolint:gomnd
		return response, fmt.Errorf("error calling API : %d", response.StatusCode)
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

type SonarCloudTime struct {
	time.Time
}

func (t *SonarCloudTime) UnmarshalJSON(b []byte) error {
	date, err := time.Parse(`"2006-01-02T15:04:05-0700"`, string(b))
	if err != nil {
		return err
	}
	t.Time = date
	return nil
}

type History struct {
	Time  SonarCloudTime `json:"date"`
	Value float64        `json:"value,string"`
}

type Measure struct {
	Metric  string    `json:"metric"`
	History []History `json:"history"`
}

type MeasureResponse struct {
	Measures []Measure `json:"measures"`
}

func (c *Client) GetMeasures(ctx context.Context, componentID string) (*MeasureResponse, error) {
	ret := &MeasureResponse{}
	response, err := c.get(
		ctx,
		fmt.Sprintf("/api/measures/search_history?component=%s&metrics=coverage", componentID),
		ret)
	if err != nil {
		return nil, errors.Wrapf(err, "error retrieving measures '%s'", componentID)
	}
	defer response.Body.Close()
	return ret, nil
}
