# Stage 1: Build Go binary
FROM golang:1.25.14-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Enable Go modules
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org,direct

# Copy only go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the source code
COPY . .

# Build the binary with architecture support
ENV TARGETARCH=arm64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-w -s" -o server .

# Stage 2: Run container
# เปลี่ยนเป็น Alpine เพื่อลดขนาด image
FROM alpine:3.20

# ติดตั้ง dependency ที่จำเป็น
RUN apk add --no-cache ca-certificates tzdata

# Set timezone to Asia/Bangkok
RUN ln -sf /usr/share/zoneinfo/Asia/Bangkok /etc/localtime \
 && echo "Asia/Bangkok" > /etc/timezone


WORKDIR /app

COPY --from=builder /app/server .

# สร้างโฟลเดอร์ /temp และตั้งสิทธิ์
RUN mkdir -p /temp && chmod 777 /temp


ENV PORT=8080 \
    TIMEZONE=Asia/Bangkok \
    USE_LLM=false \
    LLM_BASE_URL=http://127.0.0.1:4000/v1 \
    LLM_MODEL=local-chat

EXPOSE 8080

ENTRYPOINT ["./server"]
