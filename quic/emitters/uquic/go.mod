module github.com/ne-tort/utls-fingerprint-lab/quic/emitters/uquic

go 1.24.0

require (
	github.com/refraction-networking/uquic v0.0.0
	github.com/refraction-networking/utls v1.7.4-0.20250521174854-63aeec73c564
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/refraction-networking/clienthellod v0.5.0-alpha2 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
)

replace github.com/refraction-networking/uquic => ../../../_refs/uquic
