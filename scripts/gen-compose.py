from pathlib import Path
import re

root = Path(__file__).resolve().parent.parent
text = (root / "targets.yaml").read_text(encoding="utf-8")
ids = []
cur = None
for line in text.splitlines():
    m = re.match(r"\s+-\s+id:\s+(\S+)", line)
    if m:
        cur = m.group(1)
        continue
    if cur and re.match(r"\s+status:\s+active\b", line):
        ids.append(cur)
        cur = None

aliases = ["fp.lab.local"]
for i in ids:
    aliases.append(f"{i}.fp.lab.local")
    aliases.append(f"verify-{i}.fp.lab.local")

alias_block = "\n".join(f"          - {a}" for a in aliases)

# Compose uses $$ for literal $. Keep as $$ in output.
compose = f"""name: utls-lab

services:
  capture:
    build: ./capture
    ports:
      - "8443:8443"
    volumes:
      - ./captures:/captures
      - ./profiles:/profiles
      - ./certs:/app/certs
    environment:
      - TZ=UTC
    networks:
      default:
        aliases:
{alias_block}

  tools:
    build: ./tools
    volumes:
      - ./:/work
      - ./profiles:/profiles
      - ./targets.yaml:/work/targets.yaml:ro
    working_dir: /work
    profiles: ["tools"]

  verify-profile:
    build: ./tools
    depends_on: [capture]
    volumes:
      - ./profiles:/profiles
    environment:
      - PROFILE_ID=${{PROFILE_ID:-chromium-stable}}
    entrypoint: ["sh", "-c"]
    command:
      - |
        set -e
        ID="$${{PROFILE_ID:-chromium-stable}}"
        verify -profile "/profiles/$$ID" -dial capture:8443
    profiles: ["verify"]

  curl-impersonate-one:
    image: lexiforest/curl-impersonate:alpine
    depends_on: [capture]
    entrypoint: ["sh", "-c"]
    command:
      - |
        set -e
        : "$${{WRAP:?WRAP required}}"
        : "$${{TARGET_ID:?TARGET_ID required}}"
        "$$WRAP" -sk --http1.1 \\
          --connect-to "$$TARGET_ID.fp.lab.local:8443:capture:8443" \\
          "https://$$TARGET_ID.fp.lab.local:8443/" \\
          -H "X-Target-Id: $$TARGET_ID" \\
          -o /dev/null -w "HTTP %{{http_code}}\\n"
    profiles: ["capture"]

  openssl3:
    image: debian:bookworm-slim
    depends_on: [capture]
    entrypoint: ["bash", "-lc"]
    command:
      - |
        set -e
        for i in $$(seq 1 30); do (echo >/dev/tcp/capture/8443) 2>/dev/null && break; sleep 1; done
        apt-get update -qq && apt-get install -y -qq openssl >/dev/null
        echo | openssl s_client -connect capture:8443 -servername openssl3.fp.lab.local 2>/dev/null | head -20
        echo "openssl3 done"
    profiles: ["capture"]

  curl-openssl:
    image: curlimages/curl:8.11.1
    depends_on: [capture]
    entrypoint: ["sh", "-c"]
    command:
      - |
        for i in $$(seq 1 30); do nc -z capture 8443 2>/dev/null && break; sleep 1; done
        curl -sk --connect-to curl-openssl.fp.lab.local:8443:capture:8443 \\
          https://curl-openssl.fp.lab.local:8443/ \\
          -H 'X-Target-Id: curl-openssl'
        echo
    profiles: ["capture"]

  go-nethttp:
    build: ./clients/go-http
    depends_on: [capture]
    environment:
      - TARGET_ID=go-nethttp
      - DIAL_HOST=capture:8443
      - INSECURE=1
    profiles: ["capture"]

  python-requests:
    image: python:3.12-slim
    depends_on: [capture]
    entrypoint: ["bash", "-lc"]
    command:
      - |
        set -e
        python - <<'PY'
        import socket, ssl
        host = "python-requests.fp.lab.local"
        raw = socket.create_connection(("capture", 8443), timeout=15)
        ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE
        tls = ctx.wrap_socket(raw, server_hostname=host)
        req = (
            b"GET / HTTP/1.1\\r\\n"
            b"Host: " + host.encode() + b"\\r\\n"
            b"X-Target-Id: python-requests\\r\\n"
            b"Connection: close\\r\\n\\r\\n"
        )
        tls.sendall(req)
        print(tls.recv(4096).decode("utf-8", "replace")[:300])
        tls.close()
        PY
    profiles: ["capture"]

  python-httpx:
    image: python:3.12-slim
    depends_on: [capture]
    entrypoint: ["bash", "-lc"]
    command:
      - |
        set -e
        python - <<'PY'
        import socket, ssl
        id = "python-httpx"
        sni = id + ".fp.lab.local"
        raw = socket.create_connection(("capture", 8443), timeout=15)
        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE
        tls = ctx.wrap_socket(raw, server_hostname=sni)
        req = (f"GET / HTTP/1.1\\r\\nHost: {{sni}}\\r\\nX-Target-Id: {{id}}\\r\\nConnection: close\\r\\n\\r\\n").encode()
        tls.sendall(req)
        print(tls.recv(512).decode("utf-8", "replace")[:300])
        tls.close()
        PY
    profiles: ["capture"]

  chromium-stable:
    image: zenika/alpine-chrome:124
    depends_on: [capture]
    shm_size: "1gb"
    entrypoint: ["sh", "-c"]
    command:
      - |
        set -e
        for i in $$(seq 1 30); do getent hosts chromium-stable.fp.lab.local >/dev/null 2>&1 && break; sleep 1; done
        chromium-browser --headless=new --no-sandbox --disable-gpu --disable-dev-shm-usage \\
          --ignore-certificate-errors --allow-insecure-localhost --dump-dom \\
          https://chromium-stable.fp.lab.local:8443/ || true
        echo "chromium-stable done"
    profiles: ["capture"]

  firefox-release:
    image: debian:bookworm-slim
    depends_on: [capture]
    shm_size: "1gb"
    entrypoint: ["bash", "-lc"]
    command:
      - |
        set -e
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y -qq firefox-esr ca-certificates fonts-liberation libgtk-3-0 libdbus-glib-1-2 >/dev/null
        for i in $$(seq 1 30); do getent hosts firefox-release.fp.lab.local >/dev/null && break; sleep 1; done
        timeout 45 firefox-esr --headless --screenshot=/tmp/fp.png \\
          https://firefox-release.fp.lab.local:8443/ || true
        echo "firefox-release done"
    profiles: ["capture"]

  okhttp4:
    build: ./clients/okhttp4
    depends_on: [capture]
    profiles: ["capture"]

  node-undici:
    image: node:22-bookworm-slim
    depends_on: [capture]
    volumes:
      - ./clients/node-undici/get.js:/get.js:ro
    entrypoint: ["node", "/get.js"]
    profiles: ["capture"]

  java-jdk:
    build: ./clients/java-jdk
    depends_on: [capture]
    environment:
      - TARGET_ID=java-jdk
      - DIAL_HOST=capture:8443
    profiles: ["capture"]

  rust-rustls:
    build: ./clients/rustls
    depends_on: [capture]
    environment:
      - TARGET_ID=rust-rustls
      - DIAL_HOST=capture:8443
    profiles: ["capture"]

  gnutls-cli:
    image: debian:bookworm-slim
    depends_on: [capture]
    entrypoint: ["bash", "-lc"]
    command:
      - |
        set -e
        apt-get update -qq && apt-get install -y -qq gnutls-bin >/dev/null
        printf 'GET / HTTP/1.1\\r\\nHost: gnutls-cli.fp.lab.local\\r\\nX-Target-Id: gnutls-cli\\r\\nConnection: close\\r\\n\\r\\n' \\
          | gnutls-cli --insecure --sni-hostname=gnutls-cli.fp.lab.local capture:8443 || true
        echo "gnutls-cli done"
    profiles: ["capture"]

  php-curl:
    image: php:8.3-cli
    depends_on: [capture]
    entrypoint: ["bash", "-lc"]
    command:
      - |
        set -e
        IP=$$(getent hosts capture | awk '{{print $$1; exit}}')
        export CAPTURE_IP="$$IP"
        php -r '
          $$ip = getenv("CAPTURE_IP");
          $$ch = curl_init("https://php-curl.fp.lab.local:8443/");
          curl_setopt_array($$ch, [
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_SSL_VERIFYPEER => false,
            CURLOPT_SSL_VERIFYHOST => 0,
            CURLOPT_RESOLVE => ["php-curl.fp.lab.local:8443:" . $$ip],
            CURLOPT_HTTPHEADER => ["X-Target-Id: php-curl"],
          ]);
          $$body = curl_exec($$ch);
          echo curl_getinfo($$ch, CURLINFO_HTTP_CODE), " ", substr((string)$$body, 0, 120), PHP_EOL;
          if ($$body === false) {{ fwrite(STDERR, curl_error($$ch).PHP_EOL); exit(1); }}
        '
    profiles: ["capture"]

  dotnet-http:
    image: mcr.microsoft.com/dotnet/sdk:8.0
    depends_on: [capture]
    entrypoint: ["bash", "-lc"]
    command:
      - |
        set -e
        cat >/tmp/Program.cs <<'CS'
        using System.Net.Security;
        using System.Net.Sockets;
        using System.Security.Authentication;
        using System.Text;
        var id = "dotnet-http";
        var sni = id + ".fp.lab.local";
        var tcp = new TcpClient();
        await tcp.ConnectAsync("capture", 8443);
        var ssl = new SslStream(tcp.GetStream(), false, static (_,_,_,_) => true);
        await ssl.AuthenticateAsClientAsync(new SslClientAuthenticationOptions {{
          TargetHost = sni,
          EnabledSslProtocols = SslProtocols.Tls12 | SslProtocols.Tls13,
        }});
        var req = Encoding.ASCII.GetBytes($"GET / HTTP/1.1\\r\\nHost: {{sni}}\\r\\nX-Target-Id: {{id}}\\r\\nConnection: close\\r\\n\\r\\n");
        await ssl.WriteAsync(req);
        var buf = new byte[512];
        var n = await ssl.ReadAsync(buf);
        Console.WriteLine(Encoding.ASCII.GetString(buf, 0, n));
        CS
        dotnet new console -n DotFp -o /tmp/DotFp --force >/dev/null
        cp /tmp/Program.cs /tmp/DotFp/Program.cs
        dotnet run --project /tmp/DotFp -c Release
    profiles: ["capture"]
"""

out = root / "compose.yaml"
out.write_text(compose, encoding="utf-8", newline="\n")
c = out.read_text(encoding="utf-8")
assert "$$WRAP" in c, "missing $$WRAP"
assert "$$(seq 1 30)" in c, "missing seq"
assert "$$ip = getenv" in c, "missing php $$ip"
print(f"wrote {out} ({out.stat().st_size} bytes), aliases={len(aliases)}")
