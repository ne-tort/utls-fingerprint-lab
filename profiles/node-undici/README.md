# uTLS profile: node-undici

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d521000_ca35c0050d43_8e6e362c5eac`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=node-undici`
