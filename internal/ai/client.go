package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"kubeguide/internal/config"
)

type Client struct {
	config     *config.AIConfig
	httpClient *http.Client
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func NewClient(cfg *config.AIConfig) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) AnalyzePod(ctx context.Context, podYAML string) (string, error) {
	if c.config.APIKey == "" {
		return "", fmt.Errorf("AI API key is not configured. Please set KUBEGUIDE_AI_API_KEY environment variable or configure it in ~/.config/kubeguide/config.yaml")
	}

	systemPrompt := `You are a Kubernetes expert assistant. Analyze the provided pod YAML and identify issues that might be causing failures.

Focus on:
1. Resource constraints (CPU/memory limits and requests)
2. Image pull issues 
3. Configuration problems (environment variables, secrets, configmaps)
4. Health check configurations
5. Security context issues
6. Volume mount problems
7. Common misconfigurations

Provide a concise analysis with:
- Root cause identification
- Specific recommendations to fix issues
- Best practices suggestions

Keep the response focused and actionable.`

	userPrompt := fmt.Sprintf("Please analyze this Kubernetes pod YAML and help identify why it might be failing:\n\n```yaml\n%s\n```", podYAML)

	if c.config.Provider == "anthropic" {
		return c.sendAnthropicRequest(ctx, systemPrompt, userPrompt)
	}

	messages := []ChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userPrompt,
		},
	}

	req := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		Temperature: 0.1, // Low temperature for focused, consistent responses
		MaxTokens:   1000,
	}

	return c.sendRequest(ctx, req)
}

func (c *Client) sendRequest(ctx context.Context, req ChatRequest) (string, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.config.BaseURL + "/chat/completions"
	
	// Handle Anthropic API which uses a different endpoint
	if c.config.Provider == "anthropic" {
		endpoint = c.config.BaseURL + "/v1/messages"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	
	// Set appropriate headers based on provider
	if c.config.Provider == "anthropic" {
		httpReq.Header.Set("x-api-key", c.config.APIKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
	} else {
		httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d to %s (provider: %s): %s", resp.StatusCode, endpoint, c.config.Provider, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return chatResp.Choices[0].Message.Content, nil
}

type AnthropicRequest struct {
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
	Messages  []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	System string `json:"system,omitempty"`
}

type AnthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) sendAnthropicRequest(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	req := AnthropicRequest{
		Model:     c.config.Model,
		MaxTokens: 1000,
		System:    systemPrompt,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.config.BaseURL + "/v1/messages"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Anthropic API request failed with status %d to %s: %s", resp.StatusCode, endpoint, string(body))
	}

	var anthResp AnthropicResponse
	if err := json.Unmarshal(body, &anthResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if anthResp.Error != nil {
		return "", fmt.Errorf("Anthropic API error: %s", anthResp.Error.Message)
	}

	if len(anthResp.Content) == 0 {
		return "", fmt.Errorf("no content returned from Anthropic API")
	}

	return anthResp.Content[0].Text, nil
}