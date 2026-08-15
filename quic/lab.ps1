#!/usr/bin/env pwsh
# Single entry for QUIC Initial fingerprint lab (Docker-first).
param(
    [Parameter(Position = 0)]
    [ValidateSet("build", "build-emitters", "capture-listen", "parse", "list", "up", "matrix", "matrix-ext", "roundtrip", "live", "ja4", "hy2", "tuic", "host-browsers", "compare", "unify", "exp-stable", "export", "help")]
    [string]$Action = "help",
    [string]$Path = "",
    [string]$Listen = ":4433",
    [string]$Target = "unknown",
    [ValidateSet("aioquic", "curl", "chromium", "chromiumfresh", "firefox", "yandex", "all")]
    [string]$Client = "aioquic"
)

$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot
Set-Location $Root
$env:DOCKER_BUILDKIT = "1"

# Docker writes progress to stderr; under Stop that becomes a terminating error.
# Pass args as an array so -f/--profile are not bound as PowerShell parameters.
function Invoke-Docker([string[]]$DockerArgs) {
    $prev = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        & docker @DockerArgs 2>&1 | ForEach-Object { "$_" }
        if ($LASTEXITCODE -ne 0) {
            throw "docker $($DockerArgs -join ' ') failed (exit $LASTEXITCODE)"
        }
    } finally {
        $ErrorActionPreference = $prev
    }
}

function Build-CaptureHost {
    $bin = Join-Path $Root "capture\bin"
    New-Item -ItemType Directory -Force -Path $bin | Out-Null
    Push-Location (Join-Path $Root "capture")
    try {
        go build -trimpath -ldflags="-s -w" -o (Join-Path $bin "quic-capture.exe") .
    } finally { Pop-Location }
}

function Ensure-HostBin {
    $exe = Join-Path $Root "capture\bin\quic-capture.exe"
    if (-not (Test-Path $exe)) { Build-CaptureHost }
    return $exe
}

function Ensure-LinuxBins {
    if (-not (Test-Path (Join-Path $Root "bin\quic-capture"))) {
        & (Join-Path $Root "scripts\build-emitters.ps1")
    }
}

function Ensure-SingBoxLinux {
    $sb = Join-Path $Root "bin\sing-box"
    if (-not (Test-Path $sb)) {
        Write-Host "sing-box linux missing -> build-emitters (includes sing-box)"
        & (Join-Path $Root "scripts\build-emitters.ps1")
    }
    if (-not (Test-Path $sb)) { throw "quic/bin/sing-box not built" }
}

