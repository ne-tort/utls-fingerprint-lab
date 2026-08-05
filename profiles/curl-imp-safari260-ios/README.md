# uTLS profile: curl-imp-safari260-ios

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d2013h1_de3eb69493ac_c258b721e490`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=curl-imp-safari260-ios`
