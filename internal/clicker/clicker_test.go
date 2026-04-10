package clicker

import (
	"image"
	"testing"
)

func TestRegionToScreen(t *testing.T) {
	region := image.Rect(100, 200, 500, 600) // Screen region top-left at (100, 200)

	// A coordinate (50, 50) relative to the captured region
	relX := 50
	relY := 50

	absX, absY := RegionToScreen(region, relX, relY)

	expectedX := 150 // 100 + 50
	expectedY := 250 // 200 + 50

	if absX != expectedX || absY != expectedY {
		t.Errorf("expected (%d, %d), got (%d, %d)", expectedX, expectedY, absX, absY)
	}
}
