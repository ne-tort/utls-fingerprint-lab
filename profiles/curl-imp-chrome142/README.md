# uTLS profile: curl-imp-chrome142

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d1513h1_acb858a92679_0a20fe35d3a5`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=curl-imp-chrome142`
