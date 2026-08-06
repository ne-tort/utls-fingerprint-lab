package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	utls "github.com/metacubex/utls"
)

var (
	listenAddr = flag.String("listen", ":8443", "listen address")
	outDir     = flag.String("out", "/captures", "output directory")
	profileDir = flag.String("profiles", "/profiles", "importable profiles directory")
	targetHint = flag.String("default-target", "unknown", "default target_id if client omits X-Target-Id")
	seq        uint64
)

func main() {
	flag.Parse()
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatal(err)
	}
	cert, err := loadOrMakeCert("certs")
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("capture listening on %s → %s", *listenAddr, *outDir)

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handle(c, cert)
	}
}

func handle(c net.Conn, cert tls.Certificate) {
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(30 * time.Second))

	rawHello, rest, err := readClientHello(c)
	if err != nil {
		log.Printf("read ClientHello from %s: %v", c.RemoteAddr(), err)
		return
	}

	meta, err := analyze(rawHello)
	if err != nil {
		log.Printf("analyze: %v", err)
		meta = &Meta{Notes: "analyze error: " + err.Error()}
	}
	meta.Peer = c.RemoteAddr().String()
	meta.CapturedAt = time.Now().UTC().Format(time.RFC3339)
	meta.TargetID = targetFromSNI(meta.SNI, *targetHint)

	id := fmt.Sprintf("%d-%s", atomic.AddUint64(&seq, 1), sanitize(meta.TargetID))
	dir := filepath.Join(*outDir, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("mkdir: %v", err)
		return
	}
	binPath := filepath.Join(dir, "clienthello.bin")
	if err := os.WriteFile(binPath, rawHello, 0o644); err != nil {
		log.Printf("write bin: %v", err)
		return
	}
	meta.ClientHelloPath = "clienthello.bin"
	sum := sha256.Sum256(rawHello)
	meta.ClientHelloSHA256 = hex.EncodeToString(sum[:])

	// uTLS Fingerprinter expects full TLS record or handshake message.
	fp := &utls.Fingerprinter{AllowBluntMimicry: true, AlwaysAddPadding: false}
	if spec, err := fp.FingerprintClientHello(rawHello); err != nil {
		meta.UTLSFingerprinterOK = false
		meta.Notes = strings.TrimSpace(meta.Notes + " fingerprinter: " + err.Error())
		_ = os.WriteFile(filepath.Join(dir, "spec_error.txt"), []byte(err.Error()), 0o644)
	} else {
		meta.UTLSFingerprinterOK = true
		_ = os.WriteFile(filepath.Join(dir, "spec.txt"), []byte(fmt.Sprintf("%#v\n", spec)), 0o644)
	}

	enc, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, "meta.json"), enc, 0o644)
	if err := writeImportableProfile(*profileDir, meta.TargetID, dir, rawHello, enc, meta); err != nil {
		log.Printf("profile write: %v", err)
	} else {
		log.Printf("profile updated profiles/%s", sanitize(meta.TargetID))
	}
	log.Printf("captured %s ja4=%s ja3_hash=%s peer=%s", id, meta.JA4, meta.JA3Hash, meta.Peer)

	// Complete handshake so -k HTTP clients succeed (replay buffered bytes).
	replay := &prefixConn{Conn: c, prefix: append(rawHello, rest...)}
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// Prefer HTTP/1.1 when offered; advertise h2 so h2-only ClientHellos
		// (gRPC, some HTTP/2 stacks) still complete. Lab then speaks HTTP/1.1
		// framing either way so verify can read X-Captured-JA4 headers.
		NextProtos: []string{"http/1.1", "h2"},
		MinVersion: tls.VersionTLS12,
	}
	sc := tls.Server(replay, tlsCfg)
	if err := sc.Handshake(); err != nil {
		// Client may abort after CH (cert pin etc.) — fingerprint already saved.
		log.Printf("handshake after capture (%s): %v", id, err)
		return
	}
	capturedMeta := meta
	br := bufio.NewReader(sc)
	req, err := http.ReadRequest(br)
	if err != nil {
		log.Printf("http read after capture (%s): %v", id, err)
		return
	}
	var body strings.Builder
	body.WriteString("utls-fingerprint-lab ok\n")
	hdr := make(http.Header)
	hdr.Set("Content-Type", "text/plain; charset=utf-8")
	hdr.Set("Connection", "close")
	hdr.Set("X-Captured-JA4", capturedMeta.JA4)
	hdr.Set("X-Captured-JA3-Hash", capturedMeta.JA3Hash)
	hdr.Set("X-Captured-Target", capturedMeta.TargetID)
	hdr.Set("X-Captured-ID", id)
	if t := req.Header.Get("X-Target-Id"); t != "" {
		hdr.Set("X-Echo-Target", t)
	}
	resp := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        hdr,
		Body:          io.NopCloser(strings.NewReader(body.String())),
		ContentLength: int64(body.Len()),
		Close:         true,
	}
	if err := resp.Write(sc); err != nil {
		log.Printf("http write after capture (%s): %v", id, err)
	}
	_ = sc.Close()
}

