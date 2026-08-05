package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	id := env("TARGET_ID", "go-nethttp")
	dialHost := env("DIAL_HOST", "capture:8443")
	sni := id + ".fp.lab.local"
	insecure := os.Getenv("INSECURE") != "" && os.Getenv("INSECURE") != "0"

	transport := &http.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := net.Dialer{Timeout: 15 * time.Second}
			raw, err := d.DialContext(ctx, "tcp", dialHost)
			if err != nil {
				return nil, err
			}
			cfg := &tls.Config{
				InsecureSkipVerify: insecure,
				ServerName:         sni,
				MinVersion:         tls.VersionTLS12,
				NextProtos:         []string{"h2", "http/1.1"},
			}
			c := tls.Client(raw, cfg)
			if err := c.HandshakeContext(ctx); err != nil {
				_ = raw.Close()
				return nil, err
			}
			return c, nil
		},
		ForceAttemptHTTP2: true,
	}
	client := &http.Client{Transport: transport, Timeout: 20 * time.Second}
	req, err := http.NewRequest(http.MethodGet, "https://"+sni+"/", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("X-Target-Id", id)
	resp, err := client.Do(req)
	if err != nil {
		// ClientHello is captured before verify/handshake finishes on server side.
		fmt.Fprintf(os.Stderr, "request error (CH may still be saved): %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	fmt.Printf("%s %s\n", resp.Status, string(body))
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
