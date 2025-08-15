# Multi-stage build for the complete application (AMD64 & ARM64)

# Stage 1: Build the frontend
FROM node:18-slim AS frontend-build
WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package*.json ./

# Clear npm cache and install dependencies
RUN npm cache clean --force && \
    rm -rf node_modules package-lock.json && \
    npm install

# Copy frontend source
COPY frontend/ ./

# Build the frontend
RUN npm run build

# Stage 2: Build the backend
FROM golang:1.21-bullseye AS backend-build
WORKDIR /app/backend

# Install dependencies for CGO + SQLite
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files and download deps
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source
COPY backend/ ./

# Build the backend (CGO enabled for SQLite)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 3: Final runtime image
FROM debian:bullseye-slim
WORKDIR /app

# Runtime deps (include sqlite3 which pulls libsqlite3-0)
RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Data dir
RUN mkdir -p /app/data

# Copy backend binary
COPY --from=backend-build /app/backend/main /app/

# Copy built frontend
COPY --from=frontend-build /app/frontend/build /app/frontend/build

# Environment
ENV GIN_MODE=release
ENV PORT=8080
ENV DB_PATH=/app/data/changelog.db

# Expose port and persist data
EXPOSE 8080
VOLUME ["/app/data"]

# Start app
CMD ["./main"]