type Meta struct {
	TargetID            string   `json:"target_id"`
	CapturedAt          string   `json:"captured_at"`
	Peer                string   `json:"peer"`
	SNI                 string   `json:"sni"`
	ALPN                []string `json:"alpn"`
	TLSVersionRecord    string   `json:"tls_version_record"`
	TLSVersionHello     string   `json:"tls_version_hello"`
	JA3                 string   `json:"ja3"`
	JA3Hash             string   `json:"ja3_hash"`
	JA4                 string   `json:"ja4"`
	CipherSuites        []string `json:"cipher_suites"`
	Extensions          []string `json:"extensions"`
	SupportedGroups     []string `json:"supported_groups"`
	SignatureAlgorithms []string `json:"signature_algorithms"`
	ClientHelloSHA256   string   `json:"clienthello_sha256"`
	ClientHelloPath     string   `json:"clienthello_path"`
	UTLSFingerprinterOK bool     `json:"utls_fingerprinter_ok"`
	Notes               string   `json:"notes"`
}

func analyze(record []byte) (*Meta, error) {
	if len(record) < 5 {
		return nil, fmt.Errorf("short record")
	}
	recVer := binary.BigEndian.Uint16(record[1:3])
	hs := record[5:]
	if len(hs) < 4 || hs[0] != 0x01 {
		return nil, fmt.Errorf("not ClientHello handshake")
	}
	bodyLen := int(hs[1])<<16 | int(hs[2])<<8 | int(hs[3])
	body := hs[4:]
	if len(body) < bodyLen {
		return nil, fmt.Errorf("truncated hello body")
	}
	body = body[:bodyLen]
	if len(body) < 34 {
		return nil, fmt.Errorf("hello too short")
	}
	helloVer := binary.BigEndian.Uint16(body[0:2])
	// skip random 32
	i := 34
	if i >= len(body) {
		return nil, fmt.Errorf("no session id")
	}
	sidLen := int(body[i])
	i++
	i += sidLen
	if i+2 > len(body) {
		return nil, fmt.Errorf("no ciphers")
	}
	csLen := int(binary.BigEndian.Uint16(body[i : i+2]))
	i += 2
	if i+csLen > len(body) {
		return nil, fmt.Errorf("bad cipher len")
	}
	var ciphers []uint16
	for j := 0; j+2 <= csLen; j += 2 {
		ciphers = append(ciphers, binary.BigEndian.Uint16(body[i+j:i+j+2]))
	}
	i += csLen
	if i >= len(body) {
		return nil, fmt.Errorf("no compression")
	}
	compLen := int(body[i])
	i++
	i += compLen
	var (
		extIDs   []uint16
		groups   []uint16
		sigs     []uint16
		versions []uint16
		sni      string
		alpn     []string
	)
	if i+2 <= len(body) {
		extLen := int(binary.BigEndian.Uint16(body[i : i+2]))
		i += 2
		end := i + extLen
		if end > len(body) {
			end = len(body)
		}
		for i+4 <= end {
			eid := binary.BigEndian.Uint16(body[i : i+2])
			el := int(binary.BigEndian.Uint16(body[i+2 : i+4]))
			i += 4
			if i+el > end {
				break
			}
			data := body[i : i+el]
			i += el
			extIDs = append(extIDs, eid)
			switch eid {
			case 0x0000: // SNI
				if len(data) >= 5 {
					n := int(binary.BigEndian.Uint16(data[3:5]))
					if 5+n <= len(data) {
						sni = string(data[5 : 5+n])
					}
				}
			case 0x000a: // supported_groups
				if len(data) >= 2 {
					ll := int(binary.BigEndian.Uint16(data[0:2]))
					for j := 2; j+2 <= 2+ll && j+2 <= len(data); j += 2 {
						groups = append(groups, binary.BigEndian.Uint16(data[j:j+2]))
					}
				}
			case 0x000d: // signature_algorithms
				if len(data) >= 2 {
					ll := int(binary.BigEndian.Uint16(data[0:2]))
					for j := 2; j+2 <= 2+ll && j+2 <= len(data); j += 2 {
						sigs = append(sigs, binary.BigEndian.Uint16(data[j:j+2]))
					}
				}
			case 0x0010: // ALPN
				if len(data) >= 2 {
					ll := int(binary.BigEndian.Uint16(data[0:2]))
					j := 2
					for j < 2+ll && j < len(data) {
						n := int(data[j])
						j++
						if j+n > len(data) {
							break
						}
						alpn = append(alpn, string(data[j:j+n]))
						j += n
					}
				}
			case 0x002b: // supported_versions
				if len(data) >= 1 {
					ll := int(data[0])
					for j := 1; j+2 <= 1+ll && j+2 <= len(data); j += 2 {
						versions = append(versions, binary.BigEndian.Uint16(data[j:j+2]))
					}
				}
			}
		}
	}

	ja3Str, ja3Hash := buildJA3(helloVer, ciphers, extIDs, groups, nil)
	ja4 := buildJA4(versions, helloVer, sni != "", ciphers, extIDs, sigs, alpn)

	m := &Meta{
		SNI:                 sni,
		ALPN:                alpn,
		TLSVersionRecord:    fmt.Sprintf("0x%04x", recVer),
		TLSVersionHello:     fmt.Sprintf("0x%04x", helloVer),
		JA3:                 ja3Str,
		JA3Hash:             ja3Hash,
		JA4:                 ja4,
		CipherSuites:        hexList16(ciphers),
		Extensions:          hexList16(extIDs),
		SupportedGroups:     hexList16(groups),
		SignatureAlgorithms: hexList16(sigs),
	}
	return m, nil
}

