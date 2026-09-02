package mirror

// The focus-ring verdict: a focused control wears one ring colour per
// scheme, on every storey. The sheet carries that as a single token both
// its ground-floor and its raised rules name, and the Gio side as a
// derivation that takes the scheme and nothing else; this pair scores the
// two against each other on a raised storey, which is where a
// ground-derived ring would part from a scheme-derived one: a text field
// focused inside a level-2 dialog. Like TestCalibration and the other mirror
// verdicts, it only delivers a verdict on the authoritative machine;
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
	gio := golden.Capture(t, fieldSize, onColor(tokens.DefaultLight.SurfaceAt(tokens.Level2), input.Render(
		shaper, "Placeholder",
		tokens.DefaultLight, tokens.Spacing, tokens.Radius,
		tokens.DefaultTypography.BodyLarge, tokens.Comfortable,
		input.RenderState{Focused: true, Level: tokens.Level2},
	)))
	if !FixtureExists("fixtures/" + fixture) {
		t.Fatalf("no embedded fixture %q — a typo here would screenshot a 404 page", fixture)
	}
	web := CaptureBrowser(t, srv.URL+"/fixtures/"+fixture, fieldSize)
	d := Distance(gio, web)
	t.Logf("distance gio vs %s: %.4f (Tolerance %.4f)", fixture, d, Tolerance)
	if d > Tolerance {
		t.Errorf("pair %s scored %.4f > %.4f: the sheet's ring does not read as the component's own on a raised storey", fixture, d, Tolerance)
	}
}
