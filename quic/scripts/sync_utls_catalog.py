#!/usr/bin/env python3
"""Curated quic-utls-profile-v1 catalog writer + strict match gate.

Source of truth for shorts lives in CATALOG below; writes catalog/utls/*.json
and index.json. Random slots and auth channel are explicit.
"""
from __future__ import annotations

import argparse
from pathlib import Path

from quic_utls_lib import INDEX_FORMAT, PROFILE_FORMAT, write_json

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "catalog" / "utls"

# Shared random slot templates (identity ignores value; match uses GREASE token).
RANDOM_CHROME = [
    {
        "id": "tp_grease_id",
        "layer": "transport_parameters",
        "stable_for_identity": True,
        "policy": "token_GREASE",
        "notes": "Wire id is random large 31*N+27; identity stores GREASE token",
    },
    {
        "id": "tp_grease_value",
        "layer": "transport_parameters",
        "stable_for_identity": False,
        "policy": "random_0_15_or_auth_8_15",
        "auth_overlay": True,
        "notes": "Without quic_auth: random 0..15 bytes. With auth: HMAC tag len 8..15",
    },
    {
        "id": "tp_order",
        "layer": "transport_parameters",
        "stable_for_identity": True,
        "policy": "shuffle",
        "notes": "Identity match uses set, not order",
    },
    {
        "id": "tp_0x11_version_grease",
        "layer": "transport_parameters",
        "stable_for_identity": False,
        "policy": "chrome_version_information_grease",
        "notes": "Inside 0x11 available-versions; not a separate TP id",
    },
    {
        "id": "ch_ech_grease",
        "layer": "client_hello",
        "stable_for_identity": False,
        "policy": "ech_grease",
    },
    {
        "id": "ch_ext_shuffle",
        "layer": "client_hello",
        "stable_for_identity": False,
        "policy": "extension_shuffle",
    },
    {
        "id": "initial_dcid",
        "layer": "initial",
        "stable_for_identity": False,
        "policy": "random_bytes",
    },
    {
        "id": "initial_pn",
        "layer": "initial",
        "stable_for_identity": False,
        "policy": "engine_pn",
    },
]

RANDOM_QUICGO = [
    {
        "id": "tp_grease_id",
        "layer": "transport_parameters",
        "stable_for_identity": True,
        "policy": "token_GREASE",
        "notes": "Stock quic-go: short-ish 31*N+27 id; still GREASE token in identity",
    },
    {
        "id": "tp_grease_value",
        "layer": "transport_parameters",
        "stable_for_identity": False,
        "policy": "random_0_15",
        "auth_overlay": False,
        "notes": "Value entropy present; demux could verify HMAC (len 8..15) but emit hook today is parrot-only",
    },
    {
        "id": "initial_dcid",
        "layer": "initial",
        "stable_for_identity": False,
        "policy": "random_bytes",
    },
    {
        "id": "initial_scid",
        "layer": "initial",
        "stable_for_identity": False,
        "policy": "random_len_nonzero",
        "notes": "scid_len not pinned for quic-go shorts",
    },
    {
        "id": "initial_pn",
        "layer": "initial",
        "stable_for_identity": False,
        "policy": "engine_pn",
    },
    {
        "id": "ch_keyshare_secrets",
        "layer": "client_hello",
        "stable_for_identity": False,
        "policy": "ephemeral_keys",
        "notes": "CH bytes change; extension ORDER is deterministic (no shuffle)",
    },
    {
        "id": "ch_ext_order",
        "layer": "client_hello",
        "stable_for_identity": True,
        "policy": "fixed",
        "notes": "Plain quic-go does NOT shuffle extensions — must stay stable across runs",
    },
    {
        "id": "initial_frame_layout",
        "layer": "initial",
        "stable_for_identity": True,
        "policy": "fixed",
        "notes": "No Chrome chaos frames on plain path",
    },
]

# Auth channel: demux match does NOT need fingerprint / quic.client short.
AUTH_CHROME = {
    "capable": True,
    "channel": "tp_grease_value",
    "crypto": "demux-quic-v1",
    "value_len_min": 8,
    "value_len_max": 15,
    "emit_hook": "Config.ChromeGREASEValue",
    "demux_match": {"quic": {"auth": True}},
    "requires_demux_fingerprint": False,
    "requires_emit_style": "chrome_parrot_grease",
    "notes": (
        "Demux: match {quic:{auth:true}} only — no transport_params / client fingerprint. "
        "Client emit today still needs ChromeParrot so GREASE is chrome-shaped + overridable."
    ),
}

