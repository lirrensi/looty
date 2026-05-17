# Looty Latest Release Downloader (Windows)
# Usage: irm https://raw.githubusercontent.com/lirrensi/looty/main/scripts/get-latest.ps1 | iex

$Repo = "lirrensi/looty"
$ApiUrl = "https://api.github.com/repos/$Repo/releases/latest"

# Detect platform
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "arm64" }
$Platform = "windows"
$Suffix = "$Platform-$Arch"
$Ext = "zip"
$AssetName = "looty-$Suffix.$Ext"

Write-Host "📦 Detected platform: $Platform-$Arch"
Write-Host "🐱 Fetching latest Looty release..."

# Get download URL from GitHub API
try {
    $Response = Invoke-RestMethod -Uri $ApiUrl -UseBasicParsing
    $Asset = $Response.assets | Where-Object { $_.name -eq $AssetName }
    
    if (-not $Asset) {
        Write-Host "❌ Could not find asset: $AssetName"
        Write-Host "Available assets:"
        $Response.assets | Where-Object { $_.name -like "looty-*" } | ForEach-Object { Write-Host "  - $($_.name)" }
        exit 1
    }
    
    $DownloadUrl = $Asset.browser_download_url
} catch {
    Write-Host "❌ Failed to fetch release info: $_"
    exit 1
}

# Download
$TempFile = Join-Path $env:TEMP $AssetName
Write-Host "⬇️  Downloading: $AssetName"
Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempFile -UseBasicParsing

# Extract
$ExtractPath = Join-Path $env:TEMP "looty-extract"
if (Test-Path $ExtractPath) { Remove-Item -Recurse -Force $ExtractPath }
New-Item -ItemType Directory -Path $ExtractPath | Out-Null

Expand-Archive -Path $TempFile -DestinationPath $ExtractPath -Force

# Install to user's bin directory
$InstallDir = "$env:LOCALAPPDATA\Programs\Looty"
if (-not (Test-Path $InstallDir)) { New-Item -ItemType Directory -Path $InstallDir | Out-Null }

$SourceBinary = Join-Path $ExtractPath "looty.exe"
$DestBinary = Join-Path $InstallDir "looty.exe"

Move-Item -Path $SourceBinary -Destination $DestBinary -Force

# Cleanup
Remove-Item -Path $TempFile -Force
Remove-Item -Path $ExtractPath -Recurse -Force

# Add to PATH if not already there
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    Write-Host "✅ Installed to $DestBinary"
    Write-Host "🔄 Please restart your terminal for PATH changes to take effect"
} else {
    Write-Host "✅ Updated looty at $DestBinary"
}

Write-Host "🎉 Looty is ready! Run: looty --help"
Write-Host "💡 Make sure to restart your terminal if this is your first install"
