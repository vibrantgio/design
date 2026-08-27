package mirror

// The AQ1.3 storey-ring verdict: a focus ring derives against the storey the
// control stands on, not against the page. The sheet carries this as a pair
// of inherited custom properties a raised surface declares beside its own
// fill (--ground-focus-ring, --ground-border), and the Gio side as the Ground
// field every control's render state now carries; this pair scores the two
// against each other in the configuration that was under the floor — a text
// field focused inside a level-2 dialog, where the ground floor's ring rung
// measures 2.92:1 against the fill it is drawn on. Like TestCalibration and
// the earlier mirror verdicts, it only delivers a verdict on the
// authoritative machine; elsewhere one half of the harness skips loudly.
//
// The text field is the specimen because its ring is its own promoted border,
// which both sides place identically — Gio thickens the border inward, the
// sheet draws the second pixel as an inset shadow. The checkbox and the radio
// put their ring in the slack outside the glyph and CSS puts an outline on the
// glyph's edge, a geometry difference that predates this task and that no
// colour verdict should be asked to see through.

import (
	"testing"

	"github.com/vibrantgio/components/golden"
	"github.com/vibrantgio/components/input"
	"github.com/vibrantgio/theme/tokens"
)

func TestStoreyFocusRingMirrors(t *testing.T) {
	srv := Serve(t)
	shaper := tokens.DefaultTypography.DeterministicShaper()

	const fixture = "textfield-dialog-focus.html"
	gio := golden.Capture(t, fieldSize, onColor(tokens.DefaultLight.SurfaceAt(tokens.Level2), input.Render(
		shaper, "Placeholder",
		tokens.DefaultLight, tokens.Spacing, tokens.Radius,
		tokens.DefaultTypography.BodyLarge, tokens.Comfortable,
		input.RenderState{Focused: true, Ground: tokens.Level2},
	)))
	if !FixtureExists("fixtures/" + fixture) {
		t.Fatalf("no embedded fixture %q — a typo here would screenshot a 404 page", fixture)
	}
	web := CaptureBrowser(t, srv.URL+"/fixtures/"+fixture, fieldSize)
	d := Distance(gio, web)
	t.Logf("distance gio vs %s: %.4f (Tolerance %.4f)", fixture, d, Tolerance)
	if d > Tolerance {
		t.Errorf("pair %s scored %.4f > %.4f: the sheet's storey-local ring does not read as the component's own", fixture, d, Tolerance)
	}
}

// The pair carried a recorded lag between AU1.2 and AU1.3 and no longer does.
// AU1.2 moved the elevation ladder onto the light side of the Background pin
// (ADR-022), so a level-2 dialog fills #fbfbfb where it filled #d4d4d4 and the
// sheet's --color-dialog-focus-ring re-derived against the new storey to
// primary 600 (#8c59f4); the Gio half did not follow, because
// components/internal/focus's Ground still resolved a storey through the
// retired tokens.Elevation.SurfaceStep and answered a neutral-ramp rung. It
// scored 0.0249 against a 0.0170 Tolerance for exactly one task.
//
// AU1.3 re-pointed Ground at tokens.ColorTokens.SurfaceAt, both halves now
// answer primary 600, and the pair scores 0.0103 — inside Tolerance with the
// ordinary margin, so the ceiling and its explanation are deleted rather than
// widened.
