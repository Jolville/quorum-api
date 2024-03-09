# syntax=docker/dockerfile:1

FROM golang:1.22

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

ADD . /app

RUN CGO_ENABLED=0 GOOS=linux go build -o /quorum-api

EXPOSE 8080

CMD ["/quorum-api"]
