// Package mirror is the golden comparison harness that scores the published
// HTML/CSS design bundle against the Gio components it mirrors.
//
// The bundle is a second implementation of the design system and will drift
// unless something holds it. This package is that something: it renders a
// bundle page in a real browser engine, renders the corresponding Gio
// component through components/golden, and asks a perceptual metric whether
// the two read as the same component.
//
// # The external browser dependency
//
// The browser half of this harness drives a headless Chromium binary over the
// DevTools protocol via chromedp, so the whole comparison stays inside one
// `go test` run. The binary is external, installed with
//
//	brew install --cask chromium
//
// and pinned by version: [ChromiumVersion] records the build the stored
// verdicts were read with, and the harness refuses to run under any other.
// A brew upgrade therefore does not silently move the goalposts — it fails
// the test run, and re-recording the const is an explicit re-baseline.
//
// # Which machine is authoritative
//
// Rene's Mac (darwin/arm64) carrying Chromium 153.0.8008.0 is where verdicts
// are read. CI cannot be authoritative and must not be trusted if it goes
// green: the runner opens no headless Gio window, so every Gio-side capture
// answers t.Skipf and a skipped test passes. This harness skips loudly, in
// both directions: no Chromium at
// [ChromiumPath] skips the browser half, no headless Gio skips the Gio half,
// and either way the log says which machine can actually run the comparison.
//
// # The metric, and why not PixelDiff
//
// golden.PixelDiff counts exact byte mismatches, which is right between two
// Gio renders and useless across two renderers: Chromium and Gio shape text
// and antialias edges differently by construction. [Distance] instead
// box-downscales both images to coarse cells and averages the Euclidean RGB
// distance per cell, so shaping and AA noise averages away while a wrong
// colour or size still moves the number.
//
// # The calibration
//
// The harness read these scores on the authoritative machine (Chromium
// 153.0.8008.0, darwin/arm64), at the platform's own control scale: a 24 dp
// push button set in the platform's 13 pt regular system font, a 27 dp text
// field and a 20 dp stacked row.
//
//	Gio filled button vs its bundle mirror:             0.0053
//	stability: same page captured twice in Chromium:    0.0000
//	vs the mirror with the wrong colour (systemRed):    0.6553
//	vs the mirror with the wrong size (compact 19 dp):  0.1455
//
// [Tolerance] is 0.0223. Every matching pair in the package scores at most
// 0.0150. The nearest wrong variant, the wrong size, scores 0.1455. Every
// pair in the package is scored against that one number.
//
// # What the metric does not see
//
// At the platform scale (13 pt regular text on a 24 dp control) the metric
// does not see a radius error or a weight error. A corner or a stem of that
// size changes too few pixels to move the 4×4 cell means above the matching
// pairs. The harness scored each variant below against the matching Gio
// button:
//
//	the mirror with a pill radius:     0.0149
//	the mirror with a square corner:   0.0090
//	the mirror with a bold title:      0.0126
//
// All three land under [Tolerance], so the calibration scores no radius or
// weight variant. The metric sees a wrong colour and a wrong size.
package mirror
