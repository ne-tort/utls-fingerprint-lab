// QUIC Initial capture / parse tool for quic-raw-initial-v1 profiles.
// Uses gaukas/clienthellod (same stack as quic.tlsfingerprint.io).
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gaukas/clienthellod"
)

func main() {
	listen := flag.String("listen", ":4433", "UDP listen address")
	outDir := flag.String("out", "captures", "raw capture output directory")
	profileDir := flag.String("profiles", "profiles", "importable profiles directory")
	targetHint := flag.String("default-target", "unknown", "default profile id")
	promote := flag.Bool("promote", true, "write profiles/<id>/ after successful gather")
	idle := flag.Duration("gather-idle", 800*time.Millisecond, "idle time after last datagram to finalize peer")
	parsePath := flag.String("parse", "", "offline: parse a single Initial .bin and print JSON")
	flag.Parse()

	if *parsePath != "" {
		if err := parseFile(*parsePath); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatal(err)
	}
	if *promote {
		if err := os.MkdirAll(*profileDir, 0o755); err != nil {
			log.Fatal(err)
		}
	}

	pc, err := net.ListenPacket("udp", *listen)
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()
	log.Printf("quic capture listening UDP %s → %s (profiles=%v)", *listen, *outDir, *promote)

	type peerState struct {
		gather   *clienthellod.GatheredClientInitials
		timer    *time.Timer
		mu       sync.Mutex
		datagram [][]byte
	}

	var peers sync.Map

	buf := make([]byte, 65535)
	for {
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			log.Printf("read: %v", err)
			continue
		}
		pkt := append([]byte(nil), buf[:n]...)
		key := addr.String()

		ci, err := clienthellod.UnmarshalQUICClientInitialPacket(pkt)
		if err != nil {
			log.Printf("skip non-initial from %s: %v", key, err)
			continue
		}

		v, _ := peers.LoadOrStore(key, &peerState{})
		st := v.(*peerState)
		st.mu.Lock()
		if st.gather == nil {
			st.gather = clienthellod.GatherClientInitialsWithDeadline(time.Now().Add(30 * time.Second))
			st.datagram = nil
		}
		st.datagram = append(st.datagram, pkt)
		if err := st.gather.AddPacket(ci); err != nil {
			log.Printf("gather add %s: %v", key, err)
		}
		done := st.gather.Completed()
		if st.timer != nil {
			st.timer.Stop()
		}
		hint := *targetHint
		out := *outDir
		prof := *profileDir
		doPromote := *promote
		idleCopy := *idle
		finalize := func() {
			st.mu.Lock()
			defer st.mu.Unlock()
			peers.Delete(key)
			if st.gather == nil {
				return
			}
			if err := finishPeer(key, hint, out, prof, doPromote, st.datagram, st.gather); err != nil {
				log.Printf("finish %s: %v", key, err)
			}
		}
		if done {
			st.mu.Unlock()
			finalize()
			continue
		}
		st.timer = time.AfterFunc(idleCopy, finalize)
		st.mu.Unlock()
	}
}

func parseFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	ci, err := clienthellod.UnmarshalQUICClientInitialPacket(raw)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	gci := clienthellod.GatherClientInitialsWithDeadline(time.Now().Add(30 * time.Second))
	if err := gci.AddPacket(ci); err != nil {
		return err
	}
	out := map[string]any{
		"packet":    ci,
		"completed": gci.Completed(),
		"gathered":  gci,
	}
	if gci.Completed() {
		if fp, err := clienthellod.GenerateQUICFingerprint(gci); err == nil {
			out["fingerprint"] = fp
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func finishPeer(peer, hint, outDir, profileDir string, promote bool, datagrams [][]byte, gather *clienthellod.GatheredClientInitials) error {
	id := sanitize(hint)
	if gather != nil && gather.ClientHello != nil && gather.ClientHello.ServerName != "" {
		if fromSNI := targetFromSNI(gather.ClientHello.ServerName); fromSNI != "" {
			id = fromSNI
		}
	}
	if id == "" || id == "unknown" {
		id = "peer-" + sanitize(strings.ReplaceAll(peer, ":", "-"))
	}
	ts := time.Now().UTC().Format("20060102T150405Z")
	capID := fmt.Sprintf("%s-%s", ts, id)
	capPath := filepath.Join(outDir, capID)
	if err := os.MkdirAll(filepath.Join(capPath, "initials"), 0o755); err != nil {
		return err
	}

	for i, d := range datagrams {
		name := fmt.Sprintf("%03d.bin", i)
		if err := os.WriteFile(filepath.Join(capPath, "initials", name), d, 0o644); err != nil {
			return err
		}
	}

	summary := map[string]any{
		"peer":        peer,
		"target_id":   id,
		"captured_at": time.Now().UTC().Format(time.RFC3339),
		"datagrams":   len(datagrams),
		"completed":   gather.Completed(),
		"gathered":    gather,
	}
	if gather.Completed() {
		if fp, err := clienthellod.GenerateQUICFingerprint(gather); err == nil {
			summary["fingerprint"] = fp
		}
	}
	if len(datagrams) > 0 {
		sum := sha256.Sum256(datagrams[0])
		summary["first_sha256"] = hex.EncodeToString(sum[:])
	}

	metaPath := filepath.Join(capPath, "meta.json")
	b, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(metaPath, b, 0o644); err != nil {
		return err
	}
	log.Printf("captured %s (%d datagrams, completed=%v) → %s", peer, len(datagrams), gather.Completed(), capPath)

	if !promote {
		return nil
	}
	return promoteProfile(profileDir, id, datagrams, gather, summary)
}

func promoteProfile(profileDir, id string, datagrams [][]byte, gather *clienthellod.GatheredClientInitials, summary map[string]any) error {
	dir := filepath.Join(profileDir, id)
	initDir := filepath.Join(dir, "initials")
	if err := os.MkdirAll(initDir, 0o755); err != nil {
		return err
	}
	for i, d := range datagrams {
		if err := os.WriteFile(filepath.Join(initDir, fmt.Sprintf("%03d.bin", i)), d, 0o644); err != nil {
			return err
		}
	}
	metaBytes, _ := json.MarshalIndent(summary, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, "meta.json"), metaBytes, 0o644)

	header := map[string]any{}
	tpIDs := []string{}
	hexID := ""
	family := "unknown"
	var dcidLen, scidLen any
	if gather != nil && len(gather.Packets) > 0 && gather.Packets[0].Header != nil {
		h := gather.Packets[0].Header
		headerBytes, _ := json.MarshalIndent(h, "", "  ")
		_ = os.WriteFile(filepath.Join(dir, "header.json"), headerBytes, 0o644)
		header["raw"] = h
	}
	if gather != nil && gather.TransportParameters != nil {
		tp := gather.TransportParameters
		tpBytes, _ := json.MarshalIndent(tp, "", "  ")
		_ = os.WriteFile(filepath.Join(dir, "tp.json"), tpBytes, 0o644)
		for _, idn := range tp.QTPIDs {
			if idn == clienthellod.QTP_GREASE {
				tpIDs = append(tpIDs, "GREASE")
				continue
			}
			tpIDs = append(tpIDs, fmt.Sprintf("0x%x", idn))
		}
		if hasTP(tpIDs, "0x11") && hasTP(tpIDs, "0x3128") {
			family = "chrome"
		}
	}
	if gather != nil && gather.ClientHello != nil {
		raw := gather.ClientHello.Raw()
		if len(raw) > 0 {
			_ = os.WriteFile(filepath.Join(dir, "clienthello.bin"), raw, 0o644)
		}
	}
	if gather != nil && gather.Completed() {
		if fp, err := clienthellod.GenerateQUICFingerprint(gather); err == nil && fp != nil {
			hexID = fp.HexID
		}
		if hexID == "" {
			hexID = gather.HexID
		}
	}
	emit := map[string]any{"emit_kind": "match_only"}
	switch {
	case id == "chromeparrot" || strings.Contains(id, "hy2parrot") || strings.HasPrefix(id, "chromeparrot"):
		emit = map[string]any{"emit_kind": "sagernet_chrome_parrot"}
		family = "chrome"
	case id == "quicgo" || strings.Contains(id, "hy2plain") || strings.Contains(id, "plain"):
		emit = map[string]any{"emit_kind": "sagernet_plain"}
		family = "quic-go"
	case strings.Contains(id, "uquic146"):
		emit = map[string]any{"emit_kind": "uquic_preset", "uquic_preset": "chrome-146"}
		family = "chrome"
	case strings.Contains(id, "uquic115"):
		emit = map[string]any{"emit_kind": "uquic_preset", "uquic_preset": "chrome-115"}
		family = "chrome"
	case strings.Contains(id, "uquicff") || strings.Contains(id, "firefox"):
		emit = map[string]any{"emit_kind": "uquic_preset", "uquic_preset": "firefox-116"}
		family = "firefox"
	case strings.Contains(id, "aioquic"):
		emit = map[string]any{"emit_kind": "match_only", "notes": "python aioquic"}
		family = "aioquic"
	case strings.Contains(id, "curlquiche") || strings.Contains(id, "curl"):
		emit = map[string]any{"emit_kind": "match_only", "notes": "curl+quiche"}
		family = "quiche"
	case strings.Contains(id, "chromium"):
		emit = map[string]any{"emit_kind": "match_only", "notes": "live chromium H3; freshness vs chromeparrot"}
		family = "chrome"
	}

	prof := map[string]any{
		"format":  "quic-raw-initial-v1",
		"id":      id,
		"family":  family,
		"version": 0,
		"track":   "pinned",
		"expected": map[string]any{
			"ja4":             "",
			"clienthellod_id": hexID,
			"tp_id_set":       tpIDs,
			"dcid_len":        dcidLen,
			"scid_len":        scidLen,
			"header":          header,
		},
		"emit":  emit,
		"notes": "auto-promoted from capture; review family/expected/emit (docs/REPLAY_AND_EMIT.md)",
	}
	pb, _ := json.MarshalIndent(prof, "", "  ")
	return os.WriteFile(filepath.Join(dir, "profile.json"), pb, 0o644)
}

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, s)
	return strings.Trim(s, "-")
}

// targetFromSNI maps "chromeparrot.fp.lab" → "chromeparrot".
func targetFromSNI(sni string) string {
	sni = strings.ToLower(strings.TrimSpace(sni))
	if i := strings.IndexByte(sni, '.'); i > 0 {
		sni = sni[:i]
	}
	return sanitize(sni)
}

func hasTP(ids []string, want string) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}
