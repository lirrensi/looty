// FILE: cmd/blip/daemon_detach_windows.go
// PURPOSE: Provide Windows-specific detached process attributes for daemon child launches.
// OWNS: Best-effort Windows background detachment flags for Looty daemon mode.
// EXPORTS: detachedProcessAttr
// DOCS: agent_chat/plan_daemon-mode_2026-05-17.md

//go:build windows

package main

import "syscall"

func detachedProcessAttr() *syscall.SysProcAttr {
	// Best effort: CREATE_NEW_PROCESS_GROUP avoids some console coupling while keeping
	// startup handoff reliable. Full Windows service-style detachment varies by shell.
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
