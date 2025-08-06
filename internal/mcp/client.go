package mcp

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPClient provides an interactive interface to the MCP server
type MCPClient struct {
	server  *NetworkMCPServer
	verbose bool
}

// NewMCPClient creates a new MCP client
func NewMCPClient(ebpfServerURL string, verbose bool) *MCPClient {
	return &MCPClient{
		server:  NewNetworkMCPServer(ebpfServerURL, verbose),
		verbose: verbose,
	}
}

// StartInteractiveMode starts an interactive session with the MCP server
func (c *MCPClient) StartInteractiveMode(ctx context.Context) error {
	fmt.Println("🔗 Network Telemetry MCP Interactive Mode")
	fmt.Println("=========================================")
	fmt.Println()

	// Start the MCP server
	if err := c.server.Start(ctx); err != nil {
		return fmt.Errorf("failed to start MCP server: %v", err)
	}

	fmt.Println("Available commands:")
	fmt.Println("  summary      - Get network connection summary")
	fmt.Println("  list         - List recent connections")
	fmt.Println("  dropsummary  - Get packet drop summary")
	fmt.Println("  droplist     - List recent packet drops")
	fmt.Println("  analyze      - Analyze connection patterns")
	fmt.Println("  insights     - Get AI insights about network behavior")
	fmt.Println("  intelligent  - Get intelligent AI analysis with automatic tool usage")
	fmt.Println("  tools        - Show available MCP tools")
	fmt.Println("  help         - Show this help message")
	fmt.Println("  quit/exit    - Exit interactive mode")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("netspy-mcp> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "quit" || input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if err := c.handleCommand(ctx, input); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		fmt.Println()
	}

	return scanner.Err()
}

// handleCommand processes user commands in interactive mode
func (c *MCPClient) handleCommand(ctx context.Context, input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]

	switch command {
	case "help":
		c.showHelp()
		return nil

	case "tools":
		c.showTools()
		return nil

	case "summary":
		return c.handleSummaryCommand(ctx, parts[1:])

	case "list":
		return c.handleListCommand(ctx, parts[1:])

	case "dropsummary":
		return c.handleDropSummaryCommand(ctx, parts[1:])

	case "droplist":
		return c.handleDropListCommand(ctx, parts[1:])

	case "analyze":
		return c.handleAnalyzeCommand(ctx, parts[1:])

	case "insights":
		return c.handleInsightsCommand(ctx, parts[1:])

	case "intelligent":
		return c.handleIntelligentCommand(ctx, parts[1:])

	default:
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", command)
	}
}

// showHelp displays help information
func (c *MCPClient) showHelp() {
	fmt.Println("Network Telemetry MCP Commands:")
	fmt.Println()
	fmt.Println("summary [--pid <pid>] [--process <n>] [--duration <seconds>]")
	fmt.Println("  Get a summary of network connections")
	fmt.Println("  Examples:")
	fmt.Println("    summary --pid 1234")
	fmt.Println("    summary --process curl --duration 120")
	fmt.Println()
	fmt.Println("list [--pid <pid>] [--process <n>] [--max-events <count>]")
	fmt.Println("  List recent network connection events")
	fmt.Println("  Examples:")
	fmt.Println("    list")
	fmt.Println("    list --process nginx --max-events 20")
	fmt.Println()
	fmt.Println("dropsummary [--pid <pid>] [--process <n>] [--duration <seconds>]")
	fmt.Println("  Get a summary of packet drop events")
	fmt.Println("  Examples:")
	fmt.Println("    dropsummary --pid 1234")
	fmt.Println("    dropsummary --process nginx --duration 300")
	fmt.Println()
	fmt.Println("droplist [--pid <pid>] [--process <n>] [--max-events <count>]")
	fmt.Println("  List recent packet drop events")
	fmt.Println("  Examples:")
	fmt.Println("    droplist")
	fmt.Println("    droplist --process nginx --max-events 15")
	fmt.Println()
	fmt.Println("analyze [--pid <pid>] [--process <n>]")
	fmt.Println("  Analyze network connection patterns")
	fmt.Println("  Examples:")
	fmt.Println("    analyze --process ssh")
	fmt.Println()
	fmt.Println("insights <summary_text>")
	fmt.Println("  Get AI-powered insights about network behavior")
	fmt.Println("  Examples:")
	fmt.Println("    insights \"curl made 5 connections in 60 seconds\"")
	fmt.Println()
	fmt.Println("intelligent <query>")
	fmt.Println("  Get intelligent AI analysis with automatic tool usage and comprehensive insights")
	fmt.Println("  Examples:")
	fmt.Println("    intelligent \"Analyze the network behavior of process nginx\"")
	fmt.Println("    intelligent \"What's happening with my network connections?\"")
	fmt.Println("    intelligent \"Are there any packet drops or connection issues?\"")
}

// showTools displays available MCP tools
func (c *MCPClient) showTools() {
	fmt.Println("Available MCP Tools:")
	fmt.Println()
	fmt.Println("• get_network_summary: Get a summary of network connections for a specific process or PID")
	fmt.Println("• list_connections: List recent network connection events")
	fmt.Println("• get_packet_drop_summary: Get a summary of packet drop events for a specific process or PID")
	fmt.Println("• list_packet_drops: List recent packet drop events")
	fmt.Println("• analyze_patterns: Analyze network connection patterns and provide insights")
	fmt.Println("• ai_insights: Get AI-powered insights about network behavior using OpenAI GPT-3.5-turbo")
	fmt.Println("• intelligent_analysis: Get intelligent AI analysis with automatic tool usage and comprehensive insights")
}

