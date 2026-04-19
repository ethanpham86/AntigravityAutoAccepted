// Package clicker provides mouse control using Windows API.
// Supports both physical (SendInput) and background (PostMessage) clicking.
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

	procSetCursorPos     = user32.NewProc("SetCursorPos")
	procSendInput        = user32.NewProc("SendInput")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")

	// PostMessage APIs
	procWindowFromPhysicalPoint = user32.NewProc("WindowFromPhysicalPoint")
	procScreenToClient          = user32.NewProc("ScreenToClient")
	procPostMessageW    = user32.NewProc("PostMessageW")
)

const (
	inputMouse             = 0
	mouseeventfAbsolute    = 0x8000
	mouseeventfMove        = 0x0001
	mouseeventfLeftDown    = 0x0002
	mouseeventfLeftUp      = 0x0004
	mouseeventfVirtualDesk = 0x4000 // Coordinates map to entire virtual desktop

	wm_LButtonDown = 0x0201
	wm_LButtonUp   = 0x0202
	mk_LButton     = 0x0001
)

type mouseInput struct {
	Dx        int32
	Dy        int32
	MouseData uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type input struct {
	Type uint32
	Mi   mouseInput
}

type point struct {
	X int32
	Y int32
}

// ClickAt routes to either background click or physical click
func ClickAt(x, y int, background bool) error {
	if background {
		err := backgroundClickAt(x, y)
		if err == nil {
			return nil
		}
		fmt.Printf("[WARN] BackgroundClick failed (%v) - Falling back to physical click!\n", err)
		// If background click fails, fallback to physical click
	}
	return physicalClickAt(x, y)
}

func backgroundClickAt(x, y int) error {
	pt := point{X: int32(x), Y: int32(y)}
	
	// Pass POINT struct by value (packed into 64-bit uint)
	pt64 := *(*uint64)(unsafe.Pointer(&pt))
	hwnd, _, _ := procWindowFromPhysicalPoint.Call(uintptr(pt64))
	if hwnd == 0 {
		return fmt.Errorf("WindowFromPhysicalPoint failed at (%d, %d)", x, y)
	}

	// Clone point to convert to client coordinates
	clientPt := pt
	ret, _, err := procScreenToClient.Call(hwnd, uintptr(unsafe.Pointer(&clientPt)))
	if ret == 0 {
		return fmt.Errorf("ScreenToClient failed: %v", err)
	}

	// Construct lParam: MAKELPARAM(x, y)
	lParam := uintptr((uint32(clientPt.Y) & 0xFFFF) << 16 | (uint32(clientPt.X) & 0xFFFF))

	// Mouse Move - Qt/Electron require mouse to move before accepting click
	const wm_MouseMove = 0x0200
	procPostMessageW.Call(hwnd, uintptr(wm_MouseMove), 0, lParam)
	time.Sleep(20 * time.Millisecond)

	// Mouse Down
	procPostMessageW.Call(hwnd, uintptr(wm_LButtonDown), uintptr(mk_LButton), lParam)
	
	time.Sleep(30 * time.Millisecond)
	
	// Mouse Up
	procPostMessageW.Call(hwnd, uintptr(wm_LButtonUp), 0, lParam)

	return nil
}

// physicalClickAt moves the cursor to absolute screen coordinates and performs a left click.
func physicalClickAt(x, y int) error {
	screenW, _, _ := procGetSystemMetrics.Call(78)
	screenH, _, _ := procGetSystemMetrics.Call(79)
	originXu, _, _ := procGetSystemMetrics.Call(76)
	originYu, _, _ := procGetSystemMetrics.Call(77)

	originX := int(int32(originXu))
	originY := int(int32(originYu))

	if x < originX || x >= originX+int(screenW) || y < originY || y >= originY+int(screenH) {
		return fmt.Errorf("coordinates (%d, %d) outside bounds", x, y)
	}

	ret, _, err := procSetCursorPos.Call(uintptr(x), uintptr(y))
	if ret == 0 {
		return fmt.Errorf("SetCursorPos(%d, %d) failed: %v", x, y, err)
	}

	time.Sleep(50 * time.Millisecond)

	down := input{
		Type: inputMouse,
		Mi: mouseInput{
			Flags: mouseeventfLeftDown,
		},
	}
	ret, _, err = procSendInput.Call(1, uintptr(unsafe.Pointer(&down)), uintptr(unsafe.Sizeof(down)))
	if ret == 0 {
		return fmt.Errorf("SendInput(down) failed: %v", err)
	}

	time.Sleep(30 * time.Millisecond)

	up := input{
		Type: inputMouse,
		Mi: mouseInput{
			Flags: mouseeventfLeftUp,
		},
	}
	ret, _, err = procSendInput.Call(1, uintptr(unsafe.Pointer(&up)), uintptr(unsafe.Sizeof(up)))
	if ret == 0 {
		return fmt.Errorf("SendInput(up) failed: %v", err)
	}

	return nil
}

// RegionToScreen converts a point relative to a capture region into absolute screen coordinates.
func RegionToScreen(region image.Rectangle, relX, relY int) (int, int) {
	return region.Min.X + relX, region.Min.Y + relY
}
