"""Minimal aioquic H3 client that dials capture UDP (for fingerprint dump)."""
from __future__ import annotations

import argparse
import asyncio
import logging
import ssl

from aioquic.asyncio.client import connect
from aioquic.h3.connection import H3_ALPN
from aioquic.quic.configuration import QuicConfiguration

LOG = logging.getLogger("emit-aioquic")


async def main() -> None:
    p = argparse.ArgumentParser()
    p.add_argument("--host", default="127.0.0.1")
    p.add_argument("--port", type=int, default=4433)
    p.add_argument("--sni", default="aioquic.fp.lab")
    p.add_argument("--alpn", default="h3")
    args = p.parse_args()

    conf = QuicConfiguration(is_client=True, alpn_protocols=[args.alpn])
    conf.verify_mode = ssl.CERT_NONE
    conf.server_name = args.sni

    try:
        async with connect(args.host, args.port, configuration=conf) as client:
            # Blackhole capture: connection will fail; Initial already sent.
            await asyncio.sleep(0.5)
            _ = client
    except Exception as e:  # noqa: BLE001 — expected against capture blackhole
        LOG.info("dial finished/err (ok for dump): %s", e)
    print(f"ok aioquic sni={args.sni}")


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    asyncio.run(main())
