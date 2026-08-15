#!/usr/bin/env pwsh
# Single entry for QUIC Initial fingerprint lab (Docker-first).
param(
    [Parameter(Position = 0)]
    [ValidateSet("build", "build-emitters", "capture-listen", "parse", "list", "up", "matrix", "compare", "help")]
    [string]$Command = "help",
    [string]$Path = "",
    [string]$Listen = ":4433",
    [string]$Target = "unknown",
    [string]$Profile = "matrix"
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

switch ($Command) {
    "help" {
        @"
quic lab (Docker-first):
  build-emitters   linux bins → bin/ (capture + chromeparrot + uquic)
  up               docker compose up capture
  matrix           build-emitters + compose --profile matrix run emitters + compare
  compare          python scripts/compare_profiles.py
  capture-listen   host UDP peek (dev)
  parse -Path f    offline parse
  list             targets.yaml ids
  build            host capture exe only
"@
    }
    "build" { Build-CaptureHost; Write-Host "ok" }
    "build-emitters" {
        & (Join-Path $Root "scripts\build-emitters.ps1")
    }
    "capture-listen" {
        $exe = Ensure-HostBin
        $cap = Join-Path $Root "captures"
        $prof = Join-Path $Root "profiles"
        New-Item -ItemType Directory -Force -Path $cap, $prof | Out-Null
        & $exe -listen $Listen -out $cap -profiles $prof -default-target $Target -promote
    }
    "parse" {
        if (-not $Path) { throw "parse requires -Path" }
        $exe = Ensure-HostBin
        & $exe -parse $Path
    }
    "list" {
        Select-String -Path (Join-Path $Root "targets.yaml") -Pattern '^\s+- id:' | ForEach-Object { ($_ -split 'id:')[1].Trim() }
    }
    "up" {
        if (-not (Test-Path (Join-Path $Root "bin\quic-capture"))) {
            & (Join-Path $Root "scripts\build-emitters.ps1")
        }
        docker compose -f compose.yaml up --build -d capture
    }
    "matrix" {
        & (Join-Path $Root "scripts\build-emitters.ps1")
        New-Item -ItemType Directory -Force -Path (Join-Path $Root "captures"), (Join-Path $Root "profiles") | Out-Null
        docker compose -f compose.yaml up --build -d capture
        Start-Sleep -Seconds 2
        foreach ($svc in @(
            "emit-chromeparrot",
            "emit-quicgo-plain",
            "emit-uquic-chrome146",
            "emit-uquic-chrome115",
            "emit-uquic-firefox116"
        )) {
            Write-Host "=== $svc ==="
            docker compose -f compose.yaml --profile matrix run --rm $svc
            Start-Sleep -Milliseconds 1500
        }
        Start-Sleep -Seconds 2
        python (Join-Path $Root "scripts\compare_profiles.py")
    }
    "compare" {
        python (Join-Path $Root "scripts\compare_profiles.py") @Args
    }
}
