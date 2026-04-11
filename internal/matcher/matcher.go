package matcher

import (
	"image"
	"path/filepath"
	"strings"
	"sync"
)

// Template represents a loaded image to strictly match against the screen.
type Template struct {
	Name  string
	Image image.Image
}

// Match represents a found template location.
type Match struct {
	TemplateName string
	Bounds       image.Rectangle
	Confidence   float64
}

// OpaquePixel stores the relative coordinate and pre-shifted RGB values of a valid template pixel
type OpaquePixel struct {
	tx, ty int
	r, g, b float64
}

// FastTemplate contains precomputed data to hyper-accelerate SAD matching
type FastTemplate struct {
	Name             string
	Width, Height    int
	OpaquePixels     []OpaquePixel
	CoarsePixels     []OpaquePixel
	AllowedCoarseDiff float64
}

// MatchSingle runs a multi-threaded Sum of Absolute Differences (SAD) template match.
// It searches for all template inside the screen. Returns matches, and the best miss info (name, confidence).
func MatchSingle(screen image.Image, templates []Template, threshold float64) ([]Match, string, float64) {
	var matches []Match
	var mu sync.Mutex
	
	globalBestConf := 0.0
	globalBestName := ""

	if len(templates) == 0 || screen == nil {
		return matches, globalBestName, globalBestConf
	}

	screenBounds := screen.Bounds()
	sW, sH := screenBounds.Dx(), screenBounds.Dy()

	// For very large screens, multi-threading drastically improves performance
	var wg sync.WaitGroup
	numWorkers := 4
	yChunks := sH / numWorkers
	if yChunks == 0 {
		yChunks = 1
		numWorkers = 1
	}

	for _, tmpl := range templates {
		tBounds := tmpl.Image.Bounds()
		tW, tH := tBounds.Dx(), tBounds.Dy()
		tMinX, tMinY := tBounds.Min.X, tBounds.Min.Y

		if tW > sW || tH > sH {
			continue // Template larger than screen
		}

		// Pre-compute Opaque Pixels for 1000x Speedup & Alpha Masking
		var opaquePixels []OpaquePixel
		for ty := 0; ty < tH; ty++ {
			for tx := 0; tx < tW; tx++ {
				tr, tg, tb, ta := tmpl.Image.At(tMinX+tx, tMinY+ty).RGBA()
				if (ta >> 8) >= 128 {
					opaquePixels = append(opaquePixels, OpaquePixel{
						tx: tx, ty: ty,
						r: float64(tr >> 8), g: float64(tg >> 8), b: float64(tb >> 8),
					})
				}
			}
		}

		if len(opaquePixels) == 0 {
			continue // Completely transparent template
		}

		maxDiffPerPixel := 255.0 * 3.0 // R, G, B differences combined
		maxTotalDiff := maxDiffPerPixel * float64(len(opaquePixels))
		allowedDiff := maxTotalDiff * (1.0 - threshold)

		// Create uniformly distributed CoarsePixels (Up to 200 pixels)
		var coarsePixels []OpaquePixel
		coarseTarget := 200
		if len(opaquePixels) > coarseTarget {
			step := len(opaquePixels) / coarseTarget
			for i := 0; i < len(opaquePixels); i += step {
				coarsePixels = append(coarsePixels, opaquePixels[i])
				if len(coarsePixels) == coarseTarget {
					break
				}
			}
		} else {
			coarsePixels = opaquePixels
		}
		allowedCoarseDiff := maxDiffPerPixel * float64(len(coarsePixels)) * (1.0 - threshold)

		fastTmpl := FastTemplate{
			Name:              tmpl.Name,
			Width:             tW,
			Height:            tH,
			OpaquePixels:      opaquePixels,
			CoarsePixels:      coarsePixels,
			AllowedCoarseDiff: allowedCoarseDiff,
		}

		for w := 0; w < numWorkers; w++ {
			startY := w * yChunks
			endY := startY + yChunks
			if w == numWorkers-1 {
				endY = sH - tH + 1
			} else {
				if endY > sH-tH+1 {
					endY = sH - tH + 1
				}
			}

			if startY >= endY {
				continue
			}

			wg.Add(1)
			go func(sy, ey int, fTmpl FastTemplate) {
				defer wg.Done()
				localMatches, localHighestConf := findMatchesInRegion(screen, fTmpl, sy, ey, allowedDiff, maxTotalDiff, maxDiffPerPixel)
				
				mu.Lock()
				if len(localMatches) > 0 {
					matches = append(matches, localMatches...)
				}
				if localHighestConf > globalBestConf {
					globalBestConf = localHighestConf
					nameBase := filepath.Base(fTmpl.Name)
					globalBestName = strings.TrimSuffix(nameBase, filepath.Ext(nameBase))
				}
				mu.Unlock()
			}(startY, endY, fastTmpl)
		}
	}
	wg.Wait()
	return matches, globalBestName, globalBestConf
}

