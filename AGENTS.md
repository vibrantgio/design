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
root module imports nothing else in the organization. Nothing in the
organization imports it. Both directions are measured rather than typed —
`scripts/check-layers.sh --edges` reports the graph and
`scripts/sync-agents.sh` renders these sentences from it — so correcting
them here changes nothing.

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
