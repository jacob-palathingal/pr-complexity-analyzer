# ── Stage 1: Build the Go binary ─────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Cache dependencies before copying source (layer cache friendly).
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /pr-complexity .


# ── Stage 2: Runtime image ────────────────────────────────────────────────────
# python:3.12-slim gives us pip and a small footprint (~130 MB).
FROM python:3.12-slim

# Install radon once at image build time — no setup needed at runtime.
RUN pip install --no-cache-dir radon==6.*

# git is required for the diff engine.
RUN apt-get update && apt-get install -y --no-install-recommends git \
    && rm -rf /var/lib/apt/lists/*

# Copy the compiled binary from the builder stage.
COPY --from=builder /pr-complexity /usr/local/bin/pr-complexity

# Mount the target repo at /repo — users pass -v $(pwd):/repo
WORKDIR /repo

ENTRYPOINT ["pr-complexity"]
CMD ["--help"]
