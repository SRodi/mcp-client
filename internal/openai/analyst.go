package openai

import (
	"context"
	"fmt"
	"strings"
)

// IntelligentNetworkAnalyst provides AI-powered network analysis with tool integration
type IntelligentNetworkAnalyst struct {
	conversationManager *ConversationManager
	mcpExecutor         MCPToolExecutor
}

// NewIntelligentNetworkAnalyst creates a new intelligent network analyst
func NewIntelligentNetworkAnalyst(mcpExecutor MCPToolExecutor) *IntelligentNetworkAnalyst {
	analyst := &IntelligentNetworkAnalyst{
		conversationManager: NewConversationManager(mcpExecutor),
		mcpExecutor:         mcpExecutor,
	}

	// Set up the system prompt with tool context
	analyst.setupSystemContext()
	return analyst
}

// setupSystemContext configures the system context for network analysis
func (ina *IntelligentNetworkAnalyst) setupSystemContext() {
	systemPrompt := `You are an expert network connectivity analyst with access to real-time network telemetry tools. Your role is to:

1. **Analyze Network Behavior**: Examine connection patterns, frequencies, and destinations
2. **Identify Issues**: Detect anomalies, packet drops, and connectivity problems  
3. **Provide Insights**: Offer actionable recommendations for optimization and monitoring
4. **Use Tools Comprehensively**: ALWAYS use multiple tools to gather comprehensive data

## Available Tools - USE MULTIPLE TOOLS FOR COMPLETE ANALYSIS:
- **get_network_summary**: Get aggregated connection statistics for processes
- **list_connections**: View detailed connection events 
- **get_packet_drop_summary**: Analyze packet loss patterns
- **list_packet_drops**: See specific drop events
- **analyze_patterns**: Get automated pattern analysis

## Required Analysis Approach:
1. **ALWAYS start with get_network_summary** to get overall network health
2. **Use list_connections** to see detailed connection data
3. **Check for packet drops** with get_packet_drop_summary
4. **If drops exist, use list_packet_drops** for details
5. **Use analyze_patterns** for behavioral insights
6. **Provide comprehensive analysis** based on ALL gathered data

IMPORTANT: For ANY network analysis query, you MUST use multiple tools to provide complete insights. Don't rely on just one tool - use at least 3-4 tools for thorough analysis.

When a user asks about network behavior, automatically gather data from multiple sources before providing analysis. Always be thorough and comprehensive.`

	ina.conversationManager.AddSystemMessage(systemPrompt)
}

// AnalyzeNetworkQuery processes a network analysis query with intelligent tool usage
func (ina *IntelligentNetworkAnalyst) AnalyzeNetworkQuery(ctx context.Context, query string) (string, error) {
	// Enhance the query with context about what the user might want
	enhancedQuery := ina.enhanceUserQuery(query)
	
	// Process the message with function calling capabilities
	response, err := ina.conversationManager.ProcessMessage(ctx, enhancedQuery)
	if err != nil {
		return "", fmt.Errorf("failed to analyze network query: %v", err)
	}

	return response, nil
}

// enhanceUserQuery adds context to help the AI understand what tools to use
func (ina *IntelligentNetworkAnalyst) enhanceUserQuery(query string) string {
	queryLower := strings.ToLower(query)
	
	// Add helpful context based on query content
	if strings.Contains(queryLower, "summary") || strings.Contains(queryLower, "overview") {
		return query + "\n\nPlease use network summary tools to provide comprehensive data."
	}
	
	if strings.Contains(queryLower, "drop") || strings.Contains(queryLower, "loss") || strings.Contains(queryLower, "packet") {
		return query + "\n\nPlease check for packet drops and analyze any connectivity issues."
	}
	
	if strings.Contains(queryLower, "pattern") || strings.Contains(queryLower, "behavior") {
		return query + "\n\nPlease analyze connection patterns and provide behavioral insights."
	}
	
	if strings.Contains(queryLower, "connection") || strings.Contains(queryLower, "network") {
		return query + "\n\nPlease gather network connection data and analyze the results."
	}
	
	// Default enhancement
	return query + "\n\nPlease use appropriate network analysis tools to gather relevant data before providing insights."
}

// AnalyzeProcess provides focused analysis for a specific process
func (ina *IntelligentNetworkAnalyst) AnalyzeProcess(ctx context.Context, processName string, pid int, duration int) (string, error) {
	var query string
	
	if processName != "" {
		query = fmt.Sprintf("Please analyze the network behavior of process '%s' over the last %d seconds. I want to understand its connection patterns, any issues, and optimization opportunities.", processName, duration)
	} else if pid > 0 {
		query = fmt.Sprintf("Please analyze the network behavior of process ID %d over the last %d seconds. I want to understand its connection patterns, any issues, and optimization opportunities.", pid, duration)
	} else {
		query = fmt.Sprintf("Please analyze overall network activity over the last %d seconds. Show me connection patterns, any issues, and recommendations.", duration)
	}
	
	return ina.AnalyzeNetworkQuery(ctx, query)
}

// GetNetworkHealth provides a comprehensive network health assessment
func (ina *IntelligentNetworkAnalyst) GetNetworkHealth(ctx context.Context, duration int) (string, error) {
	query := fmt.Sprintf(`Please provide a comprehensive network health assessment over the last %d seconds. Include:

1. Connection summary and patterns
2. Any packet drops or connectivity issues  
3. Overall network performance indicators
4. Specific recommendations for improvement
5. Any security concerns or anomalies

Use all relevant tools to gather complete data for this analysis.`, duration)

	return ina.AnalyzeNetworkQuery(ctx, query)
}

// StartNewConversation clears the conversation history and starts fresh
func (ina *IntelligentNetworkAnalyst) StartNewConversation() {
	ina.conversationManager.ClearConversation()
	ina.setupSystemContext()
}

// GetConversationHistory returns the current conversation
func (ina *IntelligentNetworkAnalyst) GetConversationHistory() []ChatMessage {
	return ina.conversationManager.GetConversationHistory()
}

// ContinueConversation continues an existing conversation
func (ina *IntelligentNetworkAnalyst) ContinueConversation(ctx context.Context, message string) (string, error) {
	return ina.conversationManager.ProcessMessage(ctx, message)
}
