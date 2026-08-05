# uTLS profile: curl-imp-tor145

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d1511h1_77608d807651_748f4c70de1c`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=curl-imp-tor145`
