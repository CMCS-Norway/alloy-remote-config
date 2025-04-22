FROM golang:1.23.2 AS builder

# Set the working directory
WORKDIR /app

# Copy our server implementation
COPY main.go /app/

# Initialize Go module and build
RUN go mod init github.com/cmcs-norway/alloy-remote-config-server
RUN go mod tidy
RUN go build -o /app/alloy-remote-config-server .

# Create a minimal runtime image
FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/alloy-remote-config-server /app/

# Default config location
VOLUME /configs

# Default port
EXPOSE 8080

ENTRYPOINT ["/app/alloy-remote-config-server"]
CMD ["--storage-type=file", "--storage-path=/configs", "--http-listen-addr=:8080"]
