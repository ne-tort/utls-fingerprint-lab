# uTLS profile: rust-rustls

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d1009h2_b9b30c653583_3a8073edd8ef`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=rust-rustls`
