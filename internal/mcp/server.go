package mcp

import (
	"context"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/srodi/netspy/internal/netclient"
	"github.com/srodi/netspy/internal/openai"
	"github.com/srodi/netspy/internal/utils"
)

// NetworkMCPServer implements an MCP server for network telemetry using the official SDK
type NetworkMCPServer struct {
	server          *mcp.Server
	httpClient      *netclient.Client
	ebpfServerURL   string
	verbose         bool
	registeredTools map[string]*mcp.Tool // Store registered tools for discovery
}

// NewNetworkMCPServer creates a new MCP server for network telemetry using the official SDK
func NewNetworkMCPServer(ebpfServerURL string, verbose bool) *NetworkMCPServer {
	s := &NetworkMCPServer{
		httpClient:      netclient.NewClientWithVerbose(ebpfServerURL, verbose),
		ebpfServerURL:   ebpfServerURL,
		verbose:         verbose,
		registeredTools: make(map[string]*mcp.Tool),
	}

	// Create the implementation info
	impl := &mcp.Implementation{
		Name:    "network-telemetry",
		Title:   "Network Telemetry MCP Server",
		Version: "1.0.0",
	}

	// Create server options
	opts := &mcp.ServerOptions{
		Instructions: "Network telemetry analysis server providing real-time network connectivity analytics and AI-powered insights.",
	}

	// Create the MCP server with proper configuration
	s.server = mcp.NewServer(impl, opts)

	// Register tools with the official SDK
	s.registerTools()

	return s
}

// registerTools registers all available MCP tools using the official SDK
func (s *NetworkMCPServer) registerTools() {
	// Register get_network_summary tool
	networkSummaryTool := &mcp.Tool{
		Name:        "get_network_summary",
		Description: "Get a summary of network connections for a specific process or PID",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"pid": {
					Type:        "integer",
					Description: "Process ID to analyze (optional, use either pid or process_name)",
				},
				"process_name": {
					Type:        "string",
					Description: "Process name to analyze (optional, use either pid or process_name)",
				},
				"duration": {
					Type:        "integer",
					Description: "Duration in seconds to analyze (default: 60)",
					Default:     []byte("60"),
				},
			},
		},
	}
	mcp.AddTool(s.server, networkSummaryTool, s.handleGetNetworkSummary)
	s.registeredTools["get_network_summary"] = networkSummaryTool

	// Register list_connections tool
	listConnectionsTool := &mcp.Tool{
		Name:        "list_connections",
		Description: "List recent network connection events",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"pid": {
					Type:        "integer",
					Description: "Filter by process ID (optional)",
				},
				"process_name": {
					Type:        "string",
					Description: "Filter by process name (optional)",
				},
				"max_events": {
					Type:        "integer",
					Description: "Maximum number of events to return (default: 10)",
					Default:     []byte("10"),
				},
			},
		},
	}
	mcp.AddTool(s.server, listConnectionsTool, s.handleListConnections)
	s.registeredTools["list_connections"] = listConnectionsTool

	// Register analyze_patterns tool
	analyzePatternsTool := &mcp.Tool{
		Name:        "analyze_patterns",
		Description: "Analyze network connection patterns and provide insights about connection behavior, frequency, and destinations",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"pid": {
					Type:        "integer",
					Description: "Process ID to analyze (optional, use either pid or process_name)",
				},
				"process_name": {
					Type:        "string",
					Description: "Process name to analyze (optional, use either pid or process_name)",
				},
				"duration": {
					Type:        "integer",
					Description: "Duration in seconds to analyze (default: 60)",
					Default:     []byte("60"),
				},
			},
		},
	}
	mcp.AddTool(s.server, analyzePatternsTool, s.handleAnalyzePatterns)
	s.registeredTools["analyze_patterns"] = analyzePatternsTool

	// Register intelligent_analysis tool - this uses OpenAI with function calling
	contextualAnalysisTool := &mcp.Tool{
		Name:        "intelligent_analysis",
		Description: "Get AI-powered network analysis with automatic tool usage and comprehensive insights",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"query": {
					Type:        "string",
					Description: "Natural language query about network behavior or analysis needed",
				},
				"process_name": {
					Type:        "string",
					Description: "Process name to focus analysis on (optional)",
				},
				"pid": {
					Type:        "integer",
					Description: "Process ID to focus analysis on (optional)",
				},
				"duration": {
					Type:        "integer",
					Description: "Duration in seconds for analysis (default: 60)",
					Default:     []byte("60"),
				},
			},
			Required: []string{"query"},
		},
	}
	mcp.AddTool(s.server, contextualAnalysisTool, s.handleContextualAnalysis)
	s.registeredTools["contextual_analysis"] = contextualAnalysisTool

	// Register ai_insights tool
	aiInsightsTool := &mcp.Tool{
		Name:        "ai_insights",
		Description: "Get AI-powered insights about network behavior using OpenAI GPT-3.5-turbo",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"summary_text": {
					Type:        "string",
					Description: "Network summary text to analyze",
				},
			},
			Required: []string{"summary_text"},
		},
	}
	mcp.AddTool(s.server, aiInsightsTool, s.handleAIInsights)
	s.registeredTools["ai_insights"] = aiInsightsTool

	// Register get_packet_drop_summary tool
	packetDropSummaryTool := &mcp.Tool{
		Name:        "get_packet_drop_summary",
		Description: "Get a summary count of all packet drop events for a specific process or PID. This includes all drops from the eBPF server (including irrelevant ones like SK_FREE). For filtered relevant drops, use list_packet_drops instead.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"pid": {
					Type:        "integer",
					Description: "Process ID to analyze (optional, use either pid or process_name)",
				},
				"process_name": {
					Type:        "string",
					Description: "Process name to analyze (optional, use either pid or process_name)",
				},
				"duration": {
					Type:        "integer",
					Description: "Duration in seconds to analyze (default: 60)",
					Default:     []byte("60"),
				},
			},
		},
	}
	mcp.AddTool(s.server, packetDropSummaryTool, s.handleGetPacketDropSummary)
	s.registeredTools["get_packet_drop_summary"] = packetDropSummaryTool

	// Register list_packet_drops tool
	listPacketDropsTool := &mcp.Tool{
		Name:        "list_packet_drops",
		Description: "List recent packet drop events. Automatically filters out irrelevant drops (like SK_FREE) to focus on meaningful network issues (like TCP, UDP, routing problems). Only shows packet drops that indicate actual connectivity or performance problems.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"pid": {
					Type:        "integer",
					Description: "Filter by process ID (optional)",
				},
				"process_name": {
					Type:        "string",
					Description: "Filter by process name (optional)",
				},
				"max_events": {
					Type:        "integer",
					Description: "Maximum number of events to return (default: 10)",
					Default:     []byte("10"),
				},
			},
		},
	}
	mcp.AddTool(s.server, listPacketDropsTool, s.handleListPacketDrops)
	s.registeredTools["list_packet_drops"] = listPacketDropsTool
}

