// Package hotkey provides global keyboard shortcut registration via Windows API.
// F6 = Toggle Pause/Resume, F7 = Stop.
package hotkey

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Action represents a hotkey action.
type Action int

const (
	ActionTogglePause Action = iota
	ActionStop
)

const (
	wmHotkey = 0x0312

	// Virtual key codes
	vkF6 = 0x75
	vkF7 = 0x76

	// Hotkey IDs
	idTogglePause = 1
	idStop        = 2
)

var (
	user32              = windows.NewLazySystemDLL("user32.dll")
	procRegisterHotKey  = user32.NewProc("RegisterHotKey")
	procGetMessageW     = user32.NewProc("GetMessageW")
	procUnregisterHotKey = user32.NewProc("UnregisterHotKey")
)

type msg struct {
	HWND    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

// Listen registers F6 and F7 as global hotkeys and blocks, sending
// actions to the provided channel. Call from a dedicated goroutine.
// The channel is closed when the function returns.
func Listen(ch chan<- Action) error {
	defer close(ch)

	ret, _, err := procRegisterHotKey.Call(0, idTogglePause, 0, vkF6)
	if ret == 0 {
		return fmt.Errorf("RegisterHotKey F6 failed: %v", err)
	}
	defer procUnregisterHotKey.Call(0, idTogglePause)

	ret, _, err = procRegisterHotKey.Call(0, idStop, 0, vkF7)
	if ret == 0 {
		return fmt.Errorf("RegisterHotKey F7 failed: %v", err)
	}
	defer procUnregisterHotKey.Call(0, idStop)

	var m msg
	for {
		// GetMessage blocks until a message is available.
		// Returns 0 for WM_QUIT, -1 on error.
		ret, _, _ := procGetMessageW.Call(
			uintptr(unsafe.Pointer(&m)), 0, 0, 0,
		)
		if int32(ret) <= 0 {
			return nil // WM_QUIT or error
		}

		if m.Message == wmHotkey {
			switch m.WParam {
			case idTogglePause:
				ch <- ActionTogglePause
			case idStop:
				ch <- ActionStop
			}
		}
	}
}
