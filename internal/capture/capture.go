// Package capture provides screen region capture using Windows GDI API.
// No CGO — pure syscall Windows GDI calls.
package capture

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")

	procGetDC             = user32.NewProc("GetDC")
	procReleaseDC         = user32.NewProc("ReleaseDC")
	procCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject      = gdi32.NewProc("SelectObject")
	procBitBlt            = gdi32.NewProc("BitBlt")
	procDeleteObject      = gdi32.NewProc("DeleteObject")
	procDeleteDC          = gdi32.NewProc("DeleteDC")
	procGetDIBits         = gdi32.NewProc("GetDIBits")
)

const (
	srccopy    = 0x00CC0020
	biRGB      = 0
	dibRGBColors = 0
)

// bitmapInfoHeader is the BITMAPINFOHEADER structure.
type bitmapInfoHeader struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

// CaptureRegion captures a screen rectangle and returns it as an image.RGBA
// upscaled by the given scale factor. Passing scale=1 returns exact 1:1 size.
func CaptureRegion(rect image.Rectangle, scale int) (*image.RGBA, error) {
	width := rect.Dx()
	height := rect.Dy()
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid region: width=%d, height=%d", width, height)
	}

	// Get the device context for the entire screen
	hdcScreen, _, _ := procGetDC.Call(0)
	if hdcScreen == 0 {
		return nil, fmt.Errorf("GetDC failed")
	}
	defer procReleaseDC.Call(0, hdcScreen)

	// Create a compatible DC
	hdcMem, _, _ := procCreateCompatibleDC.Call(hdcScreen)
	if hdcMem == 0 {
		return nil, fmt.Errorf("CreateCompatibleDC failed")
	}
	defer procDeleteDC.Call(hdcMem)

	// Create a compatible bitmap
	hBitmap, _, _ := procCreateCompatibleBitmap.Call(hdcScreen, uintptr(width), uintptr(height))
	if hBitmap == 0 {
		return nil, fmt.Errorf("CreateCompatibleBitmap failed")
	}
	defer procDeleteObject.Call(hBitmap)

	// Select the bitmap into the memory DC
	procSelectObject.Call(hdcMem, hBitmap)

	// BitBlt from screen to memory DC
	ret, _, err := procBitBlt.Call(
		hdcMem, 0, 0, uintptr(width), uintptr(height),
		hdcScreen, uintptr(rect.Min.X), uintptr(rect.Min.Y),
		srccopy,
	)
	if ret == 0 {
		return nil, fmt.Errorf("BitBlt failed: %v", err)
	}

	// Prepare BITMAPINFOHEADER
	bmi := bitmapInfoHeader{
		BiSize:        uint32(unsafe.Sizeof(bitmapInfoHeader{})),
		BiWidth:       int32(width),
		BiHeight:      -int32(height), // top-down
		BiPlanes:      1,
		BiBitCount:    32,
		BiCompression: biRGB,
	}

	// Allocate buffer for pixel data
	imageSize := width * height * 4
	buf := make([]byte, imageSize)

	// Get the bitmap pixel data
	ret, _, err = procGetDIBits.Call(
		hdcMem, hBitmap, 0, uintptr(height),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bmi)),
		dibRGBColors,
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetDIBits failed: %v", err)
	}

	// ScaleFactor to improve Tesseract OCR accuracy on small UI texts, or scale=1 for exact pixel matching
	if scale < 1 {
		scale = 1
	}
	newWidth := width * scale
	newHeight := height * scale

	img := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcOffset := (y*width + x) * 4
			b := buf[srcOffset+0]
			g := buf[srcOffset+1]
			r := buf[srcOffset+2]
			
			// Nearest neighbor 3x upscale
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					newY := y*scale + dy
					newX := x*scale + dx
					dstOffset := (newY*newWidth + newX) * 4
					img.Pix[dstOffset+0] = r
					img.Pix[dstOffset+1] = g
					img.Pix[dstOffset+2] = b
					img.Pix[dstOffset+3] = 255
				}
			}
		}
	}

	return img, nil
}

// CaptureToFile captures a screen region and saves it as PNG to a temp file.
// Returns the file path.
func CaptureToFile(rect image.Rectangle, scale int) (string, error) {
	img, err := CaptureRegion(rect, scale)
	if err != nil {
		return "", fmt.Errorf("capture failed: %w", err)
	}

	tmpDir := os.TempDir()
	filePath := filepath.Join(tmpDir, "autoclick_capture.png")

	f, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("create file failed: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return "", fmt.Errorf("png encode failed: %w", err)
	}

	return filePath, nil
}
