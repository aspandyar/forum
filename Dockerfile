FROM golang:1.26-alpine AS build

WORKDIR /src

RUN apk add --no-cache build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/forum ./cmd/web

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates && adduser -D -u 10001 forum

COPY --from=build /bin/forum /app/forum
COPY init-up.sql /app/init-up.sql
COPY ui /app/ui

USER forum

EXPOSE 4000

CMD ["./forum"]
