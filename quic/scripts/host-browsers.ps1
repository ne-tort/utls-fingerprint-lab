# Capture QUIC Initials from Windows-installed Chromium browsers (host, not Docker).
# Usage: from quic/:  .\scripts\host-browsers.ps1
param(
    [int]$Port = 4433,
    [int]$PerBrowserSeconds = 25
)

$ErrorActionPreference = "Stop"
$Root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $Root

$exe = Join-Path $Root "capture\bin\quic-capture.exe"
if (-not (Test-Path $exe)) { & (Join-Path $Root "lab.ps1") build }
if (-not (Test-Path $exe)) { throw "missing $exe" }

function Resolve-Browser([string]$Id) {
    $cands = switch ($Id) {
        "winchrome" {
            @(
                "$env:ProgramFiles\Google\Chrome\Application\chrome.exe",
                "${env:ProgramFiles(x86)}\Google\Chrome\Application\chrome.exe"
            )
        }
        "winedge" {
            @(
                "${env:ProgramFiles(x86)}\Microsoft\Edge\Application\msedge.exe",
                "$env:ProgramFiles\Microsoft\Edge\Application\msedge.exe"
            )
        }
        "winyandex" {
            @(
                "$env:ProgramFiles\Yandex\YandexBrowser\Application\browser.exe",
                "${env:ProgramFiles(x86)}\Yandex\YandexBrowser\Application\browser.exe",
                "$env:LOCALAPPDATA\Yandex\YandexBrowser\Application\browser.exe"
            )
        }
    }
    foreach ($p in $cands) { if (Test-Path $p) { return $p } }
    throw "browser not found for $Id"
}

$browsers = @(
    @{ Id = "winchrome";  Sni = "winchrome.fp.lab";  Path = (Resolve-Browser "winchrome") }
    @{ Id = "winedge";    Sni = "winedge.fp.lab";    Path = (Resolve-Browser "winedge") }
    @{ Id = "winyandex";  Sni = "winyandex.fp.lab";  Path = (Resolve-Browser "winyandex") }
)

New-Item -ItemType Directory -Force -Path (Join-Path $Root "captures"), (Join-Path $Root "profiles") | Out-Null
foreach ($b in $browsers) {
    $p = Join-Path $Root "profiles\$($b.Id)"
    if (Test-Path $p) { Remove-Item -Recurse -Force $p }
}

Get-NetUDPEndpoint -LocalPort $Port -ErrorAction SilentlyContinue | ForEach-Object {
    try { Stop-Process -Id $_.OwningProcess -Force -ErrorAction SilentlyContinue } catch {}
}

$errLog = Join-Path $Root "captures\host-capture.err.log"
Write-Host "starting capture on UDP :$Port"
$cap = Start-Process -FilePath $exe -ArgumentList @(
    "-listen=:$Port",
    "-out=$(Join-Path $Root 'captures')",
    "-profiles=$(Join-Path $Root 'profiles')",
    "-default-target=unknown",
    "-promote=true",
    "-gather-idle=1500ms"
) -PassThru -WindowStyle Hidden -RedirectStandardError $errLog
Start-Sleep -Seconds 1
if ($cap.HasExited) { throw "capture exited early; see $errLog" }

