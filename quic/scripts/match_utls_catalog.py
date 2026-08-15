#!/usr/bin/env python3
"""Match lab quic-raw-initial-v1 dirs against quic-utls-profile-v1 catalog.

Stability-first: TP id-set (GREASE token), marker bits, datagram class, optional SCID.
Does not require byte-identical initials or exact JA4.
"""
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

from quic_utls_lib import (  # noqa: E402
    MARKER_IDS,
    load_catalog_for_match,
    load_observation,
    write_json,
)

# Re-exports for experiment_stable_random / extract (stable import surface).
from quic_utls_lib import (  # noqa: E402,F401
    markers_from_tp,
    norm_tp,
    tp_set,
)

load_catalog = load_catalog_for_match


def score_match(obs: dict, cat: dict) -> tuple[int, list[str]]:
    """Higher score = better. Hard fail on marker / tp_set / pinned scid mismatch."""
    reasons: list[str] = []
    score = 0
    cm = cat["markers"]
    om = obs["markers"]
    for mid in MARKER_IDS:
        k = f"has_{mid}"
        if k in cm and cm[k] is not None and cm[k] != om.get(k):
            return -1, [f"marker {k}: obs={om.get(k)} cat={cm[k]}"]
    if obs["tp_set"] == cat["tp_set"]:
        score += 100
        reasons.append("tp_set_exact")
    else:
        return -1, ["tp_set_mismatch"]
    if cat["scid_len"] is not None:
        if obs["scid_len"] is None:
            reasons.append("scid_unknown")
        elif obs["scid_len"] == cat["scid_len"]:
            score += 20
            reasons.append("scid_match")
        else:
            return -1, [f"scid obs={obs['scid_len']} cat={cat['scid_len']}"]
    dcc = cat["datagram_count_class"]
    if dcc is not None and obs["datagram_count"]:
        if obs["datagram_count"] == dcc:
            score += 10
            reasons.append("dg_class")
        else:
            reasons.append(f"dg_class_soft obs={obs['datagram_count']} cat={dcc}")
    if obs["id"] in cat["observation_refs"] or obs["id"] == cat["short"]:
        score += 5
        reasons.append("ref_hint")
    return score, reasons


def classify(obs: dict, catalog: list[dict]) -> dict:
    best = None
    best_score = -1
    best_reasons: list[str] = []
    candidates = []
    for cat in catalog:
        sc, reasons = score_match(obs, cat)
        if sc < 0:
            continue
        candidates.append({"short": cat["short"], "score": sc, "reasons": reasons})
        if sc > best_score:
            best_score = sc
            best = cat
            best_reasons = reasons
    auth = (best.get("auth") if best else None) or {}
    return {
        "lab_id": obs["id"],
        "matched_short": best["short"] if best else None,
        "status": best["status"] if best else None,
        "score": best_score if best else None,
        "reasons": best_reasons,
        "emit_kind": (best["emit"] or {}).get("emit_kind") if best else None,
        "engine": (best["emit"] or {}).get("engine") if best else None,
        "auth_capable": bool(auth.get("capable")),
        "auth_channel": auth.get("channel"),
        "requires_demux_fingerprint": auth.get("requires_demux_fingerprint"),
        "markers": obs["markers"],
        "tp_set": sorted(obs["tp_set"]),
        "datagram_count": obs["datagram_count"],
        "scid_len": obs["scid_len"],
        "candidates": candidates,
    }


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument(
        "--catalog",
        type=Path,
        default=Path(__file__).resolve().parents[1] / "catalog" / "utls",
    )
    ap.add_argument(
        "--profiles",
        type=Path,
        default=Path(__file__).resolve().parents[1] / "profiles",
    )
    ap.add_argument("--id", action="append", help="limit to lab profile id(s)")
    ap.add_argument("--json-out", type=Path, help="write full report JSON")
    ap.add_argument(
        "--strict-all",
        action="store_true",
        help="fail unless every lab profile matches a catalog short",
    )
    args = ap.parse_args()

    catalog = load_catalog(args.catalog)
    if not catalog:
        print("no catalog profiles", file=sys.stderr)
        return 2

    rows = []
    for d in sorted(args.profiles.iterdir()):
        if not d.is_dir() or not (d / "profile.json").is_file():
            continue
        if args.id and d.name not in args.id:
            continue
        # skip live exp noise and peer junk unless asked
        if not args.id and (
            d.name.startswith("exp-") or d.name.startswith("peer-")
        ):
            continue
        rows.append(classify(load_observation(d), catalog))

    matched = sum(1 for r in rows if r["matched_short"])
    emit_ready = sum(1 for r in rows if r["status"] == "emit_ready")
    auth_ready = sum(1 for r in rows if r.get("auth_capable"))
    unmatched_ids = [r["lab_id"] for r in rows if not r["matched_short"]]
    report = {
        "catalog_shorts": [c["short"] for c in catalog],
        "profiles_total": len(rows),
        "matched": matched,
        "unmatched": len(rows) - matched,
        "unmatched_ids": unmatched_ids,
        "emit_ready_matches": emit_ready,
        "auth_capable_matches": auth_ready,
        "auth_note": (
            "Demux quic.auth does not require fingerprint short; "
            "only chrome short is auth.capable for emit today"
        ),
        "rows": rows,
    }
    print(json.dumps(report, indent=2))
    if args.json_out:
        write_json(args.json_out, report)

    required = {"chrome", "quic-go", "quic-go-datagram"}
    have = {r["matched_short"] for r in rows if r["matched_short"]}
    missing_req = required - have
    if missing_req:
        print(f"FAIL missing required shorts: {sorted(missing_req)}", file=sys.stderr)
        return 1
    if args.strict_all and unmatched_ids:
        print(f"FAIL unmatched profiles: {unmatched_ids}", file=sys.stderr)
        return 1
    bad_auth = [
        r["lab_id"]
        for r in rows
        if r.get("auth_capable") and r.get("requires_demux_fingerprint")
    ]
    if bad_auth:
        print(f"FAIL auth requires demux fingerprint: {bad_auth}", file=sys.stderr)
        return 1
    print(
        f"OK matched={matched}/{len(rows)} required_shorts={sorted(required & have)} "
        f"auth_capable={auth_ready}"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
