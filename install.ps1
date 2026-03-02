# Looty Install Script for Windows
#
# One-liner install:
#   irm https://raw.githubusercontent.com/YOUR_GITHUB_USER/BlipSync/main/install.ps1 | iex
#
# Or download manually and run:
#   .\install.ps1 -Repo "YOUR_GITHUB_USER/BlipSync"

param(
    [string]$Version = "latest",
    [string]$Repo = "YOUR_GITHUB_USER/BlipSync"
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "         LOOTY INSTALLER" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Detect OS architecture
$Arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "x86_64" } else { "arm64" }

Write-Host "Detected: Windows ($Arch)" -ForegroundColor Gray

# Get latest release
$ApiUrl = "https://api.github.com/repos/$Repo/releases/$Version"

Write-Host "Finding release..." -ForegroundColor Yellow

try {
    $Response = Invoke-RestMethod -Uri $ApiUrl -UseBasicParsing
    $TagName = $Response.tag_name
    Write-Host "Found version: $TagName" -ForegroundColor Green
} catch {
    Write-Host "Error: Could not fetch release from $Repo" -ForegroundColor Red
    Write-Host "Make sure the repository exists and has releases." -ForegroundColor Yellow
    exit 1
}

# Find matching asset
$AssetName = "looty-windows-$Arch.exe"
$DownloadUrl = $Response.assets | Where-Object { $_.name -eq $AssetName } | Select-Object -First 1 -ExpandProperty browser_download_url

if (-not $DownloadUrl) {
    Write-Host "Error: $AssetName not found in release" -ForegroundColor Red
    Write-Host "Available assets:" -ForegroundColor Yellow
    $Response.assets | ForEach-Object { Write-Host "  - $($_.name)" }
    exit 1
}

# Download
Write-Host "Downloading $AssetName..." -ForegroundColor Yellow
$TempDir = Join-Path $env:TEMP "looty-install-$(Get-Random)"
New-Item -ItemType Directory -Force -Path $TempDir | Out-Null
$DownloadPath = Join-Path $TempDir "looty.exe"

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $DownloadPath -UseBasicParsing
} catch {
    Write-Host "Error downloading: $_" -ForegroundColor Red
    exit 1
}

$FileSize = (Get-Item $DownloadPath).Length / 1MB
Write-Host "Downloaded ($([math]::Round($FileSize, 2)) MB)" -ForegroundColor Green

# Install
$InstallDir = Join-Path $env:LOCALAPPDATA "Programs\looty"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Copy-Item $DownloadPath $InstallDir -Force
Write-Host "Installed to: $InstallDir" -ForegroundColor Green

# Add to PATH if needed
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
$PathSeparator = if ($CurrentPath[-1] -eq ';') { "" } else { ";" }

if ($CurrentPath -notlike "*$InstallDir*") {
    $NewPath = "$CurrentPath$PathSeparator$InstallDir"
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    Write-Host "Added to PATH!" -ForegroundColor Green
    Write-Host ""
    Write-Host "IMPORTANT: Restart your terminal or run:" -ForegroundColor Yellow
    Write-Host "  `$env:Path = [Environment]::GetEnvironmentVariable('Path','User')" -ForegroundColor Cyan
} else {
    Write-Host "Already in PATH" -ForegroundColor Gray
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  SUCCESS! Looty installed!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Now open a NEW terminal and run:" -ForegroundColor Yellow
Write-Host "  looty" -ForegroundColor Cyan
Write-Host ""
Write-Host "This will serve the CURRENT FOLDER on your network!" -ForegroundColor Gray

# Cleanup
Remove-Item $TempDir -Recurse -Force -ErrorAction SilentlyContinue