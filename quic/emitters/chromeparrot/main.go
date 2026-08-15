package main

// Emit one QUIC Initial with sagernet/quic-go ChromeParrot (same stack as hy2 default).
// Build from sing-box-lx root via quic/scripts/build-emitters.ps1 (uses module replace).

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/sagernet/quic-go"
)

func main() {
	host := flag.String("host", "127.0.0.1", "capture host")
	port := flag.Int("port", 4433, "capture UDP port")
	sni := flag.String("sni", "fp.lab.local", "TLS SNI")
	parrot := flag.Bool("chrome-parrot", true, "enable ChromeParrot (false = plain quic-go)")
	flag.Parse()

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
	if *parrot {
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
	}, &quic.Config{ChromeParrot: *parrot})
	time.Sleep(400 * time.Millisecond)
	fmt.Printf("ok chrome-parrot=%v\n", *parrot)
}
