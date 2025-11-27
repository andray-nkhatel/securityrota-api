# Stage 1: Build Vue frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend
COPY client/package*.json ./
RUN npm install --force
COPY client/ ./
ENV VITE_API_BASE_URL=/api/v1
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.23-alpine AS backend-builder

WORKDIR /app
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 3: Final image
FROM alpine:3.19

WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata

# Copy Go binary
COPY --from=backend-builder /app/main .

# Copy frontend build
COPY --from=frontend-builder /frontend/dist ./static

EXPOSE 8080

ENV GIN_MODE=release
ENV DB_HOST=postgres
ENV DB_PORT=5432
ENV DB_USER=rota
ENV DB_NAME=securityrota
# DB_PASSWORD and JWT_SECRET should be set at runtime via docker-compose or -e flag

CMD ["./main"]
