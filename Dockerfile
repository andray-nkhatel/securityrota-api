# Stage 1: Build Go backend
FROM golang:1.23-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 2: Final image
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy the Go binary
COPY --from=backend-builder /app/main .

# Copy pre-built frontend (built by build script)
COPY static ./static

# Expose port
EXPOSE 8080

# Environment variables
ENV GIN_MODE=release
ENV DB_HOST=postgres
ENV DB_PORT=5432
ENV DB_USER=rota
ENV DB_PASSWORD=rotapass
ENV DB_NAME=securityrota
ENV JWT_SECRET=change-this-in-production

# Run the application
CMD ["./main"]
