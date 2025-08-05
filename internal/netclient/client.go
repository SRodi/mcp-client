package netclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client wraps HTTP client for communication with the REST API
type Client struct {
	httpClient *http.Client
	baseURL    string
	verbose    bool
}

// NewClient creates a new HTTP client
func NewClient(baseURL string) *Client {
	return NewClientWithVerbose(baseURL, false)
}

// NewClientWithVerbose creates a new HTTP client with verbose logging control
func NewClientWithVerbose(baseURL string, verbose bool) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
		verbose: verbose,
	}
}

// Connect validates connection to the HTTP API server
func (c *Client) Connect(ctx context.Context) error {
	if c.verbose {
		log.Printf("Attempting to connect to HTTP API server at %s", c.baseURL)
	}

	// Test connection with health check endpoint
	healthURL := c.baseURL + "/health"
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to HTTP API server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check failed with status %d: %s", resp.StatusCode, string(body))
	}

	if c.verbose {
		log.Printf("Successfully connected to HTTP API server at %s", c.baseURL)
	}

	return nil
}

// HealthCheck checks if the API server is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	healthURL := c.baseURL + "/health"
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// ConnectionSummaryRequest represents the request body for connection summary
type ConnectionSummaryRequest struct {
	PID             int    `json:"pid,omitempty"`
	Command         string `json:"command,omitempty"`
	ProcessName     string `json:"process_name,omitempty"`
	DurationSeconds int    `json:"duration_seconds"`
}

