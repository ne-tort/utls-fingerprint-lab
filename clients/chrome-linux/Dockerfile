# syntax=docker/dockerfile:1.7
FROM debian:bookworm-slim
ENV DEBIAN_FRONTEND=noninteractive
# Bust when labctl builds with image_policy=pull (fresh Chrome .deb).
ARG CACHEBUST=1
RUN echo "cachebust=$CACHEBUST" \
 && apt-get update \
 && apt-get install -y --no-install-recommends \
      ca-certificates wget fonts-liberation \
      libasound2 libatk-bridge2.0-0 libatk1.0-0 libatspi2.0-0 \
      libcairo2 libcups2 libdbus-1-3 libdrm2 libgbm1 libgtk-3-0 \
      libnspr4 libnss3 libpango-1.0-0 libx11-6 libxcb1 libxcomposite1 \
      libxdamage1 libxext6 libxfixes3 libxkbcommon0 libxrandr2 libvulkan1 \
      xdg-utils \
 && wget -q -O /tmp/chrome.deb https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb \
 && apt-get install -y --no-install-recommends /tmp/chrome.deb \
 && rm -f /tmp/chrome.deb \
 && rm -rf /var/lib/apt/lists/*
COPY run.sh /run.sh
RUN chmod +x /run.sh
ENTRYPOINT ["bash", "/run.sh"]
