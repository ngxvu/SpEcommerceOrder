FROM golang:1.20-alpine AS builder
WORKDIR /src

# copy go.mod so dependency download works (context is repo root)
COPY go.mod go.sum ./
RUN go mod download

# copy entire repo and build the order binary (adjust package path if needed)
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/order ./order

FROM alpine:3.18
RUN adduser -D app
USER app
COPY --from=builder /app/order /app/order
EXPOSE 9090 50051
ENTRYPOINT ["/app/order"]