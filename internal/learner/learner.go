// Package learner scans sample images in the img/ directory at startup,
// runs OCR on each, and auto-extracts unique keywords to merge with config.
package learner

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	_ "image/jpeg"

	"github.com/ethanpham86/AutoClickAccepted/internal/logger"
	"github.com/ethanpham86/AutoClickAccepted/internal/matcher"
	"github.com/ethanpham86/AutoClickAccepted/internal/ocr"
)

// LearnFromImages scans the imgDir for image files (png, jpg, jpeg, bmp),
// runs Tesseract OCR on each (optionally upscaled), and returns all unique
// detected words that can serve as auto-click keywords, alongside exactly matched templates.
func LearnFromImages(imgDir string) ([]string, []matcher.Template, error) {
	entries, err := os.ReadDir(imgDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil // img/ folder doesn't exist, nothing to learn
		}
		return nil, nil, fmt.Errorf("read img dir: %w", err)
	}

	wordSet := make(map[string]bool)
	var templates []matcher.Template
	imageExtensions := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".bmp": true,
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !imageExtensions[ext] {
			continue
		}

		imgPath := filepath.Join(imgDir, entry.Name())

		// Extract true 1:1 scale template for exact matching
		if f, err := os.Open(imgPath); err == nil {
			if srcImg, _, err := image.Decode(f); err == nil {
				templates = append(templates, matcher.Template{
					Name:  entry.Name(),
					Image: srcImg,
				})
			}
			f.Close()
		}

		// Upscale the image before OCR for better accuracy
		scaledPath, err := upscaleImage(imgPath)
		if err != nil {
			logger.Error("[LEARN] ✗ Failed to upscale %s: %v", entry.Name(), err)
			continue
		}
		defer os.Remove(scaledPath)

		matches, err := ocr.DetectText(scaledPath)
		if err != nil {
			logger.Error("[LEARN] ✗ OCR failed for %s: %v", entry.Name(), err)
			continue
		}

		for _, m := range matches {
			// Accept all words for learning; real-time matching applies its own threshold
			if m.Confidence < 0 {
				continue
			}
			word := strings.TrimSpace(m.Text)
			// Filter out noise: too short, only symbols, numeric, or common stop words
			if len(word) < 3 {
				continue
			}
			if isNoise(word) || isStopWord(word) {
				continue
			}
			wordSet[word] = true
		}

		// Warning if too many words are found in a single image (indicates full screen screenshot rather than cropped button)
		var localSet []string
		for _, m := range matches {
			if m.Confidence >= 0 && len(strings.TrimSpace(m.Text)) >= 3 && !isNoise(strings.TrimSpace(m.Text)) && !isStopWord(strings.TrimSpace(m.Text)) {
				localSet = append(localSet, strings.TrimSpace(m.Text))
			}
		}

		if len(localSet) > 10 {
			logger.Error("[LEARN] ⚠ BÁO ĐỘNG: Tìm thấy %d từ trong file %s", len(localSet), entry.Name())
			logger.Error("[LEARN] ⚠ Bức ảnh này quá lớn (Full screen)! Chức năng Auto-Learner chỉ hoạt động với ẢNH CẮT NHỎ CỦA NÚT BẤM (Cropped Button). Hãy xoá bức ảnh này đi!")
		} else {
			logger.Info("[LEARN] ✓ Scanned %s → %d words extracted", entry.Name(), len(localSet))
		}
	}

	// Convert set to slice
	var words []string
	for w := range wordSet {
		words = append(words, w)
	}
	return words, templates, nil
}

// MergeKeywords merges learned keywords with existing config keywords.
// Duplicates (case-insensitive) are removed.
func MergeKeywords(existing, learned []string) []string {
	seen := make(map[string]bool)
	var merged []string

	for _, kw := range existing {
		key := strings.ToLower(strings.TrimSpace(kw))
		if !seen[key] {
			seen[key] = true
			merged = append(merged, kw)
		}
	}

	for _, kw := range learned {
		key := strings.ToLower(strings.TrimSpace(kw))
		if !seen[key] {
			seen[key] = true
			merged = append(merged, kw)
		}
	}

	return merged
}

// upscaleImage loads an image, upscales it 3x using nearest neighbor,
// saves to a temp file and returns the path.
func upscaleImage(srcPath string) (string, error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("open image: %w", err)
	}
	defer f.Close()

	srcImg, _, err := image.Decode(f)
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}

	bounds := srcImg.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	scale := 3
	newWidth := width * scale
	newHeight := height * scale

	img := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := srcImg.At(x, y).RGBA()
			r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)
			relX := x - bounds.Min.X
			relY := y - bounds.Min.Y
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					dstY := relY*scale + dy
					dstX := relX*scale + dx
					offset := (dstY*newWidth + dstX) * 4
					img.Pix[offset+0] = r8
					img.Pix[offset+1] = g8
					img.Pix[offset+2] = b8
					img.Pix[offset+3] = a8
				}
			}
		}
	}

	tmpFile, err := os.CreateTemp("", "learn_scaled_*.png")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer tmpFile.Close()

	if err := png.Encode(tmpFile, img); err != nil {
		return "", fmt.Errorf("encode png: %w", err)
	}

	return tmpFile.Name(), nil
}

// isNoise returns true if the word is just symbols, numbers, or common OCR garbage.
func isNoise(word string) bool {
	noiseChars := "~!@#$%^&*()_+-=[]{}|;':\",.<>?\\`"
	cleanWord := word
	for _, c := range noiseChars {
		cleanWord = strings.ReplaceAll(cleanWord, string(c), "")
	}
	if len(cleanWord) < 3 {
		return true
	}
	// Contains path separators = filesystem path, not a button
	if strings.Contains(word, "\\") || strings.Contains(word, "/") {
		return true
	}
	// Pure number
	allDigit := true
	for _, c := range cleanWord {
		if c < '0' || c > '9' {
			allDigit = false
			break
		}
	}
	return allDigit
}

// isStopWord returns true if the word is a common English stop word
// that should NOT be used as a click keyword.
func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "is": true, "in": true, "at": true, "to": true,
		"for": true, "on": true, "by": true, "an": true, "or": true,
		"and": true, "of": true, "it": true, "if": true, "no": true,
		"not": true, "this": true, "that": true, "with": true, "from": true,
		"than": true, "then": true, "now": true, "i'm": true, "was": true,
		"are": true, "has": true, "had": true, "have": true, "will": true,
		"you": true, "your": true, "use": true, "over": true, "where": true,
		"like": true, "less": true, "more": true, "also": true, "very": true,
		"file": true, "files": true, "tool": true, "tools": true, "doc": true,
		"general": true, "specific": true, "rather": true, "ones": true,
		"possible.": true, "methods.": true, "actions.": true, "directly,": true,
		"broader,": true, "selection,": true, "goal": true, "aiming": true,
		"focusing": true, "optimize": true, "emphasizing": true, "leveraging": true,
		"resorting": true, "direct": true, "targeted": true, "usage": true,
		"thinking": true, "thought": true, "edited": true, "running": true,
		"background": true, "command": true, "changes": true, "prioritizing": true,
	}
	return stopWords[strings.ToLower(word)]
}
