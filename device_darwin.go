//go:build darwin
// +build darwin

package main

import (
	"os/exec"
	"strings"
)

func isAdmin() bool {
	return true 
}

func elevate() error {
	return nil
}

func getDeviceInfo() (string, string) {
	cmdUUID := exec.Command("sh", "-c", "ioreg -rd1 -c IOPlatformExpertDevice | awk -F'\"' '/IOPlatformUUID/{print $4}'")
	outUUID, _ := cmdUUID.Output()
	uuid := strings.TrimSpace(string(outUUID))
	if uuid == "" {
		uuid = "UNKNOWN-MAC-UUID"
	}

	cmdName := exec.Command("scutil", "--get", "ComputerName")
	outName, _ := cmdName.Output()
	name := strings.TrimSpace(string(outName))
	if name == "" {
		name = "MacBook"
	}

	return uuid, name
}