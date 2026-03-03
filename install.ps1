$ErrorActionPreference = "Stop"

$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "amd64" } else { "arm64" }
$url = "https://github.com/lirrensi/looty/releases/latest/download/looty-windows-${arch}.zip"
$zip = "$env:TEMP\looty.zip"
$tempDir = "$env:TEMP\looty"

Invoke-WebRequest -Uri $url -OutFile $zip
Expand-Archive -Path $zip -Force -DestinationPath $tempDir

$installDir = "$env:LOCALAPPDATA\looty"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

Move-Item "$tempDir\looty-windows-${arch}.exe" "$installDir\looty.exe" -Force

$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*looty*") {
    [Environment]::SetEnvironmentVariable("Path", "$path;$installDir", "User")
}

Remove-Item $zip,$tempDir -Recurse -Force -EA SilentlyContinue

Write-Host "Installed looty to $installDir - restart your terminal"