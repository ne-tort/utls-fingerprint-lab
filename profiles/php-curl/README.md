# uTLS profile: php-curl

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d3011h2_8d44cdc55eec_8537cf56674e`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=php-curl`
