package mirror

// The I2.1 status-tag verdicts: the .tag.success/.tag.warning/.tag.error
// classes — the status vocabulary the G3.2 validation found missing, which
// the composing agent answered by inline-styling an invented variant —
// captured from the real patterns/tag chip and compared against browser
// captures of per-specimen fixtures wearing exactly the published sheet's
// classes. Like TestCalibration and the earlier mirror verdicts, these only
// deliver a verdict on the authoritative machine; elsewhere one half of the
// harness skips loudly.

import (
	"image"
	"testing"

	"github.com/vibrantgio/components/golden"
	"github.com/vibrantgio/patterns/tag"
	"github.com/vibrantgio/theme/tokens"
)

// tagSize is the status-chip capture viewport: room for a label-small pill
// (16 dp line box plus twice the 4 dp vertical padding) with ground around
// it, so the 1 dp level outline is read against the Surface pin on every
// side.
//
// The size is calibrated, not arbitrary. A tag is the first intrinsic-width
// specimen in the harness: unlike a button (which fills its viewport on
// both sides, so only glyph interiors disagree), the chip's width is its
// label's, and the two shapers disagree about label-small most of all —
// the browser applies the role's 0.5px tracking, Gio's typeset does not —
// so the whole right edge of the pill lands shifted, which the box filter
// reads as displaced cells. Measured on the authoritative machine
// (Chromium 153.0.8008.0, 2026-08-15), per-level matches against the three
// fixtures at candidate viewports, with wrong-level cross-pairs beside
// them:
//
//	120x40: matches 0.0180–0.0241, cross-pairs 0.0347–0.0452
//	160x48: matches 0.0113–0.0150, cross-pairs 0.0217–0.0282
//	200x56: matches 0.0077–0.0103, cross-pairs 0.0149–0.0194
//
// 160x48 is the one viewport where every match clears Tolerance (0.017)
// and every wrong-level pair still fails it — smaller punishes the edge
// shift a correct mirror cannot avoid, larger dilutes the chip until a
// wrong level reads as a match.
var tagSize = image.Pt(160, 48)

// TestStatusTagMirrors scores each I2.1 specimen pair: the patterns/tag
// status chip in a given level against the browser render of the matching
// fixture, both at the same viewport over the Surface pin — the ground a
// resting chip sits on, and the base its 20% level tint blends over. Every
// distance is logged; each must land under Tolerance for the page to count
// as a mirror of the chip rather than a drawing of one.
func TestStatusTagMirrors(t *testing.T) {
	srv := Serve(t)
	shaper := tokens.DefaultTypography.DeterministicShaper()

	cases := []struct {
		fixture string
		label   string
		variant tag.Variant
	}{
		{"tag-success.html", "Passing", tag.Success},
		{"tag-warning.html", "Degraded", tag.Warning},
		{"tag-error.html", "Failing", tag.Error},
	}

	for _, c := range cases {
		t.Run(c.fixture, func(t *testing.T) {
			gio := golden.Capture(t, tagSize, onColor(tokens.DefaultLight.Surface, tag.Render(
				shaper, c.label, c.variant,
				tokens.DefaultLight, tokens.Spacing, tokens.Radius,
				tokens.DefaultTypography.LabelSmall,
			)))
			if !FixtureExists("fixtures/" + c.fixture) {
				t.Fatalf("no embedded fixture %q — a typo here would screenshot a 404 page", c.fixture)
			}
			web := CaptureBrowser(t, srv.URL+"/fixtures/"+c.fixture, tagSize)
			d := Distance(gio, web)
			t.Logf("distance gio vs %s: %.4f (Tolerance %.4f)", c.fixture, d, Tolerance)
			if d > Tolerance {
				t.Errorf("pair %s scored %.4f > Tolerance %.4f: the mirror does not read as the chip", c.fixture, d, Tolerance)
			}
		})
	}
}
