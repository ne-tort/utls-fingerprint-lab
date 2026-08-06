# syntax=docker/dockerfile:1.7
FROM node:22-bookworm
ENV DEBIAN_FRONTEND=noninteractive
ENV ELECTRON_EXTRA_LAUNCH_ARGS="--no-sandbox --ignore-certificate-errors"
# Prefer npm registry binary host; GitHub releases are flaky behind some NATs.
ENV ELECTRON_MIRROR=https://npmmirror.com/mirrors/electron/
WORKDIR /app
COPY package.json ./
RUN npm install --omit=dev
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      libnss3 libatk-bridge2.0-0 libgtk-3-0 libgbm1 libasound2 fonts-liberation \
      xvfb \
 && rm -rf /var/lib/apt/lists/*
COPY main.js run.sh ./
RUN chmod +x run.sh
ENTRYPOINT ["bash", "/app/run.sh"]