try {
    foreach ($b in $browsers) {
        Write-Host "=== $($b.Id) -> $($b.Sni) ==="
        $ud = Join-Path $env:TEMP "quic-fp-$($b.Id)-$(Get-Random)"
        New-Item -ItemType Directory -Force -Path $ud | Out-Null
        # Call operator keeps spaces inside --host-resolver-rules intact
        # (Start-Process -ArgumentList re-parses and breaks MAP rules).
        $job = Start-Job -ScriptBlock {
            param($Bin, $Ud, $Sni, $Port)
            & $Bin `
                "--headless=new" `
                "--disable-gpu" `
                "--no-first-run" `
                "--no-default-browser-check" `
                "--disable-extensions" `
                "--user-data-dir=$Ud" `
                "--enable-quic" `
                "--quic-version=h3" `
                "--origin-to-force-quic-on=${Sni}:${Port}" `
                "--host-resolver-rules=MAP $Sni 127.0.0.1" `
                "--ignore-certificate-errors" `
                "--dump-dom" `
                "https://${Sni}:${Port}/" 2>&1 | Out-Null
        } -ArgumentList $b.Path, $ud, $b.Sni, $Port

        $waitSec = if ($b.Id -eq "winyandex") { [Math]::Max($PerBrowserSeconds, 35) } else { $PerBrowserSeconds }
        $deadline = (Get-Date).AddSeconds($waitSec)
        while ((Get-Date) -lt $deadline) {
            if (Test-Path (Join-Path $Root "profiles\$($b.Id)\profile.json")) { break }
            if ($job.State -ne "Running") {
                Start-Sleep -Seconds 2
                break
            }
            Start-Sleep -Milliseconds 400
        }
        # Yandex may finish dump-dom after job ends; give promote a moment
        Start-Sleep -Seconds 2
        if (-not (Test-Path (Join-Path $Root "profiles\$($b.Id)\profile.json"))) {
            Start-Sleep -Seconds 3
        }
        Stop-Job $job -ErrorAction SilentlyContinue
        Remove-Job $job -Force -ErrorAction SilentlyContinue
        # kill leftover browser processes for this user-data-dir
        Get-CimInstance Win32_Process -ErrorAction SilentlyContinue |
            Where-Object { $_.CommandLine -and $_.CommandLine -like "*$ud*" } |
            ForEach-Object { Stop-Process -Id $_.ProcessId -Force -ErrorAction SilentlyContinue }
        Start-Sleep -Seconds 1
        if (Test-Path (Join-Path $Root "profiles\$($b.Id)\profile.json")) {
            Write-Host "OK $($b.Id)"
        } else {
            Write-Warning "no profile for $($b.Id)"
        }
        Remove-Item -Recurse -Force $ud -ErrorAction SilentlyContinue
    }
} finally {
    if (-not $cap.HasExited) { Stop-Process -Id $cap.Id -Force -ErrorAction SilentlyContinue }
}

$cmp = Join-Path $Root "scripts\compare_structural.py"
Write-Host "`n=== compare ==="
foreach ($id in @("winchrome", "winedge", "winyandex")) {
    $prof = Join-Path $Root "profiles\$id"
    if (-not (Test-Path (Join-Path $prof "profile.json"))) { continue }
    $j = Get-Content (Join-Path $prof "profile.json") -Raw | ConvertFrom-Json
    Write-Host ("{0} tp=[{1}]" -f $id, ($j.expected.tp_id_set -join ","))
    Write-Host "--- $id vs chromeparrot ---"
    python $cmp $prof (Join-Path $Root "profiles\chromeparrot")
}
if ((Test-Path (Join-Path $Root "profiles\winchrome\profile.json")) -and
    (Test-Path (Join-Path $Root "profiles\chromiumfresh\profile.json"))) {
    Write-Host "--- winchrome vs chromiumfresh ---"
    python $cmp (Join-Path $Root "profiles\winchrome") (Join-Path $Root "profiles\chromiumfresh")
}
if ((Test-Path (Join-Path $Root "profiles\winyandex\profile.json")) -and
    (Test-Path (Join-Path $Root "profiles\yandex\profile.json"))) {
    Write-Host "--- winyandex vs docker yandex ---"
    python $cmp (Join-Path $Root "profiles\winyandex") (Join-Path $Root "profiles\yandex")
}
if ((Test-Path (Join-Path $Root "profiles\winchrome\profile.json")) -and
    (Test-Path (Join-Path $Root "profiles\winedge\profile.json"))) {
    Write-Host "--- winchrome vs winedge ---"
    python $cmp (Join-Path $Root "profiles\winchrome") (Join-Path $Root "profiles\winedge")
}
