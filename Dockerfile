FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/api ./cmd/api

FROM gcr.io/distroless/static:nonroot

WORKDIR /app
COPY --from=builder /app/bin/api /app/api

USER nonroot:nonroot
EXPOSE 8080 50051
ENTRYPOINT ["/app/api"]
