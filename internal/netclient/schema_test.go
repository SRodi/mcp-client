package netclient

import (
	"testing"
	"time"
)

func TestConnectionInfo_ToConnectionEvent(t *testing.T) {
	tests := []struct {
		name     string
		input    ConnectionInfo
		expected ConnectionEvent
	}{
		{
			name: "basic connection with IP:port destination",
			input: ConnectionInfo{
				PID:             1234,
				Command:         "nginx",
				Destination:     "192.168.1.100:80",
				DestinationIP:   "192.168.1.100",
				DestinationPort: 80,
				Protocol:        "TCP",
				ReturnCode:      0,
				Time:            "2024-01-01T12:00:00Z",
			},
			expected: ConnectionEvent{
				PID:             1234,
				Command:         "nginx",
				DestinationIP:   "192.168.1.100",
				DestinationPort: 80,
				Protocol:        "TCP",
				ReturnCode:      0,
				Destination:     "192.168.1.100:80",
			},
		},
		{
			name: "connection with Unix socket (empty destination)",
			input: ConnectionInfo{
				PID:         5678,
				Command:     "curl",
				Destination: "",
				Protocol:    "Unknown(0)",
				ReturnCode:  0,
				Time:        "2024-01-01T12:01:00Z",
			},
			expected: ConnectionEvent{
				PID:             5678,
				Command:         "curl",
				DestinationIP:   "",
				DestinationPort: 0,
				Protocol:        "Unknown(0)",
				ReturnCode:      0,
				Destination:     "",
			},
		},
		{
			name: "connection with API-provided destination info",
			input: ConnectionInfo{
				PID:             9876,
				Command:         "ssh",
				Destination:     "example.com:22",
				DestinationIP:   "example.com",
				DestinationPort: 22,
				Protocol:        "TCP",
				ReturnCode:      0,
				Time:            "2024-01-01T12:02:00Z",
			},
			expected: ConnectionEvent{
				PID:             9876,
				Command:         "ssh",
				DestinationIP:   "example.com",
				DestinationPort: 22,
				Protocol:        "TCP",
				ReturnCode:      0,
				Destination:     "example.com:22",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToConnectionEvent()

			if result.PID != tt.expected.PID {
				t.Errorf("PID = %v, want %v", result.PID, tt.expected.PID)
			}
			if result.Command != tt.expected.Command {
				t.Errorf("Command = %v, want %v", result.Command, tt.expected.Command)
			}
			if result.DestinationIP != tt.expected.DestinationIP {
				t.Errorf("DestinationIP = %v, want %v", result.DestinationIP, tt.expected.DestinationIP)
			}
			if result.DestinationPort != tt.expected.DestinationPort {
				t.Errorf("DestinationPort = %v, want %v", result.DestinationPort, tt.expected.DestinationPort)
			}
			if result.Protocol != tt.expected.Protocol {
				t.Errorf("Protocol = %v, want %v", result.Protocol, tt.expected.Protocol)
			}
			if result.ReturnCode != tt.expected.ReturnCode {
				t.Errorf("ReturnCode = %v, want %v", result.ReturnCode, tt.expected.ReturnCode)
			}
			if result.Destination != tt.expected.Destination {
				t.Errorf("Destination = %v, want %v", result.Destination, tt.expected.Destination)
			}

			// Test that timestamp parsing worked (should be a valid time)
			if result.WallTime.IsZero() {
				t.Error("WallTime should not be zero")
			}
			
			// Test that the time matches what we expect (2024-01-01 12:xx:00 UTC)
			expectedTime, err := time.Parse(time.RFC3339, tt.input.Time)
			if err != nil {
				t.Fatalf("Failed to parse expected time %s: %v", tt.input.Time, err)
			}
			if !result.WallTime.Equal(expectedTime) {
				t.Errorf("WallTime = %v, want %v", result.WallTime, expectedTime)
			}
		})
	}
}

func TestParseDestination(t *testing.T) {
	tests := []struct {
		name             string
		destination      string
		expectedIP       string
		expectedPort     uint16
	}{
		{"IPv4 with port", "192.168.1.1:80", "192.168.1.1", 80},
		{"hostname with port", "example.com:443", "example.com", 443},
		{"IPv6 with port", "[2001:db8::1]:22", "[2001:db8::1]", 22},
		{"IP without port", "192.168.1.1", "192.168.1.1", 0},
		{"empty destination", "", "", 0},
		{"invalid port", "example.com:invalid", "example.com", 0},
		{"high port number", "example.org:65535", "example.org", 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test via ConnectionInfo.ToConnectionEvent() which calls parseDestination
			conn := ConnectionInfo{
				PID:         1234,
				Command:     "test",
				Destination: tt.destination,
				Protocol:    "tcp",
				ReturnCode:  0,
				Time:        "2024-01-01T12:00:00Z",
			}

			result := conn.ToConnectionEvent()

			if result.DestinationIP != tt.expectedIP {
				t.Errorf("DestinationIP = %v, want %v", result.DestinationIP, tt.expectedIP)
			}
			if result.DestinationPort != tt.expectedPort {
				t.Errorf("DestinationPort = %v, want %v", result.DestinationPort, tt.expectedPort)
			}
		})
	}
}
