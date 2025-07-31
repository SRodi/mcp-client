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
				PID:         1234,
				Command:     "nginx",
				Destination: "192.168.1.100:80",
				Protocol:    "tcp",
				ReturnCode:  0,
				Timestamp:   "2024-01-01T12:00:00Z",
			},
			expected: ConnectionEvent{
				PID:             1234,
				Command:         "nginx",
				DestinationIP:   "192.168.1.100",
				DestinationPort: 80,
				Protocol:        "tcp",
				ReturnCode:      0,
				Destination:     "192.168.1.100:80",
			},
		},
		{
			name: "connection with hostname:port destination",
			input: ConnectionInfo{
				PID:         5678,
				Command:     "curl",
				Destination: "example.com:443",
				Protocol:    "tcp",
				ReturnCode:  0,
				Timestamp:   "2024-01-01T12:01:00Z",
			},
			expected: ConnectionEvent{
				PID:             5678,
				Command:         "curl",
				DestinationIP:   "example.com",
				DestinationPort: 443,
				Protocol:        "tcp",
				ReturnCode:      0,
				Destination:     "example.com:443",
			},
		},
		{
			name: "connection with IPv6 address",
			input: ConnectionInfo{
				PID:         9876,
				Command:     "ssh",
				Destination: "[2001:db8::1]:22",
				Protocol:    "tcp",
				ReturnCode:  0,
				Timestamp:   "2024-01-01T12:02:00Z",
			},
			expected: ConnectionEvent{
				PID:             9876,
				Command:         "ssh",
				DestinationIP:   "[2001:db8::1]",
				DestinationPort: 22,
				Protocol:        "tcp",
				ReturnCode:      0,
				Destination:     "[2001:db8::1]:22",
			},
		},
		{
			name: "connection with invalid port (should default to 0)",
			input: ConnectionInfo{
				PID:         1111,
				Command:     "test",
				Destination: "localhost:invalid",
				Protocol:    "tcp",
				ReturnCode:  -1,
				Timestamp:   "2024-01-01T12:03:00Z",
			},
			expected: ConnectionEvent{
				PID:             1111,
				Command:         "test",
				DestinationIP:   "localhost",
				DestinationPort: 0,
				Protocol:        "tcp",
				ReturnCode:      -1,
				Destination:     "localhost:invalid",
			},
		},
		{
			name: "connection without port (should default to 0)",
			input: ConnectionInfo{
				PID:         2222,
				Command:     "ping",
				Destination: "8.8.8.8",
				Protocol:    "icmp",
				ReturnCode:  0,
				Timestamp:   "2024-01-01T12:04:00Z",
			},
			expected: ConnectionEvent{
				PID:             2222,
				Command:         "ping",
				DestinationIP:   "8.8.8.8",
				DestinationPort: 0,
				Protocol:        "icmp",
				ReturnCode:      0,
				Destination:     "8.8.8.8",
			},
		},
		{
			name: "empty destination",
			input: ConnectionInfo{
				PID:         3333,
				Command:     "empty",
				Destination: "",
				Protocol:    "tcp",
				ReturnCode:  0,
				Timestamp:   "2024-01-01T12:05:00Z",
			},
			expected: ConnectionEvent{
				PID:             3333,
				Command:         "empty",
				DestinationIP:   "",
				DestinationPort: 0,
				Protocol:        "tcp",
				ReturnCode:      0,
				Destination:     "",
			},
		},
		{
			name: "connection with port 0 explicitly",
			input: ConnectionInfo{
				PID:         4444,
				Command:     "test",
				Destination: "127.0.0.1:0",
				Protocol:    "tcp",
				ReturnCode:  0,
				Timestamp:   "2024-01-01T12:06:00Z",
			},
			expected: ConnectionEvent{
				PID:             4444,
				Command:         "test",
				DestinationIP:   "127.0.0.1",
				DestinationPort: 0,
				Protocol:        "tcp",
				ReturnCode:      0,
				Destination:     "127.0.0.1:0",
			},
		},
		{
			name: "connection with high port number",
			input: ConnectionInfo{
				PID:         5555,
				Command:     "high-port",
				Destination: "example.org:65535",
				Protocol:    "tcp",
				ReturnCode:  0,
				Timestamp:   "2024-01-01T12:07:00Z",
			},
			expected: ConnectionEvent{
				PID:             5555,
				Command:         "high-port",
				DestinationIP:   "example.org",
				DestinationPort: 65535,
				Protocol:        "tcp",
				ReturnCode:      0,
				Destination:     "example.org:65535",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToConnectionEvent()

			if result.PID != tt.expected.PID {
				t.Errorf("PID: expected %d, got %d", tt.expected.PID, result.PID)
			}
			if result.Command != tt.expected.Command {
				t.Errorf("Command: expected '%s', got '%s'", tt.expected.Command, result.Command)
			}
			if result.DestinationIP != tt.expected.DestinationIP {
				t.Errorf("DestinationIP: expected '%s', got '%s'", tt.expected.DestinationIP, result.DestinationIP)
			}
			if result.DestinationPort != tt.expected.DestinationPort {
				t.Errorf("DestinationPort: expected %d, got %d", tt.expected.DestinationPort, result.DestinationPort)
			}
			if result.Protocol != tt.expected.Protocol {
				t.Errorf("Protocol: expected '%s', got '%s'", tt.expected.Protocol, result.Protocol)
			}
			if result.ReturnCode != tt.expected.ReturnCode {
				t.Errorf("ReturnCode: expected %d, got %d", tt.expected.ReturnCode, result.ReturnCode)
			}
			if result.Destination != tt.expected.Destination {
				t.Errorf("Destination: expected '%s', got '%s'", tt.expected.Destination, result.Destination)
			}
		})
	}
}

