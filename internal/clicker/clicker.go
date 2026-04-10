// Package clicker provides mouse control using Windows SendInput API.
package clicker

import (
	"fmt"
	"image"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32 = windows.NewLazySystemDLL("user32.dll")

	procSetCursorPos    = user32.NewProc("SetCursorPos")
	procSendInput       = user32.NewProc("SendInput")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
)

const (
	inputMouse          = 0
	mouseeventfAbsolute = 0x8000
	mouseeventfMove     = 0x0001
	mouseeventfLeftDown = 0x0002
	mouseeventfLeftUp   = 0x0004
	smCxscreen          = 0
	smCyscreen          = 1
)

// mouseInput matches the MOUSEINPUT structure.
type mouseInput struct {
	Dx        int32
	Dy        int32
	MouseData uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

// input matches the INPUT structure for SendInput.
type input struct {
	Type uint32
	Mi   mouseInput
	_    [8]byte // padding to match union size
}

// ClickAt moves the cursor to absolute screen coordinates and performs a left click.
func ClickAt(x, y int) error {
	// Move cursor
	ret, _, err := procSetCursorPos.Call(uintptr(x), uintptr(y))
	if ret == 0 {
		return fmt.Errorf("SetCursorPos(%d, %d) failed: %v", x, y, err)
	}

	// Small delay for the cursor to settle
	time.Sleep(50 * time.Millisecond)

	// Get screen dimensions for absolute coordinate normalization
	screenW, _, _ := procGetSystemMetrics.Call(smCxscreen)
	screenH, _, _ := procGetSystemMetrics.Call(smCyscreen)

	absX := int32(float64(x) * 65535.0 / float64(screenW))
	absY := int32(float64(y) * 65535.0 / float64(screenH))

	// Send mouse down
	down := input{
		Type: inputMouse,
		Mi: mouseInput{
			Dx:    absX,
			Dy:    absY,
			Flags: mouseeventfAbsolute | mouseeventfMove | mouseeventfLeftDown,
		},
	}

	ret, _, err = procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&down)),
		uintptr(unsafe.Sizeof(down)),
	)
	if ret == 0 {
		return fmt.Errorf("SendInput(down) failed: %v", err)
	}

	time.Sleep(30 * time.Millisecond)

	// Send mouse up
	up := input{
		Type: inputMouse,
		Mi: mouseInput{
			Dx:    absX,
			Dy:    absY,
			Flags: mouseeventfAbsolute | mouseeventfMove | mouseeventfLeftUp,
		},
	}

	ret, _, err = procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&up)),
		uintptr(unsafe.Sizeof(up)),
	)
	if ret == 0 {
		return fmt.Errorf("SendInput(up) failed: %v", err)
	}

	return nil
}

// RegionToScreen converts a point relative to a capture region into absolute screen coordinates.
func RegionToScreen(region image.Rectangle, relX, relY int) (int, int) {
	return region.Min.X + relX, region.Min.Y + relY
}
