//go:build darwin
// +build darwin

package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
)

// Mac ke liye window states
const (
	SW_MINIMIZE = 6
	SW_RESTORE  = 9
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

	// 3. Mac me Cmd+V simulate karna
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

// ==========================================
// MAC WINDOW MANAGEMENT (AppleScript)
// ==========================================

// SEB ki window check karta hai
func findSEBWindow() uintptr {
	script := `tell application "System Events" to return exists (process "Safe Exam Browser")`
	cmd := exec.Command("osascript", "-e", script)
	out, _ := cmd.Output()
	if strings.TrimSpace(string(out)) == "true" {
		return 1 // Return 1 matlab SEB mil gaya
	}
	return 0
}

// Window ko minimize ya restore karta hai
func showWindow(hwnd uintptr, cmd int) {
	if hwnd == 0 {
		return
	}

	if cmd == SW_MINIMIZE {
		// SEB ko hide karo (Minimize)
		script := `tell application "System Events" to set visible of process "Safe Exam Browser" to false`
		exec.Command("osascript", "-e", script).Run()
	} else if cmd == SW_RESTORE {
		// SEB ko wapas dikhao
		script := `tell application "System Events" to set visible of process "Safe Exam Browser" to true`
		exec.Command("osascript", "-e", script).Run()
	}
}

// SEB ko foreground me laata hai
func setForeground(hwnd uintptr) {
	if hwnd == 0 {
		return
	}
	// SEB ko activate karo (saamne lao)
	script := `tell application "Safe Exam Browser" to activate`
	exec.Command("osascript", "-e", script).Run()
}
