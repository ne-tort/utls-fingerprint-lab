# uTLS profile: builtin-safari

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d2012h2_de3eb69493ac_14788d8d241b`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=builtin-safari`
