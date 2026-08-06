# syntax=docker/dockerfile:1.7
FROM debian:bookworm-slim
ENV DEBIAN_FRONTEND=noninteractive
ARG CACHEBUST=1
RUN echo "cachebust=$CACHEBUST" \
 && apt-get update \
 && apt-get install -y --no-install-recommends \
      ca-certificates curl gnupg fonts-liberation \
      libasound2 libatk-bridge2.0-0 libgtk-3-0 libnss3 libgbm1 \
 && curl -fsS https://brave-browser-apt-release.s3.brave.com/brave-browser-archive-keyring.gpg \
      | gpg --dearmor -o /usr/share/keyrings/brave-browser-archive-keyring.gpg \
 && echo "deb [signed-by=/usr/share/keyrings/brave-browser-archive-keyring.gpg] https://brave-browser-apt-release.s3.brave.com/ stable main" \
      > /etc/apt/sources.list.d/brave-browser-release.list \
 && apt-get update \
 && apt-get install -y --no-install-recommends brave-browser \
 && rm -rf /var/lib/apt/lists/*
COPY run.sh /run.sh
RUN chmod +x /run.sh
ENTRYPOINT ["bash", "/run.sh"]
