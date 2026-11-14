//go:build windows
// +build windows

package main

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32              = windows.NewLazySystemDLL("user32.dll")
	procSendInput       = user32.NewProc("SendInput")
	procGetForeground   = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW  = user32.NewProc("GetWindowTextW")
	procShowWindow      = user32.NewProc("ShowWindow")
	procSetForeground   = user32.NewProc("SetForegroundWindow")
	procFindWindow      = user32.NewProc("FindWindowW")
)

const (
	INPUT_KEYBOARD    = 1
	KEYEVENTF_KEYUP   = 0x0002
	KEYEVENTF_UNICODE = 0x0004

	SW_MINIMIZE = 6
	SW_RESTORE  = 9
)

type KEYBDINPUT struct {
	Vk        uint16
	Scan      uint16
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type INPUT struct {
	Type uint32
	_    uint32
	Ki   KEYBDINPUT
	_    [8]byte
}

func sendUnicodeRune(r rune) error {
	down := INPUT{Type: INPUT_KEYBOARD}
	down.Ki.Flags = KEYEVENTF_UNICODE
	down.Ki.Scan = uint16(r)
	ret, _, err := procSendInput.Call(1, uintptr(unsafe.Pointer(&down)), unsafe.Sizeof(down))
	if ret == 0 {
		return fmt.Errorf("SendInput down failed: %v", err)
	}

	up := INPUT{Type: INPUT_KEYBOARD}
	up.Ki.Flags = KEYEVENTF_UNICODE | KEYEVENTF_KEYUP
	up.Ki.Scan = uint16(r)
	ret, _, err = procSendInput.Call(1, uintptr(unsafe.Pointer(&up)), unsafe.Sizeof(up))
	if ret == 0 {
		return fmt.Errorf("SendInput up failed: %v", err)
	}
	return nil
}

func sendVK(vk uint16) error {
	down := INPUT{Type: INPUT_KEYBOARD}
	down.Ki.Vk = vk
	ret, _, err := procSendInput.Call(1, uintptr(unsafe.Pointer(&down)), unsafe.Sizeof(down))
	if ret == 0 {
		return fmt.Errorf("SendInput vk down failed: %v", err)
	}

	up := INPUT{Type: INPUT_KEYBOARD}
	up.Ki.Vk = vk
	up.Ki.Flags = KEYEVENTF_KEYUP
	ret, _, err = procSendInput.Call(1, uintptr(unsafe.Pointer(&up)), unsafe.Sizeof(up))
	if ret == 0 {
		return fmt.Errorf("SendInput vk up failed: %v", err)
	}
	return nil
}

func getForegroundTitle() string {
	hwnd, _, _ := procGetForeground.Call()
	if hwnd == 0 {
		return ""
	}
	buf := make([]uint16, 512)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 512)
	return windows.UTF16ToString(buf)
}

func injectText(text string, startDelaySec int, charDelayMs int, sebOnly bool, onStatus func(string)) error {
	if text == "" {
		return errors.New("empty text")
	}

	for i := startDelaySec; i > 0; i-- {
		if onStatus != nil {
			onStatus(fmt.Sprintf("⏳ %ds me typing shuru hogi... target window pe click karo!", i))
		}
		time.Sleep(1 * time.Second)
	}

	if sebOnly {
		title := getForegroundTitle()
		if !strings.Contains(strings.ToLower(title), "safe exam browser") && !strings.Contains(strings.ToLower(title), "seb") {
			return fmt.Errorf("foreground window SEB nahi hai: %q", title)
		}
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	lines := strings.Split(text, "\n")
	totalLines := len(lines)

	for i, line := range lines {
		if onStatus != nil {
			onStatus(fmt.Sprintf("⌨️  Typing line %d/%d...", i+1, totalLines))
		}

		for _, r := range line {
			var err error
			if r == '\t' {
				err = sendVK(0x09)
			} else {
				err = sendUnicodeRune(r)
			}
			if err != nil {
				return fmt.Errorf("char %q par fail: %w", r, err)
			}
			time.Sleep(time.Duration(charDelayMs) * time.Millisecond)
		}

		if i < totalLines-1 {
			sendVK(0x0D)
			time.Sleep(100 * time.Millisecond)
		}
	}

	if onStatus != nil {
		onStatus(fmt.Sprintf("✅ Done! %d lines typed successfully.", totalLines))
	}
	return nil
}

func findSEBWindow() uintptr {
	titlePtr, _ := syscall.UTF16PtrFromString("Safe Exam Browser")
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	return hwnd
}

func showWindow(hwnd uintptr, cmd int) {
	procShowWindow.Call(hwnd, uintptr(cmd))
}

func setForeground(hwnd uintptr) {
	procSetForeground.Call(hwnd)
}