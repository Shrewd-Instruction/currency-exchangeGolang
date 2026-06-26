FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /currency-exchange .
FROM golang:1.22-alpine

WORKDIR /root/

COPY --from=builder /currency-exchange .

EXPOSE 8080

CMD ["./currency-exchange"]
