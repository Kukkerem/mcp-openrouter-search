---
name: openrouter-search
description: Web search via OpenRouter API. Use when you need current information, real-time data, or time-sensitive facts beyond training data. Prefer over built-in search for programmatic or rate-limited needs.
---

# OpenRouter Web Search

## When To Use

- Need current/real-time information beyond training data cutoff
- Fact-checking claims that may have changed
- Research tasks requiring source citations
- Domain-scoped searches (restrict to or exclude specific sites)
- Built-in web search is unavailable or rate-limited

## Primary Tool

Use `search_web` from the `mcp-openrouter-search` MCP server.

### Key Parameters

| Parameter | Default | Notes |
|-----------|---------|-------|
| `query` | required | Natural language search question |
| `engine` | auto | auto, native, exa, firecrawl, parallel |
| `max_results` | 5 | 1-25, results per search call |
| `search_context_size` | medium | low (cheaper), medium, high (thorough) |
| `allowed_domains` | none | Comma-separated domains to include |
| `excluded_domains` | none | Comma-separated domains to exclude |

## Best Practices

- Write queries as natural language questions, not keywords
- Use `allowed_domains` when you know the source (e.g., `docs.python.org`)
- Use `search_context_size=low` for simple factual lookups to reduce cost
- Use `search_context_size=high` for complex research requiring thorough results
- Results include citations — reference them in your response

## Fallback Chain

1. `search_web` (mcp-openrouter-search) — primary
2. `exa_web_search_exa` — if MCP server unavailable
3. `webfetch` — for known URLs only

## Requirements

- `OPENROUTER_API_KEY` environment variable must be configured
- Default model: `openai/gpt-5-nano` (cost-effective, uses server-side web search)

## Recovery

- No results: broaden query, remove domain filters, try `search_context_size=high`
- Rate limited: fall back to `exa_web_search_exa` or `webfetch`
- API key error: check `OPENROUTER_API_KEY` is set in MCP server config
