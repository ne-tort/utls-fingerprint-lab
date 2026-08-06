package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Dial capture with the gRPC-Go TLS stack so ClientHello matches typical gRPC traffic.
// The RPC layer will fail; capture already has the ClientHello after WithBlock dial.
func main() {
	id := env("TARGET_ID", "grpc-go")
	dial := env("DIAL_HOST", "capture:8443")
	sni := id + ".fp.lab.local"

	tlsCfg := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         sni,
		MinVersion:         tls.VersionTLS12,
		NextProtos:         []string{"h2"},
	}
	creds := credentials.NewTLS(tlsCfg)

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	//nolint:staticcheck // DialContext+WithBlock forces outbound TLS for lab capture
	conn, err := grpc.DialContext(ctx, dial,
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
		grpc.WithAuthority(sni),
		grpc.WithReturnConnectionError(),
	)
	if err != nil {
		// Expected: capture is not a gRPC server; ClientHello should still be on disk.
		fmt.Fprintf(os.Stderr, "grpc dial (expected fail after CH): %v\n", err)
		fmt.Println("grpc-go dial attempted")
		return
	}
	_ = conn.Close()
	fmt.Println("grpc-go connected (unexpected)")
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
