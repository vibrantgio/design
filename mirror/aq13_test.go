package mirror

// The focus-ring verdict: a focused control wears the platform's keyboard
// focus indicator, which is a colour at a coverage over whatever the ring
// lies on. The sheet names that coverage and lets the browser composite it;
// the Gio side flattens it against the surface the control was told it
// stands on. This pair scores the two against each other where they could
// part — a text field focused inside a dialog, so the ring lands on the
// dialog's fill rather than on the page. Like TestCalibration and the other
// mirror verdicts, it only delivers a verdict on the authoritative machine;
// elsewhere one half of the harness skips loudly.
//
// The text field is the specimen because its ring is its own promoted border,
// which both sides place identically — Gio thickens the border inward, the
// sheet draws the second pixel as an inset shadow. The checkbox and the radio
// put their ring in the slack outside the glyph while CSS puts an outline on
// the glyph's edge, a geometry difference no colour verdict should be asked
// to see through.

import (
	"testing"

	"github.com/vibrantgio/components/golden"
	"github.com/vibrantgio/components/input"
	"github.com/vibrantgio/theme/tokens"
)

func TestFocusRingMirrors(t *testing.T) {
	srv := Serve(t)
	shaper := tokens.DefaultTypography.DeterministicShaper()

	const fixture = "textfield-dialog-focus.html"
	gio := golden.Capture(t, fieldSize, onColor(tokens.PlatformLight.WindowBackground, input.Render(
		shaper, "Placeholder",
		tokens.PlatformLight, tokens.Spacing, tokens.Radius,
		tokens.DefaultTypography.BodyLarge, tokens.Comfortable,
		input.RenderState{Focused: true, Surface: tokens.PlatformLight.WindowBackground},
	)))
	if !FixtureExists("fixtures/" + fixture) {
		t.Fatalf("no embedded fixture %q — a typo here would screenshot a 404 page", fixture)
	}
	web := CaptureBrowser(t, srv.URL+"/fixtures/"+fixture, fieldSize)
	d := Distance(gio, web)
	t.Logf("distance gio vs %s: %.4f (Tolerance %.4f)", fixture, d, Tolerance)
	if d > Tolerance {
		t.Errorf("pair %s scored %.4f > %.4f: the sheet's ring does not read as the component's own on a hosted surface", fixture, d, Tolerance)
	}
}
