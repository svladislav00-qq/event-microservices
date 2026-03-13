# ---------- STAGE 1: BUILD ----------
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc g++ make ca-certificates git

WORKDIR /app

# go modules
COPY go.mod go.sum ./
COPY vendor ./vendor

# исходники
COPY event ./event

# сборка
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod=vendor -o app ./event/cmd/event

# ---------- STAGE 2: RUNTIME ----------
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]