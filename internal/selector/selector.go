// Package selector provides interactive screen region selection.
package selector

import (
	"bufio"
	"fmt"
	"image"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32         = windows.NewLazySystemDLL("user32.dll")
	procGetCursorPos = user32.NewProc("GetCursorPos")
)

// point matches the Windows POINT structure.
type point struct {
	X int32
	Y int32
}

// getCursorPos returns the current cursor position.
func getCursorPos() (int, int, error) {
	var pt point
	ret, _, err := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if ret == 0 {
		return 0, 0, fmt.Errorf("GetCursorPos failed: %v", err)
	}
	return int(pt.X), int(pt.Y), nil
}

// SelectRegion interactively prompts the user to select a screen region.
// User moves mouse to top-left corner and presses Enter, then bottom-right and Enter.
func SelectRegion() (image.Rectangle, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║        SCREEN REGION SELECTION                  ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Println("║  1. Move mouse to the TOP-LEFT corner           ║")
	fmt.Println("║  2. Press [Enter] in this console               ║")
	fmt.Println("║  3. Move mouse to the BOTTOM-RIGHT corner       ║")
	fmt.Println("║  4. Press [Enter] in this console               ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Println()

	// Get top-left
	fmt.Print(">>> Move mouse to TOP-LEFT corner, then press [Enter]: ")
	_, _ = reader.ReadString('\n')
	x1, y1, err := getCursorPos()
	if err != nil {
		return image.Rectangle{}, fmt.Errorf("failed to get top-left position: %w", err)
	}
	fmt.Printf("    ✓ Top-Left captured: (%d, %d)\n", x1, y1)

	// Get bottom-right
	fmt.Print("\n>>> Move mouse to BOTTOM-RIGHT corner, then press [Enter]: ")
	_, _ = reader.ReadString('\n')
	x2, y2, err := getCursorPos()
	if err != nil {
		return image.Rectangle{}, fmt.Errorf("failed to get bottom-right position: %w", err)
	}
	fmt.Printf("    ✓ Bottom-Right captured: (%d, %d)\n", x2, y2)

	// Normalize so Min is always top-left
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	rect := image.Rect(x1, y1, x2, y2)
	if rect.Dx() < 10 || rect.Dy() < 10 {
		return image.Rectangle{}, fmt.Errorf("selected region too small: %dx%d", rect.Dx(), rect.Dy())
	}

	fmt.Printf("\n    ✓ Region: %dx%d pixels at (%d,%d)\n\n", rect.Dx(), rect.Dy(), rect.Min.X, rect.Min.Y)
	return rect, nil
}