// handleSummaryCommand processes the summary command
func (c *MCPClient) handleSummaryCommand(ctx context.Context, args []string) error {
	arguments := c.parseArguments(args)

	params := &mcp.CallToolParamsFor[map[string]any]{
		Arguments: arguments,
	}

	result, err := c.server.handleGetNetworkSummary(ctx, nil, params)
	if err != nil {
		return err
	}

	c.printResult(result)
	return nil
}

// handleListCommand processes the list command
func (c *MCPClient) handleListCommand(ctx context.Context, args []string) error {
	arguments := c.parseArguments(args)

	params := &mcp.CallToolParamsFor[map[string]any]{
		Arguments: arguments,
	}

	result, err := c.server.handleListConnections(ctx, nil, params)
	if err != nil {
		return err
	}

	c.printResult(result)
	return nil
}

// handleAnalyzeCommand processes the analyze command
func (c *MCPClient) handleAnalyzeCommand(ctx context.Context, args []string) error {
	arguments := c.parseArguments(args)

	params := &mcp.CallToolParamsFor[map[string]any]{
		Arguments: arguments,
	}

	result, err := c.server.handleAnalyzePatterns(ctx, nil, params)
	if err != nil {
		return err
	}

	c.printResult(result)
	return nil
}

// handleInsightsCommand processes the insights command
func (c *MCPClient) handleInsightsCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("insights command requires summary text as argument")
	}

	summaryText := strings.Join(args, " ")
	// Remove quotes if present
	if strings.HasPrefix(summaryText, "\"") && strings.HasSuffix(summaryText, "\"") {
		summaryText = strings.Trim(summaryText, "\"")
	}

	arguments := map[string]any{
		"summary_text": summaryText,
	}

	params := &mcp.CallToolParamsFor[map[string]any]{
		Arguments: arguments,
	}

	result, err := c.server.handleAIInsights(ctx, nil, params)
	if err != nil {
		return err
	}

	c.printResult(result)
	return nil
}

// handleIntelligentCommand processes the intelligent analysis command
func (c *MCPClient) handleIntelligentCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("intelligent command requires a query as argument")
	}

	query := strings.Join(args, " ")
	// Remove quotes if present
	if strings.HasPrefix(query, "\"") && strings.HasSuffix(query, "\"") {
		query = strings.Trim(query, "\"")
	}

	arguments := map[string]any{
		"query": query,
	}

	params := &mcp.CallToolParamsFor[map[string]any]{
		Arguments: arguments,
	}

	result, err := c.server.handleIntelligentAnalysis(ctx, nil, params)
	if err != nil {
		return err
	}

	c.printResult(result)
	return nil
}

// parseArguments parses command line arguments into a map
func (c *MCPClient) parseArguments(args []string) map[string]any {
	arguments := make(map[string]any)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "--") {
			key := arg[2:] // Remove --

			// Map command line argument names to MCP tool parameter names
			if key == "process" {
				key = "process_name"
			}

			// Check if there's a value following this flag
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				value := args[i+1]
				i++ // Skip the value in next iteration

				// Try to convert to number if possible
				if intVal, err := strconv.Atoi(value); err == nil {
					arguments[key] = intVal
				} else {
					arguments[key] = value
				}
			} else {
				// Boolean flag
				arguments[key] = true
			}
		}
	}

	return arguments
}

// printResult prints the result from an MCP tool call
func (c *MCPClient) printResult(result *mcp.CallToolResult) {
	for _, content := range result.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			fmt.Println(textContent.Text)
		}
	}
}

// RunSingleCommand executes a single MCP command and returns the result
func (c *MCPClient) RunSingleCommand(ctx context.Context, toolName string, arguments map[string]any) (*mcp.CallToolResult, error) {
	// Start the MCP server
	if err := c.server.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start MCP server: %v", err)
	}

	params := &mcp.CallToolParamsFor[map[string]any]{
		Arguments: arguments,
	}

	switch toolName {
	case "get_network_summary":
		return c.server.handleGetNetworkSummary(ctx, nil, params)
	case "list_connections":
		return c.server.handleListConnections(ctx, nil, params)
	case "get_packet_drop_summary":
		return c.server.handleGetPacketDropSummary(ctx, nil, params)
	case "list_packet_drops":
		return c.server.handleListPacketDrops(ctx, nil, params)
	case "analyze_patterns":
		return c.server.handleAnalyzePatterns(ctx, nil, params)
	case "ai_insights":
		return c.server.handleAIInsights(ctx, nil, params)
	case "intelligent_analysis":
		return c.server.handleIntelligentAnalysis(ctx, nil, params)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// handleDropSummaryCommand processes the dropsummary command
func (c *MCPClient) handleDropSummaryCommand(ctx context.Context, args []string) error {
	arguments := c.parseArguments(args)

	params := &mcp.CallToolParamsFor[map[string]any]{
		Arguments: arguments,
	}

	result, err := c.server.handleGetPacketDropSummary(ctx, nil, params)
	if err != nil {
		return fmt.Errorf("packet drop summary failed: %v", err)
	}

	for _, content := range result.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			fmt.Println(textContent.Text)
		}
	}

	return nil
}

// handleDropListCommand processes the droplist command
func (c *MCPClient) handleDropListCommand(ctx context.Context, args []string) error {
	arguments := c.parseArguments(args)

	params := &mcp.CallToolParamsFor[map[string]any]{
		Arguments: arguments,
	}

	result, err := c.server.handleListPacketDrops(ctx, nil, params)
	if err != nil {
		return fmt.Errorf("packet drop list failed: %v", err)
	}

	for _, content := range result.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			fmt.Println(textContent.Text)
		}
	}

	return nil
}
