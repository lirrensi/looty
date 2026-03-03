# Plan: Discovery & Install Improvements
_Two changes: (1) looty.html always copied to home folder for easy access, (2) smarter network discovery with mDNS, caching, and tiered parallel scanning._

---

# Checklist
- [x] Step 1: Add zeroconf dependency to go.mod
- [x] Step 2: Create mDNS server in internal/server/mdns.go
- [x] Step 3: Integrate mDNS into server.Start()
- [x] Step 4: Modify main.go to copy looty.html to home folder
- [x] Step 5: Rewrite discovery.js with new discovery logic
- [x] Step 6: Update install.sh to copy looty.html to ~/looty/
- [x] Step 7: Update install.ps1 to copy looty.html to %USERPROFILE%\looty\
- [x] Step 8: Rebuild frontend and test

---

## Context

Current state:
- `go.mod`: Module is `github.com/lirrensi/looty`, uses fsnotify and gorilla/websocket
- `cmd/blip/main.go`: Extracts looty.html to exe directory only
- `internal/server/server.go`: `Start(serveDir, port)` starts HTTP server, file watcher, WebSocket hub
- `web/src/components/discovery.js`: Scans only one detected subnet (192.168.1.x etc), tries IPs 1-254 sequentially-ish
- `install.sh`: Installs binary to `~/.local/bin/`, no looty.html copy
- `install.ps1`: Installs binary to `%LOCALAPPDATA%\looty\`, no looty.html copy

Goal:
- looty.html in TWO places: alongside exe AND in home folder root
- Discovery: cache → mDNS (.local) → tiered parallel scan (32 IPs × multiple subnets) → full scan → manual

---

## Prerequisites

- Go 1.25.5 installed
- Node.js/npm installed (for frontend build)
- Project root: `C:\Users\rx\001_Code\105_DeadProjects\BlipSync`

## Scope Boundaries

OUT OF SCOPE:
- Any changes to file browser, clipboard, or WebSocket functionality
- Any changes to API endpoints
- Any changes to styling/UI appearance

---

## Steps

### Step 1: Add zeroconf dependency to go.mod

Open `go.mod`. Add the zeroconf import line to the require block.

**Action:** Run this command from project root:
```
go get github.com/grandcat/zeroconf
```

✅ Success: `go.mod` contains `github.com/grandcat/zeroconf` in require section, `go.sum` updated.
❌ If failed: Report the error output. Do not proceed.

---

### Step 2: Create mDNS server in internal/server/mdns.go

Create new file `internal/server/mdns.go` with the following content:

```go
package server

import (
	"log"

	"github.com/grandcat/zeroconf"
)

var mdnsServer *zeroconf.Server

// StartMDNS announces the looty service on the local network as "looty.local"
func StartMDNS(port int) error {
	var err error
	mdnsServer, err = zeroconf.Register(
		"looty",       // instance name
		"_http._tcp",  // service type
		"local.",      // domain
		port,          // port
		nil,           // TXT records
		nil,           // interfaces (nil = all)
	)
	if err != nil {
		return err
	}
	log.Println("mDNS: Announcing as looty.local")
	return nil
}

// StopMDNS shuts down the mDNS server
func StopMDNS() {
	if mdnsServer != nil {
		mdnsServer.Shutdown()
		mdnsServer = nil
	}
}
```

✅ Success: File `internal/server/mdns.go` exists with the above content.
❌ If failed: Report the error. Do not proceed.

---

### Step 3: Integrate mDNS into server.Start()

Open `internal/server/server.go`. Modify the `Start` function to call `StartMDNS` and ensure cleanup.

**Changes:**

1. Find the `Start` function (line 61). After `StartWatcher(serveDir)` (line 66), add:
```go
	// Start mDNS announcement
	if err := StartMDNS(port); err != nil {
		log.Printf("Warning: mDNS failed: %v", err)
	}