AUTH_NONE = {
    "capable": False,
    "channel": None,
    "demux_match": None,
    "requires_demux_fingerprint": False,
    "notes": "No stable auth overlay for this persona yet",
}

AUTH_LATENT_GREASE = {
    "capable": False,
    "channel": "tp_grease_value",
    "crypto": "demux-quic-v1",
    "value_len_min": 8,
    "value_len_max": 15,
    "emit_hook": None,
    "demux_match": {"quic": {"auth": True}},
    "requires_demux_fingerprint": False,
    "requires_emit_style": None,
    "notes": (
        "Has RFC GREASE slot; demux verify is fingerprint-agnostic. "
        "Emit overlay not wired (would need non-parrot ChromeGREASEValue or structured apply)."
    ),
}


def profile(**kwargs):
    kwargs.setdefault("format", PROFILE_FORMAT)
    kwargs.setdefault("version", 1)
    if "family" not in kwargs:
        raise ValueError(f"profile {kwargs.get('short')!r} requires family=")
    return kwargs


CATALOG: list[dict] = [
    profile(
        short="chrome",
        family="chrome",
        aliases=["chromeparrot", "hy2-default", "tuic-default"],
        status="emit_ready",
        identity={
            "tp_id_set": [
                "0x1", "0x3", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                "0xf", "0x11", "GREASE", "0x20", "0x3128",
            ],
            "tp_id_set_match": "set",
            "markers": {"has_0x11": True, "has_0x3128": True, "has_0x20": True, "has_0x2ab2": False},
            "scid_len": 0,
            "dcid_len": None,
            "datagram_count_class": 2,
            "ja4_hint": None,
        },
        emit={
            "emit_kind": "sagernet_chrome_parrot",
            "engine": {"chrome_parrot": True, "enable_datagrams": True, "zero_length_scid": True},
            "structured": {
                "initial": {
                    "src_conn_id_length": 0,
                    "dest_conn_id_length": None,
                    "frame_policy": "chrome_random_frames",
                },
                "transport_parameters": {
                    "id_order": [
                        "0x1", "0x3", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                        "0xf", "0x11", "GREASE", "0x20", "0x3128",
                    ],
                    "grease": {"policy": "random_0_15_or_auth_8_15"},
                    "value_policy": "chrome_parrot_defaults",
                },
                "client_hello": {"policy": "engine_default", "alpn": ["h3"]},
            },
        },
        random={"slots": RANDOM_CHROME},
        auth=AUTH_CHROME,
        observation_refs=[
            "chromeparrot", "chromeparrot-b", "chromiumfresh",
            "winchrome", "winedge", "hy2parrot", "tuicparrot",
        ],
        notes="Product default. Only short with auth.capable emit today.",
    ),
    profile(
        short="quic-go",
        family="quic-go",
        aliases=["quicgo", "plain"],
        status="emit_ready",
        identity={
            "tp_id_set": [
                "0x1", "0x3", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                "0xb", "0xe", "0xf", "GREASE",
            ],
            "tp_id_set_match": "set",
            "markers": {"has_0x11": False, "has_0x3128": False, "has_0x20": False, "has_0x2ab2": False},
            "scid_len": None,
            "dcid_len": None,
            "datagram_count_class": 2,
            "ja4_hint": None,
        },
        emit={
            "emit_kind": "sagernet_plain",
            "engine": {"chrome_parrot": False, "enable_datagrams": False, "zero_length_scid": False},
            "structured": {
                "initial": {"frame_policy": "engine_default"},
                "transport_parameters": {
                    "id_order": [
                        "0x1", "0x3", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                        "0xb", "0xe", "0xf", "GREASE",
                    ],
                    "grease": {"policy": "random_0_15"},
                    "value_policy": "quic_go_defaults",
                },
                "client_hello": {"policy": "engine_default", "alpn": ["h3"]},
            },
        },
        random={"slots": RANDOM_QUICGO},
        auth=AUTH_LATENT_GREASE,
        observation_refs=["quicgo"],
        notes="Stock quic-go without DATAGRAM. Auth latent in GREASE value; product forbids quic_auth without parrot.",
    ),
    profile(
        short="quic-go-datagram",
        family="quic-go",
        aliases=["quicgodg", "hy2plain", "tuicplain"],
        status="emit_ready",
        identity={
            "tp_id_set": [
                "0x1", "0x3", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                "0xb", "0xe", "0xf", "GREASE", "0x20",
            ],
            "tp_id_set_match": "set",
            "markers": {"has_0x11": False, "has_0x3128": False, "has_0x20": True, "has_0x2ab2": False},
            "scid_len": None,
            "dcid_len": None,
            "datagram_count_class": 2,
            "ja4_hint": None,
        },
        emit={
            "emit_kind": "sagernet_plain",
            "engine": {"chrome_parrot": False, "enable_datagrams": True, "zero_length_scid": False},
            "structured": {
                "initial": {"frame_policy": "engine_default"},
                "transport_parameters": {
                    "id_order": [
                        "0x1", "0x3", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                        "0xb", "0xe", "0xf", "GREASE", "0x20",
                    ],
                    "grease": {"policy": "random_0_15"},
                    "value_policy": "quic_go_defaults",
                    "values": {"0x20": {"kind": "preset", "name": "enable_datagrams"}},
                },
                "client_hello": {"policy": "engine_default", "alpn": ["h3"]},
            },
        },
        random={"slots": RANDOM_QUICGO},
        auth=AUTH_LATENT_GREASE,
        observation_refs=["quicgodg", "hy2plain", "tuicplain"],
        notes="Do not strip 0x20 (module F). Auth latent only.",
    ),
    profile(
        short="yandex",
        family="yandex",
        aliases=["winyandex"],
        status="observation_only",
        identity={
            "tp_id_set": [
                "0x1", "0x3", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                "0xf", "0x11", "GREASE", "0x20",
            ],
            "tp_id_set_match": "set",
            "markers": {"has_0x11": True, "has_0x3128": False, "has_0x20": True, "has_0x2ab2": False},
            "scid_len": 0,
            "dcid_len": None,
            "datagram_count_class": 2,
            "ja4_hint": None,
        },
        emit={
            "emit_kind": "match_only",
            "engine": {"chrome_parrot": False, "enable_datagrams": True, "zero_length_scid": True},
            "structured": {
                "initial": {"src_conn_id_length": 0, "frame_policy": "unspecified"},
                "transport_parameters": {
                    "id_order": [
                        "0x1", "0x3", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                        "0xf", "0x11", "GREASE", "0x20",
                    ],
                    "grease": {"policy": "random_0_15"},
                    "value_policy": "unspecified",
                    "values": {"0x3128": {"kind": "omit"}},
                },
                "client_hello": {"policy": "unspecified", "alpn": ["h3"]},
            },
        },
        random={"slots": [
            {"id": "tp_grease_value", "layer": "transport_parameters", "stable_for_identity": False, "policy": "unknown", "auth_overlay": False},
            {"id": "initial_dcid", "layer": "initial", "stable_for_identity": False, "policy": "random_bytes"},
        ]},
        auth=AUTH_LATENT_GREASE,
        observation_refs=["yandex", "winyandex"],
        notes="≠ chrome (no 0x3128). Auth only after structured emit.",
    ),
    profile(
        short="firefox",
        family="firefox",
        aliases=[],
        status="observation_only",
        identity={
            "tp_id_set": [
                "0x1", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                "0xb", "0xe", "0xf", "0x11", "GREASE", "0x20",
            ],
            "tp_id_set_match": "set",
            "markers": {"has_0x11": True, "has_0x3128": False, "has_0x20": True, "has_0x2ab2": False},
            "scid_len": None,
            "dcid_len": None,
            "datagram_count_class": 2,
            "ja4_hint": None,
        },
        emit={
            "emit_kind": "match_only",
            "structured": {
                "initial": {"frame_policy": "unspecified"},
                "transport_parameters": {
                    "id_order": [
                        "0x1", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                        "0xb", "0xe", "0xf", "0x11", "GREASE", "0x20",
                    ],
                    "grease": {"policy": "random_0_15"},
                    "value_policy": "unspecified",
                },
                "client_hello": {
                    "policy": "unspecified",
                    "alpn": ["h3"],
                    "notes": "Live Firefox ≠ uquic-firefox-116",
                },
            },
        },
        random={"slots": [
            {"id": "tp_grease_value", "layer": "transport_parameters", "stable_for_identity": False, "policy": "unknown"},
            {"id": "initial_scid", "layer": "initial", "stable_for_identity": False, "policy": "random_len"},
        ]},
        auth=AUTH_NONE,
        observation_refs=["firefox"],
        notes="Live browser. Separate from uquic-firefox-116.",
    ),
    profile(
        short="uquic-chrome-146",
        family="uquic-chrome",
        aliases=["uquic146"],
        status="experimental",
        identity={
            "tp_id_set": [
                "0x1", "0x3", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                "0xf", "GREASE", "0x20", "0x3127", "0x3128", "0xff73db",
            ],
            "tp_id_set_match": "set",
            "markers": {"has_0x11": False, "has_0x3128": True, "has_0x20": True, "has_0x2ab2": False},
            "scid_len": 0,
            "dcid_len": None,
            "datagram_count_class": 2,
            "ja4_hint": None,
        },
        emit={
            "emit_kind": "uquic_preset",
            "uquic_preset": "chrome-146",
            "structured": {
                "initial": {"src_conn_id_length": 0, "frame_policy": "engine_default"},
                "transport_parameters": {
                    "id_order": [
                        "0x1", "0x3", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                        "0xf", "GREASE", "0x20", "0x3127", "0x3128", "0xff73db",
                    ],
                    "grease": {"policy": "random_0_15"},
                    "value_policy": "unspecified",
                },
                "client_hello": {"policy": "unspecified"},
            },
        },
        random={"slots": [
            {"id": "tp_grease_value", "layer": "transport_parameters", "stable_for_identity": False, "policy": "uquic"},
        ]},
        auth=AUTH_NONE,
        observation_refs=["uquic146", "uquic146-b"],
        notes="≠ chrome (no 0x11). Lab reference only.",
    ),
    profile(
        short="uquic-chrome-115",
        family="uquic-chrome",
        aliases=["uquic115"],
        status="experimental",
        identity={
            "tp_id_set": [
                "0x1", "0x3", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                "0xf", "GREASE", "0x20", "0x3128", "0x4752", "0xff73db",
            ],
            "tp_id_set_match": "set",
            "markers": {"has_0x11": False, "has_0x3128": True, "has_0x20": True, "has_0x2ab2": False},
            "scid_len": 0,
            "dcid_len": None,
            "datagram_count_class": 1,
            "ja4_hint": None,
        },
        emit={
            "emit_kind": "uquic_preset",
            "uquic_preset": "chrome-115",
            "structured": {
                "transport_parameters": {
                    "id_order": [
                        "0x1", "0x3", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                        "0xf", "GREASE", "0x20", "0x3128", "0x4752", "0xff73db",
                    ],
                    "grease": {"policy": "random_0_15"},
                    "value_policy": "unspecified",
                },
                "client_hello": {"policy": "unspecified", "notes": "No ML-KEM era; not product chrome"},
            },
        },
        random={"slots": [
            {"id": "tp_grease_value", "layer": "transport_parameters", "stable_for_identity": False, "policy": "uquic"},
        ]},
        auth=AUTH_NONE,
        observation_refs=["uquic115"],
        notes="Stale vs live Chrome / sagernet parrot.",
    ),
    profile(
        short="uquic-firefox-116",
        family="uquic-firefox",
        aliases=["uquicff", "uquicffa", "uquicffb", "uquicffc"],
        status="experimental",
        identity={
            "tp_id_set": [
                "0x1", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                "0xb", "0xc", "0xe", "0xf", "GREASE", "0x20", "0x2ab2", "0xff73db",
            ],
            "tp_id_set_match": "set",
            "markers": {"has_0x11": False, "has_0x3128": False, "has_0x20": True, "has_0x2ab2": True},
            "scid_len": 3,
            "dcid_len": None,
            "datagram_count_class": 1,
            "ja4_hint": None,
        },
        emit={
            "emit_kind": "uquic_preset",
            "uquic_preset": "firefox-116",
            "structured": {
                "initial": {"src_conn_id_length": 3, "frame_policy": "engine_default"},
                "transport_parameters": {
                    "id_order": [
                        "0x1", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                        "0xb", "0xc", "0xe", "0xf", "GREASE", "0x20", "0x2ab2", "0xff73db",
                    ],
                    "grease": {"policy": "random_0_15"},
                    "value_policy": "unspecified",
                },
                "client_hello": {"policy": "unspecified", "notes": "≠ live firefox short"},
            },
            "notes": "ffa/b/c share identity; preset letter is emit detail only",
        },
        random={"slots": [
            {"id": "tp_grease_value", "layer": "transport_parameters", "stable_for_identity": False, "policy": "uquic"},
        ]},
        auth=AUTH_NONE,
        observation_refs=["uquicff", "uquicffa", "uquicffb", "uquicffc"],
        notes="Lab uquic FF116 family. Marker 0x2ab2.",
    ),
    profile(
        short="aioquic",
        family="aioquic",
        aliases=[],
        status="observation_only",
        identity={
            "tp_id_set": [
                "0x1", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                "0xa", "0xb", "0xe", "0xf", "0x11",
            ],
            "tp_id_set_match": "set",
            "markers": {"has_0x11": True, "has_0x3128": False, "has_0x20": False, "has_0x2ab2": False},
            "scid_len": 8,
            "dcid_len": None,
            "datagram_count_class": 1,
            "ja4_hint": None,
        },
        emit={"emit_kind": "match_only", "notes": "python aioquic"},
        random={"slots": [
            {"id": "initial_scid", "layer": "initial", "stable_for_identity": False, "policy": "library"},
        ]},
        auth=AUTH_NONE,
        observation_refs=["aioquic"],
        notes="No GREASE TP in this capture — auth channel absent.",
    ),
    profile(
        short="curl-quiche",
        family="quiche",
        aliases=["curlquiche", "quiche"],
        status="observation_only",
        identity={
            "tp_id_set": [
                "0x1", "0x3", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                "0xa", "0xb", "0xc", "0xf",
            ],
            "tp_id_set_match": "set",
            "markers": {"has_0x11": False, "has_0x3128": False, "has_0x20": False, "has_0x2ab2": False},
            "scid_len": 20,
            "dcid_len": None,
            "datagram_count_class": 1,
            "ja4_hint": None,
        },
        emit={"emit_kind": "match_only", "notes": "curl+quiche"},
        random={"slots": [
            {"id": "initial_scid", "layer": "initial", "stable_for_identity": False, "policy": "library"},
        ]},
        auth=AUTH_NONE,
        observation_refs=["curlquiche"],
        notes="No GREASE in this capture.",
    ),
    profile(
        short="chromium-stale",
        family="chromium",
        aliases=["chromium"],
        status="observation_only",
        identity={
            "tp_id_set": [
                "0x1", "0x3", "0x4", "0x5", "0x6", "0x7", "0x8", "0x9",
                "0xf", "GREASE", "0x20", "0x4752", "0xff73db",
            ],
            "tp_id_set_match": "set",
            "markers": {"has_0x11": False, "has_0x3128": False, "has_0x20": True, "has_0x2ab2": False},
            "scid_len": None,
            "dcid_len": None,
            "datagram_count_class": 2,
            "ja4_hint": None,
        },
        emit={
            "emit_kind": "match_only",
            "notes": "zenika/alpine Chromium — stale vs chromeparrot; keep as separate short",
        },
        random={"slots": [
            {"id": "tp_grease_value", "layer": "transport_parameters", "stable_for_identity": False, "policy": "unknown"},
        ]},
        auth=AUTH_LATENT_GREASE,
        observation_refs=["chromium"],
        notes="Not an alias of chrome. Use chromiumfresh / winchrome for live.",
    ),
]


