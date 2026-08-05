# uTLS profile: curl-imp-safari184

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d2012h1_de3eb69493ac_e42f34c56612`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=curl-imp-safari184`
