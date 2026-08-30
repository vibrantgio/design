package mirror

// The status-tag verdicts: the .tag.success/.tag.warning/.tag.error
// classes — captured from the real patterns/tag chip and compared against browser
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
// (16 dp line box plus the S1 stop spent once across both edges) with ground
// around it, so the 1 dp level outline is read against the Surface pin on
// every side.
//
// The size is calibrated, not arbitrary. A tag is the first intrinsic-width
// specimen in the harness: unlike a button (which fills its viewport on
// both sides, so only glyph interiors disagree), the chip's width is its
// label's, and the two shapers disagree about label-small most of all —
// the browser applies the role's 0.5px tracking, Gio's typeset does not —
// so the whole right edge of the pill lands shifted, which the box filter
// reads as displaced cells.
//
// Measured on the authoritative machine
// (Chromium 153.0.8008.0), the worst per-level match against the three
// fixtures and the closest wrong-level cross-pair, at candidate viewports:
//
//	100x32: worst match 0.0272, closest cross-pair 0.0488
//	120x40: worst match 0.0181, closest cross-pair 0.0325
//	140x40: worst match 0.0156, closest cross-pair 0.0279
//	160x44: worst match 0.0124, closest cross-pair 0.0222
//	160x48: worst match 0.0113, closest cross-pair 0.0203
//	180x48: worst match 0.0101, closest cross-pair 0.0181
//	200x56: worst match 0.0078, closest cross-pair 0.0139
//	240x64: worst match 0.0057, closest cross-pair 0.0102
//
// The viewport has to put every match under Tolerance (0.017) and leave
// every wrong-level pair over it — smaller punishes the edge shift a
// correct mirror cannot avoid, larger dilutes the chip until a wrong level
// reads as a match, which 200x56 and 240x64 both now do. 160x48 keeps the
// widest margin on the side that can fail either way: matches clear by
// 0.0057 and cross-pairs fail by 0.0033, where 180x48 leaves the cross-pair
// only 0.0011 of room.
var tagSize = image.Pt(160, 48)

// TestStatusTagMirrors scores each status-tag specimen pair: the patterns/tag
// status chip in a given level against the browser render of the matching
// fixture, both at the same viewport over the Surface pin — the ground a
// resting chip sits on, and the pane its level container separates from.
// Every distance is logged; each must land under Tolerance for the page to
// count as a mirror of the chip rather than a drawing of one.
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
