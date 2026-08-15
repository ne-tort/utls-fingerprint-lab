#!/usr/bin/env python3
"""Extract / refresh quic-utls-profile-v1 drafts from a lab observation dir.

Writes a draft under catalog/utls/_drafts/<id>.json (does not overwrite curated shorts).
Priority: stable identity fields + structured skeleton for later quic-go apply.
"""
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

# Reuse shared lab helpers
sys.path.insert(0, str(Path(__file__).resolve().parent))
from quic_utls_lib import PROFILE_FORMAT, load_observation, write_json  # noqa: E402


def infer_emit(obs: dict) -> dict:
    m = obs["markers"]
    if m["has_0x11"] and m["has_0x3128"]:
        return {
            "emit_kind": "sagernet_chrome_parrot",
            "engine": {
                "chrome_parrot": True,
                "enable_datagrams": True,
                "zero_length_scid": True,
            },
            "status": "emit_ready",
            "short_hint": "chrome",
            "value_policy": "chrome_parrot_defaults",
            "frame_policy": "chrome_random_frames",
        }
    if not m["has_0x11"] and not m["has_0x3128"] and m["has_0x20"]:
        return {
            "emit_kind": "sagernet_plain",
            "engine": {
                "chrome_parrot": False,
                "enable_datagrams": True,
                "zero_length_scid": False,
            },
            "status": "emit_ready",
            "short_hint": "quic-go-datagram",
            "value_policy": "quic_go_defaults",
            "frame_policy": "engine_default",
        }
    if not m["has_0x11"] and not m["has_0x3128"] and not m["has_0x20"]:
        return {
            "emit_kind": "sagernet_plain",
            "engine": {
                "chrome_parrot": False,
                "enable_datagrams": False,
                "zero_length_scid": False,
            },
            "status": "emit_ready",
            "short_hint": "quic-go",
            "value_policy": "quic_go_defaults",
            "frame_policy": "engine_default",
        }
    if m["has_0x11"] and not m["has_0x3128"]:
        return {
            "emit_kind": "match_only",
            "engine": {
                "chrome_parrot": False,
                "enable_datagrams": bool(m["has_0x20"]),
                "zero_length_scid": obs.get("scid_len") == 0,
            },
            "status": "observation_only",
            "short_hint": "yandex-or-firefox",
            "value_policy": "unspecified",
            "frame_policy": "unspecified",
        }
    return {
        "emit_kind": "match_only",
        "status": "observation_only",
        "short_hint": obs.get("family") or obs["id"],
        "value_policy": "unspecified",
        "frame_policy": "unspecified",
    }


def build_draft(obs: dict) -> dict:
    inf = infer_emit(obs)
    tp = obs["tp"]
    engine = inf.get("engine")
    has_grease = "GREASE" in tp
    emit: dict = {
        "emit_kind": inf["emit_kind"],
        "structured": {
            "initial": {
                "src_conn_id_length": obs.get("scid_len"),
                "dest_conn_id_length": None,
                "frame_policy": inf["frame_policy"],
            },
            "transport_parameters": {
                "id_order": tp,
                "grease": {"policy": "random_0_15" if has_grease else "omit"},
                "value_policy": inf["value_policy"],
            },
            "client_hello": {
                "policy": "engine_default"
                if inf["status"] == "emit_ready"
                else "unspecified",
                "alpn": ["h3"],
                "notes": "Draft from observation; review before promoting via sync_utls_catalog.py",
            },
        },
        "notes": f"auto-extracted from lab id={obs['id']}",
    }
    if engine:
        emit["engine"] = engine
    auth = {
        "capable": False,
        "channel": "tp_grease_value" if has_grease else None,
        "requires_demux_fingerprint": False,
        "notes": "Draft: capable only after curated chrome emit path",
    }
    random = {
        "slots": [
            {
                "id": "tp_grease_value",
                "layer": "transport_parameters",
                "stable_for_identity": False,
                "policy": "unknown" if has_grease else "omit",
                "auth_overlay": False,
            },
            {
                "id": "initial_dcid",
                "layer": "initial",
                "stable_for_identity": False,
                "policy": "random_bytes",
            },
        ]
    }
    return {
        "format": PROFILE_FORMAT,
        "version": 1,
        "short": inf["short_hint"],
        "family": obs.get("family") or inf["short_hint"],
        "aliases": [obs["id"]],
        "status": inf["status"],
        "identity": {
            "tp_id_set": tp,
            "tp_id_set_match": "set",
            "markers": obs["markers"],
            "scid_len": obs.get("scid_len"),
            "dcid_len": None,
            "datagram_count_class": obs["datagram_count"]
            if obs["datagram_count"] in (1, 2)
            else None,
            "ja4_hint": None,
        },
        "emit": emit,
        "random": random,
        "auth": auth,
        "observation_refs": [obs["id"]],
        "notes": "DRAFT — curated shorts from scripts/sync_utls_catalog.py",
    }


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument(
        "--profiles",
        type=Path,
        default=Path(__file__).resolve().parents[1] / "profiles",
    )
    ap.add_argument(
        "--out",
        type=Path,
        default=Path(__file__).resolve().parents[1] / "catalog" / "utls" / "_drafts",
    )
    ap.add_argument("--id", action="append", help="lab profile id(s); default all")
    args = ap.parse_args()
    args.out.mkdir(parents=True, exist_ok=True)

    n = 0
    for d in sorted(args.profiles.iterdir()):
        if not d.is_dir() or not (d / "profile.json").is_file():
            continue
        if args.id and d.name not in args.id:
            continue
        if not args.id and (d.name.startswith("exp-") or d.name.startswith("peer-")):
            continue
        obs = load_observation(d)
        draft = build_draft(obs)
        outp = args.out / f"{obs['id']}.json"
        write_json(outp, draft)
        n += 1
        print(f"wrote {outp} short_hint={draft['short']} status={draft['status']}")
    print(f"OK drafts={n}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
