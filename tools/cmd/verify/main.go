package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	utls "github.com/metacubex/utls"

	"github.com/sagernet/sing-box/lx-test/utls-fingerprint-tools/internal/profile"
)

func main() {
	profileDir := flag.String("profile", "", "path to profiles/<id> directory")
	dialAddr := flag.String("dial", "capture:8443", "capture host:port")
	sni := flag.String("sni", "", "SNI (default: verify-<id>.fp.lab.local)")
	insecure := flag.Bool("insecure", true, "skip TLS verify")
	flag.Parse()
	if *profileDir == "" {
		fmt.Fprintln(os.Stderr, "usage: verify -profile profiles/<id>")
		os.Exit(2)
	}

	m, hello, err := profile.Load(*profileDir)
	if err != nil {
		fatal(err)
	}
	if *sni == "" {
		*sni = "verify-" + m.ID + ".fp.lab.local"
	}

	fp := &utls.Fingerprinter{AllowBluntMimicry: true}
	spec, err := fp.RawClientHello(hello)
	if err != nil {
		fatal(fmt.Errorf("RawClientHello: %w", err))
	}
	for _, ext := range spec.Extensions {
		if s, ok := ext.(*utls.SNIExtension); ok {
			s.ServerName = *sni
		}
	}

	raw, err := net.DialTimeout("tcp", *dialAddr, 15*time.Second)
	if err != nil {
		fatal(err)
	}
	defer raw.Close()
	_ = raw.SetDeadline(time.Now().Add(30 * time.Second))

	ucfg := &utls.Config{
		ServerName:         *sni,
		InsecureSkipVerify: *insecure,
		NextProtos:         []string{"http/1.1"},
		MinVersion:         tls.VersionTLS12,
	}
	uconn := utls.UClient(raw, ucfg, utls.HelloCustom)
	if err := uconn.ApplyPreset(spec); err != nil {
		fatal(fmt.Errorf("ApplyPreset: %w", err))
	}
	if err := uconn.Handshake(); err != nil {
		fatal(fmt.Errorf("handshake: %w", err))
	}

	req, err := http.NewRequest(http.MethodGet, "https://"+*sni+"/", nil)
	if err != nil {
		fatal(err)
	}
	req.Header.Set("X-Target-Id", "verify-"+m.ID)
	req.Header.Set("Connection", "close")
	if err := req.Write(uconn); err != nil {
		fatal(err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(uconn), req)
	if err != nil {
		fatal(fmt.Errorf("read response: %w", err))
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))

	gotJA4 := resp.Header.Get("X-Captured-JA4")
	gotJA3 := resp.Header.Get("X-Captured-JA3-Hash")
	ok := gotJA4 != "" && gotJA4 == m.Expected.JA4

	result := map[string]any{
		"profile_id":   m.ID,
		"expected_ja4": m.Expected.JA4,
		"got_ja4":      gotJA4,
		"expected_ja3": m.Expected.JA3Hash,
		"got_ja3_hash": gotJA3,
		"ja4_match":    ok,
		"ja3_match":    gotJA3 == m.Expected.JA3Hash,
		"ja3_unstable": m.Expected.JA3HashUnstable,
		"http_status":  resp.Status,
		"body":         string(body),
		"sni":          *sni,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(result)
	_ = os.WriteFile(filepath.Join(*profileDir, "verify-last.json"), mustJSON(result), 0o644)

	if !ok {
		fmt.Fprintf(os.Stderr, "VERIFY FAIL: ja4 expected %s got %q\n", m.Expected.JA4, gotJA4)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "VERIFY OK: ja4 match")
}

func mustJSON(v any) []byte {
	b, _ := json.MarshalIndent(v, "", "  ")
	return b
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
