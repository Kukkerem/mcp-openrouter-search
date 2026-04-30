package openrouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SearchParameters struct {
	Query             string   `json:"-"`
	Engine            string   `json:"engine"`
	MaxResults        int      `json:"max_results"`
	SearchContextSize string   `json:"search_context_size"`
	MaxTotalResults   *int     `json:"max_total_results,omitempty"`
	AllowedDomains    []string `json:"allowed_domains,omitempty"`
	ExcludedDomains   []string `json:"excluded_domains,omitempty"`
}

type Message struct {
	Role        string `json:"role"`
	Content     any    `json:"content"`
	Annotations []any  `json:"annotations,omitempty"`
}

type Tool struct {
	Type       string           `json:"type"`
	Parameters SearchParameters `json:"parameters"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage Usage `json:"usage"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	ServerToolUse    *struct {
		WebSearchRequests int `json:"web_search_requests"`
	} `json:"server_tool_use,omitempty"`
}

type Citation struct {
	URL     string `json:"url"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

type SearchOutput struct {
	Answer            string     `json:"answer"`
	Citations         []Citation `json:"citations"`
	Model             string     `json:"model"`
	WebSearchRequests int        `json:"web_search_requests"`
	Usage             Usage      `json:"usage"`
	Warning           string     `json:"warning,omitempty"`
}

func DoSearch(endpoint, apiKey, model string, params SearchParameters, timeoutMs int) (*ChatResponse, error) {
	reqBody := ChatRequest{
		Model: model,
		Messages: []Message{
			{Role: "system", Content: "Use the OpenRouter web search server tool before answering. Return a concise, source-grounded answer and include citations or source URLs when available."},
			{Role: "user", Content: params.Query},
		},
		Tools: []Tool{
			{Type: "openrouter:web_search", Parameters: params},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	client := &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond}
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/Kukkerem/mcp-openrouter-search")
	req.Header.Set("X-Title", "mcp-openrouter-search")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var bodyText []byte
		bodyText, _ = io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenRouter API error %d: %s", resp.StatusCode, string(bodyText))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &chatResp, nil
}
