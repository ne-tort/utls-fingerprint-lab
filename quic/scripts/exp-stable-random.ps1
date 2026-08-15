#!/usr/bin/env pwsh
# Live host experiment: emit catalog shorts twice -> capture -> stable/random asserts.
param(
    [int]$Port = 4433,
    [string[]]$Shorts = @("chrome", "quic-go", "quic-go-datagram")
)

$ErrorActionPreference = "Stop"
$Root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $Root

function Ensure-WinCapture {
    $exe = Join-Path $Root "capture\bin\quic-capture.exe"
    if (-not (Test-Path $exe)) {
        New-Item -ItemType Directory -Force -Path (Join-Path $Root "capture\bin") | Out-Null
        Push-Location (Join-Path $Root "capture")
        try { go build -trimpath -ldflags "-s -w" -o $exe . }
        finally { Pop-Location }
    }
    return $exe
}

function Ensure-WinFromProfile {
    $exe = Join-Path $Root "bin\emit-fromprofile.exe"
    $repo = Resolve-Path (Join-Path $Root "..\..\..")
    Push-Location $repo.Path
    try {
        $env:CGO_ENABLED = "0"
        go build -trimpath -ldflags "-s -w -checklinkname=0" `
            -tags "with_gvisor,with_quic,with_wireguard,with_utls" `
            -o $exe `
            ./lx-test/utls-fingerprint-docker/quic/emitters/fromprofile
    } finally {
        Pop-Location
        Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
    }
    return $exe
}

$capture = Ensure-WinCapture
$emit = Ensure-WinFromProfile
New-Item -ItemType Directory -Force -Path (Join-Path $Root "captures"), (Join-Path $Root "profiles"), (Join-Path $Root "fixtures") | Out-Null
Get-ChildItem (Join-Path $Root "profiles") -Directory -Filter "exp-*" | Remove-Item -Recurse -Force -ErrorAction SilentlyContinue

Write-Host "starting capture on :$Port"
$stdout = Join-Path $Root "fixtures\exp-capture-stdout.log"
$stderr = Join-Path $Root "fixtures\exp-capture-stderr.log"
$capProc = Start-Process -FilePath $capture -ArgumentList @(
    "-listen", ":$Port",
    "-out", (Join-Path $Root "captures"),
    "-profiles", (Join-Path $Root "profiles"),
    "-default-target", "unknown",
    "-promote",
    "-gather-idle", "600ms"
) -PassThru -WindowStyle Hidden -RedirectStandardOutput $stdout -RedirectStandardError $stderr

Start-Sleep -Seconds 1
if ($capProc.HasExited) { throw "capture exited early; see $stderr" }

$extra = @()
try {
    foreach ($short in $Shorts) {
        $prof = Join-Path $Root "catalog\utls\$short.json"
        if (-not (Test-Path $prof)) { throw "missing $prof" }
        $doc = Get-Content $prof -Raw | ConvertFrom-Json
        if ($doc.status -eq "observation_only" -or $doc.emit.emit_kind -eq "match_only") {
            Write-Host "skip $short (not dialable)"
            continue
        }
        foreach ($letter in @("a", "b")) {
            $id = "exp-$short-$letter"
            Write-Host "=== emit $short -> $id ==="
            & $emit -profile $prof -host 127.0.0.1 -port $Port -sni "$id.fp.lab"
            if ($LASTEXITCODE -ne 0) { throw "emit failed for $short ($letter)" }
            Start-Sleep -Milliseconds 1000
        }
        $extra += "--extra-pair"
        $extra += "exp-$short-a,exp-$short-b"
        Start-Sleep -Milliseconds 500
    }
    Start-Sleep -Seconds 2
} finally {
    if (-not $capProc.HasExited) {
        Stop-Process -Id $capProc.Id -Force -ErrorAction SilentlyContinue
    }
}

foreach ($short in $Shorts) {
    foreach ($letter in @("a", "b")) {
        $id = "exp-$short-$letter"
        $p = Join-Path $Root "profiles\$id\profile.json"
        if (-not (Test-Path $p)) { throw "missing promoted profile $id" }
    }
}

Write-Host "=== stable/random experiment (archived + live) ==="
$prev = $ErrorActionPreference
$ErrorActionPreference = "Continue"
try {
    python (Join-Path $Root "scripts\experiment_stable_random.py") `
        --json-out (Join-Path $Root "fixtures\stable-random-experiment.json") `
        @extra
    if ($LASTEXITCODE -ne 0) { throw "stable/random experiment failed" }
} finally {
    $ErrorActionPreference = $prev
}

Write-Host "exp-stable-random OK"
