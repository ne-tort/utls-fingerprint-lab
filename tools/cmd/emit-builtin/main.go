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
	"strings"
	"time"

	utls "github.com/metacubex/utls"
)

func main() {
	name := flag.String("id", "chrome", "builtin id (see -list)")
	dialAddr := flag.String("dial", "capture:8443", "capture host:port")
	sni := flag.String("sni", "", "SNI (default: builtin-<id>.fp.lab.local)")
	list := flag.Bool("list", false, "print known builtin ids and exit")
	flag.Parse()
	if *list {
		for _, id := range knownIDs() {
			fmt.Println(id)
		}
		return
	}
	if *sni == "" {
		*sni = "builtin-" + sanitizeID(*name) + ".fp.lab.local"
	}
	helloID, err := mapID(*name)
	if err != nil {
		fatal(err)
	}

	raw, err := net.DialTimeout("tcp", *dialAddr, 15*time.Second)
	if err != nil {
		fatal(err)
	}
	defer raw.Close()
	_ = raw.SetDeadline(time.Now().Add(30 * time.Second))

	ucfg := &utls.Config{
		ServerName:         *sni,
		InsecureSkipVerify: true,
		NextProtos:         []string{"http/1.1"},
		MinVersion:         tls.VersionTLS12,
	}
	uconn := utls.UClient(raw, ucfg, helloID)
	if err := uconn.Handshake(); err != nil {
		fatal(err)
	}
	req, _ := http.NewRequest(http.MethodGet, "https://"+*sni+"/", nil)
	req.Header.Set("X-Target-Id", "builtin-"+sanitizeID(*name))
	req.Header.Set("Connection", "close")
	_ = req.Write(uconn)
	resp, err := http.ReadResponse(bufio.NewReader(uconn), req)
	if err != nil {
		fatal(err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	out := map[string]any{
		"builtin_id": *name,
		"utls_id":    helloID.Client + "/" + helloID.Version,
		"ja4":        resp.Header.Get("X-Captured-JA4"),
		"ja3_hash":   resp.Header.Get("X-Captured-JA3-Hash"),
		"sni":        *sni,
		"status":     resp.Status,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

func sanitizeID(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), "_", "-")
}

func knownIDs() []string {
	return []string{
		"chrome", "chrome_auto", "chrome_120", "chrome_131", "chrome_133",
		"firefox", "firefox_120",
		"edge", "edge_85", "edge_106",
		"safari", "ios",
		"android",
		"360", "qq",
		"randomized",
	}
}

func mapID(name string) (utls.ClientHelloID, error) {
	switch strings.ToLower(name) {
	case "chrome", "chrome_auto":
		return utls.HelloChrome_Auto, nil
	case "chrome_120":
		return utls.HelloChrome_120, nil
	case "chrome_131":
		return utls.HelloChrome_131, nil
	case "chrome_133":
		return utls.HelloChrome_133, nil
	case "firefox", "firefox_auto":
		return utls.HelloFirefox_Auto, nil
	case "firefox_120":
		return utls.HelloFirefox_120, nil
	case "edge", "edge_auto":
		return utls.HelloEdge_Auto, nil
	case "edge_85":
		return utls.HelloEdge_85, nil
	case "edge_106":
		return utls.HelloEdge_106, nil
	case "safari":
		return utls.HelloSafari_Auto, nil
	case "ios":
		return utls.HelloIOS_Auto, nil
	case "android":
		return utls.HelloAndroid_11_OkHttp, nil
	case "360":
		return utls.Hello360_Auto, nil
	case "qq":
		return utls.HelloQQ_Auto, nil
	case "randomized":
		return utls.HelloRandomized, nil
	default:
		return utls.ClientHelloID{}, fmt.Errorf("unknown id %q (use -list)", name)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