def write_catalog() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    for p in OUT.glob("*.json"):
        p.unlink()
    index = {
        "format": INDEX_FORMAT,
        "version": 1,
        "description": "uTLS-like shorts for QUIC. Auth demux is fingerprint-agnostic (quic.auth).",
        "shorts": [],
        "notes": (
            "Dial: emit_ready. Demux auth: match quic.auth only — no fingerprint short required. "
            "See docs/UTLS_PROFILE.md. Lab SoT: catalog/utls/*.json via sync_utls_catalog.py."
        ),
    }
    for doc in CATALOG:
        path = OUT / f"{doc['short']}.json"
        write_json(path, doc)
        index["shorts"].append(
            {
                "short": doc["short"],
                "family": doc.get("family"),
                "status": doc["status"],
                "file": path.name,
                "auth_capable": bool(doc.get("auth", {}).get("capable")),
            }
        )
        print(f"wrote {path}")
    write_json(OUT / "index.json", index)
    print(f"wrote {OUT / 'index.json'} shorts={len(CATALOG)}")


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--check-only", action="store_true")
    args = ap.parse_args()
    if args.check_only:
        for doc in CATALOG:
            assert doc["format"] == PROFILE_FORMAT
            assert doc.get("family"), doc.get("short")
            assert "random" in doc and "auth" in doc
            if doc["auth"].get("capable"):
                assert "GREASE" in doc["identity"]["tp_id_set"]
                assert doc["auth"].get("requires_demux_fingerprint") is False
        print(f"OK catalog defs={len(CATALOG)}")
        return 0
    write_catalog()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
