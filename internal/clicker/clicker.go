// Package clicker provides mouse control using Windows SendInput API.
// Supports multi-monitor setups via MOUSEEVENTF_VIRTUALDESK.
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
	inputMouse              = 0
	mouseeventfAbsolute     = 0x8000
	mouseeventfMove         = 0x0001
	mouseeventfLeftDown     = 0x0002
	mouseeventfLeftUp       = 0x0004
	mouseeventfVirtualDesk  = 0x4000 // Coordinates map to entire virtual desktop
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
// WARNING: On x64 Windows, sizeof(INPUT) MUST be exactly 40 bytes.
// Type(4) + padding(4) + mouseInput(32) = 40.
// Do NOT add extra padding — SendInput will reject wrong cbSize.
type input struct {
	Type uint32
	Mi   mouseInput
}

// ClickAt moves the cursor to absolute screen coordinates and performs a left click.
// Supports multi-monitor setups by using virtual desktop coordinates.
func ClickAt(x, y int) error {
	// Get VIRTUAL screen dimensions for multi-monitor support
	screenW, _, _ := procGetSystemMetrics.Call(78) // SM_CXVIRTUALSCREEN
	screenH, _, _ := procGetSystemMetrics.Call(79) // SM_CYVIRTUALSCREEN
	originXu, _, _ := procGetSystemMetrics.Call(76) // SM_XVIRTUALSCREEN (can be negative)
	originYu, _, _ := procGetSystemMetrics.Call(77) // SM_YVIRTUALSCREEN (can be negative)

	// Cast to signed int32 first — origin can be negative when monitor is left of primary
	originX := int(int32(originXu))
	originY := int(int32(originYu))

	// Validate coordinates are within the virtual screen bounds
	if x < originX || x >= originX+int(screenW) || y < originY || y >= originY+int(screenH) {
		return fmt.Errorf("coordinates (%d, %d) outside virtual screen bounds [%d,%d]-[%d,%d]",
			x, y, originX, originY, originX+int(screenW), originY+int(screenH))
	}

	// Move cursor first
	ret, _, err := procSetCursorPos.Call(uintptr(x), uintptr(y))
	if ret == 0 {
		return fmt.Errorf("SetCursorPos(%d, %d) failed: %v", x, y, err)
	}

	// Small delay for the cursor to settle
	time.Sleep(50 * time.Millisecond)

	// Since SetCursorPos already moved the mouse to the exact display pixel,
	// we just need to send a LEFT_DOWN and LEFT_UP at the cursor's CURRENT position.
	// We do NOT use MOUSEEVENTF_ABSOLUTE or Dx/Dy here.
	
	// Send mouse down
	down := input{
		Type: inputMouse,
		Mi: mouseInput{
			Dx:    0,
			Dy:    0,
			Flags: mouseeventfLeftDown,
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
			Dx:    0,
			Dy:    0,
			Flags: mouseeventfLeftUp,
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
