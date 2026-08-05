# uTLS profile: curl-imp-firefox133

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d1714h1_5b234860e130_eeeea6562960`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=curl-imp-firefox133`
