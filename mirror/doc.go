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
// colour, radius or size still moves the number.
//
// # The calibration
//
// Read on the authoritative machine (Chromium 153.0.8008.0, darwin/arm64)
// on 2026-09-11, against the platform's own control scale — a 24 dp push
// button, a 27 dp text field floor, a 20 dp stacked row:
//
//	Gio filled button vs its bundle mirror:             0.0178
//	stability: same page captured twice in Chromium:    0.0000
//	vs the mirror with the wrong colour (systemRed):    0.6765
//	vs the mirror with the wrong radius (a pill):       0.0279
//	vs the mirror with the wrong size (compact 19 dp):  0.1502
//
// [Tolerance] is 0.0223, the geometric midpoint of the two clusters' nearest
// members: 1.25× the measured matching distance and 1.25× below the nearest
// wrong variant, the pill radius, so both clear it by the same margin.
//
// Both margins narrowed at this re-baseline, and the cause is the platform's
// scale rather than anything either side draws. The matching distance is the
// label band and almost nothing else, and a 20 dp line box fills 20 of a
// 24 dp button's rows where it filled 20 of 36; the wrong radius, meanwhile,
// is a pill on a 24 dp button, which is a 12 px corner against the correct 6
// where it used to be 18 against 6. The floor rose and the signal shrank,
// both by the height. Widening the deliberate radius error instead of moving
// the threshold was measured and rejected: an elliptical cap scores 0.1530
// and stops being the NEAREST wrong variant, which is the whole reason that
// fixture pins the threshold rather than the wrong-size one.
//
// The two specimens that are the push button's own fill under the platform's
// control text — the tonal button and the dropdown trigger — carry their own
// measured ceiling above Tolerance; see pushButtonLabelFloor for why that is
// a floor and not a disagreement.
package mirror
