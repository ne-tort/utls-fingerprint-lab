# uTLS profile: builtin-android

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t12d120600_f901973f4d4e_036209cd1ead`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=builtin-android`
