package openai

import (
	"context"
	"fmt"
	"strings"
)

// ContextualNetworkAnalyst provides AI-powered network analysis with tool integration
type ContextualNetworkAnalyst struct {
	conversationManager *ConversationManager
	mcpExecutor         MCPToolExecutor
}

// NewContextualNetworkAnalyst creates a new contextual network analyst
func NewContextualNetworkAnalyst(mcpExecutor MCPToolExecutor, verbose bool) *ContextualNetworkAnalyst {
	analyst := &ContextualNetworkAnalyst{
		conversationManager: NewConversationManager(mcpExecutor, verbose),
		mcpExecutor:         mcpExecutor,
	}

	// Set up the system prompt with tool context
	analyst.setupSystemContext()
	return analyst
}

// setupSystemContext configures the system context for network analysis
func (cna *ContextualNetworkAnalyst) setupSystemContext() {
	systemPrompt := `You are an expert network connectivity analyst with access to real-time network telemetry tools. Your role is to:

1. **Analyze Network Behavior**: Examine connection patterns, frequencies, and destinations
2. **Identify Issues**: Detect anomalies, packet drops, and connectivity problems  
3. **Provide Insights**: Offer actionable recommendations for optimization and monitoring
4. **Use Tools Comprehensively**: ALWAYS use multiple tools to gather comprehensive data

## Available Tools - YOU MUST USE MULTIPLE TOOLS FOR COMPLETE ANALYSIS:
- **get_network_summary**: Get aggregated connection statistics for processes
- **list_connections**: View detailed connection events 
- **get_packet_drop_summary**: Analyze packet loss patterns
- **list_packet_drops**: See specific drop events (if drops are found)
- **analyze_patterns**: Get automated pattern analysis

## CRITICAL: Tool Usage Requirements:
1. **ALWAYS start with get_network_summary** to get overall network health
2. **ALWAYS use list_connections** to see detailed connection data
3. **ALWAYS check get_packet_drop_summary** for packet loss patterns
4. **ALWAYS use list_packet_drops** for detailed drop events (call this even if summary shows 0 drops)
5. **ALWAYS use analyze_patterns** for behavioral insights
6. **NEVER use only one tool** - minimum 3 tools required for any analysis
7. **NEVER skip list_packet_drops** - always call it regardless of summary results

## Special Instructions:
- When user asks for "all available tools" or "comprehensive analysis", use ALL 5 tools
- When user asks for "summary", use at least: get_network_summary, list_connections, get_packet_drop_summary, analyze_patterns
- NEVER skip tools even if you think you have enough data
- Each tool provides unique insights that others cannot provide

IMPORTANT: For ANY network analysis query, you MUST use multiple tools (minimum 3-4) to provide complete insights. Don't rely on just one or two tools - use the full toolkit for thorough analysis.

When a user asks about network behavior, automatically gather data from multiple sources before providing analysis. Always be thorough and comprehensive.`

	cna.conversationManager.AddSystemMessage(systemPrompt)
}

// AnalyzeNetworkQuery processes a network analysis query with contextual tool usage
func (cna *ContextualNetworkAnalyst) AnalyzeNetworkQuery(ctx context.Context, query string) (string, error) {
	// Enhance the query with context about what the user might want
	enhancedQuery := cna.enhanceUserQuery(query)

	// Process the message with function calling capabilities
	response, err := cna.conversationManager.ProcessMessage(ctx, enhancedQuery)
	if err != nil {
		return "", fmt.Errorf("failed to analyze network query: %v", err)
	}

	return response, nil
}

