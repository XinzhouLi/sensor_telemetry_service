FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 go build -o server ./cmd/server

FROM alpine:3.22
RUN adduser -D appuser
USER appuser
WORKDIR /app
COPY --from=builder /app/server ./server

EXPOSE 8080
CMD ["./server"]
