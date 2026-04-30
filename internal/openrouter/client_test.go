package openrouter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoSearch(t *testing.T) {
	response := ChatResponse{
		ID:    "test-123",
		Model: "test-model",
		Choices: []struct {
			Message Message `json:"message"`
		}{
			{Message: Message{Role: "assistant", Content: "Test answer"}},
		},
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
			ServerToolUse: &struct {
				WebSearchRequests int `json:"web_search_requests"`
			}{WebSearchRequests: 1},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing or wrong Authorization header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing Content-Type header")
		}

		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Model != "test-model" {
			t.Errorf("model = %q, want %q", req.Model, "test-model")
		}
		if len(req.Tools) != 1 || req.Tools[0].Type != "openrouter:web_search" {
			t.Errorf("tools = %+v, want one openrouter:web_search tool", req.Tools)
		}
		if len(req.Messages) != 2 {
			t.Errorf("messages count = %d, want 2", len(req.Messages))
		}
		if req.Messages[1].Role != "user" || req.Messages[1].Content != "test query" {
			t.Errorf("user message = %+v", req.Messages[1])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	params := SearchParameters{
		Query:             "test query",
		Engine:            "auto",
		MaxResults:        5,
		SearchContextSize: "medium",
	}

	resp, err := DoSearch(server.URL, "test-key", "test-model", params, 5000)
	if err != nil {
		t.Fatalf("DoSearch: %v", err)
	}

	if resp.Model != "test-model" {
		t.Errorf("model = %q", resp.Model)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("choices count = %d", len(resp.Choices))
	}
	if resp.Usage.ServerToolUse == nil || resp.Usage.ServerToolUse.WebSearchRequests != 1 {
		t.Errorf("web_search_requests = %v", resp.Usage.ServerToolUse)
	}
}

func TestDoSearch_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "rate limited"}`))
	}))
	defer server.Close()

	params := SearchParameters{Query: "test", Engine: "auto", MaxResults: 5, SearchContextSize: "medium"}
	_, err := DoSearch(server.URL, "key", "model", params, 5000)
	if err == nil {
		t.Fatal("expected error for 429 response")
	}
}

func TestDoSearch_ToolParameters(t *testing.T) {
	var receivedReq ChatRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedReq)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{Model: "test"})
	}))
	defer server.Close()

	maxTotal := 20
	params := SearchParameters{
		Query:             "test",
		Engine:            "exa",
		MaxResults:        10,
		SearchContextSize: "high",
		MaxTotalResults:   &maxTotal,
		AllowedDomains:    []string{"example.com"},
		ExcludedDomains:   []string{"reddit.com"},
	}

	DoSearch(server.URL, "key", "model", params, 5000)

	tool := receivedReq.Tools[0]
	if tool.Type != "openrouter:web_search" {
		t.Errorf("tool type = %q", tool.Type)
	}
	if tool.Parameters.Engine != "exa" {
		t.Errorf("engine = %q", tool.Parameters.Engine)
	}
	if tool.Parameters.MaxResults != 10 {
		t.Errorf("max_results = %d", tool.Parameters.MaxResults)
	}
	if tool.Parameters.SearchContextSize != "high" {
		t.Errorf("context_size = %q", tool.Parameters.SearchContextSize)
	}
	if tool.Parameters.MaxTotalResults == nil || *tool.Parameters.MaxTotalResults != 20 {
		t.Errorf("max_total_results = %v", tool.Parameters.MaxTotalResults)
	}
	if len(tool.Parameters.AllowedDomains) != 1 || tool.Parameters.AllowedDomains[0] != "example.com" {
		t.Errorf("allowed_domains = %v", tool.Parameters.AllowedDomains)
	}
	if len(tool.Parameters.ExcludedDomains) != 1 || tool.Parameters.ExcludedDomains[0] != "reddit.com" {
		t.Errorf("excluded_domains = %v", tool.Parameters.ExcludedDomains)
	}
}
