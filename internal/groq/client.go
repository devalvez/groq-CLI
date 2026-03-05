package groq

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"groq-cli/internal/config"
)

const (
	BaseURL        = "https://api.groq.com/openai/v1"
	defaultTimeout = 120 * time.Second
)

// Model constants
const (
	ModelLlama3_70B     = "llama3-70b-8192"
	ModelLlama3_8B      = "llama3-8b-8192"
	ModelMixtral8x7B    = "mixtral-8x7b-32768"
	ModelGemma2_9B      = "gemma2-9b-it"
	ModelLlama31_70B    = "llama-3.1-70b-versatile"
	ModelLlama31_8B     = "llama-3.1-8b-instant"
	ModelLlama33_70B    = "llama-3.3-70b-versatile"
	ModelDeepSeek       = "deepseek-r1-distill-llama-70b"
)

// Client is the Groq API client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the request body for chat completions
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatResponse is the response from chat completions
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
	Error *APIError `json:"error,omitempty"`
}

// StreamChunk is a chunk from a streaming response
type StreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// APIError represents an API error
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error [%s]: %s", e.Type, e.Message)
}

// ModelsResponse lists available models
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
	Error  *APIError `json:"error,omitempty"`
}

// Model represents a Groq model
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// NewClient creates a new Groq API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// DefaultModel returns the configured default model
func DefaultModel() string {
	cfg := config.Get()
	if cfg.DefaultModel != "" {
		return cfg.DefaultModel
	}
	return ModelLlama33_70B
}

// Chat sends a chat completion request
func (c *Client) Chat(req ChatRequest) (*ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", BaseURL+"/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if chatResp.Error != nil {
		return nil, chatResp.Error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return &chatResp, nil
}

// ChatStream sends a streaming chat completion request
// onChunk is called for each content chunk received
func (c *Client) ChatStream(req ChatRequest, onChunk func(string)) error {
	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", BaseURL+"/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	c.setHeaders(httpReq)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error *APIError `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != nil {
			return errResp.Error
		}
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				onChunk(choice.Delta.Content)
			}
		}
	}

	return scanner.Err()
}

// ListModels fetches available models
func (c *Client) ListModels() ([]Model, error) {
	httpReq, err := http.NewRequest("GET", BaseURL+"/models", nil)
	if err != nil {
		return nil, err
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var modelsResp ModelsResponse
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &modelsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if modelsResp.Error != nil {
		return nil, modelsResp.Error
	}

	return modelsResp.Data, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
}
