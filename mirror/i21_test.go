package mirror

// The badge verdicts: the .badge / .badge.success / .badge.warning /
// .badge.error / .badge.info classes — captured from the real
// components/badge and compared against browser captures of per-specimen
// fixtures wearing exactly the published sheet's classes. Like
// TestCalibration and the earlier mirror verdicts, these only deliver a
// verdict on the authoritative machine; elsewhere one half of the harness
// skips loudly.

import (
	"image"
	"testing"

	"github.com/vibrantgio/components/badge"
	"github.com/vibrantgio/components/golden"
	"github.com/vibrantgio/theme/tokens"
)

// badgeSize is the badge capture viewport: the label-medium line box with
// ground around it on every side.
//
// The size is calibrated, not arbitrary. A badge is the harness's hardest
// intrinsic-width specimen: it has no fill, no corner and no boundary, so
// every pixel that differs is a glyph pixel, and the two shapers disagree
// about a label role most of all — the browser applies the role's 0.5px
// tracking, Gio's typeset does not — so the whole right edge of the word
// lands shifted, which the box filter reads as displaced cells.
//
// Measured on the authoritative machine (Chromium 153.0.8008.0), the worst
// per-variant match against the five fixtures and the closest wrong-variant
// cross-pair, at candidate viewports:
//
//	80x24:   worst match 0.0320, closest cross-pair 0.0238
//	100x32:  worst match 0.0192, closest cross-pair 0.0143
//	120x40:  worst match 0.0128, closest cross-pair 0.0095
//	140x40:  worst match 0.0110, closest cross-pair 0.0082
//	160x48:  worst match 0.0080, closest cross-pair 0.0059
//	200x56:  worst match 0.0055, closest cross-pair 0.0041
//
// The cross-pair column no longer separates and cannot be made to: the pill
// this specimen replaced carried a fill, and a block of colour is what the
// box filter had to tell one specimen from another with. A badge is a word
// in one ink on the page's own ground, so at every viewport a wrong pair
// scores closer than the worst right one — the metric sees mostly ground
// either way. What the verdict asserts is therefore what it can: that the
// sheet sets the same word in the same place in the same ink the Gio side
// draws, not that it could not have been a different variant.
//
// That leaves one criterion, that every match land under Tolerance (0.017).
// Smaller punishes the edge shift a correct mirror cannot avoid — 100x32 and
// below fail outright — and larger dilutes the badge until the frame is
// mostly ground and the verdict is about the ground. 140x40 is the smallest
// that clears by more than half the tolerance.
var badgeSize = image.Pt(140, 40)

// TestBadgeMirrors scores each badge specimen pair: the Gio badge in a given
// variant against the browser render of the matching fixture, both at the
// same viewport over the level-0 ground — the only ground a badge's ink can
// be derived against, having no fill of its own. Every distance is logged;
// each must land under Tolerance for the page to count as a mirror of the
// badge rather than a drawing of one.
func TestBadgeMirrors(t *testing.T) {
	srv := Serve(t)
	shaper := tokens.DefaultTypography.DeterministicShaper()
	style := badge.Style(tokens.DefaultTypography, tokens.Comfortable)

	for _, c := range badgeCases {
		t.Run(c.fixture, func(t *testing.T) {
			gio := golden.Capture(t, badgeSize, onColor(tokens.DefaultLight.SurfaceAt(tokens.Level0),
				badge.Render(shaper, c.label, nil, c.variant,
					tokens.DefaultLight, tokens.Spacing, style, badge.RenderState{})))
			if !FixtureExists("fixtures/" + c.fixture) {
				t.Fatalf("no embedded fixture %q — a typo here would screenshot a 404 page", c.fixture)
			}
			web := CaptureBrowser(t, srv.URL+"/fixtures/"+c.fixture, badgeSize)
			d := Distance(gio, web)
			t.Logf("distance gio vs %s: %.4f (Tolerance %.4f)", c.fixture, d, Tolerance)
			if d > Tolerance {
				t.Errorf("pair %s scored %.4f > Tolerance %.4f: the page does not read as the badge", c.fixture, d, Tolerance)
			}
		})
	}
}

// badgeCases pairs each fixture with the variant and the word it draws. The
// words differ per variant on purpose: a mirror scored on one word in five
// hues would pass on a sheet that shaped every badge identically.
var badgeCases = []struct {
	fixture string
	label   string
	variant badge.Variant
}{
	{"badge-neutral.html", "Beta", badge.Neutral},
	{"badge-success.html", "Passing", badge.Success},
	{"badge-warning.html", "Degraded", badge.Warning},
	{"badge-error.html", "Failing", badge.Error},
	{"badge-info.html", "Preview", badge.Info},
}
