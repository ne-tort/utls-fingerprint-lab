#!/usr/bin/env python3
"""Fill expected.ja4 (q…) on quic-raw-initial-v1 profiles from initials/*.bin.

Uses ja4plus (QUIC Initial decrypt → JA4). Lab-only annotation; do not vendor
into sing-box product. FoxIO JA4 string is stored as opaque expected field.

Usage (from quic/ or Docker):
  python scripts/annotate_ja4.py [--profiles DIR] [--dry-run] [--id chromeparrot]
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path


def wrap_udp(payload: bytes, sport: int = 40123, dport: int = 4433):
    from scapy.all import IP, UDP, Raw

    return IP(src="10.10.0.2", dst="10.10.0.1") / UDP(sport=sport, dport=dport) / Raw(load=payload)


def ja4_from_initials(init_dir: Path) -> str | None:
    from ja4plus import Processor

    bins = sorted(init_dir.glob("*.bin"))
    if not bins:
        return None

    proc = Processor(thread_safe=False)
    found: str | None = None
    for path in bins:
        pkt = wrap_udp(path.read_bytes())
        for result in proc.process_packet(pkt):
            if result.type == "ja4" and result.fingerprint.startswith("q"):
                found = result.fingerprint
    return found


def annotate_profile(prof_dir: Path, dry_run: bool) -> tuple[str, str | None]:
    pid = prof_dir.name
    init_dir = prof_dir / "initials"
    profile_path = prof_dir / "profile.json"
    if not profile_path.is_file():
        return pid, None
    if not init_dir.is_dir():
        return pid, None

    ja4 = ja4_from_initials(init_dir)
    if not ja4:
        return pid, None

    if dry_run:
        return pid, ja4

    data = json.loads(profile_path.read_text(encoding="utf-8"))
    expected = data.setdefault("expected", {})
    expected["ja4"] = ja4
    profile_path.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
    return pid, ja4


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument(
        "--profiles",
        type=Path,
        default=Path(__file__).resolve().parents[1] / "profiles",
        help="profiles/ directory",
    )
    ap.add_argument("--id", action="append", default=[], help="limit to profile id(s)")
    ap.add_argument("--dry-run", action="store_true")
    args = ap.parse_args()

    if not args.profiles.is_dir():
        print(f"no profiles dir: {args.profiles}", file=sys.stderr)
        return 2

    wanted = set(args.id)
    rows: list[tuple[str, str | None]] = []
    for child in sorted(args.profiles.iterdir()):
        if not child.is_dir() or child.name.startswith("."):
            continue
        if child.name.startswith("peer-"):
            continue
        if wanted and child.name not in wanted:
            continue
        rows.append(annotate_profile(child, args.dry_run))

    ok = 0
    for pid, ja4 in rows:
        if ja4:
            ok += 1
            print(f"{pid}\t{ja4}")
        else:
            print(f"{pid}\t(no ja4)", file=sys.stderr)

    print(f"# annotated {ok}/{len(rows)}", file=sys.stderr)
    return 0 if ok else 1


if __name__ == "__main__":
    raise SystemExit(main())
