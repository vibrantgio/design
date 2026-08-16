package mirror

// The I3.1 contextual-ghost verdict: a ghost hosted on a raised surface
// washes one rung above the host's own ground, not the window ground's rung
// that resolves to the very fill it sits on. The sheet carries this as the
// contextual walk (.dialog .btn.ghost:hover — neutral 400 over the level-2
// fill), and the Gio side as RenderState.Ground; this pair scores the two
// against each other in the defect's exact configuration — the modal-close
// icon ghost, hovered, on a level-2 ground. Like TestCalibration and the
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

// ghostIconSize is the raised-ghost capture viewport: the comfortable 36 dp
// icon-button square with ground around it on every side, so the hover wash
// is read against the level-2 fill it must differ from.
var ghostIconSize = image.Pt(60, 60)

// modalCross mirrors patterns/modal's crossIcon geometry — and the fixture's
// SVG: two diagonal strokes 2 dp wide, inked from 6 dp to 14 dp of the
// 20 dp glyph box. Vector clip strokes keep the capture deterministic.
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

// TestRaisedGhostMirrors scores the I3.1 pair: the Gio icon ghost hovering
// on its level-2 ground against the browser render of the dialog-hosted
// fixture, both at the same viewport. The distance is logged and must land
// under Tolerance for the sheet's contextual walk to count as a mirror of
// the component's local-ground resolution.
func TestRaisedGhostMirrors(t *testing.T) {
	srv := Serve(t)

	const fixture = "icon-ghost-dialog-hover.html"
	gio := golden.Capture(t, ghostIconSize, onColor(tokens.DefaultLight.SurfaceAt(tokens.Level2), button.RenderIcon(
		modalCross, tokens.DefaultLight, tokens.Spacing, tokens.Radius, tokens.Comfortable,
		button.RenderState{Emphasis: button.Ghost, Ground: tokens.Level2, Hovered: true},
	)))
	if !FixtureExists("fixtures/" + fixture) {
		t.Fatalf("no embedded fixture %q — a typo here would screenshot a 404 page", fixture)
	}
	web := CaptureBrowser(t, srv.URL+"/fixtures/"+fixture, ghostIconSize)
	d := Distance(gio, web)
	t.Logf("distance gio vs %s: %.4f (Tolerance %.4f)", fixture, d, Tolerance)
	if d > Tolerance {
		t.Errorf("pair %s scored %.4f > Tolerance %.4f: the mirror does not read as the raised-ghost hover", fixture, d, Tolerance)
	}
}
