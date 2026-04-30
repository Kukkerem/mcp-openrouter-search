# Multi-stage build: compile Go binary in a builder stage, then copy to a minimal image
# --- Builder stage ---
FROM golang:1.26 AS builder
WORKDIR /src

# Copy go module files first for efficient layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o mcp-openrouter-search .

# --- Runtime stage ---
FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=builder /src/mcp-openrouter-search /usr/local/bin/mcp-openrouter-search
ENTRYPOINT ["/usr/local/bin/mcp-openrouter-search"]

# --- Alternative tiny Alpine image ---
# FROM alpine:3 AS alpine-runtime
# RUN apk add --no-cache ca-certificates
# COPY --from=builder /src/mcp-openrouter-search /usr/local/bin/mcp-openrouter-search
# ENTRYPOINT ["/usr/local/bin/mcp-openrouter-search"]
