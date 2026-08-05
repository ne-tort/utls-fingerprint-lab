# uTLS profile: curl-openssl

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d3110h2_d7c3e2abb617_6bebaf5329ac`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=curl-openssl`
