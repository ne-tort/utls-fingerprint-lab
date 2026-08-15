# Yandex Browser (Chromium-based) for QUIC Initial capture.
# Official APT repo: https://repo.yandex.ru/yandex-browser/
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl gnupg fonts-liberation \
      libasound2 libatk-bridge2.0-0 libatk1.0-0 libcups2 libdbus-1-3 \
      libdrm2 libgbm1 libgtk-3-0 libnspr4 libnss3 libx11-xcb1 libxcomposite1 \
      libxdamage1 libxrandr2 xdg-utils \
    && curl -fsSL https://repo.yandex.ru/yandex-browser/YANDEX-BROWSER-KEY.GPG \
      | gpg --dearmor -o /usr/share/keyrings/yandex-browser.gpg \
    && echo "deb [arch=amd64 signed-by=/usr/share/keyrings/yandex-browser.gpg] https://repo.yandex.ru/yandex-browser/deb stable main" \
      > /etc/apt/sources.list.d/yandex-browser.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends yandex-browser-stable \
    && rm -rf /var/lib/apt/lists/*
COPY run.sh /run-yandex.sh
RUN chmod +x /run-yandex.sh
ENTRYPOINT ["/run-yandex.sh"]
