package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"time"
)

// DefaultTimeout is the default timeout for the HTTP client.
const DefaultTimeout = 30 * time.Second

// PolyenvHttpClient is a wrapper around http.PolyenvHttpClient to provide convenience methods.
type PolyenvHttpClient struct {
	httpClient *http.Client
}

// NewPolyenvHttpClient creates a new HTTP client with default settings.
func NewPolyenvHttpClient() *PolyenvHttpClient {
	return &PolyenvHttpClient{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

func (c *PolyenvHttpClient) ValidateTarget(target interface{}) (reflect.Type, error) {
	if target == nil {
		return nil, nil
	}
	tReflect := reflect.TypeOf(target)
	if tReflect.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("target must be a pointer")
	}
	// Get the type of the value stored in the interface
	tReflect = tReflect.Elem()
	return tReflect, nil
}

// create httprequest client with some default values
func (c *PolyenvHttpClient) NewRequest(ctx context.Context, method string, url string, body interface{}) (*http.Request, error) {
	var bod io.Reader = nil

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bod = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, bod)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "polyenv-cli-ty-for-making-an-awesome-product")

	return req, nil
}

// Get sends a GET request and unmarshals the response into a target struct.
func (c *PolyenvHttpClient) Get(ctx context.Context, url string, target interface{}) error {
	req, err := c.NewRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	tReflect, err := c.ValidateTarget(target)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "request", "host", req.URL.Host)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		e := resp.Body.Close()
		if e != nil {
			slog.Error("failed to close response body", "error", e)
		}
	}()

	// Read the body to be able to reuse it and to ensure it's closed
	slog.DebugContext(ctx, "response", "status", resp.StatusCode)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-successful status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received non-2xx status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Unmarshal the response body if a target is provided
	// slog.Debug("response", "body", string(bodyBytes))
	if target != nil {
		// allow 204/empty body
		if len(bodyBytes) == 0 {
			return nil
		}
		slog.Debug("unmarshalling response body to", "target", tReflect.Name())
		if err := json.Unmarshal(bodyBytes, target); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}
	}

	return nil
}

// Post sends a POST request with a JSON body and unmarshals the response into a target struct.
// 'body' is the data to be marshalled into the request body.
// 'target' is the struct to unmarshal the JSON response into. If nil, the response body is discarded.
func (c *PolyenvHttpClient) Post(ctx context.Context, url string, body interface{}, target interface{}) error {
	req, err := c.NewRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return err
	}

	tReflect, err := c.ValidateTarget(target)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "request", "host", req.URL.Host)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		e := resp.Body.Close()
		if e != nil {
			slog.Error("failed to close response body", "error", e)
		}
	}()

	// Read the body to be able to reuse it and to ensure it's closed
	slog.DebugContext(ctx, "response", "status", resp.StatusCode)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-successful status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received non-2xx status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Unmarshal the response body if a target is provided
	// slog.Debug("response", "body", string(bodyBytes))
	if target != nil {

		// allow 204/empty body
		if len(bodyBytes) == 0 {
			return nil
		}
		slog.Debug("unmarshalling response body to", "target", tReflect.Name())
		if err := json.Unmarshal(bodyBytes, target); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}
	}

	return nil
}
