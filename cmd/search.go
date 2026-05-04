package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/Kukkerem/mcp-openrouter-search/internal/config"
	"github.com/Kukkerem/mcp-openrouter-search/internal/openrouter"
	"github.com/spf13/cobra"
)

var (
	flagQuery       string
	flagModel       string
	flagEngine      string
	flagMaxResults  int
	flagMaxTotal    int
	flagContextSize string
	flagAllowed     string
	flagExcluded    string
	flagAPIKeyFile  string
	flagJSON        bool
	flagRaw         bool
)

var searchCmd = &cobra.Command{
	Use:   "search --query <query> [options]",
	Short: "Search the web via OpenRouter",
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().StringVarP(&flagQuery, "query", "q", "", "Search question or research task (required)")
	searchCmd.Flags().StringVarP(&flagModel, "model", "m", config.DefaultModel, "OpenRouter model id")
	searchCmd.Flags().StringVar(&flagEngine, "engine", config.DefaultEngine, "Search engine: auto, native, exa, firecrawl, parallel")
	searchCmd.Flags().IntVar(&flagMaxResults, "max-results", config.DefaultMaxResults, "Results per search call, 1-25")
	searchCmd.Flags().IntVar(&flagMaxTotal, "max-total-results", 0, "Cap total results across multi-search loops")
	searchCmd.Flags().StringVar(&flagContextSize, "search-context-size", config.DefaultContextSize, "Search context size: low, medium, high")
	searchCmd.Flags().StringVar(&flagAllowed, "allowed-domains", "", "Restrict search to domains (comma-separated)")
	searchCmd.Flags().StringVar(&flagExcluded, "excluded-domains", "", "Exclude domains (comma-separated)")
	searchCmd.Flags().StringVar(&flagAPIKeyFile, "api-key-file", "", "Read OpenRouter API key from file")
	searchCmd.Flags().BoolVar(&flagJSON, "json", false, "Emit structured JSON")
	searchCmd.Flags().BoolVar(&flagRaw, "raw", false, "Emit raw OpenRouter response JSON")

	searchCmd.MarkFlagRequired("query")

	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := flagQuery

	apiKey, err := config.ResolveAPIKey(flagAPIKeyFile)
	if err != nil {
		return err
	}

	if err := config.ValidateEngine(flagEngine); err != nil {
		return err
	}
	if flagMaxResults < 1 || flagMaxResults > 25 {
		return fmt.Errorf("max_results must be between 1 and 25, got %d", flagMaxResults)
	}
	if err := config.ValidateContextSize(flagContextSize); err != nil {
		return err
	}

	params := openrouter.SearchParameters{
		Query:             query,
		Engine:            flagEngine,
		MaxResults:        flagMaxResults,
		SearchContextSize: flagContextSize,
	}

	if flagMaxTotal > 0 {
		params.MaxTotalResults = &flagMaxTotal
	}

	for _, d := range splitByComma(flagAllowed) {
		params.AllowedDomains = append(params.AllowedDomains, d)
	}
	for _, d := range splitByComma(flagExcluded) {
		params.ExcludedDomains = append(params.ExcludedDomains, d)
	}

	resp, err := openrouter.DoSearch(config.OpenRouterEndpoint, apiKey, flagModel, params, flagServerTimeout)
	if err != nil {
		return err
	}

	if flagRaw {
		return printRaw(resp)
	}

	output := openrouter.BuildOutput(resp, params)

	if flagJSON {
		text, err := openrouter.FormatJSON(output)
		if err != nil {
			return err
		}
		fmt.Println(text)
		return nil
	}

	fmt.Print(openrouter.FormatText(output))
	return nil
}

func printRaw(resp *openrouter.ChatResponse) error {
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
