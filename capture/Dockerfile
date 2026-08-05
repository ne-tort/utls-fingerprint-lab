FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY bin/capture-linux /app/capture
RUN mkdir -p /captures /profiles /app/certs
EXPOSE 8443
ENTRYPOINT ["/app/capture", "-listen", ":8443", "-out", "/captures", "-profiles", "/profiles"]
