# Looty Behavioral Specification

## Abstract

Looty is a zero-configuration local file access and scratchpad sync system consisting of a desktop-hosted server and a browser-based mobile client. This specification defines the externally observable behavior of server startup, discovery, file access, scratchpad sync, TLS trust signaling, and foreground/background execution modes.

## Introduction

Looty is designed to let a user expose the current folder to nearby devices with minimal setup. A conforming implementation prioritizes immediate usability, explicit trust signaling for self-signed TLS, and consistent startup information across interactive and non-interactive execution modes.

## Scope

In scope:
- server startup behavior
- process modes
- startup metadata
- discovery behavior
- file browsing and transfer behavior
- scratchpad synchronization
- TLS and trust signaling

Out of scope:
- user authentication
- multi-folder serving
- public cloud sync
- long-term file history/versioning

## Terminology

- **Served directory**: the directory exposed by the running Looty server
- **Startup record**: the complete set of connection details emitted when Looty starts
- **Foreground mode**: server execution attached to the invoking terminal/session
- **Background mode**: server execution detached or persistent beyond the invoking terminal/session
- **Agent-managed mode**: background-capable launch where startup details are returned in machine-readable form for another process to relay
- **Friend code**: the human-readable identifier associated with an auto-generated self-signed certificate

## Normative Language

The key words MUST, MUST NOT, SHOULD, SHOULD NOT, and MAY are to be interpreted as described in RFC 2119.

## System Model

Actors:
- Desktop user
- Mobile browser user
- Service manager or OS background runner
- Agent or automation process

Top-level interfaces:
- CLI startup interface
- HTTP(S) API
- WebSocket synchronization channel
- Startup record output channel

## Conformance

A conforming Looty implementation MUST:
- serve exactly one directory per running process
- expose startup connection details at process start
- preserve TLS trust signaling when self-signed TLS is used
- provide equivalent connection information in both foreground and background-capable modes
- keep file and scratchpad behavior consistent regardless of process mode

## Behavioral Specification

### 1. Server Startup

- Looty MUST determine the served directory from the process working directory unless overridden by `-serve-dir`.
- Looty MUST determine whether to serve HTTP or HTTPS based on bind mode and TLS-related flags.
- Looty MUST generate or load all startup connection details before entering long-running serve mode.

### 2. Startup Record

The startup record MUST contain:
- served directory
- effective host or reachable addresses
- effective port
- effective protocol (`http` or `https`)
- at least one primary connection URL

When self-signed TLS is active, the startup record MUST also contain:
- certificate fingerprint
- friend code

The startup record MAY additionally contain:
- all reachable URLs
- process identifier
- start timestamp
- QR payload URL
- QR image file path
- execution mode

### 3. Foreground Mode

- In foreground mode, Looty MUST emit a human-friendly startup record to standard output.
- In foreground mode, Looty SHOULD render a QR code for the primary connection URL when terminal capabilities permit.
- In foreground mode, Looty MUST remain attached to the calling session until stopped or terminated.

### 4. Background Mode

- In background mode, Looty MUST continue serving after launch without requiring the invoking terminal to remain attached.
- In background mode, Looty MUST preserve the startup record in a retrievable form.
- The retrievable startup record MUST remain available long enough for a user or supervising process to obtain the connection details after launch.
- Background mode MUST NOT discard TLS trust material when self-signed TLS is active.
- When a JSON startup record file is requested, Looty MUST write a sibling QR image artifact for the primary connection URL.

### 5. Agent-Managed Mode

- Agent-managed launches MUST provide the same startup record semantics as background mode.
- Agent-managed launches MUST support machine-readable startup record retrieval.
- A supervising process MUST be able to relay the primary URL and any TLS trust material to the end user without scraping decorative terminal output.
- When an agent-managed launch writes a JSON startup record file, the supervising process MUST also be able to retrieve a QR image artifact from the filesystem without rendering the QR code itself.

### 6. TLS Trust Signaling

- When binding to non-loopback interfaces without explicit TLS opt-out, Looty MUST use TLS.
- When Looty uses an auto-generated self-signed certificate, it MUST generate a certificate fingerprint and friend code for that run.
- The fingerprint presented to the user MUST correspond to the active certificate for the running server instance.
- The friend code presented to the user MUST correspond to the active certificate for the running server instance.

### 7. Discovery and Direct Connection

- Plain-HTTP discovery flows MAY rely on local HTML discovery mechanisms.
- HTTPS/self-signed flows MUST support direct connection via an explicit URL.
- When HTTPS mode makes local discovery impractical or unavailable, the startup record MUST still provide enough information for manual/direct connection.

### 8. File Serving Behavior

- Looty MUST expose listing, download, and upload behavior within the served directory only.
- Looty MUST reject path traversal attempts.
- Looty MUST NOT expose parent directories outside the served directory boundary.

### 9. Scratchpad Behavior

- Looty MUST provide a shared text scratchpad across connected clients.
- Scratchpad updates MUST propagate across connected clients in near real time.

## Data and State Model

### Startup Record Fields

| Field | Required | Meaning |
|---|---|---|
| `serveDir` | Yes | Absolute or effective served directory |
| `protocol` | Yes | `http` or `https` |
| `port` | Yes | Listening port |
| `primaryUrl` | Yes | Preferred connection URL |
| `addresses` | Yes | Reachable address set |
| `mode` | Yes | Foreground, background, or agent-managed |
| `fingerprint` | Conditional | Required for self-signed TLS |
| `friendCode` | Conditional | Required for self-signed TLS |
| `pid` | Optional | Running process id |
| `startedAt` | Optional | Process start time |
| `qrImagePath` | Optional | Filesystem path to a rendered QR image artifact |

## Error Handling and Edge Cases

- If Looty cannot determine the served directory, it MUST fail startup.
- If Looty cannot bind the requested port/address, it MUST fail startup.
- If TLS is required but certificate generation or loading fails, it MUST fail startup.
- If background-capable launch succeeds but startup record persistence fails, the implementation SHOULD treat startup as failed unless an equivalent retrieval path is still guaranteed.

## Security Considerations

- Self-signed TLS requires explicit trust signaling; suppressing or losing that trust material is a security failure.
- Machine-readable startup output MUST preserve exact fingerprint values without lossy formatting.
- Background launch mechanisms SHOULD avoid exposing private key material in logs, temp files, or process arguments.
- Plain HTTP on non-loopback networks SHOULD be treated as trusted-network-only behavior.
- When binding to a non-loopback address, Looty SHOULD emit a startup advisory reminding the user to verify that the listening port is reachable through any firewall or network policy.

## References

### Normative References
- RFC 2119 — Key words for use in RFCs to Indicate Requirement Levels

### Informative References
- `docs/product.md`
- `docs/arch.md`
