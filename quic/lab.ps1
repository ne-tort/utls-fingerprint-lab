#!/usr/bin/env pwsh
# Entry point for the QUIC Initial fingerprint lab (sibling to parent TCP lab).
param(
    [Parameter(Position = 0)]
    [ValidateSet("build", "capture-listen", "parse", "list", "help")]
    [string]$Command = "help",
    [string]$Path = "",
    [string]$Listen = ":4433",
    [string]$Target = "unknown"
)

$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot
Set-Location $Root

function Build-Capture {
    $bin = Join-Path $Root "capture\bin"
    New-Item -ItemType Directory -Force -Path $bin | Out-Null
    Push-Location (Join-Path $Root "capture")
    try {
        Write-Host "building host quic-capture..."
        go build -trimpath -ldflags="-s -w" -o (Join-Path $bin "quic-capture.exe") .
        $env:GOOS = "linux"
        $env:GOARCH = "amd64"
        $env:CGO_ENABLED = "0"
        go build -trimpath -ldflags="-s -w" -o (Join-Path $bin "quic-capture-linux") .
        Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue
    }
    finally { Pop-Location }
    Write-Host "ok → capture/bin/"
}

function Ensure-HostBin {
    $exe = Join-Path $Root "capture\bin\quic-capture.exe"
    if (-not (Test-Path $exe)) { Build-Capture }
    return $exe
}

switch ($Command) {
    "help" {
        @"
quic lab commands:
  build            cross-compile capture (host exe + linux)
  capture-listen   UDP peek (-Listen :4433 -Target id)
  parse -Path f    offline parse one Initial .bin
  list             show targets.yaml ids

See README.md and docs/.
"@
    }
    "build" { Build-Capture }
    "capture-listen" {
        $exe = Ensure-HostBin
        $cap = Join-Path $Root "captures"
        $prof = Join-Path $Root "profiles"
        New-Item -ItemType Directory -Force -Path $cap, $prof | Out-Null
        & $exe -listen $Listen -out $cap -profiles $prof -default-target $Target -promote
    }
    "parse" {
        if (-not $Path) { throw "parse requires -Path <initial.bin>" }
        $exe = Ensure-HostBin
        & $exe -parse $Path
    }
    "list" {
        if (Get-Command yq -ErrorAction SilentlyContinue) {
            yq '.targets[].id' (Join-Path $Root "targets.yaml")
        } else {
            Select-String -Path (Join-Path $Root "targets.yaml") -Pattern '^\s+- id:' | ForEach-Object { ($_ -split 'id:')[1].Trim() }
        }
    }
}
