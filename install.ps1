$ErrorActionPreference = "Stop"

$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "amd64" } else { "arm64" }
$url = "https://github.com/lirrensi/looty/releases/latest/download/looty-windows-${arch}.zip"
$zip = "$env:TEMP\looty.zip"
$tempDir = "$env:TEMP\looty"

Invoke-WebRequest -Uri $url -OutFile $zip
Expand-Archive -Path $zip -Force -DestinationPath $tempDir

if (-not (Test-Path "$tempDir\looty.exe")) {
    Remove-Item $zip,$tempDir -Recurse -Force -EA SilentlyContinue
    Write-Error "Download failed: looty.exe not found in archive"
}

$installDir = "$env:LOCALAPPDATA\looty"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

Move-Item "$tempDir\looty.exe" "$installDir\looty.exe" -Force

# Create home looty folder for easy access to looty.html
$homeLootyDir = "$env:USERPROFILE\looty"
New-Item -ItemType Directory -Force -Path $homeLootyDir | Out-Null

$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*looty*") {
    [Environment]::SetEnvironmentVariable("Path", "$path;$installDir", "User")
}

Remove-Item $zip,$tempDir -Recurse -Force -EA SilentlyContinue

Write-Host "Installed looty to $installDir"
Write-Host "Run 'looty' to start - it will extract looty.html to:"
Write-Host "  1. The current directory"
Write-Host "  2. $homeLootyDir (for easy phone transfer)"
Write-Host ""
Write-Host "Restart your terminal"