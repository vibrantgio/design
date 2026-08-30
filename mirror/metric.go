package mirror

import (
	"fmt"
	"image"
	"math"
)

// cellPx is the box-filter cell size of [Distance], in source pixels. Four
// pixels is roughly the scale of antialiasing and glyph-shaping disagreement
// between two renderers: averaging over a 4×4 box washes those out while a
// wrong colour role, corner radius or control size — which move whole
// regions, not edge pixels — still shifts the cell means.
const cellPx = 4

// Tolerance is the perceptual distance below which two renders read as the
// same component, calibrated in TestCalibration from real pairs on the
// authoritative machine (see the package comment) rather than in the
// abstract (Chromium 153.0.8008.0, darwin/arm64, 2026-08-14):
//
//	Gio filled button vs its bundle-token HTML mirror:  0.0106
//	stability: same page captured twice in Chromium:    0.0000
//	vs HTML with the wrong colour role (error):         0.3403
//	vs HTML with the wrong radius (pill):               0.0271
//	vs HTML with the wrong size (compact 28dp):         0.1425
//
// The threshold sits at 0.017, the geometric midpoint of the two clusters'
// nearest members: 1.6× the measured matching distance and 1.6× below the
// nearest wrong variant (the pill radius), so both clear it with equal
// margin. The floor must be measured, not derived from byte-count divergence
// between two Gio backends: that divergence is antialiasing and gradient
// dithering, exactly the noise the box filter averages away. The measured
// Chrome-vs-Gio matching distance (0.0106 — different text shaper, different
// rasteriser, different gamma handling) IS the cross-renderer floor under
// this metric, and Tolerance stands above it; a tolerance below that number
// would fail a correct mirror.
const Tolerance = 0.017

// Distance reports the perceptual distance between two equally-sized images
// in [0,1]: both are box-downscaled to cells of [cellPx]×[cellPx] source
// pixels, and the result is the mean Euclidean RGB distance between
// corresponding cell means, normalised so 1.0 is black-vs-white in every
// cell. 0 means the downscaled images are identical.
//
// It panics on unequal sizes, for golden.PixelDiff's reason: a distance has
// no answer for two images of different shapes, and both halves of this
// harness capture at the same requested viewport, so differing bounds is a
// defect in the test, not an outcome.
func Distance(a, b *image.RGBA) float64 {
	if a.Bounds().Size() != b.Bounds().Size() {
		panic(fmt.Sprintf("mirror: Distance: images must have equal sizes, got %v and %v",
			a.Bounds().Size(), b.Bounds().Size()))
	}
	ca, cw, ch := downscale(a)
	cb, _, _ := downscale(b)
	var sum float64
	for i := 0; i < cw*ch; i++ {
		dr := ca[3*i] - cb[3*i]
		dg := ca[3*i+1] - cb[3*i+1]
		db := ca[3*i+2] - cb[3*i+2]
		sum += math.Sqrt(dr*dr + dg*dg + db*db)
	}
	// Normalise: the largest per-cell distance is √3·255 (black vs white).
	return sum / (float64(cw*ch) * math.Sqrt(3) * 255)
}

// downscale box-filters img into cells of cellPx×cellPx source pixels and
// returns the per-cell mean RGB (row-major, 3 float64 per cell) plus the
// cell-grid dimensions. Ragged edge cells average over their real pixels, so
// no border is dropped and no pixel is counted twice. Alpha is ignored: both
// capture paths deliver fully opaque frames.
func downscale(img *image.RGBA) (cells []float64, cw, ch int) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	cw = (w + cellPx - 1) / cellPx
	ch = (h + cellPx - 1) / cellPx
	cells = make([]float64, 3*cw*ch)
	counts := make([]int, cw*ch)
	for y := 0; y < h; y++ {
		row := (y / cellPx) * cw
		off := y * img.Stride
		for x := 0; x < w; x++ {
			c := row + x/cellPx
			p := off + x*4
			cells[3*c] += float64(img.Pix[p])
			cells[3*c+1] += float64(img.Pix[p+1])
			cells[3*c+2] += float64(img.Pix[p+2])
			counts[c]++
		}
	}
	for c, n := range counts {
		cells[3*c] /= float64(n)
		cells[3*c+1] /= float64(n)
		cells[3*c+2] /= float64(n)
	}
	return cells, cw, ch
}
