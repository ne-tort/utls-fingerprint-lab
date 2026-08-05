# uTLS profile: python-httpx

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d171100_fd0eb050439a_ecd0401ec68b`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=python-httpx`
