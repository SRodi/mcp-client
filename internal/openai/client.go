package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type ChatRequest struct {
	Model       string              `json:"model"`
	Messages    []ChatMessage       `json:"messages"`
	Functions   []FunctionDefinition `json:"functions,omitempty"`
	FunctionCall interface{}          `json:"function_call,omitempty"`
	Tools       []Tool              `json:"tools,omitempty"`
	ToolChoice  interface{}         `json:"tool_choice,omitempty"`
}

type Tool struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

type ChatMessage struct {
	Role         string     `json:"role"`
	Content      *string    `json:"content,omitempty"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID   string     `json:"tool_call_id,omitempty"`
	FunctionCall *struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function_call,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func AskLLM(summary string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}

	prompt := CreateNetworkInsightsPrompt(summary)
	systemContent := "You are a network connectivity analyst focused on providing actionable insights about connection patterns and application network behavior."

	reqBody := ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ChatMessage{
			{Role: "system", Content: &systemContent},
			{Role: "user", Content: &prompt},
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	if result.Choices[0].Message.Content == nil {
		return "", fmt.Errorf("empty response from OpenAI")
	}

	return *result.Choices[0].Message.Content, nil
}

// ConversationManager manages a conversation with function calling capabilities
type ConversationManager struct {
	functionManager *FunctionCallManager
	messages        []ChatMessage
	model           string
}

// NewConversationManager creates a new conversation manager with function calling
func NewConversationManager(mcpExecutor MCPToolExecutor) *ConversationManager {
	return &ConversationManager{
		functionManager: NewFunctionCallManager(mcpExecutor),
		messages:        make([]ChatMessage, 0),
		model:           "gpt-4o-mini", // Use a more capable model for function calling
	}
}

// SetModel sets the OpenAI model to use
func (cm *ConversationManager) SetModel(model string) {
	cm.model = model
}

// AddSystemMessage adds a system message to the conversation
func (cm *ConversationManager) AddSystemMessage(content string) {
	cm.messages = append(cm.messages, ChatMessage{
		Role:    "system",
		Content: &content,
	})
}

// AddUserMessage adds a user message to the conversation
func (cm *ConversationManager) AddUserMessage(content string) {
	cm.messages = append(cm.messages, ChatMessage{
		Role:    "user",
		Content: &content,
	})
}

// ProcessMessage processes a user message and handles any function calls
func (cm *ConversationManager) ProcessMessage(ctx context.Context, userMessage string) (string, error) {
	// Add user message
	cm.AddUserMessage(userMessage)

	// Make the initial request to OpenAI with function calling capabilities
	response, err := cm.sendChatRequest(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to send chat request: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	choice := response.Choices[0]
	
	// Add assistant's response to conversation
	cm.messages = append(cm.messages, choice.Message)

	// If the assistant wants to call functions, execute them
	if len(choice.Message.ToolCalls) > 0 {
		return cm.handleFunctionCalls(ctx, choice.Message.ToolCalls)
	}

	// If it's a direct response, return the content
	if choice.Message.Content != nil {
		return *choice.Message.Content, nil
	}

	return "", fmt.Errorf("no content in response")
}

// sendChatRequest sends a chat completion request to OpenAI
func (cm *ConversationManager) sendChatRequest(ctx context.Context) (*ChatResponse, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	// Convert function definitions to tools format
	tools := make([]Tool, 0, len(cm.functionManager.GetFunctions()))
	for _, fn := range cm.functionManager.GetFunctions() {
		tools = append(tools, Tool{
			Type:     "function",
			Function: fn,
		})
	}

	reqBody := ChatRequest{
		Model:    cm.model,
		Messages: cm.messages,
		Tools:    tools,
		ToolChoice: "auto", // Let the model decide when to call functions
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", result.Error.Message)
	}

	return &result, nil
}

// handleFunctionCalls executes function calls and continues the conversation
func (cm *ConversationManager) handleFunctionCalls(ctx context.Context, toolCalls []ToolCall) (string, error) {
	// Execute all function calls
	results, err := cm.functionManager.ExecuteFunctions(ctx, toolCalls)
	if err != nil {
		return "", fmt.Errorf("failed to execute functions: %v", err)
	}

	// Add function results to conversation
	for _, result := range results {
		cm.messages = append(cm.messages, ChatMessage{
			Role:       result.Role,
			Content:    &result.Content,
			ToolCallID: result.ToolCallID,
		})
	}

	// Send another request to get the final response
	response, err := cm.sendChatRequest(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get final response: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no final response from OpenAI")
	}

	choice := response.Choices[0]
	
	// Add the final response to conversation
	cm.messages = append(cm.messages, choice.Message)

	// Check if there are more function calls (recursive case)
	if len(choice.Message.ToolCalls) > 0 {
		return cm.handleFunctionCalls(ctx, choice.Message.ToolCalls)
	}

	// Return the final content
	if choice.Message.Content != nil {
		return *choice.Message.Content, nil
	}

	return "", fmt.Errorf("no content in final response")
}

// GetConversationHistory returns the current conversation history
func (cm *ConversationManager) GetConversationHistory() []ChatMessage {
	return cm.messages
}

// ClearConversation clears the conversation history
func (cm *ConversationManager) ClearConversation() {
	cm.messages = make([]ChatMessage, 0)
}
