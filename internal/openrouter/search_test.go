package openrouter

import (
	"encoding/json"
	"testing"
)

func TestExtractMessageText_String(t *testing.T) {
	got := ExtractMessageText("hello")
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestExtractMessageText_Array(t *testing.T) {
	content := []any{"hello ", map[string]any{"text": "world"}}
	got := ExtractMessageText(content)
	if got != "hello \nworld" {
		t.Errorf("got %q", got)
	}
}

func TestExtractMessageText_Nil(t *testing.T) {
	got := ExtractMessageText(nil)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestBuildOutput(t *testing.T) {
	resp := &ChatResponse{
		Model: "test-model",
		Choices: []struct {
			Message Message `json:"message"`
		}{
			{Message: Message{
				Role:    "assistant",
				Content: "The answer is 42",
				Annotations: []any{
					map[string]any{
						"url_citation": map[string]any{
							"url":   "https://example.com",
							"title": "Example",
						},
					},
				},
			}},
		},
		Usage: Usage{
			ServerToolUse: &struct {
				WebSearchRequests int `json:"web_search_requests"`
			}{WebSearchRequests: 2},
		},
	}

	output := BuildOutput(resp, SearchParameters{})

	if output.Answer != "The answer is 42" {
		t.Errorf("answer = %q", output.Answer)
	}
	if output.WebSearchRequests != 2 {
		t.Errorf("web_search_requests = %d", output.WebSearchRequests)
	}
	if len(output.Citations) != 1 {
		t.Fatalf("citations count = %d", len(output.Citations))
	}
	if output.Citations[0].URL != "https://example.com" {
		t.Errorf("citation url = %q", output.Citations[0].URL)
	}
	if output.Citations[0].Title != "Example" {
		t.Errorf("citation title = %q", output.Citations[0].Title)
	}
	if output.Warning != "" {
		t.Errorf("warning = %q, want empty", output.Warning)
	}
}

func TestBuildOutput_ZeroSearchRequests_NoCitations(t *testing.T) {
	resp := &ChatResponse{
		Model: "test",
		Choices: []struct {
			Message Message `json:"message"`
		}{
			{Message: Message{Role: "assistant", Content: "no search"}},
		},
	}

	output := BuildOutput(resp, SearchParameters{})
	if output.Warning == "" {
		t.Error("expected warning for zero search requests with no citations")
	}
}

func TestBuildOutput_ZeroSearchRequests_WithCitations(t *testing.T) {
	resp := &ChatResponse{
		Model: "test",
		Choices: []struct {
			Message Message `json:"message"`
		}{
			{Message: Message{
				Role:    "assistant",
				Content: "has citations",
				Annotations: []any{
					map[string]any{
						"url_citation": map[string]any{
							"url":   "https://example.com",
							"title": "Example",
						},
					},
				},
			}},
		},
	}

	output := BuildOutput(resp, SearchParameters{})
	if output.Warning != "" {
		t.Errorf("expected no warning when citations present, got: %q", output.Warning)
	}
}

func TestFormatText(t *testing.T) {
	output := &SearchOutput{
		Answer:            "Test answer",
		WebSearchRequests: 1,
		Citations: []Citation{
			{URL: "https://example.com", Title: "Example"},
		},
	}

	text := FormatText(output)

	if !contains(text, "Test answer") {
		t.Error("missing answer in text output")
	}
	if !contains(text, "Sources:") {
		t.Error("missing Sources header in text output")
	}
	if !contains(text, "Example - https://example.com") {
		t.Error("missing citation in text output")
	}
	if !contains(text, "[web_search_requests] 1") {
		t.Error("missing web_search_requests in text output")
	}
}

func TestFormatJSON(t *testing.T) {
	output := &SearchOutput{
		Answer:            "Test",
		WebSearchRequests: 1,
	}

	text, err := FormatJSON(output)
	if err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}

	var parsed SearchOutput
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed.Answer != "Test" {
		t.Errorf("answer = %q", parsed.Answer)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
