# syntax=docker/dockerfile:1

FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /app/cmd/weather-api-app

RUN CGO_ENABLED=0 GOOS=linux go build -o /weather-app

EXPOSE 8080

CMD ["/weather-app"]
