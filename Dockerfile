FROM gcr.io/distroless/static-debian12:nonroot
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/mcp-openrouter-search /usr/local/bin/mcp-openrouter-search
ENTRYPOINT ["/usr/local/bin/mcp-openrouter-search"]