// GetConnectionSummary gets connection statistics using the HTTP REST API
func (c *Client) GetConnectionSummary(ctx context.Context, pid int, processName string, duration int) (ConnectionSummaryOutput, error) {
	// Prepare request body
	reqBody := ConnectionSummaryRequest{
		DurationSeconds: duration,
	}

	if pid > 0 {
		reqBody.PID = pid
	}
	if processName != "" {
		reqBody.Command = processName
		reqBody.ProcessName = processName // For backward compatibility
	}

	// Marshal request body
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return ConnectionSummaryOutput{}, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	summaryURL := c.baseURL + "/api/connection-summary"
	req, err := http.NewRequestWithContext(ctx, "POST", summaryURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return ConnectionSummaryOutput{}, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if c.verbose {
		log.Printf("Calling POST %s with body: %s", summaryURL, string(jsonData))
	}

	// Make HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ConnectionSummaryOutput{}, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ConnectionSummaryOutput{}, fmt.Errorf("failed to read response: %v", err)
	}

	if c.verbose {
		log.Printf("HTTP response status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &errorResp) == nil {
			return ConnectionSummaryOutput{}, fmt.Errorf("server error (%d): %s - %s", resp.StatusCode, errorResp.Error, errorResp.Message)
		}
		return ConnectionSummaryOutput{}, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var summary ConnectionSummaryOutput
	if err := json.Unmarshal(body, &summary); err != nil {
		return ConnectionSummaryOutput{}, fmt.Errorf("failed to parse response: %v", err)
	}

	return summary, nil
}

// ListConnectionsRequest represents the request body for list connections
type ListConnectionsRequest struct {
	PID   *int `json:"pid,omitempty"`
	Limit *int `json:"limit,omitempty"`
}

// ListConnections lists all tracked connections using the HTTP REST API
func (c *Client) ListConnections(ctx context.Context, pid *int, limit *int) (ListConnectionsOutput, error) {
	// Try GET endpoint first (simpler for basic cases)
	if pid == nil && limit == nil {
		return c.listConnectionsGET(ctx, nil, nil)
	}

	// Use POST endpoint for complex queries
	return c.listConnectionsPOST(ctx, pid, limit)
}

// listConnectionsGET uses the GET endpoint for listing connections
func (c *Client) listConnectionsGET(ctx context.Context, pid *int, limit *int) (ListConnectionsOutput, error) {
	// Build URL with query parameters
	listURL := c.baseURL + "/api/list-connections"
	params := url.Values{}

	if pid != nil {
		params.Add("pid", strconv.Itoa(*pid))
	}
	if limit != nil {
		params.Add("limit", strconv.Itoa(*limit))
	}

	if len(params) > 0 {
		listURL += "?" + params.Encode()
	}

	if c.verbose {
		log.Printf("Calling GET %s", listURL)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", listURL, nil)
	if err != nil {
		return ListConnectionsOutput{}, fmt.Errorf("failed to create request: %v", err)
	}

	// Make HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ListConnectionsOutput{}, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	return c.parseListConnectionsResponse(resp)
}

// listConnectionsPOST uses the POST endpoint for listing connections
func (c *Client) listConnectionsPOST(ctx context.Context, pid *int, limit *int) (ListConnectionsOutput, error) {
	// Prepare request body
	reqBody := ListConnectionsRequest{
		PID:   pid,
		Limit: limit,
	}

	// Marshal request body
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return ListConnectionsOutput{}, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	listURL := c.baseURL + "/api/list-connections"
	req, err := http.NewRequestWithContext(ctx, "POST", listURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return ListConnectionsOutput{}, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if c.verbose {
		log.Printf("Calling POST %s with body: %s", listURL, string(jsonData))
	}

	// Make HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ListConnectionsOutput{}, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	return c.parseListConnectionsResponse(resp)
}

// parseListConnectionsResponse parses the HTTP response for list connections
func (c *Client) parseListConnectionsResponse(resp *http.Response) (ListConnectionsOutput, error) {
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ListConnectionsOutput{}, fmt.Errorf("failed to read response: %v", err)
	}

	if c.verbose {
		log.Printf("HTTP response status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &errorResp) == nil {
			return ListConnectionsOutput{}, fmt.Errorf("server error (%d): %s - %s", resp.StatusCode, errorResp.Error, errorResp.Message)
		}
		return ListConnectionsOutput{}, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var listOutput ListConnectionsOutput
	if err := json.Unmarshal(body, &listOutput); err != nil {
		return ListConnectionsOutput{}, fmt.Errorf("failed to parse response: %v", err)
	}

	return listOutput, nil
}

// GetPacketDropSummary gets packet drop statistics using the HTTP REST API
func (c *Client) GetPacketDropSummary(ctx context.Context, pid int, processName string, duration int) (PacketDropSummaryOutput, error) {
	// Prepare request body
	reqBody := PacketDropSummaryRequest{
		DurationSeconds: duration,
	}

	if pid > 0 {
		reqBody.PID = pid
	}
	if processName != "" {
		reqBody.Command = processName
		reqBody.ProcessName = processName // For backward compatibility
	}

	// Marshal request body
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return PacketDropSummaryOutput{}, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	summaryURL := c.baseURL + "/api/packet-drop-summary"
	req, err := http.NewRequestWithContext(ctx, "POST", summaryURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return PacketDropSummaryOutput{}, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if c.verbose {
		log.Printf("Calling POST %s with body: %s", summaryURL, string(jsonData))
	}

	// Make HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return PacketDropSummaryOutput{}, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PacketDropSummaryOutput{}, fmt.Errorf("failed to read response: %v", err)
	}

	if c.verbose {
		log.Printf("HTTP response status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &errorResp) == nil {
			return PacketDropSummaryOutput{}, fmt.Errorf("server error (%d): %s - %s", resp.StatusCode, errorResp.Error, errorResp.Message)
		}
		return PacketDropSummaryOutput{}, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var summary PacketDropSummaryOutput
	if err := json.Unmarshal(body, &summary); err != nil {
		return PacketDropSummaryOutput{}, fmt.Errorf("failed to parse response: %v", err)
	}

	return summary, nil
}

// ListPacketDrops lists packet drop events using the HTTP REST API
func (c *Client) ListPacketDrops(ctx context.Context) (PacketDropListOutput, error) {
	// Create HTTP request
	listURL := c.baseURL + "/api/list-packet-drops"
	req, err := http.NewRequestWithContext(ctx, "GET", listURL, nil)
	if err != nil {
		return PacketDropListOutput{}, fmt.Errorf("failed to create request: %v", err)
	}

	if c.verbose {
		log.Printf("Calling GET %s", listURL)
	}

	// Make HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return PacketDropListOutput{}, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PacketDropListOutput{}, fmt.Errorf("failed to read response: %v", err)
	}

	if c.verbose {
		log.Printf("HTTP response status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &errorResp) == nil {
			return PacketDropListOutput{}, fmt.Errorf("server error (%d): %s - %s", resp.StatusCode, errorResp.Error, errorResp.Message)
		}
		return PacketDropListOutput{}, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var listOutput PacketDropListOutput
	if err := json.Unmarshal(body, &listOutput); err != nil {
		return PacketDropListOutput{}, fmt.Errorf("failed to parse response: %v", err)
	}

	return listOutput, nil
}

// Close is a no-op for HTTP client (no persistent connection to close)
func (c *Client) Close() error {
	// HTTP client doesn't need explicit closing
	return nil
}

// GetConnectionSummaryHTTP provides backward compatibility
func (c *Client) GetConnectionSummaryHTTP(pid int, processName string, duration int) (ConnectionSummaryOutput, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return c.GetConnectionSummary(ctx, pid, processName, duration)
}

// ListConnectionsHTTP provides backward compatibility
func (c *Client) ListConnectionsHTTP(pid *int, limit *int) (ListConnectionsOutput, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return c.ListConnections(ctx, pid, limit)
}
