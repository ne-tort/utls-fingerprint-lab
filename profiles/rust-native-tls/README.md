# uTLS profile: rust-native-tls

Format: `utls-raw-clienthello-v1`

Import: `utls.Fingerprinter.RawClientHello(clienthello.bin)` then `ApplyPreset`.

Expected JA4: `t13d311000_d7c3e2abb617_d41ae481755e`

Verify: `docker compose --profile verify run --rm verify-profile` with `PROFILE_ID=rust-native-tls`
