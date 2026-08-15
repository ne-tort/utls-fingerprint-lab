#!/usr/bin/env python3
"""Shared helpers for quic-utls catalog / match / export (lab-only).

Single place for format constants, TP normalization, observation load,
and dialability rules — used by sync/match/extract/export/experiment.
"""
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

PROFILE_FORMAT = "quic-utls-profile-v1"
CATALOG_FORMAT = "quic-utls-catalog-v1"
INDEX_FORMAT = "quic-utls-catalog-index-v1"

# Marker TP ids used for identity demux / classification.
MARKER_IDS = ("0x11", "0x3128", "0x20", "0x2ab2")


def write_json(path: Path, obj: Any) -> None:
    """Write JSON with LF-only newlines (stable across Windows/Unix)."""
    text = json.dumps(obj, indent=2, ensure_ascii=False) + "\n"
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(text.encode("utf-8"))


def read_json(path: Path) -> Any:
    return json.loads(path.read_text(encoding="utf-8"))


def norm_tp(ids) -> list[str]:
    out: list[str] = []
    for x in ids or []:
        if x is None:
            continue
        if isinstance(x, int):
            out.append("GREASE" if x == 27 else hex(x))
            continue
        s = str(x).strip()
        if s.upper() == "GREASE" or s in ("27", "0x1b"):
            out.append("GREASE")
        elif s.lower().startswith("0x"):
            out.append("0x" + s[2:].lower())
        else:
            try:
                out.append(hex(int(s)))
            except ValueError:
                out.append(s.lower())
    return out


def tp_set(ids) -> frozenset[str]:
    return frozenset(norm_tp(ids))


def markers_from_tp(s: frozenset[str]) -> dict[str, bool]:
    return {f"has_{mid}": mid in s for mid in MARKER_IDS}


def infer_scid_len(hdr: dict, exp: dict, markers: dict[str, bool]) -> int | None:
    scid = exp.get("scid_len")
    if scid is not None:
        return scid
    if "source_conn_id_len" in hdr:
        return hdr["source_conn_id_len"]
    # Missing SCID field often means zero-length (Chrome / uquic chrome / yandex-like).
    if "dest_conn_id_len" not in hdr:
        return None
    if markers.get("has_0x3128") or (
        markers.get("has_0x11") and not markers.get("has_0x3128")
    ):
        return 0
    return None


def load_observation(d: Path) -> dict:
    prof = read_json(d / "profile.json")
    exp = prof.get("expected") or {}
    tp = norm_tp(exp.get("tp_id_set") or [])
    s = frozenset(tp)
    markers = markers_from_tp(s)
    hdr = read_json(d / "header.json") if (d / "header.json").is_file() else {}
    initials = list((d / "initials").glob("*.bin")) if (d / "initials").is_dir() else []
    return {
        "id": prof.get("id") or d.name,
        "family": prof.get("family"),
        "tp": tp,
        "tp_set": s,
        "markers": markers,
        "scid_len": infer_scid_len(hdr, exp, markers),
        "dcid_len": exp.get("dcid_len", hdr.get("dest_conn_id_len")),
        "datagram_count": len(initials),
        "ja4": exp.get("ja4"),
        "emit": prof.get("emit"),
    }


def iter_curated_profiles(catalog_dir: Path):
    """Yield (path, doc) for curated quic-utls-profile-v1 JSON files."""
    for p in sorted(catalog_dir.glob("*.json")):
        if p.name.startswith("_") or p.name == "index.json":
            continue
        doc = read_json(p)
        fmt = doc.get("format")
        if fmt != PROFILE_FORMAT:
            raise ValueError(f"{p}: refuse format {fmt!r} (want {PROFILE_FORMAT})")
        yield p, doc


def load_catalog_for_match(catalog_dir: Path) -> list[dict]:
    """Catalog entries shaped for match_utls_catalog / experiment."""
    profiles = []
    for p, doc in iter_curated_profiles(catalog_dir):
        ident = doc["identity"]
        tp = norm_tp(ident.get("tp_id_set") or [])
        markers = ident.get("markers") or markers_from_tp(frozenset(tp))
        profiles.append(
            {
                "path": str(p),
                "short": doc["short"],
                "status": doc["status"],
                "tp_set": frozenset(tp),
                "markers": markers,
                "scid_len": ident.get("scid_len"),
                "datagram_count_class": ident.get("datagram_count_class"),
                "emit": doc.get("emit"),
                "auth": doc.get("auth") or {},
                "random": doc.get("random") or {},
                "observation_refs": doc.get("observation_refs") or [],
                "doc": doc,
            }
        )
    return profiles


def is_dialable(doc: dict) -> bool:
    status = doc.get("status")
    emit = doc.get("emit") or {}
    kind = emit.get("emit_kind")
    return status == "emit_ready" and kind not in (None, "match_only")


def family_of(doc: dict) -> str:
    if doc.get("family"):
        return doc["family"]
    short = doc["short"]
    if short.startswith("quic-go"):
        return "quic-go"
    if short.startswith("uquic-chrome"):
        return "uquic-chrome"
    if short.startswith("uquic-firefox"):
        return "uquic-firefox"
    if short == "chromium-stale":
        return "chromium"
    if short in ("chrome", "yandex", "firefox", "aioquic", "curl-quiche"):
        return short
    return short
