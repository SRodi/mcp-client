package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/srodi/netspy/internal/mcp"
)

func main() {
	var (
		ebpfServerURL = flag.String("server", "http://localhost:8080", "eBPF server URL")
		verbose       = flag.Bool("verbose", false, "Enable verbose logging")
		mcpTool       = flag.String("tool", "", "Run a specific MCP tool (get_network_summary, list_connections, analyze_patterns, ai_insights)")
		pid           = flag.Int("pid", 0, "Process ID to monitor")
		processName   = flag.String("process", "", "Process name to monitor")
		duration      = flag.Int("duration", 60, "Duration in seconds for monitoring")
		maxEvents     = flag.Int("max-events", 100, "Maximum number of events to retrieve")
		summaryText   = flag.String("summary-text", "", "Summary text for AI insights")
		help          = flag.Bool("help", false, "Show help information")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Setup context
	ctx := context.Background()

	// Create MCP client
	mcpClient := mcp.NewMCPClient(*ebpfServerURL, *verbose)

	// If a specific tool is requested, run it and exit
	if *mcpTool != "" {
		arguments := buildMCPArguments(*pid, *processName, *duration, *maxEvents, *summaryText)
		result, err := mcpClient.RunSingleCommand(ctx, *mcpTool, arguments)
		if err != nil {
			log.Fatalf("MCP tool execution failed: %v", err)
		}

		// Print result
		for _, content := range result.Content {
			if textContent, ok := content.(*mcpsdk.TextContent); ok {
				fmt.Println(textContent.Text)
			}
		}
		return
	}

	// Default: start interactive MCP mode
	fmt.Println("🔗 Network Telemetry MCP Server")
	fmt.Println("Starting interactive mode...")
	fmt.Println()

	if err := mcpClient.StartInteractiveMode(ctx); err != nil {
		log.Fatalf("MCP interactive mode failed: %v", err)
	}
}

func showHelp() {
	fmt.Println("Network Telemetry MCP Client")
	fmt.Println("============================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  netspy [OPTIONS]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --server URL          eBPF server URL (default: http://localhost:8080)")
	fmt.Println("  --verbose             Enable verbose logging")
	fmt.Println("  --help                Show this help message")
	fmt.Println()
	fmt.Println("Tool Execution (run specific tool and exit):")
	fmt.Println("  --tool TOOL           Run specific MCP tool:")
	fmt.Println("                          get_network_summary")
	fmt.Println("                          list_connections")
	fmt.Println("                          get_packet_drop_summary")
	fmt.Println("                          list_packet_drops")
	fmt.Println("                          analyze_patterns")
	fmt.Println("                          ai_insights")
	fmt.Println()
	fmt.Println("Tool Parameters:")
	fmt.Println("  --pid PID             Process ID to monitor")
	fmt.Println("  --process NAME        Process name to monitor")
	fmt.Println("  --duration SECONDS    Duration in seconds (default: 60)")
	fmt.Println("  --max-events COUNT    Maximum events to retrieve (default: 100)")
	fmt.Println("  --summary-text TEXT   Summary text for AI insights")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Interactive mode")
	fmt.Println("  netspy")
	fmt.Println()
	fmt.Println("  # Run specific tool")
	fmt.Println("  netspy --tool get_network_summary --process curl --duration 120")
	fmt.Println("  netspy --tool list_connections --pid 1234")
	fmt.Println("  netspy --tool get_packet_drop_summary --process nginx --duration 300")
	fmt.Println("  netspy --tool list_packet_drops --pid 1234")
	fmt.Println("  netspy --tool ai_insights --summary-text \"High network activity detected\"")
	fmt.Println()
	fmt.Println("Interactive Commands:")
	fmt.Println("  summary [--pid PID] [--process NAME] [--duration SECONDS]")
	fmt.Println("  list [--pid PID] [--process NAME] [--max-events COUNT]")
	fmt.Println("  dropsummary [--pid PID] [--process NAME] [--duration SECONDS]")
	fmt.Println("  droplist [--pid PID] [--process NAME] [--max-events COUNT]")
	fmt.Println("  analyze [--pid PID] [--process NAME]")
	fmt.Println("  insights <summary_text>")
	fmt.Println("  tools                 Show available MCP tools")
	fmt.Println("  help                  Show command help")
	fmt.Println("  quit/exit             Exit interactive mode")
}

func buildMCPArguments(pid int, processName string, duration, maxEvents int, summaryText string) map[string]any {
	arguments := make(map[string]any)

	if pid > 0 {
		arguments["pid"] = pid
	}
	if processName != "" {
		arguments["process_name"] = processName
	}
	if duration > 0 {
		arguments["duration"] = duration
	}
	if maxEvents > 0 {
		arguments["max_events"] = maxEvents
	}
	if summaryText != "" {
		arguments["summary_text"] = summaryText
	}

	return arguments
}
