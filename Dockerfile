# --- Build Stage ---
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependency layer
COPY go.mod ./
# COPY go.sum ./  # Uncomment once go.sum exists
RUN go mod download

# Copy source files
COPY . .

# Build args passed from Docker or Makefile
ARG VERSION=v0.0.0-dev
ARG COMMIT=unknown

# Build minimal statically-linked binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-X 'github.com/dchittibala/snipsnap/pkg/fileops.Version=${VERSION}' \
              -X 'github.com/dchittibala/snipsnap/pkg/fileops.Commit=${COMMIT}' \
              -s -w" \
    -o /bin/snipsnap ./cmd/snipsnap

# --- Final Runtime Stage ---
FROM alpine:3.21

RUN apk --no-cache add ca-certificates \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

# Copy binary from builder
COPY --from=builder /bin/snipsnap /usr/local/bin/snipsnap

USER appuser

ENTRYPOINT ["snipsnap"]
CMD ["--help"]