func buildJA3(version uint16, ciphers, extensions, groups []uint16, points []uint8) (string, string) {
	// Classic JA3: SSLVersion,Ciphers,Extensions,EllipticCurves,EllipticCurvePointFormats
	parts := []string{
		fmt.Sprintf("%d", version),
		joinUint16(ciphers, true),
		joinUint16(extensions, true),
		joinUint16(groups, true),
		joinUint8(points),
	}
	s := strings.Join(parts, ",")
	sum := sha256.Sum256([]byte(s)) // note: original JA3 uses MD5; we keep both
	_ = sum
	h := md5Hex(s)
	return s, h
}

func md5Hex(s string) string {
	// inline md5 to avoid extra import noise in analyze path — use crypto/md5
	return md5String(s)
}

func buildJA4(supportedVersions []uint16, helloVer uint16, hasSNI bool, ciphers, exts, sigs []uint16, alpn []string) string {
	// Simplified FoxIO JA4_a / JA4_b / JA4_c for TCP.
	ver := helloVer
	for _, v := range supportedVersions {
		if !isGREASE16(v) && v > ver {
			ver = v
		}
	}
	verStr := "00"
	switch ver {
	case 0x0304:
		verStr = "13"
	case 0x0303:
		verStr = "12"
	case 0x0302:
		verStr = "11"
	case 0x0301:
		verStr = "10"
	}
	sniFlag := "i"
	if hasSNI {
		sniFlag = "d"
	}
	var cClean []uint16
	for _, c := range ciphers {
		if !isGREASE16(c) {
			cClean = append(cClean, c)
		}
	}
	var eClean []uint16
	for _, e := range exts {
		if !isGREASE16(e) && e != 0x0000 && e != 0x0010 { // JA4 excludes SNI and ALPN from ext list
			eClean = append(eClean, e)
		}
	}
	alpnFirst := "00"
	if len(alpn) > 0 && len(alpn[0]) > 0 {
		a := alpn[0]
		if len(a) == 1 {
			alpnFirst = string(a[0]) + "0"
		} else {
			alpnFirst = string(a[0]) + string(a[len(a)-1])
		}
		// replace non-alnum
		alpnFirst = sanitizeALPN(alpnFirst)
	}
	a := fmt.Sprintf("t%s%s%02d%02d%s", verStr, sniFlag, min(len(cClean), 99), min(len(eClean), 99), alpnFirst)

	// JA4_b: sha256 of cipher list sorted? FoxIO: unsorted as appeared, hex comma-separated, first 12 of sha256
	bSrc := joinUint16HexLower(cClean, false)
	bSum := sha256.Sum256([]byte(bSrc))
	b := hex.EncodeToString(bSum[:])[:12]

	// JA4_c: sorted extensions + '_' + unsorted sigalgs
	eSorted := append([]uint16(nil), eClean...)
	sortUint16(eSorted)
	cSrc := joinUint16HexLower(eSorted, false)
	if len(sigs) > 0 {
		var sClean []uint16
		for _, s := range sigs {
			if !isGREASE16(s) {
				sClean = append(sClean, s)
			}
		}
		cSrc += "_" + joinUint16HexLower(sClean, false)
	}
	cSum := sha256.Sum256([]byte(cSrc))
	c := hex.EncodeToString(cSum[:])[:12]
	return a + "_" + b + "_" + c
}

