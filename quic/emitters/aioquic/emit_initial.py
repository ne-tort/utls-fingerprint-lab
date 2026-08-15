#!/usr/bin/env python3
"""Fire one aioquic H3 Client Initial at the lab capture UDP peer.

Capture only needs the outbound Initial flight; the peer need not complete
the handshake. Default URL uses https:// so aioquic builds a real ClientHello
with ALPN h3 and lab SNI for promote-from-SNI.
"""

from __future__ import annotations

import argparse
import asyncio
import ssl
import sys

from aioquic.asyncio.client import connect
from aioquic.quic.configuration import QuicConfiguration


async def hit(host: str, port: int, sni: str, alpn: list[str], wait_ms: int) -> None:
    cfg = QuicConfiguration(is_client=True, alpn_protocols=alpn)
    cfg.verify_mode = ssl.CERT_NONE
    cfg.server_name = sni
    # Short idle: we only care that the Initial datagram(s) left the socket.
    cfg.idle_timeout = max(0.5, wait_ms / 1000.0)

    try:
        async with connect(host, port, configuration=cfg) as proto:
            await asyncio.sleep(wait_ms / 1000.0)
            _ = proto
    except Exception as exc:  # noqa: BLE001 — expected against bare capture
        # Capture may never answer; treat timeout/connection errors as OK.
        print(f"aioquic dial finished with: {type(exc).__name__}: {exc}", file=sys.stderr)
    print(f"ok emitter=aioquic sni={sni} host={host}:{port}")


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--host", default="capture")
    p.add_argument("--port", type=int, default=4433)
    p.add_argument("--sni", default="aioquic.fp.lab")
    p.add_argument("--alpn", default="h3", help="comma-separated ALPN list")
    p.add_argument("--wait-ms", type=int, default=400)
    args = p.parse_args()
    alpn = [x.strip() for x in args.alpn.split(",") if x.strip()]
    asyncio.run(hit(args.host, args.port, args.sni, alpn, args.wait_ms))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
