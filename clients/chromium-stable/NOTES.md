# chromium-stable

Browser target runs from `compose.yaml` service `chromium-stable`
(image `zenika/alpine-chrome:124`), not a custom client image.

```powershell
./lab.ps1 capture -Id chromium-stable
./lab.ps1 verify -Id chromium-stable
```

SNI: `chromium-stable.fp.lab.local` (DNS alias on `capture`).  
Registry: `targets.yaml` → group `browsers`.
