package netclient

import (
	"strconv"
	"strings"
	"time"
)

// ConnectionInfo represents connection event information (matches server output)
type ConnectionInfo struct {
	PID         uint32 `json:"pid"`
	Command     string `json:"command"`
	Destination string `json:"destination"`
	Protocol    string `json:"protocol"`
	ReturnCode  int32  `json:"return_code"`
	Timestamp   string `json:"timestamp"`
}

// ConnectionSummaryOutput matches the server's connection summary output
type ConnectionSummaryOutput struct {
	Total   int    `json:"total_attempts"`
	PID     int    `json:"pid,omitempty"`
	Command string `json:"command,omitempty"`
	Seconds int    `json:"duration"`
	Message string `json:"message,omitempty"`
}

// ListConnectionsOutput matches the server's list connections output
type ListConnectionsOutput struct {
	TotalPIDs   int                         `json:"total_pids"`
	Connections map[string][]ConnectionInfo `json:"connections"`
	Truncated   bool                        `json:"truncated"`
	Message     string                      `json:"message,omitempty"`
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

// Convert ConnectionInfo to ConnectionEvent for backward compatibility
func (ci ConnectionInfo) ToConnectionEvent() ConnectionEvent {
	wallTime, _ := time.Parse("2006-01-02T15:04:05Z", ci.Timestamp)

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
