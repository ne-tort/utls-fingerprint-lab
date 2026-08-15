#!/usr/bin/env python3
"""Experiments: stable identity must transfer; random slots must vary.

Compares lab observation dirs against catalog/utls contracts:
  - same short → stable fields equal (tp_id_set/markers/dg class/scid when pinned)
  - same emit roundtrip (a vs b) → bytes/CH-order/frames differ; normalized CH id equal
  - different shorts → markers / tp sets separate families
  - auth: TP hex_id stays equal across runs (grease value not in identity) while
    initials differ → auth overlay into GREASE value would not break TP identity

Exit 0 only if all assertions pass.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(Path(__file__).resolve().parent))
from match_utls_catalog import (  # noqa: E402
    classify,
    load_catalog,
    load_observation,
    norm_tp,
)


def sha16(p: Path) -> str | None:
    if not p.is_file():
        return None
    return hashlib.sha256(p.read_bytes()).hexdigest()[:16]


def meta_blob(d: Path) -> dict:
    mp = d / "meta.json"
    if not mp.is_file():
        return {}
    return json.loads(mp.read_text(encoding="utf-8"))


def ch_info(d: Path) -> dict:
    m = meta_blob(d)
    ch = (
        ((m.get("fingerprint") or {}).get("ClientInitials") or {}).get("client_hello")
        or (m.get("gathered") or {}).get("client_hello")
        or {}
    )
    return {
        "ext": ch.get("extensions"),
        "ext_norm": ch.get("extensions_normalized"),
        "hex_id": ch.get("hex_id"),
        "norm_hex_id": ch.get("norm_hex_id"),
        "ciphers": ch.get("cipher_suites"),
        "groups": ch.get("supported_groups"),
        "key_share": ch.get("key_share"),
        "alpn": ch.get("alpn"),
    }


def frames(d: Path) -> list:
    m = meta_blob(d)
    pkts = ((m.get("gathered") or {}).get("packets")) or []
    return [p.get("frames") for p in pkts]


def initials(d: Path) -> dict:
    files = sorted((d / "initials").glob("*.bin")) if (d / "initials").is_dir() else []
    return {
        "n": len(files),
        "sha": [sha16(p) for p in files],
        "len": [p.stat().st_size for p in files],
    }


def load_profile_dir(d: Path) -> dict:
    obs = load_observation(d)
    tp = {}
    if (d / "tp.json").is_file():
        tp = json.loads((d / "tp.json").read_text(encoding="utf-8"))
    return {
        "path": str(d),
        "obs": obs,
        "tp_hex": tp.get("hex_id"),
        "tp_values": {
            k: v
            for k, v in tp.items()
            if k not in ("tpids", "hex_id", "num_id")
        },
        "ch": ch_info(d),
        "frames": frames(d),
        "initials": initials(d),
        "ch_bin": sha16(d / "clienthello.bin"),
        "first_sha": (meta_blob(d).get("first_sha256") or "")[:16] or None,
    }


def assert_true(cond: bool, msg: str, failures: list[str]) -> None:
    if not cond:
        failures.append(msg)


def exp_same_short_stable(catalog, profiles_root: Path, failures: list[str], report: dict) -> None:
    rows = []
    for cat in catalog:
        refs = [r for r in cat["observation_refs"] if (profiles_root / r).is_dir()]
        if len(refs) < 2:
            continue
        loaded = [load_profile_dir(profiles_root / r) for r in refs]
        base = loaded[0]
        for other in loaded[1:]:
            ok_tp = base["obs"]["tp_set"] == other["obs"]["tp_set"]
            ok_m = base["obs"]["markers"] == other["obs"]["markers"]
            ok_dg = (
                cat["datagram_count_class"] is None
                or (
                    base["obs"]["datagram_count"] == other["obs"]["datagram_count"]
                    == cat["datagram_count_class"]
                )
            )
            ok_scid = True
            if cat["scid_len"] is not None:
                ok_scid = (
                    base["obs"]["scid_len"] == other["obs"]["scid_len"] == cat["scid_len"]
                )
            # chrome family: pinned TP values (idle/windows) should transfer
            ok_vals = True
            if cat["short"] == "chrome" and base["tp_values"] and other["tp_values"]:
                ok_vals = base["tp_values"] == other["tp_values"]
            row = {
                "short": cat["short"],
                "a": base["obs"]["id"],
                "b": other["obs"]["id"],
                "tp_set": ok_tp,
                "markers": ok_m,
                "datagram_class": ok_dg,
                "scid": ok_scid,
                "tp_named_values": ok_vals,
                "tp_hex_id_eq": base["tp_hex"] == other["tp_hex"],
            }
            rows.append(row)
            assert_true(ok_tp, f"{cat['short']}: tp_set {refs}", failures)
            assert_true(ok_m, f"{cat['short']}: markers {base['obs']['id']} vs {other['obs']['id']}", failures)
            assert_true(ok_dg, f"{cat['short']}: datagram class", failures)
            assert_true(ok_scid, f"{cat['short']}: scid", failures)
            assert_true(ok_vals, f"{cat['short']}: TP named values drift {base['obs']['id']} vs {other['obs']['id']}", failures)
    report["same_short_stable"] = rows


def exp_roundtrip_random(
    pairs: list[tuple[str, str]],
    profiles_root: Path,
    failures: list[str],
    report: dict,
    catalog: list[dict] | None = None,
) -> None:
    """Random expectations depend on emit style (parrot shuffles; plain does not)."""
    short_by_id: dict[str, str] = {}
    if catalog:
        for cat in catalog:
            for ref in cat.get("observation_refs") or []:
                short_by_id[ref] = cat["short"]
            short_by_id[cat["short"]] = cat["short"]

    def expect_for(pair_id: str) -> dict[str, bool]:
        # default: require byte-level entropy; CH shuffle/chaos only for chrome parrot
        short = short_by_id.get(pair_id)
        if short is None and pair_id.startswith("exp-"):
            name = pair_id[len("exp-") :]
            short = name[:-2] if name.endswith(("-a", "-b")) else name
        if short == "chrome" or pair_id.startswith("chromeparrot") or pair_id.startswith("uquic146"):
            return {
                "initials_sha": True,
                "clienthello_bin": True,
                "ch_raw_hex_id": True,
                "ch_ext_order": True,
                "frame_chaos": True,
            }
        # sagernet plain / datagram: wire bytes vary (CID, keys, GREASE value);
        # CH extension order and Initial frame layout are deterministic.
        return {
            "initials_sha": True,
            "clienthello_bin": True,
            "ch_raw_hex_id": False,
            "ch_ext_order": False,
            "frame_chaos": False,
        }

    rows = []
    for a_id, b_id in pairs:
        da, db = profiles_root / a_id, profiles_root / b_id
        if not da.is_dir() or not db.is_dir():
            failures.append(f"missing roundtrip pair {a_id}/{b_id}")
            continue
        a, b = load_profile_dir(da), load_profile_dir(db)
        expect = expect_for(a_id)
        stable_tp = a["obs"]["tp_set"] == b["obs"]["tp_set"]
        stable_markers = a["obs"]["markers"] == b["obs"]["markers"]
        stable_dg = a["obs"]["datagram_count"] == b["obs"]["datagram_count"]
        stable_tp_hex = a["tp_hex"] == b["tp_hex"]
        stable_ch_norm = a["ch"].get("norm_hex_id") and a["ch"]["norm_hex_id"] == b["ch"]["norm_hex_id"]
        stable_ext_norm = a["ch"].get("ext_norm") == b["ch"].get("ext_norm")
        stable_ciphers = a["ch"].get("ciphers") == b["ch"].get("ciphers")
        stable_groups = a["ch"].get("groups") == b["ch"].get("groups")

        observed_random = {
            "initials_sha": a["initials"]["sha"] != b["initials"]["sha"],
            "clienthello_bin": bool(a["ch_bin"] and b["ch_bin"] and a["ch_bin"] != b["ch_bin"]),
            "ch_raw_hex_id": a["ch"].get("hex_id") != b["ch"].get("hex_id"),
            "ch_ext_order": a["ch"].get("ext") != b["ch"].get("ext"),
            "frame_chaos": a["frames"] != b["frames"],
        }

        row = {
            "pair": [a_id, b_id],
            "expect_random": expect,
            "stable": {
                "tp_set": stable_tp,
                "markers": stable_markers,
                "datagrams": stable_dg,
                "tp_hex_id": stable_tp_hex,
                "ch_norm_hex_id": bool(stable_ch_norm),
                "ch_ext_normalized": stable_ext_norm,
                "ciphers": stable_ciphers,
                "groups": stable_groups,
            },
            "random": observed_random,
            "evidence": {
                "tp_hex": [a["tp_hex"], b["tp_hex"]],
                "ch_norm": [a["ch"].get("norm_hex_id"), b["ch"].get("norm_hex_id")],
                "ch_raw": [a["ch"].get("hex_id"), b["ch"].get("hex_id")],
                "init_sha": [a["initials"]["sha"], b["initials"]["sha"]],
            },
        }
        rows.append(row)
        for k, v in row["stable"].items():
            assert_true(v, f"roundtrip {a_id}/{b_id} stable.{k} failed", failures)
        for k, want in expect.items():
            got = observed_random[k]
            if want:
                assert_true(got, f"roundtrip {a_id}/{b_id} random.{k} did not vary", failures)
            else:
                # deterministic slots must stay equal (wrong randomness would be a bug)
                assert_true(
                    not got,
                    f"roundtrip {a_id}/{b_id} unexpected entropy in {k} (plain path should be stable here)",
                    failures,
                )
    report["roundtrip_random"] = rows


def exp_cross_short_separation(catalog, profiles_root: Path, failures: list[str], report: dict) -> None:
    """Different shorts must not share full tp_set (except intentional aliases)."""
    seen: dict[frozenset, str] = {}
    rows = []
    for cat in catalog:
        key = frozenset(norm_tp(cat["doc"]["identity"]["tp_id_set"]))
        if key in seen:
            failures.append(f"duplicate tp_set between shorts {seen[key]} and {cat['short']}")
        seen[key] = cat["short"]
        # sample one lab obs and ensure it classifies to this short
        for ref in cat["observation_refs"]:
            d = profiles_root / ref
            if not d.is_dir():
                continue
            obs = load_observation(d)
            hit = classify(obs, catalog)
            ok = hit["matched_short"] == cat["short"]
            rows.append({"lab_id": ref, "expected": cat["short"], "got": hit["matched_short"], "ok": ok})
            assert_true(ok, f"classify {ref} → {hit['matched_short']} want {cat['short']}", failures)
            break
    report["cross_short"] = rows


def exp_auth_identity_invariant(profiles_root: Path, failures: list[str], report: dict) -> None:
    """Chrome roundtrip: TP hex_id identical while UDP/CH bytes differ.

    Implies GREASE *value* (auth overlay target) is outside TP identity hex —
    demux can match quic.auth without fingerprint short; emit auth won't break
    stable TP id-set / hex_id used for family identity.
    """
    a = load_profile_dir(profiles_root / "chromeparrot")
    b = load_profile_dir(profiles_root / "chromeparrot-b")
    row = {
        "tp_hex_stable": a["tp_hex"] == b["tp_hex"],
        "tp_set_stable": a["obs"]["tp_set"] == b["obs"]["tp_set"],
        "initials_vary": a["initials"]["sha"] != b["initials"]["sha"],
        "auth_channel": "tp_grease_value",
        "requires_demux_fingerprint": False,
        "note": (
            "clienthellod TP hex_id ignores grease value entropy; "
            "identity match uses GREASE token + markers"
        ),
    }
    report["auth_identity"] = row
    assert_true(row["tp_hex_stable"], "auth invariant: tp hex should be stable across chrome runs", failures)
    assert_true(row["tp_set_stable"], "auth invariant: tp_set stable", failures)
    assert_true(row["initials_vary"], "auth invariant: initials must still vary (entropy present)", failures)


def exp_catalog_random_slots(catalog, failures: list[str], report: dict) -> None:
    rows = []
    for cat in catalog:
        slots = (cat.get("random") or {}).get("slots") or (cat["doc"].get("random") or {}).get("slots") or []
        unstable = [s for s in slots if not s.get("stable_for_identity")]
        stable = [s for s in slots if s.get("stable_for_identity")]
        auth = cat.get("auth") or cat["doc"].get("auth") or {}
        row = {
            "short": cat["short"],
            "unstable_slots": [s["id"] for s in unstable],
            "stable_slots": [s["id"] for s in stable],
            "auth_capable": bool(auth.get("capable")),
            "auth_overlay_slots": [s["id"] for s in slots if s.get("auth_overlay")],
        }
        rows.append(row)
        assert_true(len(unstable) > 0, f"{cat['short']}: no unstable random slots declared", failures)
        if auth.get("capable"):
            assert_true(
                "tp_grease_value" in row["auth_overlay_slots"],
                f"{cat['short']}: auth.capable but no auth_overlay on tp_grease_value",
                failures,
            )
            assert_true(
                auth.get("requires_demux_fingerprint") is False,
                f"{cat['short']}: auth must not require demux fingerprint",
                failures,
            )
    report["catalog_random_slots"] = rows


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--profiles", type=Path, default=ROOT / "profiles")
    ap.add_argument("--catalog", type=Path, default=ROOT / "catalog" / "utls")
    ap.add_argument("--json-out", type=Path, default=ROOT / "fixtures" / "stable-random-experiment.json")
    ap.add_argument(
        "--extra-pair",
        action="append",
        default=[],
        help="extra roundtrip pair idA,idB (repeatable)",
    )
    args = ap.parse_args()

    catalog = load_catalog(args.catalog)
    # attach doc.random for slot checks
    for c in catalog:
        c["random"] = c["doc"].get("random")
        c["auth"] = c["doc"].get("auth")

    failures: list[str] = []
    report: dict = {"experiment": "stable-vs-random", "version": 1}

    exp_catalog_random_slots(catalog, failures, report)
    exp_same_short_stable(catalog, args.profiles, failures, report)
    pairs = [("chromeparrot", "chromeparrot-b"), ("uquic146", "uquic146-b")]
    for raw in args.extra_pair:
        parts = [p.strip() for p in raw.split(",")]
        if len(parts) != 2:
            failures.append(f"bad --extra-pair {raw}")
            continue
        pairs.append((parts[0], parts[1]))
    exp_roundtrip_random(pairs, args.profiles, failures, report, catalog=catalog)
    exp_cross_short_separation(catalog, args.profiles, failures, report)
    exp_auth_identity_invariant(args.profiles, failures, report)

    # live exp-* profiles must classify to their short
    live_rows = []
    for d in sorted(args.profiles.glob("exp-*")):
        if not d.is_dir() or not (d / "profile.json").is_file():
            continue
        # exp-<short>-a|b  (short may contain dashes: quic-go-datagram)
        name = d.name[len("exp-") :]
        if name.endswith("-a") or name.endswith("-b"):
            short = name[:-2]
        else:
            continue
        obs = load_observation(d)
        hit = classify(obs, catalog)
        ok = hit.get("matched_short") == short
        live_rows.append({"lab_id": d.name, "expected": short, "got": hit.get("matched_short"), "ok": ok})
        assert_true(ok, f"live classify {d.name} → {hit.get('matched_short')} want {short}", failures)
    report["live_classify"] = live_rows

    report["failures"] = failures
    report["ok"] = len(failures) == 0
    text = json.dumps(report, indent=2)
    args.json_out.parent.mkdir(parents=True, exist_ok=True)
    args.json_out.write_text(text + "\n", encoding="utf-8")
    print(text)
    if failures:
        print(f"FAIL n={len(failures)}", file=sys.stderr)
        for f in failures:
            print(f"  - {f}", file=sys.stderr)
        return 1
    print("OK stable/random experiment")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
