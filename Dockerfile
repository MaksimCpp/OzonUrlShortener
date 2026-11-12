FROM golang:1.27.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.22

RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=builder /out/server ./server

USER app
ENTRYPOINT ["./server"]
