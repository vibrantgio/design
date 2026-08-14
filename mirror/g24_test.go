package mirror

// The G2.4 overlay-page verdicts: the .scrim/.dialog, .popover, .tooltip
// and .toast classes — the vocabulary components/dialog.html composes with —
// captured from the real patterns widgets (patterns/modal, patterns/popover,
// patterns/tooltip, patterns/toast) and compared against browser captures of
// per-specimen fixtures wearing exactly the published sheet's classes. Like
// TestCalibration and the G2.1–G2.3 verdicts, these only deliver a verdict
// on the authoritative machine; elsewhere one half of the harness skips
// loudly.

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"

	"github.com/vibrantgio/components/golden"
	"github.com/vibrantgio/patterns/modal"
	"github.com/vibrantgio/patterns/popover"
	"github.com/vibrantgio/patterns/toast"
	"github.com/vibrantgio/patterns/tooltip"
	"github.com/vibrantgio/theme/tokens"
	"github.com/vibrantgio/theme/typeset"
)

// overlaySize is the shared overlay capture viewport — the patterns overlay
// goldens' 320x240 canvas (modal_test.go, popover_test.go, tooltip_test.go
// and toast_test.go all pin the same one).
var overlaySize = image.Pt(320, 240)

// grow mirrors modal_test.go's fillRect: a sharp-edged solid stand-in
// filling its width at a fixed height, used as the dialog body so neither
// renderer has a text-wrap decision to disagree on inside the surface.
func grow(c color.NRGBA, heightDp float32) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(unit.Dp(heightDp)))
		paint.FillShape(gtx.Ops, c, clip.Rect{Max: size}.Op())
		return layout.Dimensions{Size: size}
	}
}

// chip mirrors the overlay tests' fixedRect: a sharp-edged solid stand-in
// with explicit dims — the popover anchor, the tooltip trigger and the
// decision dialog's footer actions.
func chip(c color.NRGBA, widthDp, heightDp float32) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		size := image.Pt(gtx.Dp(unit.Dp(widthDp)), gtx.Dp(unit.Dp(heightDp)))
		paint.FillShape(gtx.Ops, c, clip.Rect{Max: size}.Op())
		return layout.Dimensions{Size: size}
	}
}

// bodyLine mirrors popover_test.go's textContent: one non-wrapping
// body-medium line at the Text pin, drawn through theme/typeset so the line
// box is the role's LineHeight — which is exactly what makes the browser's
// line box comparable.
func bodyLine(shaper *text.Shaper, s string) layout.Widget {
	style := tokens.DefaultTypography.BodyMedium
	return func(gtx layout.Context) layout.Dimensions {
		m := op.Record(gtx.Ops)
		paint.ColorOp{Color: tokens.DefaultLight.Text}.Add(gtx.Ops)
		material := m.Stop()
		f := typeset.Font(style, font.Normal)
		lbl := typeset.Label(style, 1)
		gtx.Constraints.Min = image.Point{}
		return typeset.Layout(gtx, shaper, lbl, f, unit.Sp(style.Size), s, material)
	}
}

// onColor wraps a widget in a fill of the given ground — the bg pin for the
// scrimmed and anchored specimens, but the Surface pin for the toast stack,
// which its goldens composite over Surface so the tinted fill is read
// against the ground app panes are painted with.
func onColor(bg color.NRGBA, w layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, bg, clip.Rect{Max: gtx.Constraints.Max}.Op())
		return w(gtx)
	}
}

// The specimen colours the patterns goldens pin: the grey body slot, the
// blue anchor/trigger/cancel chip and the red discard chip.
var (
	slotGrey    = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	chipBlue    = color.NRGBA{R: 80, G: 160, B: 220, A: 255}
	chipRed     = color.NRGBA{R: 220, G: 100, B: 100, A: 255}
	lightBg     = tokens.DefaultLight.Background
	lightGround = tokens.DefaultLight.Surface
)

// TestOverlayMirrors scores each G2.4 specimen pair: the patterns widget
// against the browser render of the matching fixture, both at the same
// viewport. Every distance is logged; each must land under Tolerance for
// the page to count as a mirror of the pattern rather than a drawing of
// one.
func TestOverlayMirrors(t *testing.T) {
	srv := Serve(t)
	shaper := tokens.DefaultTypography.DeterministicShaper()

	cases := []struct {
		fixture string
		bg      color.NRGBA
		gio     layout.Widget
	}{
		{"dialog-panel.html", lightBg, modal.Render(
			shaper,
			modal.Props{Title: "Preferences", Body: grow(slotGrey, 40), Shaper: shaper},
			true,
			tokens.DefaultLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.TitleMedium, tokens.Comfortable,
		)},
		{"dialog-decision.html", lightBg, modal.Render(
			shaper,
			modal.Props{
				Title:    "Discard changes?",
				Body:     grow(slotGrey, 40),
				Actions:  []layout.Widget{chip(chipBlue, 60, 28), chip(chipRed, 60, 28)},
				Decision: &modal.Decision{Destructive: true},
				Shaper:   shaper,
			},
			true,
			tokens.DefaultLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.TitleMedium, tokens.Comfortable,
		)},
		{"popover-bottom.html", lightBg, popover.Render(
			popover.Props{
				Anchor:    chip(chipBlue, 60, 28),
				Content:   bodyLine(shaper, "Sort ascending"),
				Placement: popover.Bottom,
			},
			true,
			tokens.DefaultLight, tokens.Spacing, tokens.Radius,
		)},
		{"tooltip-top.html", lightBg, tooltip.Render(
			shaper,
			tooltip.Props{Text: "Save", Trigger: chip(chipBlue, 60, 28), Placement: tooltip.Top, Shaper: shaper},
			true,
			tokens.DefaultLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.LabelSmall,
		)},
		{"toast-stack.html", lightGround, toast.Render(
			shaper,
			toast.Props{Position: toast.TopRight, Shaper: shaper},
			[]toast.Toast{
				{ID: 1, Level: toast.Info, Text: "Syncing tokens"},
				{ID: 2, Level: toast.Success, Text: "Workspace saved"},
				{ID: 3, Level: toast.Warning, Text: "Connection is slow"},
			},
			tokens.DefaultLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.LabelMedium,
		)},
	}

	for _, c := range cases {
		t.Run(c.fixture, func(t *testing.T) {
			gio := golden.Capture(t, overlaySize, onColor(c.bg, c.gio))
			if !FixtureExists("fixtures/" + c.fixture) {
				t.Fatalf("no embedded fixture %q — a typo here would screenshot a 404 page", c.fixture)
			}
			web := CaptureBrowser(t, srv.URL+"/fixtures/"+c.fixture, overlaySize)
			d := Distance(gio, web)
			t.Logf("distance gio vs %s: %.4f (Tolerance %.4f)", c.fixture, d, Tolerance)
			if d > Tolerance {
				t.Errorf("pair %s scored %.4f > Tolerance %.4f: the mirror does not read as the pattern", c.fixture, d, Tolerance)
			}
		})
	}
}
