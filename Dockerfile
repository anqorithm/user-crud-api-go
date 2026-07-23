FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache pkgconfig postgresql-dev

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /api ./cmd/api

FROM alpine:3.19

RUN apk add --no-cache ca-certificates libpq

COPY --from=builder /api /usr/local/bin/

EXPOSE 8000

CMD ["api"]