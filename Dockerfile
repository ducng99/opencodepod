FROM golang:1.26-alpine3.23 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o opencodepod-server ./cmd/server

FROM alpine:3.23

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/opencodepod-server /app/opencodepod-server

EXPOSE 8080
ENTRYPOINT ["/app/opencodepod-server"]
