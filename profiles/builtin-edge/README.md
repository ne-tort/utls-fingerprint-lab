# uTLS profile: builtin-edge

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d1513h2_acb858a92679_de4a06bb82e3`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=builtin-edge`
