package mirror

// The calibration and proof for the harness. These tests only deliver a
// verdict on the authoritative machine (see the package comment): elsewhere
// either the pinned Chromium is missing (CaptureBrowser skips loudly) or
// headless Gio is unavailable (golden.Capture skips loudly). A green run on a
// machine that skipped is NOT a verdict — read the log.

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"

	"github.com/vibrantgio/components/button"
	"github.com/vibrantgio/components/golden"
	"github.com/vibrantgio/theme/tokens"
)

// mirrorSize is the shared capture size: the Gio button fills its window's
// width and renders Comfortable's 36 dp control height, so a 220×36 window
// is exactly the button — and the fixtures draw a 220×36 .btn at the same
// viewport.
var mirrorSize = image.Pt(220, 36)

// gioFilledButton captures the reference render: components/button's filled
// button, normal state, DefaultLight, Comfortable density, drawn with the
// DeterministicShaper (a golden must pin its faces — a system-shaped
// render would measure the machine, not the mirror) over the scheme's
// background pin, exactly as components' own emphasis goldens are recorded.
func gioFilledButton(t *testing.T) *image.RGBA {
	t.Helper()
	shaper := tokens.DefaultTypography.DeterministicShaper()
	w := button.Render(
		shaper, "Save Changes",
		tokens.DefaultLight, tokens.Spacing, tokens.Radius,
		tokens.DefaultTypography.LabelLarge, tokens.Comfortable,
		button.RenderState{},
	)
	return golden.Capture(t, mirrorSize, func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, tokens.DefaultLight.Background,
			clip.Rect{Max: gtx.Constraints.Max}.Op())
		return w(gtx)
	})
}

func captureFixture(t *testing.T, srv string, name string) *image.RGBA {
	t.Helper()
	if !FixtureExists("fixtures/" + name) {
		t.Fatalf("no embedded fixture %q — a typo here would screenshot a 404 page", name)
	}
	return CaptureBrowser(t, srv+"/fixtures/"+name, mirrorSize)
}

// TestCalibration is the harness's proof, and where Tolerance's numbers come
// from: the matching Gio/HTML pair must land under Tolerance, every
// deliberately wrong variant above it, and a re-render of the same page at
// (near) zero. Each measured distance is logged so a re-baseline (Chromium
// upgrade, token change) can re-read the clusters from `go test -v`.
func TestCalibration(t *testing.T) {
	gio := gioFilledButton(t)
	srv := Serve(t)

	t.Run("match", func(t *testing.T) {
		web := captureFixture(t, srv.URL, "button.html")
		d := Distance(gio, web)
		t.Logf("distance gio vs bundle-token mirror: %.4f (Tolerance %.4f)", d, Tolerance)
		if d > Tolerance {
			t.Errorf("matching pair scored %.4f > Tolerance %.4f: the mirror no longer reads as the component", d, Tolerance)
		}
	})

	for _, wrong := range []string{
		"button-wrong-color.html",
		"button-wrong-radius.html",
		"button-wrong-size.html",
	} {
		t.Run(wrong, func(t *testing.T) {
			web := captureFixture(t, srv.URL, wrong)
			d := Distance(gio, web)
			t.Logf("distance gio vs %s: %.4f (Tolerance %.4f)", wrong, d, Tolerance)
			if d <= Tolerance {
				t.Errorf("wrong variant %s scored %.4f <= Tolerance %.4f: the metric cannot see this mistake", wrong, d, Tolerance)
			}
		})
	}

	t.Run("stability", func(t *testing.T) {
		a := captureFixture(t, srv.URL, "button.html")
		b := captureFixture(t, srv.URL, "button.html")
		d := Distance(a, b)
		t.Logf("distance same page captured twice: %.4f", d)
		if d > Tolerance/4 {
			t.Errorf("two captures of the same page scored %.4f — the browser side is not deterministic enough to carry a verdict", d)
		}
	})
}

// TestChromiumLookupFailure exercises the branch a Chromium-less machine
// takes: the version lookup must come back as an error (which assertChromium
// turns into a loud skip), never as an empty version that would then fail
// the pin assertion with a misleading message.
func TestChromiumLookupFailure(t *testing.T) {
	if _, err := chromiumVersion("/nonexistent/chromium/binary"); err == nil {
		t.Fatal("chromiumVersion on a nonexistent path returned no error; a machine without Chromium would not skip")
	}
}

// TestDistance pins the metric's edges: identical images are 0, black vs
// white is 1, and unequal sizes panic (a defect in the calling test, per
// golden.PixelDiff's argument).
func TestDistance(t *testing.T) {
	a := image.NewRGBA(image.Rect(0, 0, 16, 16))
	b := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for i := range a.Pix {
		a.Pix[i] = 0xFF
	}
	for i := 3; i < len(b.Pix); i += 4 {
		b.Pix[i] = 0xFF // opaque black
	}
	if d := Distance(a, a); d != 0 {
		t.Errorf("Distance(a, a) = %v, want 0", d)
	}
	if d := Distance(a, b); d < 0.999 || d > 1.001 {
		t.Errorf("Distance(white, black) = %v, want 1", d)
	}
	defer func() {
		if recover() == nil {
			t.Error("Distance on unequal sizes did not panic")
		}
	}()
	Distance(a, image.NewRGBA(image.Rect(0, 0, 8, 8)))
}
