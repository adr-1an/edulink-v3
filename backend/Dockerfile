FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGET=api

RUN CGO_ENABLED=0 go build -o /out/app ./cmd/${TARGET}

FROM alpine:latest

WORKDIR /app

COPY --from=builder /out/app /app/app

EXPOSE 8080

CMD ["/app/app"]
