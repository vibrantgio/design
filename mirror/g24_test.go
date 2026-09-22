package mirror

// The overlay-page verdicts: the .scrim/.dialog, .popover, .tooltip
// and .toast classes — the vocabulary components/dialog.html composes with —
// captured from the real Gio packages (patterns/modal, patterns/popover,
// components/tooltip, patterns/notifications) and compared against browser
// captures of per-specimen fixtures wearing exactly the published sheet's
// classes. Like
// TestCalibration and the other mirror verdicts, these only deliver a verdict
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
	"github.com/vibrantgio/components/toast"
	"github.com/vibrantgio/components/tooltip"
	"github.com/vibrantgio/patterns/modal"
	"github.com/vibrantgio/patterns/notifications"
	"github.com/vibrantgio/patterns/popover"
	vgcolor "github.com/vibrantgio/theme/color"
	"github.com/vibrantgio/theme/tokens"
	"github.com/vibrantgio/theme/typeset"
)

// overlaySize is the shared overlay capture viewport — the patterns overlay
// goldens' 320x240 viewport (modal_test.go, popover_test.go, tooltip_test.go
// and notifications_test.go all pin the same one).
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

// actionChip is a footer action's stand-in: a flat rectangle filling the box
// the footer lays an action out in, heightDp tall. It states no width of its
// own, matching the fixture's divs, which take the footer's width from the
// sheet.
func actionChip(c color.NRGBA, heightDp float32) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(unit.Dp(heightDp)))
		paint.FillShape(gtx.Ops, c, clip.Rect{Max: size}.Op())
		return layout.Dimensions{Size: size}
	}
}

// bodyLine mirrors popover_test.go's textContent: one non-wrapping
// body-medium line in the platform's label, flattened onto the window
// background the popover is filled with, drawn through theme/typeset so the
// line box is the role's LineHeight — which is exactly what makes the
// browser's line box comparable.
func bodyLine(shaper *text.Shaper, s string) layout.Widget {
	style := tokens.DefaultTypography.BodyMedium
	return func(gtx layout.Context) layout.Dimensions {
		m := op.Record(gtx.Ops)
		paint.ColorOp{Color: vgcolor.Flatten(tokens.PlatformLight.Label, tokens.PlatformLight.WindowBackground)}.Add(gtx.Ops)
		material := m.Stop()
		f := typeset.Font(style, font.Normal)
		lbl := typeset.Label(style, 1)
		gtx.Constraints.Min = image.Point{}
		return typeset.Layout(gtx, shaper, lbl, f, unit.Sp(style.Size), s, material)
	}
}

// onColor wraps a layout.Widget in a fill of the given colour — the plane the
// specimen stands on, which every overlay in this set takes from the platform:
// the window background for the scrimmed and anchored specimens, and the
// content fill for the toast stack, which its goldens composite over the fill
// an app pane is painted with.
func onColor(bg color.NRGBA, w layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, bg, clip.Rect{Max: gtx.Constraints.Max}.Op())
		return w(gtx)
	}
}

// The specimen colours the patterns goldens pin: the grey body slot, the
// blue anchor/trigger/cancel chip and the red discard chip.
var (
	slotGrey     = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	chipBlue     = color.NRGBA{R: 80, G: 160, B: 220, A: 255}
	chipRed      = color.NRGBA{R: 220, G: 100, B: 100, A: 255}
	lightBg      = tokens.PlatformLight.WindowBackground
	lightSurface = tokens.PlatformLight.ControlBackground
)

// TestOverlayMirrors scores each overlay specimen pair: the patterns component
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
			tokens.PlatformLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.TitleMedium, tokens.Comfortable,
		)},
		{"dialog-decision.html", lightBg, modal.Render(
			shaper,
			modal.Props{
				Title: "Discard changes?",
				Body:  grow(slotGrey, 40),
				// Neither stand-in states a width: the footer lays every
				// action out in the platform's measured dialog button
				// width, which is the 74 the fixture's own divs carry.
				Actions:  []layout.Widget{actionChip(chipBlue, 28), actionChip(chipRed, 28)},
				Decision: &modal.Decision{Destructive: true},
				Shaper:   shaper,
			},
			true,
			tokens.PlatformLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.TitleMedium, tokens.Comfortable,
		)},
		{"popover-bottom.html", lightBg, popover.Render(
			popover.Props{
				Anchor:    chip(chipBlue, 60, 28),
				Content:   bodyLine(shaper, "Sort ascending"),
				Placement: popover.Bottom,
			},
			true,
			tokens.PlatformLight, tokens.Spacing, tokens.Radius,
		)},
		{"tooltip-top.html", lightBg, tooltip.Render(
			shaper,
			tooltip.Props{Text: "Save", Trigger: chip(chipBlue, 60, 28), Placement: tooltip.Top, Shaper: shaper},
			true,
			tokens.PlatformLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.LabelSmall,
		)},
		{"toast-stack.html", lightSurface, notifications.Render(
			shaper,
			notifications.Props{Position: notifications.TopRight, Shaper: shaper},
			[]notifications.Notification{
				{ID: 1, Status: toast.Info, Text: "Syncing tokens"},
				{ID: 2, Status: toast.Success, Text: "Workspace saved"},
				{ID: 3, Status: toast.Warning, Text: "Connection is slow"},
			},
			tokens.PlatformLight, tokens.Spacing, tokens.Radius,
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
