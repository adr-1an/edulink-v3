# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG TARGET=api

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -o /out/app ./cmd/${TARGET}

FROM alpine:latest

WORKDIR /app

COPY --from=builder /out/app /app/app

EXPOSE 8080

CMD ["/app/app"]