```

2. The current `Start` function returns `http.ListenAndServe(addr, mux)` at line 117. This never returns until server stops. mDNS will keep running. For proper cleanup on shutdown, we would need signal handling, but for now this is acceptable - the mDNS announcement stops when the process exits.

✅ Success: `server.go` calls `StartMDNS(port)` after `StartWatcher(serveDir)`.
❌ If failed: Report the error. Do not proceed.

---

### Step 4: Modify main.go to copy looty.html to home folder

Open `cmd/blip/main.go`. Modify the HTML extraction section to also copy to home folder.

**Find the block** (lines 49-66) that extracts looty.html. Replace the entire block with:

```go
	// Extract looty.html to exe's directory (so it can be copied to phone)
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("Warning: Could not get executable path: %v", err)
	}
	exeDir := filepath.Dir(execPath)
	lootyHTMLPath := filepath.Join(exeDir, "looty.html")
	html, err := server.GetHTML()
	if err != nil {
		log.Printf("Warning: Could not get embedded HTML: %v", err)
	} else {
		// Write to exe directory
		err = os.WriteFile(lootyHTMLPath, html, 0644)
		if err != nil {
			log.Printf("Warning: Could not create looty.html: %v", err)
		} else {
			fmt.Println("Extracted looty.html to exe directory")
		}

		// Also write to home folder for easy access
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Warning: Could not get home directory: %v", err)
		} else {
			homeLootyDir := filepath.Join(homeDir, "looty")
			os.MkdirAll(homeLootyDir, 0755)
			homeLootyPath := filepath.Join(homeLootyDir, "looty.html")
			err = os.WriteFile(homeLootyPath, html, 0644)
			if err != nil {
				log.Printf("Warning: Could not create looty.html in home: %v", err)
			} else {
				fmt.Printf("Also saved to: %s\n", homeLootyPath)
			}
		}
	}
```

✅ Success: `main.go` writes looty.html to BOTH exe directory and `~/looty/looty.html` (or `%USERPROFILE%\looty\looty.html` on Windows).
❌ If failed: Report the error. Do not proceed.

---

### Step 5: Rewrite discovery.js with new discovery logic

Open `web/src/components/discovery.js`. Replace the entire file content with:

```js
export function discovery() {
  return {
    status: 'searching', // searching, connected, failed
    serverIP: '',
    log: [], // Track all attempts
    cachedIPKey: 'looty_cached_ip',
    
    logMsg(msg) {
      this.log.push(`[${new Date().toLocaleTimeString()}] ${msg}`)
      console.log(`[LOOTY] ${msg}`)
    },
    
    saveCachedIP(ip) {
      try {
        localStorage.setItem(this.cachedIPKey, ip)
      } catch (e) {
        // localStorage might be unavailable
      }
    },
    
    getCachedIP() {
      try {
        return localStorage.getItem(this.cachedIPKey)
      } catch (e) {
        return null
      }
    },
    
    async findServer() {
      this.log = []
      this.logMsg('Starting server discovery...')
      
      // 1. Try cached IP first (instant reconnect)
      const cachedIP = this.getCachedIP()
      if (cachedIP) {
        this.logMsg(`Trying cached IP: ${cachedIP}:41111`)
        const found = await this.pingServer(cachedIP)
        if (found) {
          this.serverIP = cachedIP
          this.status = 'connected'
          this.logMsg(`SUCCESS: Found server at cached IP ${cachedIP}!`)
          return cachedIP
        }
        this.logMsg(`Cached IP ${cachedIP} not responding`)
      }
      
      // 2. Try mDNS hostname (zero-config)
      this.logMsg('Trying looty.local:41111 (mDNS)')
      const localFound = await this.pingServer('looty.local')
      if (localFound) {
        this.serverIP = 'looty.local'
        this.status = 'connected'
        this.saveCachedIP('looty.local')
        this.logMsg('SUCCESS: Found server via mDNS!')
        return 'looty.local'
      }
      this.logMsg('looty.local not responding')
      
      // 3. Smart parallel scan - Tier 1: first 32 IPs of common subnets
      const subnets = ['192.168.1', '192.168.0', '10.0.0', '192.168.2']
      this.logMsg(`Tier 1: Scanning first 32 IPs of ${subnets.join(', ')} in parallel...`)
      
      const tier1Promises = []
      for (const subnet of subnets) {
        for (let i = 1; i <= 32; i++) {
          tier1Promises.push(this.pingServerWithIP(`${subnet}.${i}`))
        }
      }
      
      const tier1Results = await Promise.allSettled(tier1Promises)
      const tier1Found = tier1Results.find(r => r.status === 'fulfilled' && r.value)
      
      if (tier1Found) {
        const foundIP = tier1Found.value
        this.serverIP = foundIP
        this.status = 'connected'
        this.saveCachedIP(foundIP)
        this.logMsg(`SUCCESS: Found server at ${foundIP}!`)
        return foundIP
      }
      
      // 4. Tier 2: Expand to full subnet scan
      this.logMsg('Tier 2: Expanding to full subnet scan...')
      
      const tier2Promises = []
      for (const subnet of subnets) {
        for (let i = 33; i <= 254; i++) {
          tier2Promises.push(this.pingServerWithIP(`${subnet}.${i}`))
        }
      }
      
      const tier2Results = await Promise.allSettled(tier2Promises)
      const tier2Found = tier2Results.find(r => r.status === 'fulfilled' && r.value)
      
      if (tier2Found) {
        const foundIP = tier2Found.value
        this.serverIP = foundIP
        this.status = 'connected'
        this.saveCachedIP(foundIP)
        this.logMsg(`SUCCESS: Found server at ${foundIP}!`)
        return foundIP
      }
      
      this.status = 'failed'
      this.logMsg('ERROR: No server found on network')
      return null
    },
    
    async pingServer(ip) {
      const controller = new AbortController()
      const timeout = setTimeout(() => controller.abort(), 500)
      
      try {
        const response = await fetch(`http://${ip}:41111/ping`, {
          method: 'GET',
          signal: controller.signal,
        })
        clearTimeout(timeout)
        
        if (response.ok) {
          return ip
        }
        return null
      } catch (err) {
        clearTimeout(timeout)
        return null
      }
    },
    
    // Returns the IP if found, null otherwise (for Promise.allSettled usage)
    async pingServerWithIP(ip) {
      const found = await this.pingServer(ip)
      return found ? ip : null
    },
    
    useManualIP(ip) {
      this.serverIP = ip
      this.status = 'connected'
      this.saveCachedIP(ip)
    },
  }
}
```

✅ Success: `discovery.js` contains new discovery logic with cache, mDNS, tiered scanning.
❌ If failed: Report the error. Do not proceed.

---

### Step 6: Update install.sh to copy looty.html to ~/looty/

Open `install.sh`. Add logic to also create ~/looty/ and place a note there (since the binary itself will extract looty.html on first run).

Replace entire file with:

```bash
#!/bin/sh
set -e

ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')"
OS="$(uname | sed -e 's/Darwin/macos/' -e 's/Linux/linux/')"

