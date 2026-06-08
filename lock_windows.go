//go:build windows

package main

import (
	"syscall"
)

func isProcessAlive(pid int) bool {
	const SYNCHRONIZE = 0x00100000
	handle, err := syscall.OpenProcess(SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)
	const WAIT_TIMEOUT = 0x102
	ret, _ := syscall.WaitForSingleObject(handle, 0)
	return ret == WAIT_TIMEOUT
}

func killProcess(pid int) {
	const PROCESS_TERMINATE = 0x0001
	handle, err := syscall.OpenProcess(PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return
	}
	defer syscall.CloseHandle(handle)
	syscall.TerminateProcess(handle, 0)
}
