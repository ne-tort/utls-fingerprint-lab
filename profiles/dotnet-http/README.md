# uTLS profile: dotnet-http

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d121000_3c962d0d4049_f6eeaf393d2e`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=dotnet-http`
