# uTLS profile: builtin-360

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t12d2008s2_ae54550e61e1_736b2a1ed4d3`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=builtin-360`
