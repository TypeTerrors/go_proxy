# ----------------------------
# Build Stage
# ----------------------------
FROM golang:1.24-alpine AS builder

# Install git (if needed for module downloads)
RUN apk update && apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum first, so we can cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code.
COPY . .

# Declare a build argument for the version
ARG VERSION

# Build the binary with the ldflags that inject the version into the global variable.
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-X main.Version=${VERSION}" -o main ./cmd/server

# ----------------------------
# Run Stage
# ----------------------------
FROM alpine:latest

# Install certificates (if your application makes HTTPS requests)
RUN apk add --no-cache ca-certificates

# Set working directory in the runtime image.
WORKDIR /root/

# Copy the binary from the builder stage.
COPY --from=builder /app/main .

# Expose the port (as defined in your code, e.g. 3000).
EXPOSE 80

# Run the binary.
CMD ["./main"]