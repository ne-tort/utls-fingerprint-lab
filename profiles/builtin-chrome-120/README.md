# uTLS profile: builtin-chrome-120

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d1514h2_acb858a92679_02713d6af862`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=builtin-chrome-120`
