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

// badgeSize is the badge capture viewport: the badge's own box with ground
// around it on every side.
//
// The size is calibrated, not arbitrary. The two shapers disagree about a
// label role most of all — the browser applies the role's 0.5px tracking,
// Gio's typeset does not — so the whole right edge of the word lands shifted,
// which the box filter reads as displaced cells. Too small a frame and that
// shift is most of the picture; too large and the badge is diluted until the
// verdict is about the ground.
//
// Measured on the authoritative machine (Chromium 153.0.8008.0), the worst
// per-variant match against the five fixtures and the closest wrong-variant
// cross-pair, at candidate viewports:
//
//	80x24:   worst match 0.0297, closest cross-pair 0.0571
//	100x32:  worst match 0.0178, closest cross-pair 0.0343
//	120x40:  worst match 0.0119, closest cross-pair 0.0229
//	140x40:  worst match 0.0102, closest cross-pair 0.0196
//	160x48:  worst match 0.0074, closest cross-pair 0.0143
//	200x56:  worst match 0.0051, closest cross-pair 0.0098
//
// The cross-pair column separates at every one of them, which it did not when
// the badge stood bare: the specimen was then one word in one colour on the
// page's own ground, and a block of colour is what the box filter needs to
// tell one specimen from another. The container fill is that block. Both
// columns are roughly twice what the bare badge measured, the match column
// included — a filled specimen moves the metric whichever way it is wrong.
//
// The verdict wants Tolerance (0.017) BETWEEN the two columns: every right
// pair under it and every wrong pair over it. Two viewports do that. 120x40
// is the more balanced — the match column clears by 1.43x and the cross-pair
// column by 1.35x, against 140x40's lopsided 1.67x and 1.15x — and a verdict
// is only as strong as its nearer margin.
var badgeSize = image.Pt(120, 40)

// TestBadgeMirrors scores every badge specimen pair: the Gio badge in a given
// variant against the browser render of each of the five fixtures, all at the
// same viewport over the level-0 ground — the storey the sheet's pages stand
// on, and therefore the one both halves derive the fill and the foreground
// against.
//
// Two criteria, and the second is what makes the first mean anything. Each
// matching pair must land under Tolerance, or the page is not the badge; and
// each MISMATCHED pair must land over it, or the metric cannot see which
// variant it is looking at and a page that passed would have proved only that
// something badge-shaped is there. Every distance is logged either way.
func TestBadgeMirrors(t *testing.T) {
	srv := Serve(t)
	shaper := tokens.DefaultTypography.DeterministicShaper()
	style := badge.Style(tokens.DefaultTypography, tokens.Comfortable)

	webs := make(map[string]*image.RGBA, len(badgeCases))
	for _, c := range badgeCases {
		if !FixtureExists("fixtures/" + c.fixture) {
			t.Fatalf("no embedded fixture %q — a typo here would screenshot a 404 page", c.fixture)
		}
		webs[c.fixture] = CaptureBrowser(t, srv.URL+"/fixtures/"+c.fixture, badgeSize)
	}

	for _, c := range badgeCases {
		t.Run(c.fixture, func(t *testing.T) {
			gio := golden.Capture(t, badgeSize, onColor(tokens.DefaultLight.SurfaceAt(tokens.Level0),
				badge.Render(shaper, c.label, nil, c.variant,
					tokens.DefaultLight, tokens.Spacing, tokens.Radius, style, badge.RenderState{})))
			for _, other := range badgeCases {
				d := Distance(gio, webs[other.fixture])
				t.Logf("distance gio %s vs %s: %.4f (Tolerance %.4f)", c.fixture, other.fixture, d, Tolerance)
				switch {
				case other.fixture == c.fixture && d > Tolerance:
					t.Errorf("pair %s scored %.4f > Tolerance %.4f: the page does not read as the badge", c.fixture, d, Tolerance)
				case other.fixture != c.fixture && d <= Tolerance:
					t.Errorf("the %s badge scored %.4f <= Tolerance %.4f against the %s page: the metric cannot tell the two variants apart",
						c.fixture, d, Tolerance, other.fixture)
				}
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
