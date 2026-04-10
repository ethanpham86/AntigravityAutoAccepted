// Package ocr provides Tesseract CLI wrapper for text detection with bounding boxes.
package ocr

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// TextMatch represents a detected word with its bounding box and confidence.
type TextMatch struct {
	Text       string          // Raw text extracted explicitly from Tesseract
	Keyword    string          // The targeted keyword it matched against
	Bounds     image.Rectangle // bounding box in image coordinates
	Confidence int
}

// DetectText runs Tesseract OCR on an image file and returns all detected words
// with their bounding boxes.
func DetectText(imagePath string) ([]TextMatch, error) {
	// Check tesseract is available
	tesseractPath, err := exec.LookPath("tesseract")
	if err != nil {
		// Fallback to default Windows install location
		fallbackPath := `C:\Program Files\Tesseract-OCR\tesseract.exe`
		if _, fallbackErr := os.Stat(fallbackPath); fallbackErr == nil {
			tesseractPath = fallbackPath
		} else {
			return nil, fmt.Errorf("tesseract not found on PATH or default location. Install from https://github.com/UB-Mannheim/tesseract/wiki : %w", err)
		}
	}

	// Run tesseract with TSV output for word-level bounding boxes
	// --psm 11 = Sparse text. Find as much text as possible (ideal for UI/AutoClickers)
	// -c debug_file=NUL forces Leptonica/Tesseract to suppress its C++ ObjectCache Memory Leak UI warnings
	cmd := exec.Command(tesseractPath, imagePath, "stdout", "-l", "eng", "--psm", "11", "-c", "debug_file=NUL", "tsv")

	// Disable window creation flash (Windows specific) to make it whisper quiet
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// Capture stdout and stderr separately.
	// Tesseract often writes warnings to stderr even on success.
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tesseract execution failed: %w (stderr: %s)", err, stderr.String())
	}

	return parseTSV(stdout.String())
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

		confFloat, err := strconv.ParseFloat(strings.TrimSpace(fields[10]), 64)
		if err != nil || confFloat < 0 {
			continue
		}
		conf := int(confFloat)

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
// Uses fuzzy matching: exact substring match OR Levenshtein distance ≤ 2.
// Only includes results above the confidence threshold.
func FindKeywords(matches []TextMatch, keywords []string, confidenceThreshold int) []TextMatch {
	var found []TextMatch

	for _, m := range matches {
		if m.Confidence < confidenceThreshold {
			continue
		}
		textLower := strings.ToLower(m.Text)
		// Strip common OCR artifacts (quotes, brackets, pipes)
		textCleaned := stripArtifacts(textLower)

		for _, kw := range keywords {
			kwLower := strings.ToLower(kw)

			// 1. EXACT MATCH (Highest Priority)
			if textCleaned == kwLower || textLower == kwLower {
				m.Keyword = kw
				found = append(found, m)
				break
			}

			// 2. Contains Match (Strictly guarded)
			// Only allow substring matching if the keyword is long enough (>= 5 chars)
			// e.g. "accept" inside "acceptall". Do not allow "run" inside "running"
			if len(kwLower) >= 5 && (strings.Contains(textCleaned, kwLower) || strings.Contains(textLower, kwLower)) {
				m.Keyword = kw
				found = append(found, m)
				break
			}

			// 3. Fuzzy match (Levenshtein) for single-word keywords
			// Short words (≤4 chars): max distance 1
			// Long words (≥5 chars): max distance 2
			if !strings.Contains(kwLower, " ") && len(textCleaned) >= 3 {
				lenDiff := len(textCleaned) - len(kwLower)
				if lenDiff < 0 {
					lenDiff = -lenDiff
				}
				maxDist := 1
				if len(kwLower) >= 5 && len(textCleaned) >= 5 {
					maxDist = 2
				}
				// Ensure length difference isn't too huge to avoid checking completely unrelated words
				if lenDiff <= 2 {
					if levenshtein(textCleaned, kwLower) <= maxDist || levenshtein(textLower, kwLower) <= maxDist {
						m.Keyword = kw
						found = append(found, m)
						break
					}
				}
			}
		}
	}

	return found
}

// FindMultiWordKeywords performs multi-word keyword matching by combining
// adjacent detected words and checking against multi-word keywords.
// Also performs concatenation matching for words Tesseract may merge together.
func FindMultiWordKeywords(matches []TextMatch, keywords []string, confidenceThreshold int) []TextMatch {
	var found []TextMatch

	// Single-word matches (with fuzzy)
	found = append(found, FindKeywords(matches, keywords, confidenceThreshold)...)

	// Multi-word matching: combine consecutive words on similar Y positions
	for _, kw := range keywords {
		kwWords := strings.Fields(kw)
		if len(kwWords) <= 1 {
			continue
		}

		// Strategy 1: Exact sequential word match
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
					Keyword: kw,
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

		// Strategy 2: Concatenation match (Tesseract sometimes merges words)
		// e.g. "Accept all" merged into "Acceptall" or "(Acceptall)"
		kwConcat := strings.ToLower(strings.ReplaceAll(kw, " ", ""))
		for _, m := range matches {
			if m.Confidence < confidenceThreshold {
				continue
			}
			// Strip common OCR bracket artifacts
			cleaned := strings.ToLower(m.Text)
			cleaned = strings.Trim(cleaned, "()[]{}|/\\")
			if cleaned == kwConcat || levenshtein(cleaned, kwConcat) <= 2 {
				found = append(found, TextMatch{
					Text:       kw,
					Keyword:    kw,
					Bounds:     m.Bounds,
					Confidence: m.Confidence,
				})
			}
		}
	}

	return found
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Use single row optimization
	prev := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr := make([]int, lb+1)
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = minInt(curr[j-1]+1, minInt(prev[j]+1, prev[j-1]+cost))
		}
		prev = curr
	}
	return prev[lb]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// stripArtifacts removes common OCR noise characters (quotes, brackets, pipes, backslashes).
func stripArtifacts(s string) string {
	artifacts := `"'()[]{}|\/><“”‘’`
	result := s
	for _, c := range artifacts {
		result = strings.ReplaceAll(result, string(c), "")
	}
	return strings.TrimSpace(result)
}