// handleGetNetworkSummary handles the get_network_summary tool call
func (s *NetworkMCPServer) handleGetNetworkSummary(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
	if s.verbose {
		log.Printf("MCP Server: Handling get_network_summary request")
	}

	// Parse arguments
	arguments := params.Arguments
	var pid int
	var processName string
	duration := 60

	if pidVal, exists := arguments["pid"]; exists {
		if pidFloat, ok := pidVal.(float64); ok {
			pid = int(pidFloat)
		} else if pidInt, ok := pidVal.(int); ok {
			pid = pidInt
		}
	}

	if procVal, exists := arguments["process_name"]; exists {
		if procStr, ok := procVal.(string); ok {
			processName = procStr
		}
	}

	if durVal, exists := arguments["duration"]; exists {
		if durFloat, ok := durVal.(float64); ok {
			duration = int(durFloat)
		} else if durInt, ok := durVal.(int); ok {
			duration = durInt
		}
	}

	if s.verbose {
		log.Printf("MCP Server: get_network_summary called with pid=%d, processName='%s', duration=%d", pid, processName, duration)
	}

	// Connect to eBPF server
	if err := s.httpClient.Connect(ctx); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to connect to eBPF server: %v", err),
				},
			},
		}, nil
	}

	// Get summary from eBPF server
	summary, err := s.httpClient.GetConnectionSummary(ctx, pid, processName, duration)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to get connection summary: %v", err),
				},
			},
		}, nil
	}

	// Format the response
	formattedSummary := utils.FormatConnectionSummary(pid, processName, duration, summary)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: formattedSummary,
			},
		},
	}, nil
}

