package ocr

import (
	"image"
	"reflect"
	"testing"
)

func TestParseTSV(t *testing.T) {
	// Simulated tesseract TSV output
	tsvData := `level	page_num	block_num	par_num	line_num	word_num	left	top	width	height	conf	text
5	1	1	1	1	1	10	20	50	15	95	Auto
5	1	1	1	1	2	65	20	45	15	91	Click
5	1	1	1	1	3	120	20	60	15	12	noise
5	1	1	1	1	4	190	20	80	15	99	ACCEPTED`

	matches, err := parseTSV(tsvData)
	if err != nil {
		t.Fatalf("parseTSV returned error: %v", err)
	}

	expected := []TextMatch{
		{"Auto", image.Rect(10, 20, 60, 35), 95},
		{"Click", image.Rect(65, 20, 110, 35), 91},
		{"noise", image.Rect(120, 20, 180, 35), 12}, // Low conf parsing
		{"ACCEPTED", image.Rect(190, 20, 270, 35), 99},
	}

	if len(matches) != len(expected) {
		t.Fatalf("expected %d matches, got %d", len(expected), len(matches))
	}

	for i := range expected {
		if !reflect.DeepEqual(matches[i], expected[i]) {
			t.Errorf("match %d mismatch:\nExpected: %+v\nGot: %+v", i, expected[i], matches[i])
		}
	}
}

func TestFindKeywords(t *testing.T) {
	matches := []TextMatch{
		{"Hello", image.Rect(0, 0, 10, 10), 90},
		{"ACCEPTED", image.Rect(0, 0, 10, 10), 95},
		{"World", image.Rect(0, 0, 10, 10), 80},
		{"Allow", image.Rect(0, 0, 10, 10), 40}, // Below threshold
	}

	keywords := []string{"accepted", "ALLOW"}
	threshold := 85

	found := FindMultiWordKeywords(matches, keywords, threshold)

	if len(found) != 1 {
		t.Fatalf("expected 1 match above threshold, got %d", len(found))
	}
	if found[0].Text != "ACCEPTED" {
		t.Errorf("expected 'ACCEPTED' match, got %s", found[0].Text)
	}
}

func TestFindMultiWordKeywords(t *testing.T) {
	matches := []TextMatch{
		{"Allow", image.Rect(10, 10, 50, 20), 95},
		{"Once", image.Rect(55, 10, 95, 20), 94},
	}

	keywords := []string{"Allow Once"}
	threshold := 90

	found := FindMultiWordKeywords(matches, keywords, threshold)

	if len(found) != 1 {
		t.Fatalf("expected 1 multi-word match, got %d", len(found))
	}
	if found[0].Text != "Allow Once" {
		t.Errorf("expected 'Allow Once', got %s", found[0].Text)
	}

	expectedRect := image.Rect(10, 10, 95, 20)
	if !found[0].Bounds.Eq(expectedRect) {
		t.Errorf("expected bounds %v, got %v", expectedRect, found[0].Bounds)
	}
}
