# uTLS profile: openssl3

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d310900_d7c3e2abb617_1f22a2ca17c4`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=openssl3`
