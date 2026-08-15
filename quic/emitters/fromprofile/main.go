// Emit QUIC Initial from a quic-utls-profile-v1 JSON (engine path).
// Structured apply is not implemented yet — engine flags are the stable dial path.
//
// Build from sing-box-lx root via quic/scripts/build-emitters.ps1.
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/sagernet/quic-go"
)

type profileDoc struct {
	Format string `json:"format"`
	Short  string `json:"short"`
	Status string `json:"status"`
	Emit   struct {
		EmitKind    string `json:"emit_kind"`
		UquicPreset string `json:"uquic_preset"`
		Engine      *struct {
			ChromeParrot    bool `json:"chrome_parrot"`
			EnableDatagrams bool `json:"enable_datagrams"`
			ZeroLengthSCID  bool `json:"zero_length_scid"`
		} `json:"engine"`
		Notes string `json:"notes"`
	} `json:"emit"`
	Auth struct {
		Capable                    bool   `json:"capable"`
		Channel                    string `json:"channel"`
		RequiresDemuxFingerprint   bool   `json:"requires_demux_fingerprint"`
		EmitHook                   string `json:"emit_hook"`
	} `json:"auth"`
}

func main() {
	host := flag.String("host", "127.0.0.1", "capture host")
	port := flag.Int("port", 4433, "capture UDP port")
	sni := flag.String("sni", "fp.lab.local", "TLS SNI")
	profilePath := flag.String("profile", "", "path to quic-utls-profile-v1 JSON")
	flag.Parse()
	if *profilePath == "" {
		fmt.Fprintln(os.Stderr, "usage: fromprofile -profile catalog/utls/chrome.json")
		os.Exit(2)
	}

	raw, err := os.ReadFile(*profilePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	var doc profileDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if doc.Format != "quic-utls-profile-v1" {
		fmt.Fprintf(os.Stderr, "unexpected format %q\n", doc.Format)
		os.Exit(2)
	}
	if doc.Status == "observation_only" || doc.Emit.EmitKind == "match_only" {
		fmt.Fprintf(os.Stderr, "refuse dial: short=%s status=%s emit_kind=%s\n",
			doc.Short, doc.Status, doc.Emit.EmitKind)
		os.Exit(3)
	}
	if doc.Emit.EmitKind == "uquic_preset" || doc.Emit.EmitKind == "structured" {
		fmt.Fprintf(os.Stderr, "engine path only today; emit_kind=%s not wired in fromprofile\n",
			doc.Emit.EmitKind)
		os.Exit(4)
	}
	if doc.Emit.Engine == nil {
		fmt.Fprintln(os.Stderr, "profile missing emit.engine")
		os.Exit(2)
	}
	eng := doc.Emit.Engine

	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer conn.Close()

	var tr *quic.Transport
	if eng.ZeroLengthSCID || eng.ChromeParrot {
		tr = &quic.Transport{Conn: conn, ConnectionIDGenerator: quic.ZeroLengthConnectionIDGenerator{}}
	} else {
		tr = &quic.Transport{Conn: conn}
	}
	defer tr.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _ = tr.Dial(ctx, raddr, &tls.Config{
		ServerName:         *sni,
		InsecureSkipVerify: true,
		NextProtos:         []string{"h3"},
	}, &quic.Config{
		ChromeParrot:    eng.ChromeParrot,
		EnableDatagrams: eng.EnableDatagrams,
	})
	time.Sleep(400 * time.Millisecond)
	fmt.Printf("ok short=%s emit_kind=%s chrome_parrot=%v datagrams=%v zero_scid=%v auth_capable=%v auth_channel=%q demux_fp_required=%v\n",
		doc.Short, doc.Emit.EmitKind, eng.ChromeParrot, eng.EnableDatagrams, eng.ZeroLengthSCID,
		doc.Auth.Capable, doc.Auth.Channel, doc.Auth.RequiresDemuxFingerprint)
}
