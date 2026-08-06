# uTLS profile: grpc-go

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d0909h2_8e4c1465eaca_e7c285222651`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=grpc-go`
