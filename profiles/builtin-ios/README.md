# uTLS profile: builtin-ios

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d2611h2_1de9e775a572_845d286b0d67`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=builtin-ios`
