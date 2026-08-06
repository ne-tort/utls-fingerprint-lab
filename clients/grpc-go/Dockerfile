# syntax=docker/dockerfile:1.7
FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download || true
COPY main.go ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod tidy && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /grpc-go .

FROM debian:bookworm-slim
RUN apt-get update -qq && apt-get install -y -qq ca-certificates >/dev/null \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /grpc-go /usr/local/bin/grpc-go
ENV TARGET_ID=grpc-go
ENV DIAL_HOST=capture:8443
ENTRYPOINT ["grpc-go"]