// handleListConnections handles the list_connections tool call
func (s *NetworkMCPServer) handleListConnections(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
	if s.verbose {
		log.Printf("MCP Server: Handling list_connections request")
	}

	// Parse arguments
	arguments := params.Arguments
	var pid *int
	var processName string
	maxEvents := 10

	if pidVal, exists := arguments["pid"]; exists {
		if pidFloat, ok := pidVal.(float64); ok {
			pidInt := int(pidFloat)
			pid = &pidInt
		} else if pidInt, ok := pidVal.(int); ok {
			pid = &pidInt
		}
	}

	if procVal, exists := arguments["process_name"]; exists {
		if procStr, ok := procVal.(string); ok {
			processName = procStr
		}
	}

	if maxVal, exists := arguments["max_events"]; exists {
		if maxFloat, ok := maxVal.(float64); ok {
			maxEvents = int(maxFloat)
		} else if maxInt, ok := maxVal.(int); ok {
			maxEvents = maxInt
		}
	}

	if s.verbose {
		pidStr := "nil"
		if pid != nil {
			pidStr = fmt.Sprintf("%d", *pid)
		}
		log.Printf("MCP Server: list_connections called with pid=%s, processName='%s', maxEvents=%d", pidStr, processName, maxEvents)
	}

	// Connect to eBPF server
	if err := s.httpClient.Connect(ctx); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to connect to eBPF server: %v", err),
				},
			},
		}, nil
	}

	// Get connections from eBPF server
	output, err := s.httpClient.ListConnections(ctx, pid, nil)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to list connections: %v", err),
				},
			},
		}, nil
	}

	// Convert to connection events and filter
	var allEvents []netclient.ConnectionEvent
	for _, connections := range output.EventsByPID {
		for _, conn := range connections {
			event := conn.ToConnectionEvent()
			// Filter by process name if specified
			if processName == "" || event.Command == processName {
				allEvents = append(allEvents, event)
			}
		}
	}

	// Format the response
	formattedList := utils.FormatConnectionEvents(allEvents, maxEvents)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: formattedList,
			},
		},
	}, nil
}

// handleAnalyzePatterns handles the analyze_patterns tool call
func (s *NetworkMCPServer) handleAnalyzePatterns(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
	if s.verbose {
		log.Printf("MCP Server: Handling analyze_patterns request")
	}

	// Parse arguments
	arguments := params.Arguments
	var pid *int
	var processName string

	if pidVal, exists := arguments["pid"]; exists {
		if pidFloat, ok := pidVal.(float64); ok {
			pidInt := int(pidFloat)
			pid = &pidInt
		} else if pidInt, ok := pidVal.(int); ok {
			pid = &pidInt
		}
	}

	if procVal, exists := arguments["process_name"]; exists {
		if procStr, ok := procVal.(string); ok {
			processName = procStr
		}
	}

	// Connect to eBPF server
	if err := s.httpClient.Connect(ctx); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to connect to eBPF server: %v", err),
				},
			},
		}, nil
	}

	// Get connections from eBPF server
	output, err := s.httpClient.ListConnections(ctx, pid, nil)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to list connections: %v", err),
				},
			},
		}, nil
	}

	// Convert to connection events and filter
	var filteredEvents []netclient.ConnectionEvent
	for _, connections := range output.EventsByPID {
		for _, conn := range connections {
			event := conn.ToConnectionEvent()
			if (pid == nil || event.PID == uint32(*pid)) &&
				(processName == "" || event.Command == processName) {
				filteredEvents = append(filteredEvents, event)
			}
		}
	}

	if len(filteredEvents) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "No connection events found for analysis",
				},
			},
		}, nil
	}

	// Analyze patterns
	analysis := utils.AnalyzeConnectionPatterns(filteredEvents)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: analysis,
			},
		},
	}, nil
}

// handleAIInsights handles the ai_insights tool call
func (s *NetworkMCPServer) handleAIInsights(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
	if s.verbose {
		log.Printf("MCP Server: Handling ai_insights request")
	}

	// Parse arguments
	arguments := params.Arguments
	summaryText, exists := arguments["summary_text"]
	if !exists {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Missing required parameter: summary_text",
				},
			},
		}, nil
	}

	summaryStr, ok := summaryText.(string)
	if !ok {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Invalid summary_text parameter: must be a string",
				},
			},
		}, nil
	}

	// Get AI insights using OpenAI
	insights, err := openai.AskLLM(summaryStr)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to get AI insights: %v\n(Ensure OPENAI_API_KEY environment variable is set)", err),
				},
			},
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: insights,
			},
		},
	}, nil
}

