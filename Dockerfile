# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/teeth-bot ./cmd/teeth-bot

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY --from=build /out/teeth-bot /teeth-bot
USER nobody
ENTRYPOINT ["/teeth-bot"]