switch ($Action) {
    "help" {
        @"
quic lab (Docker-first):
  build-emitters   linux bins -> bin/
  up               docker compose up capture
  matrix           emitters: parrot/uquic/aioquic + compare
  matrix-ext       firefox 116A/B/C + quicgo-datagram
  roundtrip        prove emit recipes reproduce structural identity
  live -Client X   aioquic|curl|chromium|chromiumfresh|firefox|yandex|all
  ja4              annotate expected.ja4 via ja4plus (compose.ja4.yaml)
  hy2              real hy2 outbound Initial vs chromeparrot/quicgo
  tuic             real tuic outbound Initial vs hy2 / chromeparrot
  host-browsers    Windows Chrome / Edge / Yandex on :4433
  compare          TP table over profiles/
  unify            extract drafts + match quic-utls catalog vs profiles/
  exp-stable       live emit x2 + assert stable identity / random entropy
  export           dist/export for future product sync (lab-only prep)
  capture-listen   host UDP peek (dev)
  parse -Path f    offline parse
  list             targets.yaml ids
  build            host capture exe only

Docs: docs/UTLS_PROFILE.md · docs/REPLAY_AND_EMIT.md · docs/PYTHON_VS_GO_CAPTURE.md
"@
    }
    "build" { Build-CaptureHost; Write-Host "ok" }
    "build-emitters" { & (Join-Path $Root "scripts\build-emitters.ps1") }
    "capture-listen" {
        $exe = Ensure-HostBin
        New-Item -ItemType Directory -Force -Path (Join-Path $Root "captures"), (Join-Path $Root "profiles") | Out-Null
        & $exe -listen $Listen -out (Join-Path $Root "captures") -profiles (Join-Path $Root "profiles") -default-target $Target -promote
    }
    "parse" {
        if (-not $Path) { throw "parse requires -Path" }
        & (Ensure-HostBin) -parse $Path
    }
    "list" {
        Select-String -Path (Join-Path $Root "targets.yaml") -Pattern '^\s+- id:' | ForEach-Object { ($_ -split 'id:')[1].Trim() }
    }
    "up" {
        Ensure-LinuxBins
        Invoke-Docker @("compose", "-f", "compose.yaml", "up", "--build", "-d", "capture")
    }
    "matrix" {
        Ensure-LinuxBins
        New-Item -ItemType Directory -Force -Path (Join-Path $Root "captures"), (Join-Path $Root "profiles") | Out-Null
        Invoke-Docker @("compose", "-f", "compose.yaml", "up", "--build", "-d", "capture")
        Start-Sleep -Seconds 2
        foreach ($svc in @(
            "emit-chromeparrot",
            "emit-quicgo-plain",
            "emit-uquic-chrome146",
            "emit-uquic-chrome115",
            "emit-uquic-firefox116",
            "emit-aioquic"
        )) {
            Write-Host "=== $svc ==="
            Invoke-Docker @("compose", "-f", "compose.yaml", "--profile", "matrix", "run", "--rm", "--build", $svc)
            Start-Sleep -Milliseconds 1500
        }
        Start-Sleep -Seconds 2
        python (Join-Path $Root "scripts\compare_profiles.py")
    }
    "matrix-ext" {
        Ensure-LinuxBins
        New-Item -ItemType Directory -Force -Path (Join-Path $Root "captures"), (Join-Path $Root "profiles") | Out-Null
        Invoke-Docker @("compose", "-f", "compose.yaml", "up", "--build", "-d", "capture")
        Start-Sleep -Seconds 2
        foreach ($svc in @(
            "emit-quicgo-datagram",
            "emit-uquic-firefox116a",
            "emit-uquic-firefox116b",
            "emit-uquic-firefox116c"
        )) {
            Write-Host "=== $svc ==="
            Invoke-Docker @("compose", "-f", "compose.yaml", "--profile", "matrix-ext", "run", "--rm", "--build", $svc)
            Start-Sleep -Milliseconds 1500
        }
        Start-Sleep -Seconds 2
        if (Test-Path (Join-Path $Root "profiles\hy2plain")) {
            Write-Host "=== structural quicgodg vs hy2plain ==="
            python (Join-Path $Root "scripts\compare_structural.py") `
                (Join-Path $Root "profiles\quicgodg") `
                (Join-Path $Root "profiles\hy2plain")
        }
        python (Join-Path $Root "scripts\compare_profiles.py")
    }
    "roundtrip" {
        Ensure-LinuxBins
        New-Item -ItemType Directory -Force -Path (Join-Path $Root "captures"), (Join-Path $Root "profiles") | Out-Null
        foreach ($id in @("chromeparrot", "chromeparrot-b", "uquic146", "uquic146-b")) {
            $p = Join-Path $Root "profiles\$id"
            if (Test-Path $p) { Remove-Item -Recurse -Force $p }
        }
        Invoke-Docker @("compose", "-f", "compose.yaml", "up", "--build", "-d", "capture")
        Start-Sleep -Seconds 2
        Invoke-Docker @("compose", "-f", "compose.yaml", "--profile", "matrix", "run", "--rm", "emit-chromeparrot")
        Start-Sleep -Seconds 1
        Invoke-Docker @("compose", "-f", "compose.yaml", "--profile", "roundtrip", "run", "--rm", "emit-chromeparrot-b")
        Start-Sleep -Seconds 1
        Invoke-Docker @("compose", "-f", "compose.yaml", "--profile", "matrix", "run", "--rm", "emit-uquic-chrome146")
        Start-Sleep -Seconds 1
        Invoke-Docker @("compose", "-f", "compose.yaml", "--profile", "roundtrip", "run", "--rm", "emit-uquic146-b")
        Start-Sleep -Seconds 2
        Write-Host "=== structural chromeparrot vs chromeparrot-b ==="
        python (Join-Path $Root "scripts\compare_structural.py") `
            (Join-Path $Root "profiles\chromeparrot") `
            (Join-Path $Root "profiles\chromeparrot-b")
        $c1 = $LASTEXITCODE
        Write-Host "=== structural uquic146 vs uquic146-b ==="
        python (Join-Path $Root "scripts\compare_structural.py") `
            (Join-Path $Root "profiles\uquic146") `
            (Join-Path $Root "profiles\uquic146-b")
        $c2 = $LASTEXITCODE
        if ($c1 -ne 0 -or $c2 -ne 0) { throw "roundtrip structural mismatch" }
        Write-Host "roundtrip OK"
    }
    "live" {
        Ensure-LinuxBins
        New-Item -ItemType Directory -Force -Path (Join-Path $Root "captures"), (Join-Path $Root "profiles") | Out-Null
        Invoke-Docker @("compose", "-f", "compose.yaml", "-f", "compose.live-clients.yaml", "up", "--build", "-d", "capture")
        Start-Sleep -Seconds 2
        $svcs = switch ($Client) {
            "aioquic" { @("emit-aioquic") }
            "curl" { @("emit-curl-quiche") }
            "chromium" { @("emit-chromium-h3") }
            "chromiumfresh" { @("emit-chromium-fresh") }
            "firefox" { @("emit-firefox-h3") }
            "yandex" { @("emit-yandex-h3") }
            "all" { @("emit-aioquic", "emit-curl-quiche", "emit-chromium-h3", "emit-chromium-fresh", "emit-firefox-h3", "emit-yandex-h3") }
        }
        foreach ($svc in $svcs) {
            Write-Host "=== live $svc ==="
            Invoke-Docker @("compose", "-f", "compose.yaml", "-f", "compose.live-clients.yaml", "--profile", "live", "run", "--rm", "--build", $svc)
            Start-Sleep -Seconds 2
        }
        python (Join-Path $Root "scripts\compare_profiles.py")
    }
    "ja4" {
        Invoke-Docker @("compose", "-f", "compose.ja4.yaml", "--profile", "ja4", "run", "--rm", "--build", "ja4-annotate")
    }
    "hy2" {
        Ensure-LinuxBins
        Ensure-SingBoxLinux
        New-Item -ItemType Directory -Force -Path (Join-Path $Root "captures"), (Join-Path $Root "profiles") | Out-Null
        foreach ($id in @("hy2parrot", "hy2plain")) {
            $p = Join-Path $Root "profiles\$id"
            if (Test-Path $p) { Remove-Item -Recurse -Force $p }
        }
        Invoke-Docker @("compose", "-f", "compose.yaml", "-f", "compose.hy2.yaml", "--profile", "hy2", "up", "--build", "-d", "capture", "hy2-parrot", "hy2-plain")
        Start-Sleep -Seconds 2
        Write-Host "=== trigger hy2-parrot ==="
        Invoke-Docker @("compose", "-f", "compose.yaml", "-f", "compose.hy2.yaml", "--profile", "hy2", "run", "--rm", "trigger-hy2-parrot")
        Start-Sleep -Seconds 2
        Write-Host "=== trigger hy2-plain ==="
        Invoke-Docker @("compose", "-f", "compose.yaml", "-f", "compose.hy2.yaml", "--profile", "hy2", "run", "--rm", "trigger-hy2-plain")
        Start-Sleep -Seconds 2
        Write-Host "=== structural hy2parrot vs chromeparrot ==="
        if (-not (Test-Path (Join-Path $Root "profiles\chromeparrot"))) {
            Write-Warning "profiles/chromeparrot missing - run matrix first for baseline"
        } else {
            python (Join-Path $Root "scripts\compare_structural.py") `
                (Join-Path $Root "profiles\hy2parrot") `
                (Join-Path $Root "profiles\chromeparrot")
            if ($LASTEXITCODE -ne 0) { throw "hy2parrot != chromeparrot structurally" }
        }
        Write-Host "=== structural hy2plain vs quicgo ==="
        if (Test-Path (Join-Path $Root "profiles\quicgo")) {
            python (Join-Path $Root "scripts\compare_structural.py") `
                (Join-Path $Root "profiles\hy2plain") `
                (Join-Path $Root "profiles\quicgo")
            if ($LASTEXITCODE -ne 0) {
                Write-Warning "hy2plain != quicgo (expected: hy2 may add 0x20; see RESULTS_MATRIX)"
            }
        }
        Invoke-Docker @("compose", "-f", "compose.yaml", "-f", "compose.hy2.yaml", "--profile", "hy2", "down")
        Write-Host "hy2 parity OK (parrot matched chromeparrot)"
    }
    "tuic" {
        Ensure-LinuxBins
        Ensure-SingBoxLinux
        New-Item -ItemType Directory -Force -Path (Join-Path $Root "captures"), (Join-Path $Root "profiles") | Out-Null
        foreach ($id in @("tuicparrot", "tuicplain")) {
            $p = Join-Path $Root "profiles\$id"
            if (Test-Path $p) { Remove-Item -Recurse -Force $p }
        }
        Invoke-Docker @("compose", "-f", "compose.yaml", "-f", "compose.tuic.yaml", "--profile", "tuic", "up", "--build", "-d", "capture", "tuic-parrot", "tuic-plain")
        Start-Sleep -Seconds 2
        Write-Host "=== trigger tuic-parrot ==="
        Invoke-Docker @("compose", "-f", "compose.yaml", "-f", "compose.tuic.yaml", "--profile", "tuic", "run", "--rm", "trigger-tuic-parrot")
        Start-Sleep -Seconds 2
        Write-Host "=== trigger tuic-plain ==="
        Invoke-Docker @("compose", "-f", "compose.yaml", "-f", "compose.tuic.yaml", "--profile", "tuic", "run", "--rm", "trigger-tuic-plain")
        Start-Sleep -Seconds 2
        if (Test-Path (Join-Path $Root "profiles\chromeparrot")) {
            Write-Host "=== structural tuicparrot vs chromeparrot ==="
            python (Join-Path $Root "scripts\compare_structural.py") `
                (Join-Path $Root "profiles\tuicparrot") `
                (Join-Path $Root "profiles\chromeparrot")
        }
        if (Test-Path (Join-Path $Root "profiles\hy2plain")) {
            Write-Host "=== structural tuicplain vs hy2plain ==="
            python (Join-Path $Root "scripts\compare_structural.py") `
                (Join-Path $Root "profiles\tuicplain") `
                (Join-Path $Root "profiles\hy2plain")
        }
        Invoke-Docker @("compose", "-f", "compose.yaml", "-f", "compose.tuic.yaml", "--profile", "tuic", "down")
        Write-Host "tuic parity done"
    }
    "host-browsers" {
        & (Join-Path $Root "scripts\host-browsers.ps1")
    }
    "compare" {
        python (Join-Path $Root "scripts\compare_profiles.py")
    }
    "unify" {
        New-Item -ItemType Directory -Force -Path (Join-Path $Root "catalog\utls\_drafts") | Out-Null
        New-Item -ItemType Directory -Force -Path (Join-Path $Root "fixtures") | Out-Null
        Write-Host "=== sync curated catalog/utls ==="
        $prev = $ErrorActionPreference
        $ErrorActionPreference = "Continue"
        try {
            python (Join-Path $Root "scripts\sync_utls_catalog.py")
            if ($LASTEXITCODE -ne 0) { throw "sync_utls_catalog failed" }
            Write-Host "=== extract observation -> utls drafts ==="
            python (Join-Path $Root "scripts\extract_utls_profile.py")
            if ($LASTEXITCODE -ne 0) { throw "extract_utls_profile failed" }
            Write-Host "=== match catalog/utls vs profiles/ (strict-all) ==="
            python (Join-Path $Root "scripts\match_utls_catalog.py") `
                --strict-all `
                --json-out (Join-Path $Root "fixtures\utls-catalog-match.json")
            if ($LASTEXITCODE -ne 0) { throw "unify: catalog match failed" }
        } finally {
            $ErrorActionPreference = $prev
        }
        Write-Host "unify OK - see docs/UTLS_PROFILE.md"
    }
    "exp-stable" {
        & (Join-Path $Root "scripts\exp-stable-random.ps1")
    }
    "export" {
        $prev = $ErrorActionPreference
        $ErrorActionPreference = "Continue"
        try {
            python (Join-Path $Root "scripts\export_dist.py") --sync-first
            if ($LASTEXITCODE -ne 0) { throw "export_dist failed" }
        } finally {
            $ErrorActionPreference = $prev
        }
        Write-Host "export OK -> dist/export (lab prep; product sync later)"
    }
}
