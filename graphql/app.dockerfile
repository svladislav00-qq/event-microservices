# ---------- STAGE 1: BUILD ----------
FROM golang:1.25.5-alpine AS builder

RUN apk add --no-cache gcc g++ make ca-certificates git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

COPY graphql ./graphql
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o app ./graphql/cmd/graphql

# ---------- STAGE 2: RUNTIME ----------
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]