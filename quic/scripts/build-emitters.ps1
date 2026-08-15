# Build linux emitters + capture for quic docker lab.
# ChromeParrot emitter is compiled from sing-box-lx (sagernet/quic-go with patches).
# uquic emitter uses lab _refs/uquic.
$ErrorActionPreference = "Stop"
$LabRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$RepoRoot = Resolve-Path (Join-Path $LabRoot "..\..\..")
$Out = Join-Path $LabRoot "bin"
New-Item -ItemType Directory -Force -Path $Out | Out-Null

$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"

Write-Host "building quic-capture (linux)..."
Push-Location (Join-Path $LabRoot "capture")
try {
  go build -trimpath -ldflags "-s -w" -o (Join-Path $Out "quic-capture") .
} finally { Pop-Location }

Write-Host "building chromeparrot + fromprofile emitters from sing-box-lx..."
$Tags = "with_gvisor,with_quic,with_wireguard,with_utls"
Push-Location $RepoRoot.Path
try {
  go build -trimpath -ldflags "-s -w -checklinkname=0" -tags $Tags `
    -o (Join-Path $Out "emit-chromeparrot") `
    ./lx-test/utls-fingerprint-docker/quic/emitters/chromeparrot
  go build -trimpath -ldflags "-s -w -checklinkname=0" -tags $Tags `
    -o (Join-Path $Out "emit-fromprofile") `
    ./lx-test/utls-fingerprint-docker/quic/emitters/fromprofile
} finally { Pop-Location }

Write-Host "building uquic emitter..."
Push-Location (Join-Path $LabRoot "emitters\uquic")
try {
  go mod tidy
  go build -trimpath -ldflags "-s -w" -o (Join-Path $Out "emit-uquic") .
} finally { Pop-Location }

Write-Host "building sing-box (linux, hy2 parity)..."
$SBTags = "with_gvisor,with_quic,with_wireguard,with_utls"
Push-Location $RepoRoot.Path
try {
  go build -trimpath -ldflags "-s -w -checklinkname=0" -tags $SBTags `
    -o (Join-Path $Out "sing-box") `
    ./cmd/sing-box
} finally { Pop-Location }

Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue
Write-Host "ok → $Out"
Get-ChildItem $Out | Format-Table Name, Length
