#!/usr/bin/env python3
"""Structural compare of two quic-raw-initial-v1 profile dirs (GREASE-tolerant)."""
from __future__ import annotations

import json
import sys
from pathlib import Path


def load(d: Path) -> dict:
    prof = json.loads((d / "profile.json").read_text(encoding="utf-8"))
    exp = prof.get("expected") or {}
    tp = exp.get("tp_id_set") or []
    # normalize GREASE token
    tp_n = ["GREASE" if (isinstance(x, str) and x.upper() == "GREASE") or x == 27 else str(x) for x in tp]
    initials = list((d / "initials").glob("*.bin")) if (d / "initials").is_dir() else []
    return {
        "id": prof.get("id"),
        "family": prof.get("family"),
        "tp": tuple(sorted(set(tp_n))),
        "dg": len(initials),
        "has_011": "0x11" in tp_n,
        "has_3128": "0x3128" in tp_n,
        "emit": prof.get("emit"),
    }


def main() -> int:
    if len(sys.argv) != 3:
        print("usage: compare_structural.py <profileA> <profileB>", file=sys.stderr)
        return 2
    a, b = load(Path(sys.argv[1])), load(Path(sys.argv[2]))
    ok = a["tp"] == b["tp"] and a["dg"] == b["dg"] and a["has_011"] == b["has_011"] and a["has_3128"] == b["has_3128"]
    print(json.dumps({"a": a, "b": b, "structural_match": ok}, indent=2))
    return 0 if ok else 1


if __name__ == "__main__":
    raise SystemExit(main())
