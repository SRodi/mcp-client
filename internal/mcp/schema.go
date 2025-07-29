package mcp

import "time"

// ConnectionEvent represents a single network connection event
type ConnectionEvent struct {
	PID             uint32    `json:"pid"`
	TimestampNS     uint64    `json:"timestamp_ns"`
	ReturnCode      int32     `json:"return_code"`
	Command         string    `json:"command"`
	DestinationIP   string    `json:"destination_ip"`
	DestinationPort uint16    `json:"destination_port"`
	Destination     string    `json:"destination"`
	AddressFamily   uint16    `json:"address_family"`
	Protocol        string    `json:"protocol"`
	SocketType      string    `json:"socket_type"`
	WallTime        time.Time `json:"wall_time"`
}

// ListConnectionsRequest represents the request to list all connections
type ListConnectionsRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

// ListConnectionsResponse represents the response containing all tracked connections
type ListConnectionsResponse struct {
	Result map[string][]ConnectionEvent `json:"result"`
}

// GetConnectionSummaryParams represents the parameters for connection summary requests
type GetConnectionSummaryParams struct {
	PID      int    `json:"pid,omitempty"`
	Command  string `json:"command,omitempty"`
	Duration int    `json:"duration"`
}
