# Build stage
FROM golang:1.23 AS build

# Set the working directory inside the container
WORKDIR /go/src/app

# Copy witness submodule
COPY ./witness ./witness

# Copy go.mod and go.sum to cache dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY ./cmd/serverless-witness ./cmd/serverless-witness

# Build the Go program
RUN CGO_ENABLED=0 go build -o /go/bin/app ./cmd/serverless-witness 

# Final stage using distroless
FROM gcr.io/distroless/static-debian12

# Copy the binary from the builder stage
COPY --from=build /go/bin/app /

# Expose the default port
ENV PORT=8080
EXPOSE 8080

# Documentation for other environment variables
# CONFIG: JSON config
# CONFIG_SECRET: GCP Secret Manager secret name to fetch config from
# REGION: Region for the service (required)

# Command to run the binary
CMD ["/app"]