// enhanceUserQuery adds context to help the AI understand what tools to use
func (cna *ContextualNetworkAnalyst) enhanceUserQuery(query string) string {
	queryLower := strings.ToLower(query)

	// Add helpful context based on query content
	if strings.Contains(queryLower, "all available tools") || strings.Contains(queryLower, "use all tools") {
		return query + "\n\nIMPORTANT: You MUST use ALL 5 core analysis tools for comprehensive analysis. Call these tools in this EXACT order: 1) get_network_summary with duration=300, 2) list_connections, 3) get_packet_drop_summary with duration=300, 4) list_packet_drops (MANDATORY even if no drops found), 5) analyze_patterns with duration=300. Do not skip ANY tools. Each tool provides unique data."
	}

	if strings.Contains(queryLower, "comprehensive") || strings.Contains(queryLower, "complete analysis") {
		return query + "\n\nPlease use multiple tools (at least get_network_summary, list_connections, get_packet_drop_summary, and analyze_patterns) to provide comprehensive data."
	}

	if strings.Contains(queryLower, "summary") || strings.Contains(queryLower, "overview") {
		return query + "\n\nPlease use multiple network analysis tools including get_network_summary, list_connections, and get_packet_drop_summary to provide comprehensive data."
	}

	if strings.Contains(queryLower, "drop") || strings.Contains(queryLower, "loss") || strings.Contains(queryLower, "packet") {
		return query + "\n\nPlease check for packet drops and analyze any connectivity issues using get_packet_drop_summary and list_packet_drops tools."
	}

	if strings.Contains(queryLower, "pattern") || strings.Contains(queryLower, "behavior") {
		return query + "\n\nPlease analyze connection patterns using get_network_summary, list_connections, and analyze_patterns tools."
	}

	if strings.Contains(queryLower, "connection") || strings.Contains(queryLower, "network") {
		return query + "\n\nPlease gather network connection data using get_network_summary and list_connections tools, then analyze the results."
	}

	// Default enhancement
	return query + "\n\nPlease use appropriate network analysis tools (at least 2-3 different tools) to gather relevant data before providing insights."
}

// AnalyzeProcess provides focused analysis for a specific process
func (cna *ContextualNetworkAnalyst) AnalyzeProcess(ctx context.Context, processName string, pid int, duration int) (string, error) {
	var query string

	if processName != "" {
		query = fmt.Sprintf("Please analyze the network behavior of process '%s' over the last %d seconds. I want to understand its connection patterns, any issues, and optimization opportunities.", processName, duration)
	} else if pid > 0 {
		query = fmt.Sprintf("Please analyze the network behavior of process ID %d over the last %d seconds. I want to understand its connection patterns, any issues, and optimization opportunities.", pid, duration)
	} else {
		query = fmt.Sprintf("Please analyze overall network activity over the last %d seconds. Show me connection patterns, any issues, and recommendations.", duration)
	}

	return cna.AnalyzeNetworkQuery(ctx, query)
}

// GetNetworkHealth provides a comprehensive network health assessment
func (cna *ContextualNetworkAnalyst) GetNetworkHealth(ctx context.Context, duration int) (string, error) {
	query := fmt.Sprintf(`Please provide a comprehensive network health assessment over the last %d seconds. Include:

1. Connection summary and patterns
2. Any packet drops or connectivity issues  
3. Overall network performance indicators
4. Specific recommendations for improvement
5. Any security concerns or anomalies

Use all relevant tools to gather complete data for this analysis.`, duration)

	return cna.AnalyzeNetworkQuery(ctx, query)
}

// GetComprehensiveAnalysis provides analysis using ALL available tools
func (cna *ContextualNetworkAnalyst) GetComprehensiveAnalysis(ctx context.Context, duration int) (string, error) {
	query := fmt.Sprintf(`COMPREHENSIVE ANALYSIS REQUEST: Analyze network activity over the last %d seconds using ALL available tools.

REQUIRED: You MUST call these tools in this exact order:
1. get_network_summary (for overall connection statistics)
2. list_connections (for detailed connection events)  
3. get_packet_drop_summary (for packet loss analysis)
4. list_packet_drops (for detailed drop information if any drops found)
5. analyze_patterns (for behavioral pattern analysis)

Do NOT skip any of these tools. Each provides unique insights needed for complete analysis.

After gathering all data, provide:
- Overall network health assessment
- Connection pattern analysis
- Performance issues and recommendations
- Security observations
- Optimization suggestions`, duration)

	return cna.AnalyzeNetworkQuery(ctx, query)
}

// StartNewConversation clears the conversation history and starts fresh
func (cna *ContextualNetworkAnalyst) StartNewConversation() {
	cna.conversationManager.ClearConversation()
	cna.setupSystemContext()
}

// GetConversationHistory returns the current conversation
func (cna *ContextualNetworkAnalyst) GetConversationHistory() []ChatMessage {
	return cna.conversationManager.GetConversationHistory()
}

// ContinueConversation continues an existing conversation
func (cna *ContextualNetworkAnalyst) ContinueConversation(ctx context.Context, message string) (string, error) {
	return cna.conversationManager.ProcessMessage(ctx, message)
}