// findMatchesInRegion sweeps the template across the screen y-region.
func findMatchesInRegion(screen image.Image, fTmpl FastTemplate, startY, endY int, allowedDiff, maxTotalDiff, maxDiffPerPixel float64) ([]Match, float64) {
	var matches []Match
	var localHighestConf float64 = 0.0

	screenBounds := screen.Bounds()
	sMaxX := screenBounds.Max.X
	sMinX := screenBounds.Min.X
	sMinY := screenBounds.Min.Y

	tW, tH := fTmpl.Width, fTmpl.Height
	endX := sMaxX - tW + 1

	for y := startY; y < endY; y++ {
		for x := sMinX; x < endX; x++ {
			
			// PASS 1: Coarse Grid Filter (200 uniformly sampled pixels)
			var coarseDiffAccum float64
			coarseFail := false
			coarseProcessed := 0
			for idx := range fTmpl.CoarsePixels {
				p := &fTmpl.CoarsePixels[idx]
				sr, sg, sb, _ := screen.At(x+p.tx, sMinY+y+p.ty).RGBA()

				rDiff := p.r - float64(sr>>8)
				gDiff := p.g - float64(sg>>8)
				bDiff := p.b - float64(sb>>8)

				if rDiff < 0 { rDiff = -rDiff }
				if gDiff < 0 { gDiff = -gDiff }
				if bDiff < 0 { bDiff = -bDiff }

				coarseDiffAccum += rDiff + gDiff + bDiff
				coarseProcessed++

				if coarseDiffAccum > fTmpl.AllowedCoarseDiff {
					coarseFail = true
					break
				}
			}

			if coarseFail {
				if coarseProcessed > 0 {
					avgDiff := coarseDiffAccum / float64(coarseProcessed)
					conf := 1.0 - (avgDiff / maxDiffPerPixel)
					if conf > localHighestConf { localHighestConf = conf }
				}
				continue // Skip the rest of this coordinate (Extremely fast early exit)
			}


			// PASS 2: Fine Grained Sweep (Only runs if Coarse Filter passes)
			var diffAccum float64
			earlyExit := false
			pixelsProcessed := 0

			// Lightning fast flattened SAD calculation on pre-filtered pixels
			for idx := range fTmpl.OpaquePixels {
				p := &fTmpl.OpaquePixels[idx]
				sr, sg, sb, _ := screen.At(x+p.tx, sMinY+y+p.ty).RGBA()

				rDiff := p.r - float64(sr>>8)
				gDiff := p.g - float64(sg>>8)
				bDiff := p.b - float64(sb>>8)

				if rDiff < 0 { rDiff = -rDiff }
				if gDiff < 0 { gDiff = -gDiff }
				if bDiff < 0 { bDiff = -bDiff }

				diffAccum += rDiff + gDiff + bDiff
				pixelsProcessed++

				if diffAccum > allowedDiff {
					earlyExit = true
					break
				}
			}

			// Calculate confidence
			var conf float64
			if !earlyExit {
				conf = 1.0 - (diffAccum / maxTotalDiff)
				nameBase := filepath.Base(fTmpl.Name)
				ext := filepath.Ext(nameBase)
				nameClean := strings.TrimSuffix(nameBase, ext)

				matches = append(matches, Match{
					TemplateName: nameClean,
					Bounds:       image.Rect(x, sMinY+y, x+tW, sMinY+y+tH),
					Confidence:   conf,
				})
			} else if pixelsProcessed > 0 {
				avgDiff := diffAccum / float64(pixelsProcessed)
				conf = 1.0 - (avgDiff / maxDiffPerPixel)
			}

			if conf > localHighestConf {
				localHighestConf = conf
			}
		}
	}
	return matches, localHighestConf
}
