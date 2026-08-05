# uTLS profile: curl-imp-safari260

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d2012h1_09fdf58d427f_d0a99439f9b1`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=curl-imp-safari260`
