# ---------- BUILD ----------
FROM golang:1.25.5-alpine AS builder

RUN apk add --no-cache gcc g++ make ca-certificates git

WORKDIR /app

COPY go.mod go.sum ./

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o app ./attendee/cmd/attendee

# ---------- RUNTIME ----------
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 44046 

CMD ["./app"]