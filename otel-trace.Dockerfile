# Step 1: Build stage
FROM golang:1.21 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files


# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
#RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o otel-trace ./cmd/otel-trace/main.go

# Step 2: Runtime stage
FROM alpine:latest

# Install ca-certificates in case your app makes external network requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/otel-trace .

# Command to run the executable
CMD ["./otel-trace"]