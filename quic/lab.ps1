#!/usr/bin/env pwsh
# Single entry for QUIC Initial fingerprint lab (Docker-first).
param(
    [Parameter(Position = 0)]
    [ValidateSet("build", "build-emitters", "capture-listen", "parse", "list", "up", "matrix", "roundtrip", "live", "compare", "help")]
    [string]$Command = "help",
    [string]$Path = "",
    [string]$Listen = ":4433",
    [string]$Target = "unknown",
    [ValidateSet("aioquic", "curl", "chromium", "all")]
    [string]$Client = "aioquic"
)

$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot
Set-Location $Root
$env:DOCKER_BUILDKIT = "1"

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

switch ($Command) {
    "help" {
        @"
quic lab (Docker-first):
  build-emitters   linux bins → bin/
  up               docker compose up capture
  matrix           emitters: parrot/uquic/aioquic + compare
  roundtrip        prove emit recipes reproduce structural identity
  live -Client X   live clients overlay (aioquic|curl|chromium|all)
  compare          TP table over profiles/
  capture-listen   host UDP peek (dev)
  parse -Path f    offline parse
  list             targets.yaml ids
  build            host capture exe only

Docs: docs/REPLAY_AND_EMIT.md · docs/PYTHON_VS_GO_CAPTURE.md
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
        docker compose -f compose.yaml up --build -d capture
    }
    "matrix" {
        Ensure-LinuxBins
        New-Item -ItemType Directory -Force -Path (Join-Path $Root "captures"), (Join-Path $Root "profiles") | Out-Null
        docker compose -f compose.yaml up --build -d capture
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
            docker compose -f compose.yaml --profile matrix run --rm --build $svc
            Start-Sleep -Milliseconds 1500
        }
        Start-Sleep -Seconds 2
        python (Join-Path $Root "scripts\compare_profiles.py")
    }
    "roundtrip" {
        Ensure-LinuxBins
        New-Item -ItemType Directory -Force -Path (Join-Path $Root "captures"), (Join-Path $Root "profiles") | Out-Null
        # clear prior roundtrip ids
        foreach ($id in @("chromeparrot", "chromeparrot-b", "uquic146", "uquic146-b")) {
            $p = Join-Path $Root "profiles\$id"
            if (Test-Path $p) { Remove-Item -Recurse -Force $p }
        }
        docker compose -f compose.yaml up --build -d capture
        Start-Sleep -Seconds 2
        docker compose -f compose.yaml --profile matrix run --rm emit-chromeparrot
        Start-Sleep -Seconds 1
        docker compose -f compose.yaml --profile roundtrip run --rm emit-chromeparrot-b
        Start-Sleep -Seconds 1
        docker compose -f compose.yaml --profile matrix run --rm emit-uquic-chrome146
        Start-Sleep -Seconds 1
        docker compose -f compose.yaml --profile roundtrip run --rm emit-uquic146-b
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
        docker compose -f compose.yaml -f compose.live-clients.yaml up --build -d capture
        Start-Sleep -Seconds 2
        $svcs = switch ($Client) {
            "aioquic" { @("emit-aioquic") }
            "curl" { @("emit-curl-quiche") }
            "chromium" { @("emit-chromium-h3") }
            "all" { @("emit-aioquic", "emit-curl-quiche", "emit-chromium-h3") }
        }
        foreach ($svc in $svcs) {
            Write-Host "=== live $svc ==="
            docker compose -f compose.yaml -f compose.live-clients.yaml --profile live run --rm --build $svc
            Start-Sleep -Seconds 2
        }
        python (Join-Path $Root "scripts\compare_profiles.py")
    }
    "compare" {
        python (Join-Path $Root "scripts\compare_profiles.py")
    }
}
