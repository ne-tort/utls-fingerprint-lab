FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY bin/verify-linux /usr/local/bin/verify
COPY bin/emit-builtin-linux /usr/local/bin/emit-builtin
COPY bin/labctl-linux /usr/local/bin/labctl
WORKDIR /work
ENTRYPOINT ["labctl"]
