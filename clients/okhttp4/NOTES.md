# okhttp4

JVM OkHttp 4.x — сверка с uTLS `HelloAndroid_11_OkHttp` / sing-box `fingerprint: android`.

## Run

```powershell
docker compose --profile wave3 run --rm okhttp4
$env:PROFILE_ID='okhttp4'; docker compose --profile verify run --rm -e PROFILE_ID=okhttp4 verify-profile
```
