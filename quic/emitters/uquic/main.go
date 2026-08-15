package main

// Emit one QUIC Initial using refraction-networking/uquic presets toward capture.
// Lab-only tool; does not modify sing-box.

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	quic "github.com/refraction-networking/uquic"
	tls "github.com/refraction-networking/utls"
)

func main() {
	host := flag.String("host", "127.0.0.1", "capture host")
	port := flag.Int("port", 4433, "capture UDP port")
	sni := flag.String("sni", "fp.lab.local", "TLS SNI")
	preset := flag.String("preset", "chrome-146", "uquic preset: chrome-146|chrome-115|firefox-116|firefox-116a|firefox-116b|firefox-116c")
	flag.Parse()

	id, err := presetID(*preset)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	spec, err := quic.QUICID2Spec(id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	udp, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer udp.Close()

	tr := &quic.UTransport{
		Transport: &quic.Transport{Conn: udp},
		QUICSpec:  &spec,
	}
	defer tr.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _ = tr.Dial(ctx, raddr, &tls.Config{
		ServerName:         *sni,
		InsecureSkipVerify: true,
		NextProtos:         []string{"h3"},
	}, nil)
	time.Sleep(400 * time.Millisecond)
	fmt.Printf("ok preset=%s\n", *preset)
}

func presetID(name string) (quic.QUICID, error) {
	switch name {
	case "chrome-146", "chrome":
		return quic.QUICChrome_146, nil
	case "chrome-115":
		return quic.QUICChrome_115, nil
	case "firefox-116", "firefox":
		return quic.QUICFirefox_116, nil
	case "firefox-116a":
		return quic.QUICFirefox_116A, nil
	case "firefox-116b":
		return quic.QUICFirefox_116B, nil
	case "firefox-116c":
		return quic.QUICFirefox_116C, nil
	default:
		return quic.QUICID{}, fmt.Errorf("unknown preset %q", name)
	}
}