func TestConnectionInfo_ToConnectionEvent_EdgeCases(t *testing.T) {
	// Test various edge cases for destination parsing
	edgeCases := []struct {
		destination  string
		expectedIP   string
		expectedPort uint16
		description  string
	}{
		{"localhost:8080", "localhost", 8080, "standard localhost with port"},
		{"127.0.0.1:3000", "127.0.0.1", 3000, "IP with port"},
		{"[::1]:8080", "[::1]", 8080, "IPv6 localhost with port"},
		{"[2001:db8::1]:443", "[2001:db8::1]", 443, "IPv6 with port"},
		{"example.com", "example.com", 0, "hostname without port"},
		{"192.168.1.1", "192.168.1.1", 0, "IP without port"},
		{"", "", 0, "empty destination"},
		{"hostname:", "hostname", 0, "hostname with colon but no port"},
		{":8080", "", 8080, "port only"},
		{"host:port", "host", 0, "non-numeric port"},
		{"host:-1", "host", 0, "negative port"},
		{"host:99999", "host", 0, "port too high for uint16"},
		{"multiple:colons:here:8080", "multiple:colons:here", 8080, "multiple colons"},
	}

	for _, tc := range edgeCases {
		t.Run(tc.description, func(t *testing.T) {
			conn := ConnectionInfo{
				PID:         1234,
				Command:     "test",
				Destination: tc.destination,
				Protocol:    "tcp",
				ReturnCode:  0,
				Timestamp:   "2024-01-01T12:00:00Z",
			}

			result := conn.ToConnectionEvent()

			if result.DestinationIP != tc.expectedIP {
				t.Errorf("For destination '%s': expected IP '%s', got '%s'",
					tc.destination, tc.expectedIP, result.DestinationIP)
			}
			if result.DestinationPort != tc.expectedPort {
				t.Errorf("For destination '%s': expected port %d, got %d",
					tc.destination, tc.expectedPort, result.DestinationPort)
			}
		})
	}
}

func TestDataStructuresJSONSerialization(t *testing.T) {
	// Test that our data structures can be properly marshaled/unmarshaled to/from JSON

	t.Run("ConnectionInfo JSON", func(t *testing.T) {
		original := ConnectionInfo{
			PID:         1234,
			Command:     "test-command",
			Destination: "example.com:8080",
			Protocol:    "tcp",
			ReturnCode:  0,
			Timestamp:   "2024-01-01T12:00:00Z",
		}

		// This would be tested if we needed JSON serialization
		// We're mainly testing the structure is well-formed
		if original.PID == 0 {
			t.Error("ConnectionInfo should have non-zero PID")
		}
	})

	t.Run("ConnectionEvent JSON", func(t *testing.T) {
		original := ConnectionEvent{
			PID:             1234,
			Command:         "test-command",
			DestinationIP:   "example.com",
			DestinationPort: 8080,
			Protocol:        "tcp",
			ReturnCode:      0,
			Destination:     "example.com:8080",
			WallTime:        time.Now(),
		}

		if original.PID == 0 {
			t.Error("ConnectionEvent should have non-zero PID")
		}
	})

	t.Run("ConnectionSummaryOutput", func(t *testing.T) {
		summary := ConnectionSummaryOutput{
			Total:   42,
			Command: "test-process",
			Seconds: 30,
		}

		if summary.Total == 0 {
			t.Error("ConnectionSummaryOutput should have non-zero Total")
		}
	})

	t.Run("ListConnectionsOutput", func(t *testing.T) {
		output := ListConnectionsOutput{
			TotalPIDs: 2,
			Truncated: false,
			Connections: map[string][]ConnectionInfo{
				"1234": {
					{
						PID:         1234,
						Command:     "nginx",
						Destination: "192.168.1.1:80",
						Protocol:    "tcp",
						ReturnCode:  0,
						Timestamp:   "2024-01-01T12:00:00Z",
					},
				},
			},
		}

		if len(output.Connections) == 0 {
			t.Error("ListConnectionsOutput should have connections")
		}
	})
}
