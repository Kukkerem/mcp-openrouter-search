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

With a longer OpenRouter request timeout:

```json
{
  "mcpServers": {
    "openrouter-search": {
      "command": "mcp-openrouter-search",
      "args": ["--timeout-ms", "120000"],
      "env": {
        "OPENROUTER_API_KEY": "sk-or-v1-..."
      }
    }
  }
}
```

With a specific model:

```json
{
  "mcpServers": {
    "openrouter-search": {
      "command": "mcp-openrouter-search",
      "args": ["--model", "openai/gpt-4o-mini"],
      "env": {
        "OPENROUTER_API_KEY": "sk-or-v1-..."
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
    --timeout-ms int               Request timeout in milliseconds (default 60000) [global flag]

-q, --query string                 Search question (required)
-m, --model string                 OpenRouter model id (default "openai/gpt-5-nano")
    --engine string                auto, native, exa, firecrawl, parallel (default "auto")
    --max-results int              Results per search call, 1-25 (default 5)
    --max-total-results int        Cap total results
    --search-context-size string   low, medium, high (default "medium")
    --allowed-domains string       Restrict to domains (comma-separated)
    --excluded-domains string      Exclude domains (comma-separated)
    --api-key-file string          Read API key from file
    --json                         Emit structured JSON
    --raw                          Emit raw OpenRouter response JSON
```

> `--timeout-ms` is a global flag and must be placed before the subcommand:
> `mcp-openrouter-search --timeout-ms 30000 search --query "..."`

## Docker

Build and run locally:

```bash
docker build -t mcp-openrouter-search .
docker run -e OPENROUTER_API_KEY="sk-or-v1-..." mcp-openrouter-search search --query "test"
```

Or use Docker Compose with a key file:

```bash
# Create a key file
echo "sk-or-v1-..." > openrouter-api-key.txt
docker compose up
```

Pull the prebuilt image from GitHub Container Registry:

```bash
docker pull ghcr.io/kukkerem/mcp-openrouter-search:latest
```

## GitHub Packages Install

Install as a global npm package from GitHub Packages. The package downloads the correct binary for your platform during install.

Configure npm to use GitHub Packages for the `@kukkerem` scope:

```bash
npm config set @kukkerem:registry https://npm.pkg.github.com
```

If npm asks for authentication, create a GitHub personal access token (classic) with `read:packages`, then run:

```bash
npm login --scope=@kukkerem --registry=https://npm.pkg.github.com --auth-type=legacy
```

Install the package:

```bash
npm install -g @kukkerem/mcp-openrouter-search
mcp-openrouter-search search --query "test"
```

### MCP setup via npm package

After the global install, configure your MCP client to run the installed binary:

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

## CI/CD

This project uses GitHub Actions for CI and automated releases.

- **CI** (`.github/workflows/ci.yml`): Runs tests, lint, and Nix flake checks on every push and PR.
- **Release** (`.github/workflows/release.yml`): Triggered on version tags (`v*`):
  - Builds Go binaries for Linux, macOS, Windows (amd64 + arm64) via GoReleaser
  - Builds and pushes multi-platform Docker images to `ghcr.io/kukkerem/mcp-openrouter-search`
  - Publishes the GitHub Packages npm package
  - Uses `GITHUB_TOKEN` for GitHub Packages publishing; no npmjs token is required
  - First GitHub Packages publishes default to private; adjust package visibility/access in GitHub after publishing if needed
- **Dependabot** (`.github/dependabot.yml`): Monitors and auto-PRs updates for:
  - Go modules (`go.mod`)
  - GitHub Actions
  - Nix flakes (`flake.lock`)
  
  **Note:** When Dependabot updates Go modules, the `vendorHash` in `flake.nix` must be recalculated manually:
  ```bash
  # Update go.sum and recalculate vendorHash
  go mod tidy
  nix build .#  # Will fail with the expected new hash
  # Copy the new hash from the error message and update flake.nix
  ```

To trigger a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## Releasing

### Prerequisites

- Install [GoReleaser](https://goreleaser.com/)
- Export `GITHUB_TOKEN` with `repo` and `write:packages` scopes
- Log in to ghcr.io: `docker login ghcr.io -u USERNAME`

### Manual release

```bash
goreleaser release --clean
```

### Nix build

Use the project Cachix cache to avoid rebuilding cached package outputs:

```bash
nix build \
  --extra-substituters https://mcp-openrouter-search.cachix.org \
  --extra-trusted-public-keys mcp-openrouter-search.cachix.org-1:S6bkAuk57MmpxzXAjaKAmmpesRStzfWHe8Fu3rYfkJw= \
  .#
```

Or add it to your Nix configuration:

```nix
{
  nix.settings = {
    substituters = [ "https://mcp-openrouter-search.cachix.org" ];
    trusted-public-keys = [
      "mcp-openrouter-search.cachix.org-1:S6bkAuk57MmpxzXAjaKAmmpesRStzfWHe8Fu3rYfkJw="
    ];
  };
}
```

```bash
nix build .#
```

## Nix / Home Manager

### Add as flake input

```nix
# flake.nix
inputs = {
  mcp-openrouter-search = {
    url = "github:Kukkerem/mcp-openrouter-search";
    inputs.nixpkgs.follows = "nixpkgs";
  };
};
```

Pass it into your Home Manager config:

```nix
outputs = { self, nixpkgs, home-manager, mcp-openrouter-search, ... }: {
  homeConfigurations."user@host" = home-manager.lib.homeManagerConfiguration {
    extraSpecialArgs = { inherit mcp-openrouter-search; };
    modules = [ ./home.nix ];
  };
};
```

### Configure as MCP server in OpenCode

In your `home.nix` (or any imported module):

```nix
{ pkgs, mcp-openrouter-search, ... }:
let
  openrouterSearchPackage = mcp-openrouter-search.packages.${pkgs.system}.default;
in
{
  programs.opencode = {
    enable = true;
    settings = {
      mcp = {
        openrouter-search = {
          command = pkgs.lib.getExe openrouterSearchPackage;
          args = [ "--timeout-ms" "120000" ];
          env = {
            OPENROUTER_API_KEY_FILE = "/run/secrets/openrouter-api-key";
          };
        };
      };
    };
  };
}
```

### With sops-nix for secret management

```nix
{ config, pkgs, mcp-openrouter-search, ... }:
let
  openrouterSearchPackage = mcp-openrouter-search.packages.${pkgs.system}.default;
in
{
  sops.secrets.openrouter-api-key = { };

  programs.opencode.settings.mcp.openrouter-search = {
    command = pkgs.lib.getExe openrouterSearchPackage;
    env = {
      OPENROUTER_API_KEY_FILE = config.sops.secrets.openrouter-api-key.path;
    };
  };
}
```

### Cachix binary cache

Add the cache to skip rebuilding:

```nix
nix.settings = {
  substituters = [ "https://mcp-openrouter-search.cachix.org" ];
  trusted-public-keys = [
    "mcp-openrouter-search.cachix.org-1:S6bkAuk57MmpxzXAjaKAmmpesRStzfWHe8Fu3rYfkJw="
  ];
};
```

## API Key Resolution

1. `OPENROUTER_API_KEY` environment variable
2. `OPENROUTER_API_KEY_FILE` environment variable (reads file content)
3. `--api-key-file` flag (CLI mode only, reads file content)

## License

MIT
