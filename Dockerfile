FROM golang:1.14 as builder
RUN apt-get update && apt-get install -y nocache git ca-certificates && update-ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/bin/go-chat-backend .


FROM debian:buster-slim
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
RUN mkdir public
COPY --from=builder /app/bin /app
COPY --from=builder /app/public /app/public
COPY .env .env
EXPOSE 8080 8888
CMD ["./go-chat-backend"]