package utils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/srodi/netspy/internal/netclient"
)

// FormatConnectionSummary creates a human-readable summary of connection data
func FormatConnectionSummary(pid int, processName string, duration int, summary netclient.ConnectionSummaryOutput) string {
	var target string
	if pid > 0 {
		target = fmt.Sprintf("PID %d", pid)
	} else {
		target = fmt.Sprintf("process '%s'", processName)
	}

	if summary.Count == 0 {
		return fmt.Sprintf("No network connections found for %s in the last %d seconds", target, duration)
	}

	return fmt.Sprintf("%s made %d outbound connection attempts over the last %d seconds",
		target, summary.Count, duration)
}

// FormatConnectionEvents provides a detailed view of connection events
func FormatConnectionEvents(events []netclient.ConnectionEvent, maxEvents int) string {
	if len(events) == 0 {
		return "No connection events found"
	}

	// Sort events by timestamp (most recent first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].TimestampNS > events[j].TimestampNS
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Recent connection events (%d total):\n", len(events)))

	limit := maxEvents
	if len(events) < limit {
		limit = len(events)
	}

	for i := 0; i < limit; i++ {
		event := events[i]
		sb.WriteString(fmt.Sprintf("  %s | %s:%d | %s | %s\n",
			event.WallTime.Format("15:04:05"),
			event.DestinationIP,
			event.DestinationPort,
			event.Protocol,
			event.Command))
	}

	if len(events) > maxEvents {
		sb.WriteString(fmt.Sprintf("  ... and %d more events\n", len(events)-maxEvents))
	}

	return sb.String()
}

// AnalyzeConnectionPatterns provides insights about connection patterns
func AnalyzeConnectionPatterns(events []netclient.ConnectionEvent) string {
	if len(events) == 0 {
		return "No patterns to analyze"
	}

	// Count destinations and protocols
	destinations := make(map[string]int)
	protocols := make(map[string]int)

	for _, event := range events {
		dest := fmt.Sprintf("%s:%d", event.DestinationIP, event.DestinationPort)
		destinations[dest]++
		protocols[event.Protocol]++
	}

	var sb strings.Builder
	sb.WriteString("Connection Analysis:\n")

	// Top destinations
	if len(destinations) > 0 {
		sb.WriteString("  Top destinations:\n")
		type destCount struct {
			dest  string
			count int
		}
		var sorted []destCount
		for dest, count := range destinations {
			sorted = append(sorted, destCount{dest, count})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].count > sorted[j].count
		})

		limit := 10
		if len(sorted) < limit {
			limit = len(sorted)
		}
		for i := 0; i < limit; i++ {
			sb.WriteString(fmt.Sprintf("    %s (%d connections)\n", sorted[i].dest, sorted[i].count))
		}
	}

	// Protocol distribution
	if len(protocols) > 0 {
		sb.WriteString("  Protocols: ")
		var protoStrs []string
		for proto, count := range protocols {
			protoStrs = append(protoStrs, fmt.Sprintf("%s (%d)", proto, count))
		}
		sb.WriteString(strings.Join(protoStrs, ", "))
		sb.WriteString("\n")
	}

	return sb.String()
}
