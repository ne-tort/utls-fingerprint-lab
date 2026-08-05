# uTLS profile: java-jdk

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d311100_cdf1781c2960_7e1102d2036b`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=java-jdk`
