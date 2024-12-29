# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory that matches the go module name
WORKDIR /go/src/tcddbot

# Install required build tools
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o /go/bin/tcddbot ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy necessary files with correct paths
COPY --from=builder /go/bin/tcddbot .
COPY stations.json .

# Install required runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set timezone to Istanbul
ENV TZ=Europe/Istanbul

# Run the application
CMD ["./tcddbot"]
