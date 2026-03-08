FROM golang:1.25.5-alpine AS builder
RUN apk add --no-cache gcc g++ make ca-certificates git
WORKDIR /go/src/github.com/svladislav00-q/event-microservices
COPY go.mod go.sum ./
COPY vendor vendor
COPY auth auth
COPY event event
COPY attendee attendee 
COPY graphql graphql
RUN GO111MODULE=on \
    go build -mod=vendor -o /go/bin/app ./graphql

FROM alpine:3.19
WORKDIR /usr/bin
COPY --from=builder /go/bin .
EXPOSE 8080
CMD ["./app"]