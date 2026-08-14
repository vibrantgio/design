package mirror

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"eliasnaur.com/font/roboto/robotomedium"
	"eliasnaur.com/font/roboto/robotoregular"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"

	"github.com/vibrantgio/design"
)

// ChromiumPath is where the pinned browser lives on the authoritative
// machine: the brew-cask Chromium install location on macOS. On a machine
// without it the harness skips loudly rather than letting chromedp hunt for
// whatever Chrome happens to be on PATH — a mirror comparison that silently
// changes renderer is worthless.
const ChromiumPath = "/Applications/Chromium.app/Contents/MacOS/Chromium"

// ChromiumVersion is the exact `--version` output of the pinned browser,
// asserted at startup by every browser capture. If brew upgrades Chromium the
// harness fails with instructions instead of quietly re-reading its verdicts
// under a different renderer; updating this const (and re-running the
// calibration in TestCalibration) is the explicit re-baseline that upgrade
// now requires.
const ChromiumVersion = "Chromium 153.0.8008.0"

// fixtures are the harness's own calibration pages, embedded here and served
// beside the bundle under /fixtures/. They are deliberately NOT part of
// design.Bundle: scripts/push-design.sh uploads exactly six paths
// (readme.md, theme.json, styles.css, foundations/*.html) and test fixtures
// have no business on claude.ai/design.
//
//go:embed fixtures
var fixtures embed.FS

// Serve starts an httptest server with design.Bundle at the root (so
// /styles.css and /foundations/color.html are the published bytes), the
// harness fixtures under /fixtures/, and the Roboto faces the fixtures'
// @font-face rules load under /fonts/ — the same TTFs theme's
// DeterministicShaper pins on the Gio side, so neither renderer shapes with
// a machine font. The server is shut down when the test ends.
func Serve(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(design.Bundle)))
	mux.Handle("/fixtures/", http.FileServer(http.FS(fixtures)))
	mux.HandleFunc("/fonts/roboto-regular.ttf", serveTTF(robotoregular.TTF))
	mux.HandleFunc("/fonts/roboto-medium.ttf", serveTTF(robotomedium.TTF))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func serveTTF(ttf []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "font/ttf")
		w.Write(ttf)
	}
}

// FixtureExists reports whether the embedded fixture set carries name
// (a path like "fixtures/button.html"). Tests use it to fail fast on a
// typo'd fixture path instead of screenshotting Chromium's 404 page.
func FixtureExists(name string) bool {
	_, err := fs.Stat(fixtures, name)
	return err == nil
}

// chromiumVersion runs the binary at path with --version and returns the
// trimmed output. It is the lookup-failure branch of the harness: a missing
// binary comes back as the error that makes CaptureBrowser skip.
func chromiumVersion(path string) (string, error) {
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		return "", fmt.Errorf("%s --version: %w", path, err)
	}
	return strings.TrimSpace(string(out)), nil
}

var versionOnce struct {
	sync.Once
	version string
	err     error
}

// assertChromium skips the test if the pinned Chromium is absent and fails
// it if the binary at ChromiumPath is not exactly ChromiumVersion. The
// version is read once per test binary.
func assertChromium(t *testing.T) {
	t.Helper()
	versionOnce.Do(func() {
		versionOnce.version, versionOnce.err = chromiumVersion(ChromiumPath)
	})
	if versionOnce.err != nil {
		t.Skipf("mirror: pinned Chromium not runnable at %s (%v) — the browser half of the "+
			"harness is skipped on this machine; verdicts are read on the authoritative "+
			"machine described in the package comment (brew install --cask chromium)",
			ChromiumPath, versionOnce.err)
	}
	if versionOnce.version != ChromiumVersion {
		t.Fatalf("mirror: Chromium at %s reports %q but the harness is calibrated against %q — "+
			"a renderer change invalidates every stored verdict. Re-baseline deliberately: "+
			"re-run the calibration in TestCalibration, then update ChromiumVersion and "+
			"Tolerance together.",
			ChromiumPath, versionOnce.version, ChromiumVersion)
	}
}

// CaptureBrowser renders url in the pinned headless Chromium at exactly
// size pixels — device scale factor forced to 1, so a Retina host does not
// double every length — waits for the document's fonts to finish loading,
// and returns the screenshot. It skips (loudly) when the pinned browser is
// absent and fails when its version has drifted; see assertChromium.
func CaptureBrowser(t *testing.T, url string, size image.Point) *image.RGBA {
	t.Helper()
	assertChromium(t)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(ChromiumPath),
		chromedp.DisableGPU,
		chromedp.Flag("force-device-scale-factor", "1"),
		chromedp.Flag("hide-scrollbars", true),
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, cancelTimeout := context.WithTimeout(ctx, 60*time.Second)
	defer cancelTimeout()

	var shot []byte
	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(int64(size.X), int64(size.Y), chromedp.EmulateScale(1)),
		chromedp.Navigate(url),
		// Fonts load asynchronously; a screenshot taken before
		// document.fonts settles shapes the label in a fallback face and
		// measures the race, not the mirror.
		chromedp.Evaluate(`document.fonts.ready.then(() => true)`, nil,
			func(p *runtime.EvaluateParams) *runtime.EvaluateParams { return p.WithAwaitPromise(true) }),
		chromedp.CaptureScreenshot(&shot),
	)
	if err != nil {
		t.Fatalf("mirror: chromedp: %v", err)
	}

	decoded, err := png.Decode(bytes.NewReader(shot))
	if err != nil {
		t.Fatalf("mirror: decode screenshot: %v", err)
	}
	img := toRGBA(decoded)
	if got := img.Bounds().Size(); got != size {
		t.Fatalf("mirror: screenshot is %v, wanted %v — check viewport emulation and "+
			"device scale factor", got, size)
	}
	return img
}

// toRGBA converts a decoded PNG to *image.RGBA without an alpha conversion
// for the NRGBA case, matching components/golden's loadImage: screenshots
// are fully opaque, so the bytes are the same either way.
func toRGBA(decoded image.Image) *image.RGBA {
	switch v := decoded.(type) {
	case *image.RGBA:
		return v
	case *image.NRGBA:
		return &image.RGBA{Pix: v.Pix, Stride: v.Stride, Rect: v.Rect}
	default:
		b := decoded.Bounds()
		rgba := image.NewRGBA(b)
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				rgba.Set(x, y, decoded.At(x, y))
			}
		}
		return rgba
	}
}
