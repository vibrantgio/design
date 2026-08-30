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
// distance per cell, so shaping and AA noise washes out while a wrong colour
// role, radius or size still moves the number. [Tolerance] carries the
// calibration; its comment carries the measured evidence.
package mirror
