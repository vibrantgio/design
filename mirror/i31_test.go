package mirror

// The hosted-ghost verdict: a ghost carries no fill of its own, so every
// coverage it draws lands on the surface it was put on rather than on
// anything of its own. The Gio side is told which surface that is
// (RenderState.Surface) and flattens against it; the sheet is told nothing
// and lets the browser composite against the pixels actually beneath. This
// pair scores the two against each other in the configuration where they
// could part — the modal-close icon ghost, HELD, standing on the dialog's
// fill, where the platform's press overlay is the whole of what is drawn.
// The hover state it used to score is gone with the walk: a push button on
// this platform does not tint under the pointer. Like TestCalibration and the
// earlier mirror verdicts, it only delivers a verdict on the authoritative
// machine; elsewhere one half of the harness skips loudly.

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"github.com/vibrantgio/components/button"
	"github.com/vibrantgio/components/golden"
	"github.com/vibrantgio/theme/tokens"
)

// ghostIconSize is the hosted-ghost capture viewport: the comfortable
// icon-button square with the host's fill around it on every side — the
// square plus one half-square of surround, so the press overlay is read
// against the fill it is laid on.
var ghostIconSize = image.Pt(2*buttonHeight, 2*buttonHeight)

// modalCross mirrors patterns/modal's crossIcon geometry — and the fixture's
// SVG: two diagonal strokes 2 dp wide, inset 6 dp on every side of the glyph
// box. Vector clip strokes keep the capture deterministic.
func modalCross(gtx layout.Context, sizePx int, col color.NRGBA) {
	w, h := float32(sizePx), float32(sizePx)
	pad := float32(gtx.Dp(unit.Dp(6)))
	stroke := float32(gtx.Dp(unit.Dp(2)))
	if stroke < 1 {
		stroke = 1
	}
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(f32.Pt(pad, pad))
	p.LineTo(f32.Pt(w-pad, h-pad))
	paint.FillShape(gtx.Ops, col, clip.Stroke{Path: p.End(), Width: stroke}.Op())

	p.Begin(gtx.Ops)
	p.MoveTo(f32.Pt(w-pad, pad))
	p.LineTo(f32.Pt(pad, h-pad))
	paint.FillShape(gtx.Ops, col, clip.Stroke{Path: p.End(), Width: stroke}.Op())
}

// TestHostedGhostMirrors scores the hosted-ghost pair: the Gio icon ghost held
// on the dialog's own fill against the browser render of the dialog-hosted
// fixture, both at the same viewport. The distance is logged and must land
// under Tolerance for the sheet's compositing to count as a mirror of the
// component's flattening against the surface it was told it stands on.
func TestHostedGhostMirrors(t *testing.T) {
	srv := Serve(t)

	const fixture = "icon-ghost-dialog-press.html"
	gio := golden.Capture(t, ghostIconSize, onColor(tokens.PlatformLight.WindowBackground, button.RenderIcon(
		modalCross, tokens.PlatformLight, tokens.Spacing, tokens.Radius, tokens.Comfortable,
		button.RenderState{Emphasis: button.Ghost, Surface: tokens.PlatformLight.WindowBackground, Pressed: true},
	)))
	if !FixtureExists("fixtures/" + fixture) {
		t.Fatalf("no embedded fixture %q — a typo here would screenshot a 404 page", fixture)
	}
	web := CaptureBrowser(t, srv.URL+"/fixtures/"+fixture, ghostIconSize)
	d := Distance(gio, web)
	t.Logf("distance gio vs %s: %.4f (Tolerance %.4f)", fixture, d, Tolerance)
	if d > Tolerance {
		t.Errorf("pair %s scored %.4f > Tolerance %.4f: the mirror does not read as the hosted ghost held", fixture, d, Tolerance)
	}
}
