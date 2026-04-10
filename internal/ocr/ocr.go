// Package ocr provides Tesseract CLI wrapper for text detection with bounding boxes.
package ocr

import (
	"fmt"
	"image"
	"os/exec"
	"strconv"
	"strings"
)

// TextMatch represents a detected word with its bounding box and confidence.
type TextMatch struct {
	Text       string
	Bounds     image.Rectangle // bounding box in image coordinates
	Confidence int
}

// DetectText runs Tesseract OCR on an image file and returns all detected words
// with their bounding boxes.
func DetectText(imagePath string) ([]TextMatch, error) {
	// Check tesseract is available
	tesseractPath, err := exec.LookPath("tesseract")
	if err != nil {
		return nil, fmt.Errorf("tesseract not found on PATH. Install from https://github.com/UB-Mannheim/tesseract/wiki : %w", err)
	}

	// Run tesseract with TSV output for word-level bounding boxes
	// --psm 6 = Assume a single uniform block of text
	cmd := exec.Command(tesseractPath, imagePath, "stdout", "-l", "eng", "--psm", "6", "tsv")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("tesseract execution failed: %w", err)
	}

	return parseTSV(string(output))
}

// parseTSV parses Tesseract TSV output into TextMatch slices.
// TSV columns: level page_num block_num par_num line_num word_num left top width height conf text
func parseTSV(tsv string) ([]TextMatch, error) {
	var matches []TextMatch
	lines := strings.Split(strings.TrimSpace(tsv), "\n")

	for i, line := range lines {
		if i == 0 {
			continue // skip header
		}

		fields := strings.Split(strings.TrimSpace(line), "\t")
		if len(fields) < 12 {
			continue
		}

		text := strings.TrimSpace(fields[11])
		if text == "" || text == " " {
			continue
		}

		conf, err := strconv.Atoi(strings.TrimSpace(fields[10]))
		if err != nil || conf < 0 {
			continue
		}

		left, _ := strconv.Atoi(strings.TrimSpace(fields[6]))
		top, _ := strconv.Atoi(strings.TrimSpace(fields[7]))
		width, _ := strconv.Atoi(strings.TrimSpace(fields[8]))
		height, _ := strconv.Atoi(strings.TrimSpace(fields[9]))

		matches = append(matches, TextMatch{
			Text:       text,
			Bounds:     image.Rect(left, top, left+width, top+height),
			Confidence: conf,
		})
	}

	return matches, nil
}

// FindKeywords searches detected text matches for any of the given keywords.
// Returns matches where the detected text contains a keyword (case-insensitive).
// Only includes results above the confidence threshold.
func FindKeywords(matches []TextMatch, keywords []string, confidenceThreshold int) []TextMatch {
	var found []TextMatch

	for _, m := range matches {
		if m.Confidence < confidenceThreshold {
			continue
		}
		textLower := strings.ToLower(m.Text)
		for _, kw := range keywords {
			kwLower := strings.ToLower(kw)
			if strings.Contains(textLower, kwLower) || strings.Contains(kwLower, textLower) {
				found = append(found, m)
				break
			}
		}
	}

	return found
}

// FindMultiWordKeywords performs multi-word keyword matching by combining
// adjacent detected words and checking against multi-word keywords.
func FindMultiWordKeywords(matches []TextMatch, keywords []string, confidenceThreshold int) []TextMatch {
	var found []TextMatch

	// Single-word matches
	found = append(found, FindKeywords(matches, keywords, confidenceThreshold)...)

	// Multi-word matching: combine consecutive words on similar Y positions
	for _, kw := range keywords {
		kwWords := strings.Fields(kw)
		if len(kwWords) <= 1 {
			continue
		}

		for i := 0; i <= len(matches)-len(kwWords); i++ {
			matched := true
			for j, kwWord := range kwWords {
				idx := i + j
				if idx >= len(matches) {
					matched = false
					break
				}
				if matches[idx].Confidence < confidenceThreshold {
					matched = false
					break
				}
				if !strings.EqualFold(matches[idx].Text, kwWord) {
					matched = false
					break
				}
			}
			if matched {
				// Create a combined bounding box
				first := matches[i]
				last := matches[i+len(kwWords)-1]
				combined := TextMatch{
					Text: kw,
					Bounds: image.Rect(
						first.Bounds.Min.X,
						first.Bounds.Min.Y,
						last.Bounds.Max.X,
						last.Bounds.Max.Y,
					),
					Confidence: first.Confidence,
				}
				found = append(found, combined)
			}
		}
	}

	return found
}
