# uTLS profile: okhttp4

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d1112h2_d3eabd2901df_89e9e4bb3419`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=okhttp4`
