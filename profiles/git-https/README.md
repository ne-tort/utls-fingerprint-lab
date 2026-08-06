# uTLS profile: git-https

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d3010h2_8d44cdc55eec_882d495ac381`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=git-https`
