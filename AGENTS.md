# AGENTS.md — design

The published bundle of the Vibrant Gio design system: `readme.md`,
`theme.json`, `styles.css` and the three foundation pages under
`foundations/` — the six paths that live at `claude.ai/design`. The Go
module embeds them as `Bundle`, an `embed.FS`, so a consumer reads a page
by name rather than by a relative path that breaks the moment a test runs
from somewhere else. The repository also carries the system's architecture
rationale — `DESIGN.md` and its archived first edition `DESIGN-v1.md`, the
two hand-written files beside the generated bundle — so the rationale lives
beside the artifact it explains.

**Layer.** Outside ADR-001's tier table: an application at the top of the
stack, which the tier rule exempts and which may import any layer of the
design system. Its bundle is generated: emitted by theme's `cmd/vg-tokens`
and pushed to `claude.ai/design` by the org's `scripts/push-design.sh`, so
a change to a token lands in theme first and is re-emitted here rather than
edited here — only the two `DESIGN*.md` documents are edited in place. Its
root module imports nothing else in the organization. That direction is
measured rather than typed — `scripts/check-layers.sh --edges` reports the
graph and `scripts/sync-agents.sh` renders these sentences from it — so
correcting them here changes nothing. The other direction is measured too
and deliberately not written down: the gate checks the graph both ways, but
a public API's consumers are unknowable, so this file says what its module
needs and never who needs it.

**Read the canonical guide before you write code against this module.** It is
the organization's one agent guide — the module inventory with current tags,
the application skeleton, the MVU loop and rx semantics, typography, and the
pitfalls that are not guessable. It lives exactly once, in `vibrantgio/.github`,
and this file links it rather than copying it:

    https://raw.githubusercontent.com/vibrantgio/.github/master/llms.txt

**Module.** `github.com/vibrantgio/design`, one module at the repository
root.

**Build and test.** From the repository root:

    go build ./... && go test ./...

**The mirror harness, and the organization's first non-Go dependency.**
`mirror/` is the G1.1 golden comparison harness: it renders a bundle page in
headless Chromium (driven by chromedp), renders the corresponding Gio
component through `components/golden` with theme's `DeterministicShaper()`,
and scores the pair with a perceptual box-downscale metric
(`mirror.Distance`, threshold `mirror.Tolerance`, both calibrated from
measured pairs — the numbers live on the const). Until this package,
everything in the organization built and tested with a Go toolchain alone;
the harness adds an external browser binary, installed with
`brew install --cask chromium` and pinned by `mirror.ChromiumVersion` — the
harness asserts the binary's `--version` at startup and fails loudly on a
mismatch, so a brew upgrade is an explicit re-baseline (re-run
`TestCalibration`, update the const and `Tolerance` together), never a
silent renderer swap.

**Verdicts are read on one machine.** Rene's Mac (darwin/arm64, the machine
carrying Chromium 153.0.8008.0) is authoritative. CI cannot be: a runner
opens no headless Gio window, so the Gio half answers `t.Skipf` and a
skipped test passes — the F5.7 trap. The harness skips loudly when Chromium
or headless Gio is missing; a green run whose log shows skips is not a
verdict. `go test ./...` still passes on a Chromium-less machine by design.

**Fixtures are not bundle.** `mirror/fixtures/*.html` are embedded
calibration pages (one faithful token-built mirror, three deliberately
wrong variants) served beside the bundle in tests. `scripts/push-design.sh`
uploads exactly six paths; the fixtures are not among them and must never
be.
