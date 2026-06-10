# ── Stage 1: Build the Go binary ─────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Cache dependencies before copying source (layer cache friendly).
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /pr-complexity .


# ── Stage 2: Runtime image ────────────────────────────────────────────────────
# python:3.12-slim gives us pip for Radon while keeping the runtime compact.
FROM python:3.12-slim

LABEL org.opencontainers.image.source="https://github.com/jacob-palathingal/pr-complexity-analyzer" \
      org.opencontainers.image.description="CLI for pull-request cyclomatic complexity deltas" \
      org.opencontainers.image.licenses="MIT"

# git is required for the diff engine; ca-certificates keeps git usable against
# HTTPS remotes in CI containers.
RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Python analyzer dependency once at image build time.
RUN pip install --no-cache-dir radon==6.*

# Copy the compiled binary from the builder stage.
COPY --from=builder /pr-complexity /usr/local/bin/pr-complexity

# Mount the target repo at /repo — users pass -v $(pwd):/repo.
WORKDIR /repo

ENTRYPOINT ["pr-complexity"]
CMD ["--help"]
