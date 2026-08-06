# uTLS profile: brave

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d1515h2_acb858a92679_a87ad97598a9`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=brave`
