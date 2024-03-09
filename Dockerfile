# syntax=docker/dockerfile:1

FROM golang:1.22-alpine as golang
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify
ADD . /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /quorum-api .

FROM alpine:latest
COPY --from=golang /quorum-api .
EXPOSE 8080
CMD ["/quorum-api"]