// Start starts the MCP server
func (s *NetworkMCPServer) Start(ctx context.Context) error {
	if s.verbose {
		log.Printf("Starting Network Telemetry MCP Server")
	}

	// Test connection to eBPF server
	if err := s.httpClient.Connect(ctx); err != nil {
		log.Printf("Warning: Could not connect to eBPF server at %s: %v", s.ebpfServerURL, err)
		log.Printf("Make sure the eBPF server is running with: sudo ./bin/ebpf-server --http --port 8080")
	} else {
		log.Printf("Successfully connected to eBPF server at %s", s.ebpfServerURL)
	}

	return nil
}

// handleGetPacketDropSummary handles the get_packet_drop_summary tool call
func (s *NetworkMCPServer) handleGetPacketDropSummary(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
	if s.verbose {
		log.Printf("MCP Server: Handling get_packet_drop_summary request")
	}

	// Parse arguments
	arguments := params.Arguments
	var pid int
	var processName string
	duration := 60

	if pidVal, exists := arguments["pid"]; exists {
		if pidFloat, ok := pidVal.(float64); ok {
			pid = int(pidFloat)
		} else if pidInt, ok := pidVal.(int); ok {
			pid = pidInt
		}
	}

	if procVal, exists := arguments["process_name"]; exists {
		if procStr, ok := procVal.(string); ok {
			processName = procStr
		}
	}

	if durVal, exists := arguments["duration"]; exists {
		if durFloat, ok := durVal.(float64); ok {
			duration = int(durFloat)
		} else if durInt, ok := durVal.(int); ok {
			duration = durInt
		}
	}

	// Get packet drop summary from eBPF server
	summary, err := s.httpClient.GetPacketDropSummary(ctx, pid, processName, duration)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to get packet drop summary: %v", err),
				},
			},
		}, nil
	}

	// Format the response
	var target string
	if pid > 0 {
		target = fmt.Sprintf("PID %d", pid)
	} else if processName != "" {
		target = fmt.Sprintf("process '%s'", processName)
	} else {
		target = "all processes"
	}

	var result string
	if summary.Count == 0 {
		result = fmt.Sprintf("No packet drops found for %s in the last %d seconds", target, duration)
	} else {
		result = fmt.Sprintf("%s had %d packet drops over the last %d seconds", target, summary.Count, duration)
	}

	if summary.QueryTime != "" {
		result += fmt.Sprintf(" (query time: %s)", summary.QueryTime)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: result,
			},
		},
	}, nil
}

// handleListPacketDrops handles the list_packet_drops tool call
func (s *NetworkMCPServer) handleListPacketDrops(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
	if s.verbose {
		log.Printf("MCP Server: Handling list_packet_drops request")
	}

	// Parse arguments
	arguments := params.Arguments
	var pid *int
	var processName string
	maxEvents := 10

	if pidVal, exists := arguments["pid"]; exists {
		if pidFloat, ok := pidVal.(float64); ok {
			pidInt := int(pidFloat)
			pid = &pidInt
		} else if pidInt, ok := pidVal.(int); ok {
			pid = &pidInt
		}
	}

	if procVal, exists := arguments["process_name"]; exists {
		if procStr, ok := procVal.(string); ok {
			processName = procStr
		}
	}

	if maxVal, exists := arguments["max_events"]; exists {
		if maxFloat, ok := maxVal.(float64); ok {
			maxEvents = int(maxFloat)
		} else if maxInt, ok := maxVal.(int); ok {
			maxEvents = maxInt
		}
	}

	// Get packet drops from eBPF server
	output, err := s.httpClient.ListPacketDrops(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to list packet drops: %v", err),
				},
			},
		}, nil
	}

	// Convert and filter packet drops
	var filteredDrops []string
	count := 0
	for _, drops := range output.EventsByPID {
		for _, drop := range drops {
			// Filter by PID if specified
			if pid != nil && drop.PID != uint32(*pid) {
				continue
			}
			// Filter by process name if specified
			if processName != "" && drop.Command != processName {
				continue
			}

			if count >= maxEvents {
				break
			}

			dropInfo := fmt.Sprintf("PID %d (%s): packet dropped - %s", drop.PID, drop.Command, drop.Reason)
			filteredDrops = append(filteredDrops, dropInfo)
			count++
		}
		if count >= maxEvents {
			break
		}
	}

	var result string
	if len(filteredDrops) == 0 {
		result = "No packet drop events found"
		if pid != nil {
			result += fmt.Sprintf(" for PID %d", *pid)
		}
		if processName != "" {
			result += fmt.Sprintf(" for process '%s'", processName)
		}
	} else {
		result = fmt.Sprintf("Recent packet drop events (%d total):\n", output.TotalEvents)
		for i, drop := range filteredDrops {
			result += fmt.Sprintf("%d. %s\n", i+1, drop)
		}
		if output.TotalEvents > len(filteredDrops) {
			result += fmt.Sprintf("... and %d more events", output.TotalEvents-len(filteredDrops))
		}
	}

	if output.QueryTime != "" {
		result += fmt.Sprintf("\nQuery time: %s", output.QueryTime)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: result,
			},
		},
	}, nil
}

