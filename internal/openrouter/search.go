package openrouter

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ExtractMessageText(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var parts []string
		for _, part := range v {
			switch p := part.(type) {
			case string:
				parts = append(parts, p)
			case map[string]any:
				if t, ok := p["text"].(string); ok {
					parts = append(parts, t)
				} else if t, ok := p["content"].(string); ok {
					parts = append(parts, t)
				}
			}
		}
		return strings.Join(parts, "\n")
	default:
		return ""
	}
}

func ExtractCitations(msg Message) []Citation {
	var citations []Citation
	seen := make(map[string]bool)

	collectFromAnnotations := func(annotations []any) {
		for _, ann := range annotations {
			m, ok := ann.(map[string]any)
			if !ok {
				continue
			}
			cit := extractCitationFromMap(m)
			if cit != nil && !seen[cit.URL] {
				seen[cit.URL] = true
				citations = append(citations, *cit)
			}
		}
	}

	if msg.Annotations != nil {
		collectFromAnnotations(msg.Annotations)
	}

	switch v := msg.Content.(type) {
	case []any:
		for _, part := range v {
			if m, ok := part.(map[string]any); ok {
				if ann, ok := m["annotations"].([]any); ok {
					collectFromAnnotations(ann)
				}
			}
		}
	}

	return citations
}

func extractCitationFromMap(m map[string]any) *Citation {
	cit := m["url_citation"]
	if cit == nil {
		cit = m["urlCitation"]
	}
	if cit == nil {
		cit = m
	}

	c, ok := cit.(map[string]any)
	if !ok {
		return nil
	}

	url, _ := c["url"].(string)
	if url == "" {
		return nil
	}

	result := &Citation{URL: url}
	if title, ok := c["title"].(string); ok {
		result.Title = title
	}
	if content, ok := c["content"].(string); ok {
		result.Content = content
	}
	return result
}

func BuildOutput(resp *ChatResponse, params SearchParameters) *SearchOutput {
	var msg Message
	if len(resp.Choices) > 0 {
		msg = resp.Choices[0].Message
	}

	answer := ExtractMessageText(msg.Content)
	webSearchRequests := 0
	if resp.Usage.ServerToolUse != nil {
		webSearchRequests = resp.Usage.ServerToolUse.WebSearchRequests
	}

	output := &SearchOutput{
		Answer:            answer,
		Citations:         ExtractCitations(msg),
		Model:             resp.Model,
		WebSearchRequests: webSearchRequests,
		Usage:             resp.Usage,
	}

	if webSearchRequests == 0 {
		output.Warning = "OpenRouter reported zero web search requests."
	}

	return output
}

func FormatText(output *SearchOutput) string {
	var b strings.Builder

	if output.Answer == "" {
		b.WriteString("[OpenRouter web search returned no content]\n")
	} else {
		b.WriteString(output.Answer)
		b.WriteString("\n")
	}

	if output.Warning != "" {
		b.WriteString("\n[warning] " + output.Warning + "\n")
	}

	if len(output.Citations) > 0 {
		b.WriteString("\nSources:\n")
		for _, cit := range output.Citations {
			if cit.Title != "" {
				fmt.Fprintf(&b, "- %s - %s\n", cit.Title, cit.URL)
			} else {
				fmt.Fprintf(&b, "- %s\n", cit.URL)
			}
		}
	}

	fmt.Fprintf(&b, "\n[web_search_requests] %d\n", output.WebSearchRequests)

	return b.String()
}

func FormatJSON(output *SearchOutput) (string, error) {
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal output: %w", err)
	}
	return string(data), nil
}
