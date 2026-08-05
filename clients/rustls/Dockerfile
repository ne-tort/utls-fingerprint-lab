# syntax=docker/dockerfile:1.7
FROM rust:1.86-bookworm AS build
WORKDIR /app
COPY Cargo.toml .
COPY src ./src
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/usr/local/cargo/git \
    cargo build --release \
 && cp target/release/rustls-get /rustls-get

FROM debian:bookworm-slim
RUN apt-get update -qq && apt-get install -y -qq ca-certificates >/dev/null \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /rustls-get /usr/local/bin/rustls-get
ENV TARGET_ID=rust-rustls
ENV DIAL_HOST=capture:8443
ENTRYPOINT ["rustls-get"]
