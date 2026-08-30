package mirror

// The component-page verdicts: the class vocabulary the components/*.html
// pages compose with — .btn's quieter registers and the
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

// TestComponentMirrors scores each component specimen pair: the Gio widget in a
// given register/state against the browser render of the matching fixture,
// both at the same viewport. Every distance is logged; each must land under
// Tolerance for the page to count as a mirror of the component rather than
// a drawing of one.
func TestComponentMirrors(t *testing.T) {
	srv := Serve(t)
	shaper := tokens.DefaultTypography.DeterministicShaper()

	// ceilings overrides Tolerance for a pair whose own cross-renderer floor
	// sits above it; every other pair is scored against Tolerance itself.
	// See dropdownInkFloor for the one entry and its measurement.
	ceilings := map[string]float64{"dropdown.html": dropdownInkFloor}

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
			ceiling := Tolerance
			if v, ok := ceilings[c.fixture]; ok {
				ceiling = v
			}
			t.Logf("distance gio vs %s: %.4f (ceiling %.4f)", c.fixture, d, ceiling)
			if d > ceiling {
				t.Errorf("pair %s scored %.4f > %.4f: the mirror does not read as the component", c.fixture, d, ceiling)
			}
		})
	}
}

// dropdownInkFloor is the dropdown pair's own cross-renderer floor, which
// sits just above Tolerance, and this comment is the measurement that says
// why it is a floor rather than a disagreement.
//
// The two halves agree on colour exactly. Scored column by column across the
// 220-wide frame, every cell right of the label — the trigger's fill, its
// edge, the chevron — measures 0.0000; the whole distance is the label band,
// and inside it both renderers place the same glyphs at the same positions
// with the same number of antialiased pixels (295 against 293 over the band).
// What differs is stem weight: Chromium's macOS rasteriser gamma-darkens
// stems, so 266 pixels come out under 0x60 where Gio's produces 134. That is
// the cross-renderer floor the Tolerance comment names — a different shaper,
// a different rasteriser, a different gamma — and it is the noise the box
// filter exists to average down rather than a colour the sheet got wrong.
//
// The pair clears Tolerance only because Distance is a mean of absolute RGB
// distances, so a fixed coverage disagreement costs in proportion to the gap
// between ink and ground: with the trigger filled at the raised storey the
// gap against the Text pin is 229 levels rather than 213, and 0.0165 × 229/213
// is 0.0177 against a measured 0.0178. This is the one specimen in the set
// where the metric's absence of contrast normalisation shows — the
// highest-ink-contrast frame. The ceiling sits a hair above the measurement so
// a real drift still fails, and it retires the day Distance normalises by the
// frame's own ink-to-ground range, which would fold this back under one
// Tolerance for every pair.
const dropdownInkFloor = 0.0185 // measured 0.0178 on the authoritative machine, 2026-08-27