curl -sL "https://github.com/lirrensi/looty/releases/latest/download/looty-${OS}-${ARCH}.tar.gz" | tar xz

mkdir -p ~/.local/bin
mv looty ~/.local/bin/

# Create ~/looty directory for easy access to looty.html
mkdir -p ~/looty

echo "Installed looty to ~/.local/bin"
echo "Run 'looty' to start - it will extract looty.html to:"
echo "  1. The current directory"
echo "  2. ~/looty/ (for easy phone transfer)"
echo ""
echo "Restart your terminal or run: export PATH=\"$HOME/.local/bin:\$PATH\""
```

✅ Success: `install.sh` creates `~/looty/` directory and prints helpful message.
❌ If failed: Report the error. Do not proceed.

---

### Step 7: Update install.ps1 to copy looty.html to %USERPROFILE%\looty\

Open `install.ps1`. Add logic to create the home looty folder.

Replace entire file with:

```powershell
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
```

✅ Success: `install.ps1` creates `%USERPROFILE%\looty\` directory and prints helpful message.
❌ If failed: Report the error. Do not proceed.

---

### Step 8: Rebuild frontend and test

**Action:** Run the following commands from project root:

```bash
cd web
npm install
npm run build
```

Then copy the output to the server assets:
```bash
cp dist/index.html ../internal/server/assets/index.html
```

Then build the Go binary:
```bash
cd ..
go build -o looty.exe ./cmd/blip
```

✅ Success: `looty.exe` exists in project root, `internal/server/assets/index.html` is updated.
❌ If failed: Report the full error output. Do not proceed.

---

## Verification

1. Run `looty.exe` from project root
2. Verify console output shows:
   - "mDNS: Announcing as looty.local"
   - "Extracted looty.html to exe directory"
   - "Also saved to: C:\Users\rx\looty\looty.html"
3. Open `web/dist/index.html` in a browser (or run from a local server)
4. Verify discovery attempts show in order: cached IP → looty.local → tier 1 scan → tier 2 scan
5. Check that `C:\Users\rx\looty\looty.html` exists

## Rollback

If critical failure:
```bash
git checkout -- go.mod go.sum cmd/blip/main.go internal/server/server.go internal/server/mdns.go web/src/components/discovery.js install.sh install.ps1
rm -f internal/server/mdns.go
```
