---
name: looty-file-share
description: Run Looty on a remote server to share any folder via browser. Use this whenever the user needs to make a directory accessible over the network, download and start the binary, or verify a self-signed certificate before trusting the connection.
---

# Looty — share any folder over the network

Looty is a single-binary file server. Drop it in **any** folder, run it, and that folder is instantly available in a browser. No config, no accounts, no cloud.

That is the whole point. You do not create a special folder — you go to wherever the files already are.

## Step 0 — Is Looty installed?

Before anything else, check if the `looty` binary exists on the target machine and you know its location.

- `which looty` / `where looty`
- If the path is known from a previous run, use that.

### Not installed? Pick one:

**One-liner from the repo (preferred):**

macOS / Linux:
```bash
curl -sL https://raw.githubusercontent.com/lirrensi/looty/main/install.sh | sh
```

Windows (PowerShell):
```powershell
irm https://raw.githubusercontent.com/lirrensi/looty/main/install.ps1 | iex
```

**Git clone + build (if you need the source or the one-liner doesn't work):**

```bash
git clone https://github.com/lirrensi/looty.git
cd looty
make build
# binary lands at ./looty (or ./looty.exe on Windows)
```

**GitHub releases** (grab the binary directly for your OS/arch):
- `https://github.com/lirrensi/looty/releases/latest`

## GitHub repo

Always useful to have the link handy so you or the user can check how it works, read the README, or clone:

```
https://github.com/lirrensi/looty
```

## How to share a folder

1. `cd` into whatever folder you want to share.
2. Run `looty` — that's it. The server binds to `:41111`, and the current directory is served.

That folder is now accessible from any device on the network.

### Background / daemon mode

If the server should keep running after the terminal closes:

```bash
looty -daemon -json-file startup.json
```

This writes a startup record with the URL, addresses, TLS fingerprint (if applicable), and friend code into `startup.json` so you can retrieve it later.

### Running on a remote server (raw IP)

When Looty binds to a non-loopback address it automatically enables HTTPS with a self-signed certificate and prints a fingerprint:

```bash
looty -host 0.0.0.0 -daemon -json-file startup.json
```

The self-signed cert is expected — do not treat it as an error.

### Running behind a reverse proxy

Bind to localhost only and let the proxy handle public HTTPS:

```bash
looty -host 127.0.0.1 -daemon -json-file startup.json
```

Return the proxy URL, not the Looty URL.

## Verifying the connection

- Open the URL in a browser.
- Confirm the folder contents are listed.
- Confirm the scheme (http / https) matches the exposure mode.

### If the certificate is self-signed

The browser will show a warning. That is normal and expected for a self-signed cert. The user needs to verify the fingerprint matches what was printed at startup.

1. In the browser address bar, click the lock / Not Secure / page-info icon.
2. Open the certificate details.
3. Look for the **SHA-256 fingerprint** (also called thumbprint).
4. Compare it to the fingerprint from the startup output or `startup.json`.

If the two fingerprints match, the connection is safe even though the browser warns. The only way to be sure is to compare fingerprints — tell the user that directly.

## What to return

After setup, hand back:
- the working directory (the folder being shared)
- the access URL
- whether it's behind a proxy or direct IP
- TLS status and fingerprint (if self-signed)
- the GitHub repo link so the user can read the docs

## Useful phrases

- "Go to whatever folder has the files you want to share, `cd` there, and run `looty`. Done."
- "Looty doesn't need a dedicated folder — it serves whatever directory you run it in."
- "The self-signed certificate warning is normal. Compare the fingerprint to confirm it's the right server."
- "The repo is at `github.com/lirrensi/looty` if you want to read how it works or build from source."
