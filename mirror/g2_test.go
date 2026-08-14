package mirror

// The G2.1 component-page verdicts: the class vocabulary the new
// components/*.html pages compose with — .btn's quieter registers and the
// form controls — captured from the real Gio widgets (components/button,
// components/input) and compared against browser captures of per-specimen
// fixtures wearing exactly the published sheet's classes. Like
// TestCalibration, these only deliver a verdict on the authoritative
// machine; elsewhere one half of the harness skips loudly.

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"

	"github.com/vibrantgio/components/button"
	"github.com/vibrantgio/components/golden"
	"github.com/vibrantgio/components/input"
	"github.com/vibrantgio/theme/tokens"
)

// onBackground wraps a widget in a fill of the light scheme's Background
// pin, matching the fixtures' body { background: var(--color-bg) }.
func onBackground(w layout.Widget) func(layout.Context) layout.Dimensions {
	return func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, tokens.DefaultLight.Background,
			clip.Rect{Max: gtx.Constraints.Max}.Op())
		return w(gtx)
	}
}

// glyphSize is the checkbox/radio capture viewport: the comfortable 36 dp
// control row the 20 dp glyph is centred in — exactly what drawCheckbox and
// drawRadio return.
var glyphSize = image.Pt(36, 36)

// fieldSize is the text-field/dropdown capture viewport: 220 wide like the
// button captures, 40 tall because BodyLarge's 24 dp line box plus twice
// the 8 dp vertical padding beats the 36 dp control-height floor.
var fieldSize = image.Pt(220, 40)

// TestComponentMirrors scores each G2.1 specimen pair: the Gio widget in a
// given register/state against the browser render of the matching fixture,
// both at the same viewport. Every distance is logged; each must land under
// Tolerance for the page to count as a mirror of the component rather than
// a drawing of one.
func TestComponentMirrors(t *testing.T) {
	srv := Serve(t)
	shaper := tokens.DefaultTypography.DeterministicShaper()

	cases := []struct {
		fixture string
		size    image.Point
		gio     layout.Widget
	}{
		{"button-tonal.html", mirrorSize, button.Render(
			shaper, "Save Changes",
			tokens.DefaultLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.LabelLarge, tokens.Comfortable,
			button.RenderState{Emphasis: button.Tonal},
		)},
		{"button-ghost.html", mirrorSize, button.Render(
			shaper, "Save Changes",
			tokens.DefaultLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.LabelLarge, tokens.Comfortable,
			button.RenderState{Emphasis: button.Ghost},
		)},
		{"textfield.html", fieldSize, input.Render(
			shaper, "Placeholder",
			tokens.DefaultLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.BodyLarge, tokens.Comfortable,
			input.RenderState{},
		)},
		{"dropdown.html", fieldSize, input.RenderDropdown(
			shaper,
			tokens.DefaultLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.BodyLarge, tokens.Comfortable,
			input.DropdownRenderState{Options: []string{"Comfortable", "Compact"}},
		)},
		{"checkbox.html", glyphSize, input.RenderCheckbox(
			tokens.DefaultLight, tokens.Spacing, tokens.Radius,
			input.CheckboxRenderState{},
		)},
		{"checkbox-checked.html", glyphSize, input.RenderCheckbox(
			tokens.DefaultLight, tokens.Spacing, tokens.Radius,
			input.CheckboxRenderState{Checked: true},
		)},
		{"radio-selected.html", glyphSize, input.RenderRadio(
			tokens.DefaultLight, tokens.Spacing, tokens.Radius,
			input.RadioRenderState{Selected: true},
		)},
	}

	for _, c := range cases {
		t.Run(c.fixture, func(t *testing.T) {
			gio := golden.Capture(t, c.size, onBackground(c.gio))
			if !FixtureExists("fixtures/" + c.fixture) {
				t.Fatalf("no embedded fixture %q — a typo here would screenshot a 404 page", c.fixture)
			}
			web := CaptureBrowser(t, srv.URL+"/fixtures/"+c.fixture, c.size)
			d := Distance(gio, web)
			t.Logf("distance gio vs %s: %.4f (Tolerance %.4f)", c.fixture, d, Tolerance)
			if d > Tolerance {
				t.Errorf("pair %s scored %.4f > Tolerance %.4f: the mirror does not read as the component", c.fixture, d, Tolerance)
			}
		})
	}
}
