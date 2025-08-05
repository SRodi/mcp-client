package utils

import (
	"testing"
	"time"

	"github.com/srodi/netspy/internal/netclient"
)

func makeEvent(pid int, cmd, ip string, port uint16, proto string, ts uint64) netclient.ConnectionEvent {
	return netclient.ConnectionEvent{
		PID:             uint32(pid),
		Command:         cmd,
		DestinationIP:   ip,
		DestinationPort: port,
		Protocol:        proto,
		TimestampNS:     ts,
		WallTime:        time.Unix(0, int64(ts)),
	}
}

func TestFormatConnectionSummary(t *testing.T) {
	summary := netclient.ConnectionSummaryOutput{}
	summary.Count = 0
	out := FormatConnectionSummary(123, "", 60, summary)
	if out == "" || out[:2] != "No" {
		t.Errorf("expected no connections message, got: %s", out)
	}

	summary.Count = 5
	out = FormatConnectionSummary(123, "", 60, summary)
	if out == "" || out[:3] != "PID" {
		t.Errorf("expected PID summary, got: %s", out)
	}

	out = FormatConnectionSummary(0, "curl", 60, summary)
	if out == "" || out[:7] != "process" {
		t.Errorf("expected process summary, got: %s", out)
	}
}

func TestFormatConnectionEvents(t *testing.T) {
	events := []netclient.ConnectionEvent{
		makeEvent(1, "curl", "1.2.3.4", 80, "tcp", 200),
		makeEvent(2, "wget", "5.6.7.8", 443, "tcp", 100),
	}
	out := FormatConnectionEvents(events, 1)
	if out == "" || out[:6] != "Recent" {
		t.Errorf("expected formatted events, got: %s", out)
	}
	out = FormatConnectionEvents(nil, 5)
	if out != "No connection events found" {
		t.Errorf("expected no events message, got: %s", out)
	}
}

func TestAnalyzeConnectionPatterns(t *testing.T) {
	events := []netclient.ConnectionEvent{
		makeEvent(1, "curl", "1.2.3.4", 80, "tcp", 200),
		makeEvent(2, "curl", "1.2.3.4", 80, "tcp", 100),
		makeEvent(3, "wget", "5.6.7.8", 443, "udp", 150),
	}
	out := AnalyzeConnectionPatterns(events)
	if out == "" || out[:10] != "Connection" {
		t.Errorf("expected analysis output, got: %s", out)
	}
	out = AnalyzeConnectionPatterns(nil)
	if out != "No patterns to analyze" {
		t.Errorf("expected no patterns message, got: %s", out)
	}
}
