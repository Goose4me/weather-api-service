# syntax=docker/dockerfile:1

FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /app/cmd/mail-sender

RUN CGO_ENABLED=0 GOOS=linux go build -o /mail-sender

EXPOSE 8080

CMD ["/mail-sender"]
