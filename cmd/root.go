package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kukkerem/mcp-openrouter-search/internal/config"
	"github.com/Kukkerem/mcp-openrouter-search/internal/openrouter"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

type SearchInput struct {
	Query             string `json:"query" jsonschema:"search question or research task"`
	Engine            string `json:"engine,omitempty" jsonschema:"search engine: auto, native, exa, firecrawl, parallel. Default: auto"`
	MaxResults        *int   `json:"max_results,omitempty" jsonschema:"results per search call, 1-25. Default: 5"`
	MaxTotalResults   *int   `json:"max_total_results,omitempty" jsonschema:"cap total results across multi-search loops"`
	SearchContextSize string `json:"search_context_size,omitempty" jsonschema:"search context size: low, medium, high. Default: medium"`
	AllowedDomains    string `json:"allowed_domains,omitempty" jsonschema:"comma-separated domains to restrict search to"`
	ExcludedDomains   string `json:"excluded_domains,omitempty" jsonschema:"comma-separated domains to exclude"`
}

type SearchOutput struct {
	Answer            string                `json:"answer" jsonschema:"the search answer text"`
	Citations         []openrouter.Citation `json:"citations" jsonschema:"source citations from the search"`
	Model             string                `json:"model" jsonschema:"the model used for the search"`
	WebSearchRequests int                   `json:"web_search_requests" jsonschema:"number of web search requests made"`
}

var Version string

var rootCmd = &cobra.Command{
	Use:   "mcp-openrouter-search",
	Short: "MCP server for web search via OpenRouter API",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

func Execute() error {
	rootCmd.Version = Version
	return rootCmd.Execute()
}

func runServer() error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-openrouter-search",
		Version: Version,
	}, &mcp.ServerOptions{
		Instructions: "Web search via OpenRouter API. Use search_web for current information beyond training data. Returns answers with source citations. Prefer over static knowledge for time-sensitive queries. Use allowed_domains/excluded_domains to scope searches. Use search_context_size=low for cheaper queries or high for thorough research.",
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_web",
		Description: "Search the web using OpenRouter's server-side web search tool. Use this when you need current information through OpenRouter quota.",
	}, handleSearch)

	return server.Run(context.Background(), &mcp.StdioTransport{})
}

func handleSearch(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, SearchOutput, error) {
	apiKey, err := config.ResolveAPIKey("")
	if err != nil {
		return nil, SearchOutput{}, fmt.Errorf("API key error: %w", err)
	}

	engine := input.Engine
	if engine == "" {
		engine = config.DefaultEngine
	}

	maxResults := config.DefaultMaxResults
	if input.MaxResults != nil {
		maxResults = *input.MaxResults
	}

	contextSize := input.SearchContextSize
	if contextSize == "" {
		contextSize = config.DefaultContextSize
	}

	params := openrouter.SearchParameters{
		Query:             input.Query,
		Engine:            engine,
		MaxResults:        maxResults,
		SearchContextSize: contextSize,
		MaxTotalResults:   input.MaxTotalResults,
	}

	if input.AllowedDomains != "" {
		params.AllowedDomains = splitByComma(input.AllowedDomains)
	}
	if input.ExcludedDomains != "" {
		params.ExcludedDomains = splitByComma(input.ExcludedDomains)
	}

	resp, err := openrouter.DoSearch(config.OpenRouterEndpoint, apiKey, config.DefaultModel, params, config.DefaultTimeoutMs)
	if err != nil {
		return nil, SearchOutput{}, err
	}

	output := openrouter.BuildOutput(resp, params)
	text := openrouter.FormatText(output)

	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}

	return result, SearchOutput{
		Answer:            output.Answer,
		Citations:         output.Citations,
		Model:             output.Model,
		WebSearchRequests: output.WebSearchRequests,
	}, nil
}

func splitByComma(s string) []string {
	parts := make([]string, 0)
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}
