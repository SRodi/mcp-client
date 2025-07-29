package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ConnectionSummaryRequest struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

type ConnectionSummaryResponse struct {
	Result struct {
		Total int `json:"total_attempts"`
	} `json:"result"`
}

// GetConnectionSummary queries the MCP server for connection statistics
// Either pid (non-zero) or processName (non-empty) should be provided, not both
func GetConnectionSummary(pid int, processName string, duration int, serverURL string) (ConnectionSummaryResponse, error) {
	params := map[string]interface{}{
		"duration": duration,
	}

	// Add either PID or command name to params
	if pid > 0 {
		params["pid"] = pid
	} else if processName != "" {
		params["command"] = processName
	} else {
		return ConnectionSummaryResponse{}, fmt.Errorf("either pid or processName must be provided")
	}

	payload := ConnectionSummaryRequest{
		Method: "get_connection_summary",
		Params: params,
	}

	var response ConnectionSummaryResponse

	data, err := json.Marshal(payload)
	if err != nil {
		return response, err
	}

	resp, err := http.Post(serverURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return response, fmt.Errorf("failed to connect to MCP server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("MCP server returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("failed to decode response: %v", err)
	}

	return response, nil
}

// ListConnections retrieves all tracked connections from the MCP server
func ListConnections(serverURL string) (ListConnectionsResponse, error) {
	payload := ListConnectionsRequest{
		Method: "list_connections",
		Params: map[string]interface{}{},
	}

	var response ListConnectionsResponse

	data, err := json.Marshal(payload)
	if err != nil {
		return response, err
	}

	resp, err := http.Post(serverURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return response, fmt.Errorf("failed to connect to MCP server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("MCP server returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("failed to decode response: %v", err)
	}

	return response, nil
}
