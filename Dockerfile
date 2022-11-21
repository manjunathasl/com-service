# build stage
FROM golang:1.19.3-alpine3.15 AS builder

WORKDIR /app

COPY . .

RUN go build -o main .

# main stage
FROM alpine:3.15

WORKDIR /app

COPY --from=builder /app .

EXPOSE 8080

CMD ["/app/main"]
