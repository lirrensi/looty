// FILE: cmd/blip/daemon_detach_unix.go
// PURPOSE: Provide Unix-specific detached process attributes for daemon child launches.
// OWNS: Best-effort Unix session detachment for Looty daemon mode.
// EXPORTS: detachedProcessAttr
// DOCS: agent_chat/plan_daemon-mode_2026-05-17.md

//go:build !windows

package main

import "syscall"

func detachedProcessAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
