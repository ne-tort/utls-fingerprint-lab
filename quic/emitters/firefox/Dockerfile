# Minimal Firefox for H3 Initial capture (prefs force QUIC without Alt-Svc).
FROM selenium/standalone-firefox:latest
USER root
COPY run.sh /run-ff.sh
RUN chmod +x /run-ff.sh && chown seluser:seluser /run-ff.sh
USER seluser
ENV HOME=/home/seluser
WORKDIR /home/seluser
ENTRYPOINT ["/run-ff.sh"]
