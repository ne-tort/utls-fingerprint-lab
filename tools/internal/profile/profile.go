package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manifest is the importable lab profile contract.
// Primary artifact: clienthello.bin (full TLS record) → utls.Fingerprinter.RawClientHello.
type Manifest struct {
	ID     string `json:"id"`
	Format string `json:"format"` // utls-raw-clienthello-v1
	Source struct {
		CapturedAt string `json:"captured_at"`
		CaptureDir string `json:"capture_dir,omitempty"`
		Notes      string `json:"notes,omitempty"`
	} `json:"source"`
	Files struct {
		ClientHelloBin string `json:"clienthello_bin"`
		MetaJSON       string `json:"meta_json"`
	} `json:"files"`
	Expected struct {
		JA4             string `json:"ja4"`
		JA3Hash         string `json:"ja3_hash,omitempty"`
		JA3HashUnstable bool   `json:"ja3_hash_unstable"`
	} `json:"expected"`
	UTLS struct {
		ImportHow string `json:"import"`
	} `json:"utls"`
}

const FormatRawV1 = "utls-raw-clienthello-v1"

func Write(dir, id, captureDir string, ja4, ja3Hash string, hello []byte, metaJSON []byte) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "clienthello.bin"), hello, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), metaJSON, 0o644); err != nil {
		return err
	}
	m := Manifest{
		ID:     id,
		Format: FormatRawV1,
	}
	m.Source.CapturedAt = time.Now().UTC().Format(time.RFC3339)
	m.Source.CaptureDir = captureDir
	m.Files.ClientHelloBin = "clienthello.bin"
	m.Files.MetaJSON = "meta.json"
	m.Expected.JA4 = ja4
	m.Expected.JA3Hash = ja3Hash
	m.Expected.JA3HashUnstable = true // GREASE / shuffle
	m.UTLS.ImportHow = "utls.Fingerprinter{AllowBluntMimicry:true}.RawClientHello(clienthello.bin) → UClient(..., HelloCustom).ApplyPreset(spec)"
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "profile.json"), b, 0o644); err != nil {
		return err
	}
	readme := fmt.Sprintf(`# uTLS profile: %s

Format: %s

## Import (Go)

`+"```"+`go
raw, _ := os.ReadFile("clienthello.bin")
fp := &utls.Fingerprinter{AllowBluntMimicry: true}
spec, err := fp.RawClientHello(raw)
uconn := utls.UClient(rawConn, &utls.Config{ServerName: sni, InsecureSkipVerify: true}, utls.HelloCustom)
err = uconn.ApplyPreset(spec)
`+"```"+`

## Verify

`+"```"+`text
docker compose --profile verify run --rm verify-profile PROFILE_ID=%s
`+"```"+`

Expected JA4: %s
`, id, FormatRawV1, id, ja4)
	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0o644)
}

func Load(dir string) (*Manifest, []byte, error) {
	b, err := os.ReadFile(filepath.Join(dir, "profile.json"))
	if err != nil {
		return nil, nil, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, nil, err
	}
	binName := m.Files.ClientHelloBin
	if binName == "" {
		binName = "clienthello.bin"
	}
	hello, err := os.ReadFile(filepath.Join(dir, binName))
	if err != nil {
		return nil, nil, err
	}
	return &m, hello, nil
}
