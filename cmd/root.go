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
	Query             string `json:"query" jsonschema:"required,description=Search question or research task"`
	Engine            string `json:"engine,omitempty" jsonschema:"description=Search engine: auto, native, exa, firecrawl, parallel. Default: auto"`
	MaxResults        *int   `json:"max_results,omitempty" jsonschema:"description=Results per search call, 1-25. Default: 5"`
	MaxTotalResults   *int   `json:"max_total_results,omitempty" jsonschema:"description=Cap total results across multi-search loops"`
	SearchContextSize string `json:"search_context_size,omitempty" jsonschema:"description=Search context size: low, medium, high. Default: medium"`
	AllowedDomains    string `json:"allowed_domains,omitempty" jsonschema:"description=Comma-separated domains to restrict search to"`
	ExcludedDomains   string `json:"excluded_domains,omitempty" jsonschema:"description=Comma-separated domains to exclude"`
}

type SearchOutput struct {
	Answer            string                `json:"answer"`
	Citations         []openrouter.Citation `json:"citations"`
	Model             string                `json:"model"`
	WebSearchRequests int                   `json:"web_search_requests"`
}

var rootCmd = &cobra.Command{
	Use:   "mcp-openrouter-search",
	Short: "MCP server for web search via OpenRouter API",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func runServer() error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-openrouter-search",
		Version: "0.1.0",
	}, nil)

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
		for _, d := range splitDomains(input.AllowedDomains) {
			params.AllowedDomains = append(params.AllowedDomains, d)
		}
	}
	if input.ExcludedDomains != "" {
		for _, d := range splitDomains(input.ExcludedDomains) {
			params.ExcludedDomains = append(params.ExcludedDomains, d)
		}
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

func splitDomains(s string) []string {
	var result []string
	for _, d := range splitByComma(s) {
		if d != "" {
			result = append(result, d)
		}
	}
	return result
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
