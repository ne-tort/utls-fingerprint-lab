# uTLS profile: gnutls-cli

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d291200_7c1bf9677551_2cc26d266019`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=gnutls-cli`
