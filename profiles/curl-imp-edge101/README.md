# uTLS profile: curl-imp-edge101

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d1513h1_acb858a92679_de4a06bb82e3`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=curl-imp-edge101`
