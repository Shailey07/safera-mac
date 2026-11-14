//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

func isAdmin() bool {
	cmd := exec.Command("net", "session")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil
}

func elevate() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	const SW_NORMAL = 1
	err = windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, SW_NORMAL)
	if err != nil {
		return fmt.Errorf("ShellExecute failed: %w", err)
	}
	return nil
}

func getDeviceInfo() (string, string) {
	cmdUUID := exec.Command("powershell", "-NoProfile", "-Command", "(Get-CimInstance Win32_ComputerSystemProduct).UUID")
	cmdUUID.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	outUUID, err := cmdUUID.Output()
	
	uuid := strings.TrimSpace(string(outUUID))
	if err != nil || uuid == "" || strings.Contains(strings.ToLower(uuid), "error") {
		uuid = "UNKNOWN-UUID"
	}

	cmdName := exec.Command("hostname")
	cmdName.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	outName, _ := cmdName.Output()
	name := strings.TrimSpace(string(outName))
	if name == "" {
		name = "Windows-PC"
	}

	return uuid, name
}