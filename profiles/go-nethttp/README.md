# uTLS profile: go-nethttp

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d1310h2_45672f90a73c_e5728521abd4`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=go-nethttp`
