$ErrorActionPreference = "Stop"

$repo = Split-Path -Parent $PSScriptRoot
$outDir = Join-Path $repo "bin"
$exe = Join-Path $outDir "simplerouter.exe"

# Locate the Go toolchain. A freshly-installed Go may be on the persistent
# Machine PATH but missing from an already-open shell, so fall back to the
# usual install locations rather than aborting with a confusing error.
$go = (Get-Command go -ErrorAction SilentlyContinue).Source
if (-not $go) {
    foreach ($candidate in @(
        "$env:ProgramFiles\Go\bin\go.exe",
        "$env:LOCALAPPDATA\Programs\Go\bin\go.exe",
        "$HOME\go\bin\go.exe",
        "C:\Go\bin\go.exe"
    )) {
        if (Test-Path $candidate) { $go = $candidate; break }
    }
}
if (-not $go) {
    throw "Could not find 'go'. Install Go (https://go.dev/dl/) or open a new terminal so PATH refreshes, then rerun this script."
}

New-Item -ItemType Directory -Force -Path $outDir | Out-Null
Push-Location $repo
try {
    & $go build -buildvcs=false -o $exe ./cmd/simplerouter
    if ($LASTEXITCODE -ne 0) { throw "go build failed (exit $LASTEXITCODE)" }
} finally {
    Pop-Location
}

$installDir = Join-Path $HOME ".local\bin"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Copy-Item -Force $exe (Join-Path $installDir "simplerouter.exe")

Write-Host "Installed simplerouter to $installDir"
