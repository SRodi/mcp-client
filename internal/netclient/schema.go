package netclient

import (
	"strconv"
	"strings"
	"time"
)

// ConnectionInfo represents connection event information (matches server output)
type ConnectionInfo struct {
	PID         uint32  `json:"pid"`
	Command     string  `json:"command"`
	Destination string  `json:"destination"`
	Protocol    string  `json:"protocol"`
	ReturnCode  int32   `json:"return_code"`
	Timestamp   float64 `json:"timestamp"`
}

// ConnectionSummaryOutput matches the server's connection summary output
type ConnectionSummaryOutput struct {
	Count           int    `json:"count"`
	PID             int    `json:"pid,omitempty"`
	Command         string `json:"command,omitempty"`
	DurationSeconds int    `json:"duration_seconds"`
	QueryTime       string `json:"query_time,omitempty"`
}

// ListConnectionsOutput matches the server's list connections output
type ListConnectionsOutput struct {
	TotalEvents int                         `json:"total_events"`
	TotalPIDs   int                         `json:"total_pids"`
	EventsByPID map[string][]ConnectionInfo `json:"events_by_pid"`
	QueryTime   string                      `json:"query_time,omitempty"`
}

// Legacy ConnectionEvent for backward compatibility with existing code
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

// PacketDropInfo represents packet drop event information
type PacketDropInfo struct {
	PID       uint32  `json:"pid"`
	Command   string  `json:"command"`
	Reason    string  `json:"reason"`
	Timestamp float64 `json:"timestamp"`
}

// PacketDropSummaryRequest represents the request body for packet drop summary
type PacketDropSummaryRequest struct {
	PID             int    `json:"pid,omitempty"`
	Command         string `json:"command,omitempty"`
	ProcessName     string `json:"process_name,omitempty"`
	DurationSeconds int    `json:"duration_seconds"`
}

// PacketDropSummaryOutput matches the server's packet drop summary output
type PacketDropSummaryOutput struct {
	Count           int    `json:"count"`
	PID             int    `json:"pid,omitempty"`
	Command         string `json:"command,omitempty"`
	DurationSeconds int    `json:"duration_seconds"`
	QueryTime       string `json:"query_time,omitempty"`
	Message         string `json:"message,omitempty"`
}

// PacketDropListOutput matches the server's packet drop list output
type PacketDropListOutput struct {
	TotalEvents int                         `json:"total_events"`
	TotalPIDs   int                         `json:"total_pids"`
	EventsByPID map[string][]PacketDropInfo `json:"events_by_pid"`
	QueryTime   string                      `json:"query_time,omitempty"`
	Message     string                      `json:"message,omitempty"`
}

// Convert ConnectionInfo to ConnectionEvent for backward compatibility
func (ci ConnectionInfo) ToConnectionEvent() ConnectionEvent {
	// Convert timestamp from Unix time (float64) to time.Time
	wallTime := time.Unix(int64(ci.Timestamp), 0)

	// Parse destination to extract IP and port
	var destIP string
	var destPort uint16

	if ci.Destination != "" {
		destIP, destPort = parseDestination(ci.Destination)
	}

	return ConnectionEvent{
		PID:             ci.PID,
		ReturnCode:      ci.ReturnCode,
		Command:         ci.Command,
		DestinationIP:   destIP,
		DestinationPort: destPort,
		Destination:     ci.Destination,
		Protocol:        ci.Protocol,
		WallTime:        wallTime,
		TimestampNS:     uint64(ci.Timestamp * 1e9), // Convert to nanoseconds
	}
}

// parseDestination parses a destination string like "127.0.0.1:8080" or "[::1]:8080"
// and returns the IP and port separately
func parseDestination(destination string) (string, uint16) {
	if destination == "" {
		return "", 0
	}

	// Handle IPv6 addresses with brackets like [::1]:8080
	if strings.HasPrefix(destination, "[") {
		closeBracket := strings.Index(destination, "]")
		if closeBracket != -1 {
			ip := destination[:closeBracket+1] // Include the closing bracket
			remaining := destination[closeBracket+1:]

			// Check if there's a port after the bracket
			if strings.HasPrefix(remaining, ":") {
				portStr := remaining[1:] // Remove the colon
				if port, err := strconv.ParseUint(portStr, 10, 16); err == nil {
					return ip, uint16(port)
				}
			}
			return ip, 0
		}
	}

	// Handle regular IPv4 addresses or hostnames like "127.0.0.1:8080" or "example.com:80"
	// Find the last colon to handle cases with multiple colons
	lastColon := strings.LastIndex(destination, ":")
	if lastColon != -1 {
		ip := destination[:lastColon]
		portStr := destination[lastColon+1:]

		if port, err := strconv.ParseUint(portStr, 10, 16); err == nil {
			return ip, uint16(port)
		}
		// If port parsing fails, return just the IP part
		return ip, 0
	}

	// No port separator found, assume it's just an IP or hostname
	return destination, 0
}
