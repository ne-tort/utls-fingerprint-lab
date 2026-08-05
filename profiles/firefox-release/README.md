# uTLS profile: firefox-release

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d1715h2_5b234860e130_3cbfd9057e0d`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=firefox-release`
