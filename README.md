# mcp-openrouter-search

MCP server + CLI for web search via the [OpenRouter](https://openrouter.ai/) API.

Single Go binary. Runs as an MCP server over stdio by default, or as a CLI with the `search` subcommand.

## Features

- MCP `search_web` tool with full parameter support
- Citation extraction from OpenRouter annotations
- Dual mode: MCP server (default) + CLI (`search` subcommand)
- API key from env var or file (works with `sops`, `age`, etc.)
- Structured JSON and raw output modes

## Install

```bash
go install github.com/Kukkerem/mcp-openrouter-search@latest
```

## MCP Server

Default mode — runs over stdio:

```json
{
  "mcpServers": {
    "openrouter-search": {
      "command": "mcp-openrouter-search",
      "env": {
        "OPENROUTER_API_KEY": "sk-or-v1-..."
      }
    }
  }
}
```

With an API key file:

```json
{
  "mcpServers": {
    "openrouter-search": {
      "command": "mcp-openrouter-search",
      "env": {
        "OPENROUTER_API_KEY_FILE": "/run/secrets/openrouter-api-key"
      }
    }
  }
}
```

### Tool: `search_web`

| Parameter             | Type   | Required | Default    | Description                                      |
|-----------------------|--------|----------|------------|--------------------------------------------------|
| `query`               | string | yes      | —          | Search question or research task                  |
| `engine`              | string | no       | `auto`     | `auto`, `native`, `exa`, `firecrawl`, `parallel`  |
| `max_results`         | number | no       | `5`        | Results per search call, 1-25                     |
| `max_total_results`   | number | no       | —          | Cap total results across multi-search loops        |
| `search_context_size` | string | no       | `medium`   | `low`, `medium`, `high`                           |
| `allowed_domains`     | string | no       | —          | Comma-separated domains to restrict search to     |
| `excluded_domains`    | string | no       | —          | Comma-separated domains to exclude                |

## CLI

```bash
mcp-openrouter-search search --query "latest Go generics features"
```

### Options

```
-q, --query string                 Search question (required)
-m, --model string                 OpenRouter model id (default "openai/gpt-5-nano")
    --engine string                auto, native, exa, firecrawl, parallel (default "auto")
    --max-results int              Results per search call, 1-25 (default 5)
    --max-total-results int        Cap total results
    --search-context-size string   low, medium, high (default "medium")
    --allowed-domain string        Restrict to domains (comma-separated)
    --excluded-domain string       Exclude domains (comma-separated)
    --api-key-file string          Read API key from file
    --timeout-ms int               Request timeout (default 60000)
    --json                         Emit structured JSON
    --raw                          Emit raw OpenRouter response JSON
```

## API Key Resolution

1. `OPENROUTER_API_KEY` environment variable
2. `OPENROUTER_API_KEY_FILE` environment variable (reads file content)
3. `--api-key-file` flag (CLI mode only, reads file content)

## License

MIT