// GetServer returns the underlying MCP server
func (s *NetworkMCPServer) GetServer() *mcp.Server {
	return s.server
}

// GetRegisteredTools implements the MCPToolDiscovery interface
func (s *NetworkMCPServer) GetRegisteredTools() map[string]*mcp.Tool {
	return s.registeredTools
}

// RunSingleCommand implements the MCPToolExecutor interface for the OpenAI function calling
func (s *NetworkMCPServer) RunSingleCommand(ctx context.Context, toolName string, arguments map[string]any) (*mcp.CallToolResult, error) {
	params := &mcp.CallToolParamsFor[map[string]any]{
		Arguments: arguments,
	}

	// Check if the tool exists in our registered tools
	if _, exists := s.registeredTools[toolName]; !exists {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Unknown tool: %s", toolName),
				},
			},
		}, nil
	}

	switch toolName {
	case "get_network_summary":
		return s.handleGetNetworkSummary(ctx, nil, params)
	case "list_connections":
		return s.handleListConnections(ctx, nil, params)
	case "get_packet_drop_summary":
		return s.handleGetPacketDropSummary(ctx, nil, params)
	case "list_packet_drops":
		return s.handleListPacketDrops(ctx, nil, params)
	case "analyze_patterns":
		return s.handleAnalyzePatterns(ctx, nil, params)
	case "ai_insights":
		return s.handleAIInsights(ctx, nil, params)
	case "contextual_analysis":
		return s.handleContextualAnalysis(ctx, nil, params)
	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Tool %s is registered but handler not implemented", toolName),
				},
			},
		}, nil
	}
}

// handleContextualAnalysis handles the contextual_analysis tool call using OpenAI function calling
func (s *NetworkMCPServer) handleContextualAnalysis(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
	if s.verbose {
		log.Printf("MCP Server: Handling contextual_analysis request")
	}

	// Parse arguments
	arguments := params.Arguments
	query, exists := arguments["query"]
	if !exists {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Error: query parameter is required for intelligent analysis",
				},
			},
		}, nil
	}

	queryStr, ok := query.(string)
	if !ok {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Error: query must be a string",
				},
			},
		}, nil
	}

	// Parse optional parameters
	var pid int
	var processName string
	duration := 60

	if pidVal, exists := arguments["pid"]; exists {
		if pidFloat, ok := pidVal.(float64); ok {
			pid = int(pidFloat)
		}
	}

	if processNameVal, exists := arguments["process_name"]; exists {
		if processNameStr, ok := processNameVal.(string); ok {
			processName = processNameStr
		}
	}

	if durationVal, exists := arguments["duration"]; exists {
		if durFloat, ok := durationVal.(float64); ok {
			duration = int(durFloat)
		} else if durInt, ok := durationVal.(int); ok {
			duration = durInt
		}
	}

	// Create the intelligent network analyst
	analyst := openai.NewContextualNetworkAnalyst(s, s.verbose)

	// If specific process parameters are provided, do focused analysis
	if processName != "" || pid > 0 {
		analysis, err := analyst.AnalyzeProcess(ctx, processName, pid, duration)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Error during intelligent analysis: %v", err),
					},
				},
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: analysis,
				},
			},
		}, nil
	}

	// Otherwise, process the general query
	analysis, err := analyst.AnalyzeNetworkQuery(ctx, queryStr)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error during intelligent analysis: %v", err),
				},
			},
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: analysis,
			},
		},
	}, nil
}
