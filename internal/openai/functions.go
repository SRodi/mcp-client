package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// FunctionDefinition represents an OpenAI function definition
type FunctionDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// ToolCall represents a function call from OpenAI
type ToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Function FunctionDetails `json:"function"`
}

// FunctionDetails contains the function name and arguments
type FunctionDetails struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolCallResult represents the result of a tool execution
type ToolCallResult struct {
	ToolCallID string `json:"tool_call_id"`
	Role       string `json:"role"`
	Content    string `json:"content"`
}

// MCPToolExecutor interface for executing MCP tools
type MCPToolExecutor interface {
	RunSingleCommand(ctx context.Context, toolName string, arguments map[string]any) (*mcp.CallToolResult, error)
}

// MCPToolDiscovery interface for discovering available MCP tools
type MCPToolDiscovery interface {
	GetRegisteredTools() map[string]*mcp.Tool
}

// FunctionCallManager manages OpenAI function calling integration with MCP tools
type FunctionCallManager struct {
	mcpExecutor MCPToolExecutor
	functions   []FunctionDefinition
}

// NewFunctionCallManager creates a new function call manager with automatic tool discovery
func NewFunctionCallManager(mcpExecutor MCPToolExecutor) *FunctionCallManager {
	fm := &FunctionCallManager{
		mcpExecutor: mcpExecutor,
		functions:   make([]FunctionDefinition, 0),
	}
	
	// If the executor also implements tool discovery, auto-discover tools
	if toolDiscovery, ok := mcpExecutor.(MCPToolDiscovery); ok {
		fm.discoverMCPTools(toolDiscovery)
	} else {
		// Fallback: register known tools manually
		fm.registerKnownMCPTools()
	}
	
	return fm
}

// discoverMCPTools automatically discovers MCP tools and converts them to OpenAI functions
func (fm *FunctionCallManager) discoverMCPTools(discovery MCPToolDiscovery) {
	tools := discovery.GetRegisteredTools()
	
	for toolName, tool := range tools {
		// Convert MCP tool to OpenAI function definition
		functionDef := fm.convertMCPToolToFunction(toolName, tool)
		fm.functions = append(fm.functions, functionDef)
	}
}

// convertMCPToolToFunction converts an MCP tool definition to OpenAI function format
func (fm *FunctionCallManager) convertMCPToolToFunction(toolName string, tool *mcp.Tool) FunctionDefinition {
	// Convert JSON schema to OpenAI parameters format
	parameters := fm.convertJSONSchemaToOpenAI(tool.InputSchema)
	
	return FunctionDefinition{
		Name:        toolName,
		Description: tool.Description,
		Parameters:  parameters,
	}
}

// convertJSONSchemaToOpenAI converts a JSON schema to OpenAI function parameters format
func (fm *FunctionCallManager) convertJSONSchemaToOpenAI(schema *jsonschema.Schema) map[string]interface{} {
	if schema == nil {
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}

	result := map[string]interface{}{
		"type": schema.Type,
	}

	if schema.Properties != nil {
		properties := make(map[string]interface{})
		for propName, propSchema := range schema.Properties {
			properties[propName] = fm.convertPropertySchema(propSchema)
		}
		result["properties"] = properties
	}

	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	return result
}

// convertPropertySchema converts a property schema to OpenAI format
func (fm *FunctionCallManager) convertPropertySchema(schema *jsonschema.Schema) map[string]interface{} {
	property := map[string]interface{}{
		"type": schema.Type,
	}

	if schema.Description != "" {
		property["description"] = schema.Description
	}

	if schema.Default != nil {
		// Parse the default value if it's JSON bytes
		var defaultValue interface{}
		if err := json.Unmarshal(schema.Default, &defaultValue); err == nil {
			property["default"] = defaultValue
		}
	}

	return property
}

// registerKnownMCPTools provides fallback registration for known tools
func (fm *FunctionCallManager) registerKnownMCPTools() {
	// This is a fallback method - only used if auto-discovery fails
	// The actual tools will be discovered automatically from the MCP server
	fmt.Println("Warning: Using fallback tool registration. MCP tool auto-discovery not available.")
}

// GetFunctions returns all registered function definitions
func (fm *FunctionCallManager) GetFunctions() []FunctionDefinition {
	return fm.functions
}

// ExecuteFunction executes a function call and returns the result
func (fm *FunctionCallManager) ExecuteFunction(ctx context.Context, functionCall ToolCall) (*ToolCallResult, error) {
	// Parse function arguments
	var arguments map[string]any
	if err := json.Unmarshal([]byte(functionCall.Function.Arguments), &arguments); err != nil {
		return nil, fmt.Errorf("failed to parse function arguments: %v", err)
	}

	// Validate function exists
	functionExists := false
	for _, fn := range fm.functions {
		if fn.Name == functionCall.Function.Name {
			functionExists = true
			break
		}
	}
	if !functionExists {
		return nil, fmt.Errorf("unknown function: %s", functionCall.Function.Name)
	}

	// Execute the MCP tool
	result, err := fm.mcpExecutor.RunSingleCommand(ctx, functionCall.Function.Name, arguments)
	if err != nil {
		return &ToolCallResult{
			ToolCallID: functionCall.ID,
			Role:       "tool",
			Content:    fmt.Sprintf("Error executing %s: %v", functionCall.Function.Name, err),
		}, nil
	}

	// Extract content from MCP result
	content := ""
	for _, c := range result.Content {
		if textContent, ok := c.(*mcp.TextContent); ok {
			content += textContent.Text + "\n"
		}
	}

	return &ToolCallResult{
		ToolCallID: functionCall.ID,
		Role:       "tool",
		Content:    content,
	}, nil
}

// ExecuteFunctions executes multiple function calls and returns their results
func (fm *FunctionCallManager) ExecuteFunctions(ctx context.Context, functionCalls []ToolCall) ([]ToolCallResult, error) {
	results := make([]ToolCallResult, 0, len(functionCalls))
	
	for _, call := range functionCalls {
		result, err := fm.ExecuteFunction(ctx, call)
		if err != nil {
			// Return error result for this function call
			results = append(results, ToolCallResult{
				ToolCallID: call.ID,
				Role:       "tool",
				Content:    fmt.Sprintf("Function execution failed: %v", err),
			})
		} else {
			results = append(results, *result)
		}
	}
	
	return results, nil
}
