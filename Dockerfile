# Dockerfile for The Game Server

# --- Build Stage ---
# Use the official Go image as a builder
FROM golang:1.24 AS builder

# Set up protoc
ENV PROTOC_VERSION=22.0

# Install build dependencies and protoc
RUN apt-get update && apt-get install -y make curl unzip && \
  ARCH=$(uname -m) && \
  case $ARCH in \
  x86_64) PROTOC_ARCH=x86_64 ;; \
  aarch64) PROTOC_ARCH=aarch_64 ;; \
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;; \
  esac && \
  PROTOC_ZIP=protoc-${PROTOC_VERSION}-linux-${PROTOC_ARCH}.zip && \
  curl -Lo /tmp/${PROTOC_ZIP} https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/${PROTOC_ZIP} && \
  unzip -d /usr/local /tmp/${PROTOC_ZIP} && \
  rm /tmp/${PROTOC_ZIP}

# Set the working directory inside the container
WORKDIR /app

# Create logs directory
RUN mkdir -p /app/logs

# Copy dependency management files
COPY go.mod go.sum ./
COPY Makefile ./

# Download Go module dependencies and install protoc plugins
RUN go mod download
RUN make deps

# Copy the entire source code into the container
COPY . .

# Generate protobuf and mock files
RUN make generate

# Build the server binary.
# CGO_ENABLED=0 is important for a static build.
# GOOS=linux is necessary because we're building on Alpine for a scratch image.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /server ./cmd/server

# --- Final Stage ---
# Use a minimal 'scratch' image for the final container
FROM scratch

# Copy the compiled binary from the builder stage
COPY --from=builder /server /server

# Expose the gRPC port the server will listen on
EXPOSE 50051

# Set the entrypoint for the container to run the server
ENTRYPOINT ["/server"] 