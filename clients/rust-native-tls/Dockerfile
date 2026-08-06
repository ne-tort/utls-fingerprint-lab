# syntax=docker/dockerfile:1.7
FROM rust:1.86-bookworm AS build
WORKDIR /app
RUN apt-get update -qq && apt-get install -y -qq pkg-config libssl-dev >/dev/null \
    && rm -rf /var/lib/apt/lists/*
COPY Cargo.toml .
COPY src ./src
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/usr/local/cargo/git \
    cargo build --release \
 && cp target/release/native-tls-get /native-tls-get

FROM debian:bookworm-slim
RUN apt-get update -qq && apt-get install -y -qq ca-certificates libssl3 >/dev/null \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /native-tls-get /usr/local/bin/native-tls-get
ENV TARGET_ID=rust-native-tls
ENV DIAL_HOST=capture:8443
ENTRYPOINT ["native-tls-get"]
