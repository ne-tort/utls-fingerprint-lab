# hy2 outbound configs for Initial dump parity

JSON mirrors demux-quic-auth client pattern (**no `quic_auth`**). Capture is
dump-only; handshake fails — Initial is enough.

```powershell
./lab.ps1 build-emitters   # builds quic/bin/sing-box (linux) among other bins
./lab.ps1 hy2              # compose.hy2.yaml → profiles/hy2parrot + hy2plain
```

Expect: `hy2parrot` ≡ matrix `chromeparrot`; `hy2plain` ≈ `quicgo` (may include
extra `0x20`).
