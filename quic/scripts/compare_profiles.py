#!/usr/bin/env python3
"""Compare quic-raw-initial-v1 profiles: TP id-set, clienthellod id, datagram count."""
from __future__ import annotations

import json
import sys
from pathlib import Path


def load_profile(d: Path) -> dict:
    prof = json.loads((d / "profile.json").read_text(encoding="utf-8"))
    meta = {}
    mp = d / "meta.json"
    if mp.exists():
        meta = json.loads(mp.read_text(encoding="utf-8"))
    tp = {}
    tp_path = d / "tp.json"
    if tp_path.exists():
        tp = json.loads(tp_path.read_text(encoding="utf-8"))
    initials = sorted((d / "initials").glob("*.bin")) if (d / "initials").is_dir() else []
    exp = prof.get("expected") or {}
    return {
        "id": prof.get("id"),
        "family": prof.get("family"),
        "tp_id_set": exp.get("tp_id_set") or tp.get("tpids") or [],
        "clienthellod_id": exp.get("clienthellod_id") or "",
        "datagrams": len(initials) or meta.get("datagrams"),
        "completed": meta.get("completed"),
        "has_0x11": "0x11" in (exp.get("tp_id_set") or []),
        "has_0x3128": "0x3128" in (exp.get("tp_id_set") or []),
    }


def main() -> int:
    root = Path(__file__).resolve().parents[1] / "profiles"
    if len(sys.argv) > 1:
        root = Path(sys.argv[1])
    rows = []
    for d in sorted(p for p in root.iterdir() if p.is_dir() and (p / "profile.json").exists()):
        rows.append(load_profile(d))
    if not rows:
        print("no profiles in", root)
        return 1
    print(f"{'id':<28} {'family':<10} {'dg':>3} {'0x11':>4} {'0x3128':>6}  tp_id_set")
    print("-" * 100)
    for r in rows:
        print(
            f"{str(r['id']):<28} {str(r['family']):<10} {str(r['datagrams']):>3} "
            f"{'Y' if r['has_0x11'] else 'n':>4} {'Y' if r['has_0x3128'] else 'n':>6}  "
            f"{','.join(str(x) for x in r['tp_id_set'])}"
        )
    # highlight chromeparrot vs others
    parrot = [r for r in rows if "chromeparrot" in str(r["id"]) or r["id"] == "hy2-parrot"]
    uquic = [r for r in rows if "uquic" in str(r["id"])]
    if parrot and uquic:
        print("\n--- chromeparrot vs uquic chrome ---")
        p = parrot[0]
        for u in uquic:
            same_tp = set(map(str, p["tp_id_set"])) == set(map(str, u["tp_id_set"]))
            print(f"  {p['id']} vs {u['id']}: tp_id_set_equal={same_tp}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
