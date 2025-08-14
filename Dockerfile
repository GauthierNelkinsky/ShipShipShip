# Multi-stage build for the complete application

# Stage 1: Build the frontend
FROM --platform=linux/amd64 node:18-slim AS frontend-build
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
FROM --platform=linux/amd64 golang:1.21-bullseye AS backend-build
WORKDIR /app/backend

# Install dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source
COPY backend/ ./

# Build the backend
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 3: Final runtime image
FROM --platform=linux/amd64 debian:bullseye-slim
WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Create data directory
RUN mkdir -p /app/data

# Copy built backend
COPY --from=backend-build /app/backend/main /app/

# Copy built frontend
COPY --from=frontend-build /app/frontend/build /app/frontend/build

# Set environment variables
ENV GIN_MODE=release
ENV PORT=8080
ENV DB_PATH=/app/data/changelog.db

# Expose port
EXPOSE 8080

# Create volume for data persistence
VOLUME ["/app/data"]

# Run the application
CMD ["./main"]
