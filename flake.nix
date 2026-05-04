{
  description = "MCP server + CLI for web search via OpenRouter API";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default =
          let
            version = "0.1.3";
          in
          pkgs.buildGoModule {
            pname = "mcp-openrouter-search";
            inherit version;
            src = ./.;
            vendorHash = "sha256-ChZ2Ist3OUzKA6K34sZn3sundyCA3E2Fh7XDYwH16n0=";
            ldflags = [
              "-s"
              "-w"
              "-X main.version=${version}"
            ];
            meta = with pkgs.lib; {
              description = "MCP server + CLI for web search via OpenRouter API";
              homepage = "https://github.com/Kukkerem/mcp-openrouter-search";
              license = licenses.mit;
              mainProgram = "mcp-openrouter-search";
            };
          };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
            goreleaser
          ];
        };
      }
    );
}
