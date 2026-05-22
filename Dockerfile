FROM golang:1.26-alpine3.23 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o opencodepod-server ./cmd/server

FROM alpine:3.23

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/opencodepod-server /app/opencodepod-server

EXPOSE 8080
ENTRYPOINT ["/app/opencodepod-server"]
