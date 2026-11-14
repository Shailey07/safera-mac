//go:build darwin
// +build darwin

package main

import (
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/atotto/clipboard"
)

func injectText(text string, startDelaySec int, charDelayMs int, sebOnly bool, onStatus func(string)) error {
	if text == "" {
		return errors.New("empty text")
	}

	for i := startDelaySec; i > 0; i-- {
		if onStatus != nil {
			onStatus(fmt.Sprintf("⏳ %ds me paste shuru hoga... Mac par target window pe click karo!", i))
		}
		time.Sleep(1 * time.Second)
	}

	// 1. Purana clipboard save karein
	originalClipboard, _ := clipboard.ReadAll()

	// 2. Naya text clipboard me daalein
	err := clipboard.WriteAll(text)
	if err != nil {
		return fmt.Errorf("clipboard fail: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	if onStatus != nil {
		onStatus("⌨️  Pasting text via Mac Accessibility...")
	}

	// 3. Mac me Cmd+V simulate karna (AppleScript ke through)
	script := `tell application "System Events" to keystroke "v" using command down`
	cmd := exec.Command("osascript", "-e", script)
	err = cmd.Run()
	
	time.Sleep(200 * time.Millisecond)

	// 4. Wapas purana clipboard restore karein
	clipboard.WriteAll(originalClipboard)

	if err != nil {
		return fmt.Errorf("paste fail hua. Kya aapne Mac ki Settings > Privacy & Security > Accessibility me SafeEra ko permission di hai? Error: %v", err)
	}

	if onStatus != nil {
		onStatus(fmt.Sprintf("✅ Done! %d characters pasted.", len([]rune(text))))
	}
	return nil
}

// Mac par window tools abhi available nahi hain, isliye khali functions
func findSEBWindow() uintptr { return 0 }
func showWindow(hwnd uintptr, cmd int) {}
func setForeground(hwnd uintptr) {}