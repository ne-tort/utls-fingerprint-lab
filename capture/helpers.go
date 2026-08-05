package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func pemEncode(typ string, der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der})
}

func targetFromSNI(sni, fallback string) string {
	sni = strings.ToLower(strings.TrimSpace(sni))
	if sni == "" {
		return fallback
	}
	const suffix = ".fp.lab.local"
	if strings.HasSuffix(sni, suffix) {
		id := strings.TrimSuffix(sni, suffix)
		if id != "" && !strings.Contains(id, ".") {
			return id
		}
	}
	if sni == "fp.lab.local" || sni == "capture" || sni == "localhost" {
		return fallback
	}
	return fallback
}

func writeImportableProfile(root, targetID, captureDir string, hello, metaJSON []byte, meta *Meta) error {
	id := sanitize(targetID)
	if id == "" || id == "unknown" {
		return nil
	}
	dir := filepath.Join(root, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "clienthello.bin"), hello, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), metaJSON, 0o644); err != nil {
		return err
	}
	profile := map[string]any{
		"id":     id,
		"format": "utls-raw-clienthello-v1",
		"source": map[string]any{
			"captured_at": meta.CapturedAt,
			"capture_dir": captureDir,
		},
		"files": map[string]string{
			"clienthello_bin": "clienthello.bin",
			"meta_json":       "meta.json",
		},
		"expected": map[string]any{
			"ja4":               meta.JA4,
			"ja3_hash":          meta.JA3Hash,
			"ja3_hash_unstable": true,
		},
		"utls": map[string]string{
			"import": "Fingerprinter{AllowBluntMimicry:true}.RawClientHello(clienthello.bin) → UClient(HelloCustom).ApplyPreset(spec)",
		},
	}
	b, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "profile.json"), b, 0o644); err != nil {
		return err
	}
	readme := "# uTLS profile: " + id + "\n\n" +
		"Format: `utls-raw-clienthello-v1`\n\n" +
		"Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.\n\n" +
		"Expected JA4: `" + meta.JA4 + "`\n\n" +
		"Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=" + id + "`\n"
	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0o644)
}

func sanitize(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	out := b.String()
	if out == "" {
		return "unknown"
	}
	return out
}

func sanitizeALPN(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('0')
		}
	}
	for b.Len() < 2 {
		b.WriteByte('0')
	}
	return b.String()[:2]
}

func isGREASE16(v uint16) bool {
	// RFC 8701
	return v&0x0f0f == 0x0a0a
}

func hexList16(vs []uint16) []string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = fmt.Sprintf("0x%04x", v)
	}
	return out
}

func joinUint16(vs []uint16, skipGREASE bool) string {
	var parts []string
	for _, v := range vs {
		if skipGREASE && isGREASE16(v) {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d", v))
	}
	return strings.Join(parts, "-")
}

func joinUint16HexLower(vs []uint16, skipGREASE bool) string {
	var parts []string
	for _, v := range vs {
		if skipGREASE && isGREASE16(v) {
			continue
		}
		parts = append(parts, fmt.Sprintf("%04x", v))
	}
	return strings.Join(parts, ",")
}

func joinUint8(vs []uint8) string {
	var parts []string
	for _, v := range vs {
		parts = append(parts, fmt.Sprintf("%d", v))
	}
	return strings.Join(parts, "-")
}

func sortUint16(vs []uint16) {
	sort.Slice(vs, func(i, j int) bool { return vs[i] < vs[j] })
}

func md5String(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type prefixConn struct {
	net.Conn
	prefix []byte
	off    int
}

func (p *prefixConn) Read(b []byte) (int, error) {
	if p.off < len(p.prefix) {
		n := copy(b, p.prefix[p.off:])
		p.off += n
		return n, nil
	}
	return p.Conn.Read(b)
}

type singleListener struct {
	c        net.Conn
	accepted bool
}

func (s *singleListener) Accept() (net.Conn, error) {
	if s.accepted {
		return nil, net.ErrClosed
	}
	s.accepted = true
	return s.c, nil
}

func (s *singleListener) Close() error { return nil }

func (s *singleListener) Addr() net.Addr { return s.c.LocalAddr() }
