FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o worker worker.go

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/worker .
CMD ["./worker"]
