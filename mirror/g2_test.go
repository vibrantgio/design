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

// glyphSize is the checkbox/radio capture viewport: the comfortable control
// row the 16 dp glyph is centred in — exactly what drawCheckbox and drawRadio
// return, which is max(ControlHeight, the glyph) square.
var glyphSize = image.Pt(buttonHeight, buttonHeight)

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

	// ceilings overrides Tolerance for a pair whose own cross-renderer floor
	// sits above it; every other pair is scored against Tolerance itself.
	// See pushButtonLabelFloor for the two entries and their measurement.
	ceilings := map[string]float64{
		"button-tonal.html": pushButtonLabelFloor,
		"dropdown.html":     pushButtonLabelFloor,
	}

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

// pushButtonLabelFloor is the cross-renderer floor shared by the two
// specimens that are the push button's own fill under the platform's control
// text at the control height — the tonal button and the dropdown trigger.
// This comment is the measurement that says why it is a floor rather than a
// disagreement.
//
// The two halves agree on colour exactly. Scored column by column across the
// 220-wide dropdown frame, every cell right of the label — the trigger's
// fill, its edge, the chevron — measures 0.0000; the whole distance is the
// label band, and inside it the two renderers place the same glyphs one pixel
// apart and weigh their stems differently, Chromium's macOS rasteriser
// gamma-darkening stems where Gio's does not. That is the cross-renderer
// floor the package comment names.
//
// It clears Tolerance only because Distance is a mean of absolute RGB
// distances with no contrast normalisation, so a fixed coverage disagreement
// costs in proportion to the gap between foreground and fill — and these are
// the highest-contrast frames in the set: labelColor flattened in sRGB over
// the push button's own #ececec fill is a 200-level gap on every channel,
// inside a frame that is the control height and therefore almost entirely
// label. The share is what moved, not the drawing: at the platform's 24 dp
// control the label is 20 of 24 rows where it was 20 of 40.
//
// The ceiling sits above both measurements and below the nearest wrong
// variant the calibration scores, so a real drift still fails. It retires the
// day Distance normalises by the frame's own foreground-to-fill range, which
// would fold both back under one Tolerance.
//
// Measured on the authoritative machine, 2026-09-11: button-tonal 0.0218,
// dropdown 0.0251, against the wrong-radius variant's 0.0279.
const pushButtonLabelFloor = 0.0265