func readClientHello(r io.Reader) (record []byte, rest []byte, err error) {
	hdr := make([]byte, 5)
	if _, err = io.ReadFull(r, hdr); err != nil {
		return nil, nil, err
	}
	if hdr[0] != 0x16 {
		return nil, nil, fmt.Errorf("not handshake record (type=%d)", hdr[0])
	}
	n := int(binary.BigEndian.Uint16(hdr[3:5]))
	body := make([]byte, n)
	if _, err = io.ReadFull(r, body); err != nil {
		return nil, nil, err
	}
	rec := append(hdr, body...)
	// If fragmented hellos appear later we ignore for v1.
	return rec, nil, nil
}

func loadOrMakeCert(dir string) (tls.Certificate, error) {
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")
	if _, err := os.Stat(certFile); err == nil {
		return tls.LoadX509KeyPair(certFile, keyFile)
	}
	_ = os.MkdirAll(dir, 0o755)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "fp.lab.local"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames: []string{
			"fp.lab.local", "capture", "localhost",
			"openssl3.fp.lab.local", "curl-openssl.fp.lab.local",
			"go-nethttp.fp.lab.local", "python-requests.fp.lab.local",
			"chromium-stable.fp.lab.local",
		},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEM := pemEncode("CERTIFICATE", der)
	keyPEM := pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))
	_ = os.WriteFile(certFile, certPEM, 0o644)
	_ = os.WriteFile(keyFile, keyPEM, 0o600)
	return tls.X509KeyPair(certPEM, keyPEM)
}
