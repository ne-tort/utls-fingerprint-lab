# uTLS profile: builtin-firefox

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d1713h2_5b234860e130_5c2c66f702b0`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=builtin-firefox`
