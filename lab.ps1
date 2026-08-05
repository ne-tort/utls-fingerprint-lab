#!/usr/bin/env pwsh
# Single entry point for the uTLS fingerprint lab.
param(
    [Parameter(Position = 0)]
    [ValidateSet("build", "list", "capture", "verify", "test", "catalog", "export", "clean", "help")]
    [string]$Command = "help",
    [string]$Id = "",
    [string]$Group = "",
    [string]$Status = "active"
)

$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot
Set-Location $Root
$env:DOCKER_BUILDKIT = "1"
$env:COMPOSE_DOCKER_CLI_BUILD = "1"
$env:COMPOSE_PROFILES = "capture,verify,tools"

function Invoke-Compose {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$ComposeArgs)
    & docker compose -f compose.yaml --project-name utls-lab @ComposeArgs
    if ($LASTEXITCODE -ne 0) { throw "docker compose failed: $($ComposeArgs -join ' ')" }
}

function Build-HostLinuxBins {
    Write-Host "cross-compiling linux tools (GOCACHE on host)..."
    $bin = Join-Path $Root "bin"
    $capBin = Join-Path $Root "capture\bin"
    $toolBin = Join-Path $Root "tools\bin"
    $goHttpBin = Join-Path $Root "clients\go-http\bin"
    New-Item -ItemType Directory -Force -Path $bin, $capBin, $toolBin, $goHttpBin | Out-Null
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"
    Push-Location (Join-Path $Root "capture")
    try { go build -trimpath -ldflags="-s -w" -o (Join-Path $capBin "capture-linux") . }
    finally { Pop-Location }
    Push-Location (Join-Path $Root "clients\go-http")
    try { go build -trimpath -ldflags="-s -w" -o (Join-Path $goHttpBin "go-http-linux") . }
    finally { Pop-Location }
    Push-Location (Join-Path $Root "tools")
    try {
        go get gopkg.in/yaml.v3@v3.0.1 2>$null | Out-Null
        go build -trimpath -ldflags="-s -w" -o (Join-Path $toolBin "verify-linux") ./cmd/verify
        go build -trimpath -ldflags="-s -w" -o (Join-Path $toolBin "emit-builtin-linux") ./cmd/emit-builtin
        go build -trimpath -ldflags="-s -w" -o (Join-Path $toolBin "labctl-linux") ./cmd/labctl
        $env:GOOS = "windows"
        go build -trimpath -ldflags="-s -w" -o (Join-Path $bin "labctl.exe") ./cmd/labctl
        $env:GOOS = "linux"
    } finally { Pop-Location }
}

function Ensure-Labctl {
    $exe = Join-Path $Root "bin\labctl.exe"
    if (-not (Test-Path $exe)) { Build-HostLinuxBins }
    return $exe
}

switch ($Command) {
    "help" {
        Write-Host @"
utls-fingerprint-lab

  ./lab.ps1 build                 Host-cross-compile + runtime images
  ./lab.ps1 list [-Status active] List targets from targets.yaml
  ./lab.ps1 capture [-Id|-Group]  Capture active targets
  ./lab.ps1 verify [-Id]          Replay-verify profiles
  ./lab.ps1 test                  Smoke subset
  ./lab.ps1 catalog|export|clean

See docs/EXTENDING.md and docs/IMPORT.md
"@
    }
    "build" {
        python (Join-Path $Root "scripts\gen-compose.py")
        Build-HostLinuxBins
        Invoke-Compose build capture tools
        Write-Host "build ok"
    }
    "list" {
        $labctl = Ensure-Labctl
        $a = @("-root", $Root, "list", "-status", $Status)
        if ($Group) { $a += @("-group", $Group) }
        & $labctl @a
    }
    "capture" {
        python (Join-Path $Root "scripts\gen-compose.py")
        Build-HostLinuxBins
        Invoke-Compose build capture tools
        $labctl = Ensure-Labctl
        $a = @("-root", $Root, "capture")
        if ($Id) { $a += @("-id", $Id) }
        if ($Group) { $a += @("-group", $Group) }
        & $labctl @a
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
        & $labctl -root $Root catalog
    }
    "verify" {
        Build-HostLinuxBins
        Invoke-Compose build capture tools
        $labctl = Ensure-Labctl
        $a = @("-root", $Root, "verify")
        if ($Id) { $a += @("-id", $Id) }
        & $labctl @a
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    }
    "catalog" {
        $labctl = Ensure-Labctl
        & $labctl -root $Root catalog
    }
    "export" {
        $labctl = Ensure-Labctl
        & $labctl -root $Root export
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    }
    "test" {
        Write-Host "== smoke test =="
        python (Join-Path $Root "scripts\gen-compose.py")
        Build-HostLinuxBins
        Invoke-Compose build capture tools
        $labctl = Ensure-Labctl
        foreach ($tid in @("openssl3", "curl-imp-chrome146", "builtin-chrome", "go-nethttp")) {
            Write-Host "--- capture $tid ---"
            & $labctl -root $Root capture -id $tid
            if ($LASTEXITCODE -ne 0) { throw "capture failed: $tid" }
        }
        foreach ($tid in @("openssl3", "curl-imp-chrome146", "builtin-chrome", "go-nethttp")) {
            Write-Host "--- verify $tid ---"
            & $labctl -root $Root verify -id $tid
            if ($LASTEXITCODE -ne 0) { throw "verify failed: $tid" }
        }
        & $labctl -root $Root catalog
        Write-Host "SMOKE OK"
    }
    "clean" {
        if (Test-Path (Join-Path $Root "captures")) {
            Remove-Item -Recurse -Force (Join-Path $Root "captures\*")
        }
        Get-ChildItem (Join-Path $Root "profiles") -Directory -Filter "verify-*" -EA SilentlyContinue | Remove-Item -Recurse -Force
        Get-ChildItem (Join-Path $Root "profiles") -Recurse -Filter "verify-last.json" -EA SilentlyContinue | Remove-Item -Force
        Write-Host "cleaned"
    }
}
