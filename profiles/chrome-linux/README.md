# uTLS profile: chrome-linux

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d1514h2_acb858a92679_806a8c22fdea`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=chrome-linux`
