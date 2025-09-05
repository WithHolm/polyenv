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

// Post sends a POST request with a JSON body and unmarshals the response into a target struct.
// 'body' is the data to be marshalled into the request body.
// 'target' is the struct to unmarshal the JSON response into. If nil, the response body is discarded.
func (c *PolyenvHttpClient) Post(ctx context.Context, url string, body interface{}, target interface{}) error {
	// Marshal the body to JSON
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	var tReflect reflect.Type
	//validate target
	if target != nil {
		tReflect = reflect.TypeOf(target)
		if tReflect.Kind() != reflect.Ptr {
			return fmt.Errorf("target must be a pointer")
		}

		// Get the type of the value stored in the interface
		tReflect = tReflect.Elem()

	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "polyenv-cli-ty-for-making-an-awesome-product") // Good practice to set a User-Agent

	slog.DebugContext(ctx, "request", "host", req.URL.Host)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

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
