# Importing lab profiles into sing-box / metacubex uTLS

## Format

`profile.json` field `format`: **`utls-raw-clienthello-v1`**

| File | Use |
|------|-----|
| `clienthello.bin` | Full TLS record containing ClientHello — **source of truth** |
| `profile.json` | Manifest: id, expected JA4, format |
| `meta.json` | Lab metadata (optional at runtime) |

## Go (metacubex/utls)

```go
package example

import (
	"os"

	utls "github.com/metacubex/utls"
)

func ApplyLabProfile(rawConn net.Conn, profileDir, sni string) (*utls.UConn, error) {
	bin, err := os.ReadFile(filepath.Join(profileDir, "clienthello.bin"))
	if err != nil {
		return nil, err
	}
	fp := &utls.Fingerprinter{AllowBluntMimicry: true}
	spec, err := fp.RawClientHello(bin)
	if err != nil {
		return nil, err
	}
	// Optional: rewrite SNI extension to the real destination.
	for _, ext := range spec.Extensions {
		if s, ok := ext.(*utls.SNIExtension); ok {
			s.ServerName = sni
		}
	}
	uconn := utls.UClient(rawConn, &utls.Config{
		ServerName:         sni,
		InsecureSkipVerify: false,
	}, utls.HelloCustom)
	if err := uconn.ApplyPreset(spec); err != nil {
		return nil, err
	}
	return uconn, nil
}
```

## sing-box future wiring (intent)

1. Ship selected `profiles/<id>/` as assets or embed.
2. Config key (proposal): `tls.utls.profile` / `fingerprint: "lab:<id>"` resolving to a lab profile pack.
3. Runtime path: same `Fingerprinter.RawClientHello` → `HelloCustom` (already proven by `tools/cmd/verify`).
4. Prefer **lab profiles** for ground-truth browsers (Chromium, curl-impersonate); keep stock `chrome`/`firefox` Hello*_Auto for lightweight defaults.

## Verification before shipping

```powershell
./lab.ps1 verify -Id chromium-stable
./lab.ps1 verify -Id curl-imp-chrome146
```

JA4 must match `profile.json` → `expected.ja4`. JA3 may drift (GREASE); marked `ja3_hash_unstable`.

## Notes

- Capture server forces ALPN `http/1.1` so verify can read JA4 headers. JA4 may show `h1` even when the real client prefers `h2`. Cipher/extension material remains valid for ClientHello mimicry.
- GREASE / extension shuffle: expect JA3 variance; pin on JA4 + structural verify.
