package mirror

// The component-page verdicts: the class vocabulary the components/*.html
// pages compose with — .btn's less pronounced variants and the
// form controls — captured from the real Gio components (components/button,
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

// onBackground wraps a layout.Widget in a fill of the platform's window
// background, matching the fixtures' body { background:
// var(--platform-window-background) }.
func onBackground(w layout.Widget) func(layout.Context) layout.Dimensions {
	return func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, tokens.PlatformLight.WindowBackground,
			clip.Rect{Max: gtx.Constraints.Max}.Op())
		return w(gtx)
	}
}

// glyphSize is the checkbox/radio capture viewport: the comfortable
// CHECKBOX ROW the 16 dp glyph is centred in — exactly what drawCheckbox and
// drawRadio return, which is the density's own footprint for these two
// controls and not the push button's height. The measured 22 against the
// button's 24 is why it is a frame of its own: a viewport a row too tall
// would score two glyphs one pixel apart.
var glyphSize = image.Pt(checkboxRowHeight, checkboxRowHeight)

// fieldSize is the text-field capture viewport: 220 wide like the button
// captures, and fieldHeight tall — BodyLarge's 24 dp line box plus twice the
// density's 2 dp vertical padding, which beats the 27 dp field-height floor.
var fieldSize = image.Pt(220, fieldHeight)

// triggerSize is the dropdown capture viewport: the same width, at the
// BUTTON's height, because components/picker's field trigger is a push
// button and takes the control height where a text field takes its own.
var triggerSize = image.Pt(220, buttonHeight)

// TestComponentMirrors scores each component specimen pair: the Gio component
// in a given variant/state against the browser render of the matching fixture,
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
			tokens.PlatformLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.LabelLarge, tokens.Comfortable,
			button.RenderState{Emphasis: button.Tonal, Surface: tokens.PlatformLight.WindowBackground},
		)},
		{"button-ghost.html", mirrorSize, button.Render(
			shaper, "Save Changes",
			tokens.PlatformLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.LabelLarge, tokens.Comfortable,
			button.RenderState{Emphasis: button.Ghost, Surface: tokens.PlatformLight.WindowBackground},
		)},
		{"textfield.html", fieldSize, input.Render(
			shaper, "Placeholder",
			tokens.PlatformLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.BodyLarge, tokens.Comfortable,
			input.RenderState{Surface: tokens.PlatformLight.WindowBackground},
		)},
		{"dropdown.html", triggerSize, input.RenderDropdown(
			shaper,
			tokens.PlatformLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.BodyLarge, tokens.Comfortable,
			input.DropdownRenderState{Options: []string{"Comfortable", "Compact"}},
		)},
		{"checkbox.html", glyphSize, input.RenderCheckbox(
			nil,
			tokens.PlatformLight, tokens.Spacing,
			tokens.DefaultTypography.BodyLarge,
			input.CheckboxRenderState{Surface: tokens.PlatformLight.WindowBackground},
		)},
		{"checkbox-checked.html", glyphSize, input.RenderCheckbox(
			nil,
			tokens.PlatformLight, tokens.Spacing,
			tokens.DefaultTypography.BodyLarge,
			input.CheckboxRenderState{Checked: true, Surface: tokens.PlatformLight.WindowBackground},
		)},
		{"radio-selected.html", glyphSize, input.RenderRadio(
			nil,
			tokens.PlatformLight, tokens.Spacing,
			tokens.DefaultTypography.BodyLarge,
			input.RadioRenderState{Selected: true, Surface: tokens.PlatformLight.WindowBackground},
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

// ONE CEILING. Every pair in this file is scored against [Tolerance] and
// nothing else. Two of them — the tonal button and the dropdown trigger, the
// push button's own fill under the platform's control text at the control
// height — carried a measured ceiling above it until 2026-09-18, on the
// reading that the two renderers place the same glyphs one pixel apart and
// weigh their stems differently inside a frame that is almost entirely label.
//
// Half of that gap was the sheet, not the rasterisers: it set every label a
// fraction wider than the component because it spent the type roles' tracking
// where the library's typeset spends none. With that fixed both pairs measure
// under Tolerance on the authoritative machine — button-tonal 0.0174 and
// dropdown 0.0176 against Tolerance's 0.0223 — with the nearest wrong variant
// the calibration scores, the wrong radius, at 0.0255. So the ceiling is one
// number again and a pair that drifts fails wherever it drifts.
