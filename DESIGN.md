# Vibrant Gio Design System — Architecture & Rationale

> **Status:** Second edition (2026-08). This document describes the system as
> shipped across the organization's repositories after the Phase B–F rework:
> the inverted layering, the generative colour model, and the deliberate
> desktop divergences from Material Design 3. The original design document is
> preserved as [DESIGN-v1.md](DESIGN-v1.md); the road from there to here is
> summarised near the end. Since the rework the release has been cut and the
> repositories renamed: the `spectrum`, `prism`, `pulse` and `cadence` this
> document's historical passages speak of are today's `theme`, `components`,
> `effects` and `patterns`. Current published versions are recorded where
> they always were — ADR-006's tag rules govern them and `llms.txt` lists
> them — never here.

**Two documents, two jobs.** This file is the design rationale — *why* the
system is shaped the way it is, for people (and agents) working **on** the
design system. The canonical guide for building applications **with** it is
[`llms.txt`](https://raw.githubusercontent.com/vibrantgio/workbench/master/llms.txt)
in the `workbench` repository (ADR-004): bootstrap skeleton,
token vocabulary, component catalogue, pitfalls. If you are writing an app,
read that first; if you are changing the system, read this. This file lives
in the `design` repository, beside the published token bundle it explains;
it spent its first life in `workbench`, and its pre-move history is in that
repository's log.

---

## What Vibrant Gio is

Vibrant Gio is a platform for building beautiful, native desktop applications
on macOS, Windows, and Linux, built on [gioui.org](https://gioui.org). The
goal is a first-class design system — analogous to Material Design for Google
— but unique to Vibrant Gio. The "vibrancy" in the name is both literal (rich
colour, depth, motion) and philosophical (alive, reactive, responsive).

It is **one coherent design system across twenty-two repositories**: nineteen
library repos, the `workbench` repo carrying the seven reference
applications, this `design` repo carrying the published token bundle and this
rationale, and the org's `.github` repo carrying the plan and the agent
guide. Coherence is the point and the discipline —
one theme observable carries the entire look (colour, typography, density,
elevation, motion), every module sits in an enforced tier, and the reference
apps all embody the same contract.

The application model is **Functional Reactive Programming** using
`reactivego/rx` and `vibrantgio/mvu`. The design system feels native to that
model — components take the theme observable and need nothing else. Two
pattern families coexist: pure FRP composition and an explicit MVU
(Model–View–Update) state machine; every component serves both through one
API (§Key architectural patterns).

### Non-goals

- **No web target.** This is a desktop-native system; we do not aim for
  browser deployment. (Tokens are *exported* to CSS for prototyping in
  Claude Design, but that surface renders specimens, not applications.)
- **No mobile target initially.** Gio supports Android/iOS but Vibrant Gio is
  optimised for desktop input modalities, window chrome, and platform
  integration.
- **No embedded / kiosk target.** No memory- or binary-size budgets shaped
  around constrained devices.
- **No CSS-like dynamic styling.** Tokens are typed Go values, not
  stringly-typed maps.
- **No hand-picked palettes.** Colour is derived from a seed by the
  generative engine (ADR-002/007); a lint fails the build on colour literals
  in the component repos.
- **Not MD3's look.** MD3's *system* is adopted where it earns its place;
  its touch-first look is deliberately rejected (ADR-005).

---

## The layering

The design-system spine, foundation first:

```
mvu  →  theme  →  components  →  effects  →  patterns  →  markdown
```

`theme` — the theme runtime and every design token — is the **foundation**
of the design system, below the components it themes. This is the inversion
ADR-001 records: the system was first built with the tokens inside `prism`
and `spectrum` layered above it, which meant the theme depended on what it
themed, `LiveTheme` hardcoded the default palettes, and there was no palette
injection point anywhere in the stack. Moving the token and theme contract
down into `spectrum` (and the animation code out of it, into
`pulse/transition`) gave the stack one direction: tokens flow up, nothing
flows down.

The full tier table judges all nineteen library modules, not just the spine.
A module may import only modules in a strictly lower tier, plus anything in
the support row:

| Tier | Modules | May import |
| --- | --- | --- |
| 0 | mvu, font, style, textdraw, backdrop, gradient, circle | support libraries only |
| 1 | theme | tier 0 |
| 2 | components | tiers 0–1 |
| 3 | effects | tiers 0–2 |
| 4 | patterns, markdown | tiers 0–3 |
| — | ivg, svg, seen, csg, kiwi, noise, traer | nothing in this table |

Notes the table needs to be read with:

- **`font` is tier 0** because theme depends on it for the default Roboto
  and Roboto Mono faces (ADR-003). Without that row the layer check would
  reject the exact edge the typography decision requires.
- **One intra-tier edge is admitted explicitly:** `style` imports `font` and
  `textdraw`, both tier 0. `style` is frozen rather than deleted (ADR-003),
  so the check permits that single edge instead of renumbering the tier.
- **The support libraries** are consumed by the design system and never
  depend on it. That is the whole of their contract.
- **Nested demo modules are exempt from being imported, not from importing.**
  `components/gallery` and `mvu/example` are separate modules that may depend on
  layers above their parent; their parents may not inherit that edge. This is
  the mechanism that keeps a demo from re-closing a cycle — the org's one
  real dependency cycle (pulse into prism) was a single demo file's doing.
- **`workbench` is eight application modules in subdirectories**, each
  released on its own cadence, beside a root module that collects them. No
  library imports any of them, and a nested module stands outside its parent
  by Go's own rules, so the root never swallows an application.

The table is enforced, not aspirational: the org's `check-layers.sh` script
walks `go list -deps` for every module and asserts only the edges the table
permits. A layering rule that is not checked is not a rule.

**Compatibility shims.** The inversion was landed without breaking any
consumer: `prism/tokens`, `prism/theme` and `prism/a11y` stayed behind as
alias packages forwarding to their spectrum homes, and `spectrum/transition`
forwarded to `pulse/transition`, for one release cycle. The breaking release
that followed the rework deleted them all — after every in-org consumer had
moved off them, which the reference apps had already done as of Phase F1.
The only import paths today are the real ones: `theme/…` and
`effects/transition`.

---

## The generative model

Colour, and increasingly everything else, is **derived, not picked**. One
seed colour generates the entire palette — both modes — and the theme
observable carries the result to every component.

### Seed → ramps → pins → semantic layer

The engine (`theme/color`, ADR-002) derives tonal palettes on two axes:
**tone is CIELAB L\***, exactly as MD3 defines it, and **hue and chroma come
from OKLCh** — HCT's architecture with OKLab substituted for CAM16, plus
chroma-reduction gamut mapping. From a seed:

```go
light, dark := tokens.FromSeed(seed) // paired ColorTokens; the light
                                     // Primary pin is the seed exactly
```

What comes out (ADR-007's model, in Claude Design's vocabulary):

- **Ramps.** Five roles — Neutral, Primary, Secondary, Tertiary, Error —
  each a **nine-step ramp, 100–900**, where *the step is the meaning*:
  100–300 tinted fills, hovers and subtle borders; 500 the mid reference and
  strong border; 700–900 text over tinted fills and pressed states.
- **Paired dark ramps, not a second table.** The generator emits light and
  dark scales in which the same step keeps the same job — a component asks
  for neutral‑200 and gets a light card on a light ground and a dark card on
  a dark one, with no second assignment table to drift.
- **Pins.** A brand colour rarely sits on the shared lightness scale (the
  default seed `#6750A4` is L\* 40 — step‑700 depth), so each accent role's
  solid fill is **pinned separately** from its ramp and reproduced exactly,
  with an `On*` colour guaranteed readable over it. Background and Text are
  pins too.
- **A thin semantic layer** — Background, Surface (neutral 200), Divider
  (neutral 300), Text — sits over the ramps so call sites read intent; reach
  into the ramps when you need a specific step.

### States are step walks

Interaction states are resolved as walks up the ramp, not alpha overlays:
hover is one step past the component's ground, pressed/selected two, clamped
at 900; pinned solid fills walk toward the 900 depth. Disabled is an opacity
(MD3's 38%); focus keeps the surface and strokes a neutral‑500 ring. Because
the dark ramp is paired, every state resolves in both modes with one rule.

### Elevation is a ladder

Elevation is to surfaces what states are to fills — a walk up the neutral
ramp, with the level naming the rung:

| Level | Surface | Step |
| --- | --- | --- |
| 0 | the app background | the Background pin (step‑100 ground) |
| 1 | card / raised-in-place surface | neutral 200 |
| 2 | menus, toasts, higher panels | neutral 300 |
| 3 | the top of the desktop ladder | neutral 400 |

Because the ramps are paired, a raised surface lightens in dark mode and
darkens in light mode with no second rule — dark-mode surface tint, the one
thing MD3's tonal elevation existed to encode, falls out of the pairing for
free. State walks compose on top with the level's step as the ground. MD3's
levels 4 and 5 survived only as shims clamping to level 3 until the breaking
release deleted them: desktop has no six-storey stack. Shadows are opt-in vibrancy,
not part of elevation (§Desktop divergences).

### Which region wears which rung

The ladder says how high a surface stands; this says which part of a window
stands where. It was implicit in what shipped for a year and written down
only in ADR-021, which is why applications derived it independently and some
derived it backwards. Three kinds of area, three rungs:

| area | rung | resolves to |
| --- | --- | --- |
| **content ground** — the document, the transcript, the list the window exists to show | 0 | the Background pin |
| **chrome furniture** — sidebars, asides, rails, toolbars, inspectors | 1 | `Surface`, neutral 200 |
| **transient surfaces and edges** — what appears and leaves, what draws a boundary or a state | 2–3 | neutral 300 / 400, `Divider` |

- **The resting content ground is level 0** — the Background pin, which is
  why the ladder's step‑0 is a sentinel rather than a ramp step and why
  `RenderState.Ground`'s zero value is documented as "the window ground".
- **Furniture stands exactly one rung up.** One step of the neutral ramp is
  the whole separation between chrome and content; two is a mistake.
- **Levels 2 and 3 are transient, or they are edges** — a dialog and a
  toast's base at level 2, a popover and a dropdown at level 3, `Divider`
  and the state walks at neutral 300. No resting expanse of a window fills
  at level 2 or deeper, however it is labelled: permanence is the test and
  size is its tell.
- **Rungs walk from the local ground, not from the window.** A card on the
  content ground is level 1 over level 0; a control inside a dialog walks
  from level 2, which is what `RenderState.Ground` exists to say. A raised
  inset inside a level‑0 body steps up from the paper it lies on. Level 0
  itself answers differently, because the Background pin is off the ramp and
  has no step to walk from: a raised thing on the content ground is drawn as
  a level‑1 surface and its state walk starts from level 1's step — which is
  what `SurfaceAt` already says ("treat interactive regions on it as level‑1
  surfaces instead") and what the button register already resolves level 0
  to. A shared surface takes its ground as a parameter, because it cannot
  know it: a pattern that paints a plane and then a band over it — a table's
  grid and its header, a tab panel and its strip — walks that band from the
  plane's own rung and never from an absolute step, which is right only for
  as long as every caller happens to rest where it was written.
- **What is chosen is Primary‑tinted; what is transient is a neutral walk.**
  The item a window is currently showing takes `Ramps.Primary.Step(300)`;
  hover, pressed and a keyboard cursor stay step walks (§States are step
  walks), so a list can show a cursor and a current item at once without
  the two colliding.
- **The titlebar wears the ground of the region it caps** — content behind
  the strip where the platform allows it, an application-painted band where
  it does not, never an unpainted native strip over a painted window. Taking
  that treatment takes on what came with the strip: the platform's window
  controls now stand inside the application's own layout, so the region that
  reaches the top-leading corner owes them a measured run and reads their
  geometry off the band it gave them; and the native drag leaves with the
  native strip, so the capping regions say where the window may be picked up.
  Where the strip crosses a seam, each side wears its own fill and both hold
  one height — two depths across one strip read as a step in the window's top
  edge.
- **The check:** walk out from the middle of the window and the rung
  numbers must never decrease.

Because the ramps are paired, that one statement dresses both schemes and
inverts itself for free: in light the window is lightest at its centre and
steps darker outward (`#F6F6F6` → `#E8E8E8` → `#D4D4D4`), in dark it is
darkest at its centre and steps lighter (`#181818` → `#222222` → `#2E2E2E`).
The hexes are evidence; the rungs are the rule. Nothing here needs a token
that does not already exist — the grammar was read off `vaultview`'s frame,
where it already held, and written down because `mindchat`'s transcript
filled most of a window with the level‑2 rung and the window read darker in
its middle than at its edges.

### Density is a theme token

`tokens.Density` carries the drawn control height and inner padding:
Comfortable (36 dp control, the default) and Compact (28 dp). Every components and
patterns control sizes from it, so switching an app to Compact is a theme
change, not a sweep. The WCAG 2.5.5 pointer target (44 dp) is deliberately
*not* part of density — it is a constant floor, and components extend their
pointer area beyond the drawn control to meet it. The numbers are measured,
not invented (§Desktop divergences).

### Motion is a theme token

`tokens.MotionScale` carries MD3's motion *semantics* at desktop pace: five
duration stops (50/150/250/400/500 ms), the MD3 standard and emphasized
easing families as cubic-bezier control points, and spring presets
(mass/stiffness/damping) for the effects physics path. Components take
durations from the theme, never from local constants — which is what makes
reduced motion free (below).

### Typography is a theme token

**The theme owns the typeface** (ADR-003). `tokens.Typography` carries one
`TextStyle` per MD3 type role — fifteen roles — plus `Code`, a sixteenth
style outside the MD3 grid (BodyMedium's metrics on the mono face; MD3 has no
code role, so the org added one). Typography also carries the face collection
and builds **one** cached `text.Shaper` shared by every component. Roboto is
the default because the default Typography names it; Roboto Mono is its
companion for code. No library source file may import `gioui.org/font/gofont`
— a lint fails `go test`, and therefore CI, in the four component repos. A
second lint does the same for `color.NRGBA` literals. The old practice does
not merely look wrong; it fails the build.

### Accessibility composes on top

- **Contrast is guaranteed by construction, in APCA terms** (ADR-007): in
  both ramps, step 900 reaches Lc 90 and step 700 Lc 60 over the 100/200
  grounds, and every pin's `On*` colour reaches Lc 60 over its pin. WCAG 2
  ratios are computed and reported — conformance claims cite them — but they
  do not gate the palette, because WCAG 2 over-rates light-on-dark pairs.
- **High contrast is derived, not hand-written:** when the OS reports
  increased contrast, the Color observable emits a variant derived from the
  *same seed* with higher floors — step 700 at Lc ≥ 90, pinned on-colours at
  Lc ≥ 75, Divider from the strong-border step.
- **Reduced motion snaps:** while the OS preference is on, the Motion
  observable emits zero durations; duration-driven components complete in
  zero frames, spring-driven components read the zeros as the snap signal.
- **Hit targets** hold the 44 dp floor at every density, and every
  interactive component participates in focus and keyboard activation.

Apps do nothing to get any of this; it arrives through the theme.

### The export surface

`theme/export` serialises a theme emission to `theme.json` (the generative
parameters — the theme is reproducible from the file alone) and a CSS token
sheet (`--color-<role>-100…900`, the pins, `--font-*`, `--space-*`,
`--radius-*`, density, elevation surface roles, motion), plus foundation
pages that render the scales at real sizes. A round-trip test parses the CSS
back and asserts every value against the Go token it came from, so the two
cannot drift. The emitted bundle is committed in the org's `design`
repository — the repository this document lives in — and uploaded to
`claude.ai/design` by the org's `push-design.sh`. This is the prototyping
surface: look-and-feel decisions are judged in a browser before they cost a
golden-image sweep.

---

## Desktop divergences from MD3

MD3 assumes touch and Android. Where that assumption shows, Vibrant Gio
diverges deliberately and says why — this is what makes the system Vibrant
Gio's rather than a port (ADR-005: MD3's *system*, not MD3's *look*).

### Density: shadcn/ui's metrics, not MD3's

MD3 is touch-first: 48 dp targets, 40 dp buttons, 56 dp text fields.
Desktop numbers were **measured, not invented** — the three-way table lives
as the doc comment of `theme/tokens/density.go` (sources fetched/measured
2026-08-05):

| metric | shadcn/ui | MD3 | macOS (AppKit, measured) |
| --- | --- | --- | --- |
| button height, default | 36 px (h-9) | 40 dp | 24 pt regular, 28 pt large |
| button height, small | 32 px (h-8) | — | 20 pt small, 16 pt mini |
| input height | 36 px (h-9) | 56 dp | 24 pt |
| stacked spacing | 8 px label→control | 8 dp grid | 8 pt system spacing |

The picks: **Comfortable = 36 dp** (the shadcn/ui default, between macOS
large and MD3's 40) and **Compact = 28 dp** (macOS's large control height).
The pre-rework hardcoded 44 dp was rejected as a control height — 44 comes
from touch guidelines; it survives as the pointer-target *floor*, independent
of density.

### Elevation: tonal first, shadows opt-in

On desktop a raised surface reads as raised by **tint first, shadow second**.
Elevation is the surface ladder above; shadows survive only as explicit
vibrancy via `effects/depth`, and the verdict on when is recorded in that
package's doc: **a shadow marks what floats and can leave** — a toast, a
popover, a menu, a drag preview — never what is raised in place, which reads
as raised by its surface step alone. The cost backs the rule: one
`depth.Shadow` issues nine paint operations per frame (eight gradient fills
plus an interior fill, measured), a surface step is one `FillShape`. The
caller audit executed this: toast kept its shadow, mindchat's floating undo
bar kept its, card's `Elevated` variant lost its shadow and became a level‑2
fill.

### Motion: a subset, at desktop pace

MD3 defines sixteen duration stops; desktop wants fewer and faster. The scale
maps the existing five stops onto MD3's duration roles — 50 ms hover
feedback up to a 500 ms ceiling — keeping MD3's easing semantics (standard
and emphasized families, accelerate/decelerate variants) and adding spring
presets MD3 has no vocabulary for. The mapping and its reasoning live in the
token doc comment, the same contract as density's table.

### Blur: owned, measured, and rationed

Gio exposes no blur primitive and no custom shaders — but
`gioui.org/gpu/headless` can render an op list offscreen and read the pixels
back, which is a real backdrop-blur pipeline built from Gio's own primitives.
The org owns the kernel rather than importing one (all three candidate
libraries measured compromised: unmaintained, anomalously slow, or silently
wrong): `effects/blur` is a parallel three-pass box approximation of a
Gaussian, correct at the edges and allocation-free in steady state.

The economics are measured (ten-core Apple Silicon; a 60 fps frame budget is
16.7 ms). Full shipped pipeline for a 1440×900 backdrop — record ops,
headless render at reduced resolution, readback, blur — by divisor:

    ÷1  1440×900  σ=8   blur 5.8  ms   pipeline 29.0 ms
    ÷2   720×450  σ=4   blur 1.8  ms   pipeline  8.1 ms
    ÷4   360×225  σ=2   blur 0.74 ms   pipeline  2.9 ms   ← the working configuration
    ÷8   180×112  σ=1   blur 0.35 ms   pipeline  1.1 ms

(The pipeline beats the exploratory spike's numbers — 69.2/12.9/3.8/1.6 ms —
in every row by reusing the headless window and readback buffer; synchronous
GPU readback dominates, so the resolution divisor is the lever that matters.
The blur destroys the detail anyway.)

Three contracts follow from the numbers:

1. **Refresh policy.** The pipeline runs synchronously on the events thread;
   it is driven by *content change* (a dialog opened, a scroll settled),
   never per frame — painting the cached result is free.
2. **Fallback.** Headless rendering is not available on every platform; the
   documented fallback is a flat tinted scrim, never a crash.
3. **What not to blur: animated glows.** A blur-based glow was prototyped,
   measured and rejected (evidence in `effects/glow`'s doc): visually better,
   but an animating blur-glow costs 0.2–0.8 ms of events-thread CPU plus an
   image-sized allocation and texture upload per glow per frame, against
   ~0.5 µs for the eight-gradient halo — and no cache holds while the radius
   or intensity animates. A correct approximation beats a slow exact answer.

### The component inventory: shadcn's, not MD3's

Patterns' inventory — shell, navbar, sidebar, table, pagination, tabs, modal,
alert, popover, tooltip, toast, card, accordion, breadcrumb, plus the
marketing sections — is shadcn/ui's inventory. MD3 has no breadcrumb, no data
table and no pricing section; conversely there is no FAB, navigation rail,
bottom sheet or snackbar here, because adopting them would make a Mac app
read as an Android port. ADR-005 ratified a choice the code had already made.

---

## Key architectural patterns

*(Corrected to the shipped code; DESIGN-v1.md documents the original
pre-migration wiring. The operational rules an app author needs — AutoConnect
counts, pitfalls, recipes — live in `llms.txt`; this section records why the
architecture holds.)*

### 1. The events goroutine is the heartbeat

Gio has enforced a synchronous frame protocol since v0.9: events must be read
and `Frame()` called on the same goroutine. `mvu.Window.Render` therefore
subscribes the `CombineLatest` of all layers on an rx goroutine and stores
each result as an **atomic snapshot**, then invalidates the window; the
events goroutine reads the snapshot when the frame event arrives and lays out
from it.

The consequence is the same claim the system was founded on, in its current
form: heavy upstream work (data processing, theme computation, layout
preparation) parallelises freely on rx goroutines, but values only reach
rendering at a frame boundary, and everything that touches Gio ops runs
single-threaded. The events goroutine is the heartbeat; everything else beats
to its rhythm. This is how FRP coexists with Gio's single-threaded
immediate-mode model — by making the frame loop the leader of the FRP graph,
not a participant in it.

### 2. Interaction state lives in `rx.Defer` closures

State allocated inside an `rx.Defer` factory is created once per subscription
and captured by reference in the map functions and widget closures below it.
It is only ever read or written from the events goroutine, so it needs no
locks — and it must never be handed to another goroutine or stored in a
subject. The scope hierarchy: Defer closure (once per subscription, owns the
lifetime) → Map closure (per emission) → widget closure (per frame, events
goroutine).

Component identity — v1's open "Experiment A" — landed as `components/keyed`:
`keyed.Defer` hands back the same per-key state across reorder, insertion and
deletion, which is what makes dynamic lists safe in the FRP style.
First-frame sentinels standardised as `components/initial`.

### 3. `MessageOp` bridges components to MVU

Widgets emit messages by adding `mvu.MessageOp{Message: …}` to the ops
buffer; the runtime collects them during the frame and delivers them to
`Update`. The collection is a registered collector on the ops buffer — the
`unsafe.Pointer` reinterpretation of Gio's internal ops layout that v1
documented as a known fragility was repaid during the Gio migration, exactly
as its repayment plan promised. One component serves both pattern families:
MVU consumers use `MessageOp`, FRP consumers pass plain callbacks and wrap
them in subjects. The bridge is event-shaped, not state-shaped.

### 4. Animation self-schedules and idles

Animated widgets tick their simulation inside the frame and request the next
frame only while active (`gtx.Execute(op.InvalidateCmd{})`); when activity
settles, nothing invalidates and Gio goes idle. Invalidation is
window-global — every widget re-lays-out — so expensive widgets cache ops
when inputs are unchanged (`components/cache`, v1's "Experiment B" landed).
Effects are explicit *variants* of components widgets (`effects/springbutton`
wraps `components/button`), opt-in per call site, never a global decorator — and
every animated component takes its durations and springs from the theme's
MotionScale, which is what makes reduced motion a zero-cost guarantee.

---

## Threading & lifecycle

1. **The events goroutine is single-threaded and authoritative.** Anything
   that mutates UI state, allocates Gio ops, or reads Gio's event queue runs
   there. Never call Gio from your own goroutine.
2. **Upstream observables may be multi-threaded.** The layer subscription
   runs on an rx goroutine; values cross into rendering only via the atomic
   snapshot, read at frame time.
3. **`Defer`-scoped state is implicitly serialised** by being touched only
   from the events goroutine. Do not pass it to goroutines.
4. **Messages cross thread boundaries via a buffered channel.** MVU updates
   run on the loop's goroutine, not the events goroutine.
5. **Subscription lifecycle.** A subscription begins at `Render`; every
   `Defer` factory runs once. It ends when the window closes
   (`app.DestroyEvent`) or the observable completes or errors — state is
   collected with the closure. Re-subscription is not supported: to restart a
   UI, build new layers for a new `Render` call. Components must not error
   except on unrecoverable failure.

---

## The road from v1

The original document ([DESIGN-v1.md](DESIGN-v1.md)) planned in phases named
Components → Theme → Effects → Patterns and recorded open experiments and known
fragilities. Its bets mostly paid off, and its debts were repaid:

- **The Gio migration ("Phase −1") happened.** The stack sits on a current
  Gio (v0.10.x), and the architecture's abstract claims survived the API
  rework — the strongest evidence they were architecture rather than
  coincidence. The `unsafe.Pointer` MessageOp hack is gone.
- **The validation experiments became packages.** Keyed identity is
  `components/keyed`, op caching is `components/cache`, cross-widget coordination
  became `components/coordination` — since deprecated: ADR-008 (recorded in
  `llms.txt` §Coordination) retired the cross-widget bus in favour of
  `mvu/stream.Value` and the arbiter pattern.
- **The layering inverted.** v1 put the token contract in components with
  spectrum above it; ADR-001 records why that was wrong and the tier table
  that replaced it.
- **The token scale outgrew its sources.** v1 aligned with Tailwind's values
  wholesale; the shipped system derives colour from a seed (ADR-002/007) and
  keeps only the 4-pt spacing scale idea. The "three design systems in a
  trench coat" token package is gone.
- **MD3 stopped being the reference for the look.** ADR-005 records the
  split; Phase E implemented it.

What survives intact from v1 is the application model itself — the FRP/MVU
duality, Defer-scoped state, the heartbeat, frame-driven physics — and the
project's voice: measured over assumed, explicit over magic, accessibility
non-optional.

---

## Decision records

The seven architecture decision records below were made and landed during the
Phase A–F rework. They are adapted from the working plan: task mechanics are
stripped, decisions, reasoning and measured evidence are kept. All seven are
**adopted and landed**, including the consequences that were still planned
when this edition was first written — the release tags were cut and the
shim-deleting breaking release happened. The records keep the repository
names the decisions were made under (`spectrum`, `prism`, `pulse`,
`cadence` — today's `theme`, `components`, `effects`, `patterns`). Decisions
made after the rework — ADR-008's coordination retirement onward — are
recorded in `llms.txt` and the working plan, not back-filled here.

### ADR-001: Spectrum is the foundation, not a consumer

**Decision.** The token and theme contract moves from `prism` down into
`spectrum`. The design-system spine becomes
`mvu → spectrum → prism → pulse → cadence → markdown` — under today's names,
`mvu → theme → components → effects → patterns → markdown` — governed by the full
tier table in §The layering, which admits all nineteen library modules and is
enforced by the org's layer-check script. `spectrum/transition` moves to
`pulse/transition` (today `effects/transition`), since it is animation code — that move removed an edge
that already existed rather than preventing a hypothetical one: the
foundation imported the effects layer, tier 1 reaching into tier 3.
`spectrum/window` may keep its `mvu` dependency; mvu carries no design
tokens.

**Why.** `spectrum` — the theme runtime — depended on `prism`, the component
library it exists to theme. The theme therefore sat above what it themes,
which is why `LiveTheme` hardcoded the default palettes and why there was no
palette injection point anywhere in the stack. Separately, `prism` and
`pulse` required each other — a demo file's doing — and that cycle pinned
half the org to a stale prism while the other half moved on; cutting it is
what made one current version resolvable everywhere. The nested-demo rule
(demo modules may import upward, their parents may not inherit the edge)
exists so no demo re-closes a cycle.

**How it stayed non-breaking.** `prism/tokens`, `prism/theme` and
`prism/a11y` remained as alias packages of type aliases and re-exported
variables; every downstream import path kept working for one release cycle,
and the breaking release then deleted the shims (§The layering,
"Compatibility shims").

### ADR-002: CIELAB tone with OKLCh hue and chroma

**Decision.** Derive tonal palettes on two axes: **tone is CIELAB L\***,
exactly as MD3 defines it, and **hue and chroma come from OKLCh**. Replace
both the colour mathematics and the hardcoded values that preceded them. This
is HCT's architecture with OKLab substituted for CAM16. *(As originally
written this ADR also kept MD3's role vocabulary and its tone-assignment
tables; the D0.1 spike amended that — ADR-007 retired the tables. The
mathematics was not reopened: all three candidate assignment models need a
perceptually even lightness axis and none supplies one.)*

**Why not what was there.** The predecessor token package was three design
systems in a trench coat: MD3 type roles, a verbatim Tailwind v3 palette
wearing MD3 semantic names, Tailwind spacing and radius keys, MD3 elevation
levels, and CSS easing names. No single system's design logic survived the
mix — which is precisely why nothing felt designed together.

**Why not HCT.** MD3's own space carries CAM16 and viewing-condition
machinery that buys little on a desktop screen and is substantial to
implement correctly.

**Why not plain OKLCh.** OKLab's L is not CIELAB L\*. Deriving tones from it
means "tone 40" stops meaning what Google means by it. Keeping L\* as the
tone axis kept that vocabulary for free — and even after ADR-007 retired the
role tables, the axis still pays: the generator was validated against MD3's
published palettes, tone numbers stay comparable with Google's, and the spike
reproduced the seed exactly at tone 40 on this axis.

**Why OKLCh for the other two axes.** Holding CIELAB `a,b` constant while
sweeping L\* does not hold *perceived* hue constant — the blue shift is
exactly why Google built HCT rather than using CIELAB directly. OKLab fixes
it in a short, testable conversion chain, with no dependency and no
viewing-condition model.

**On `reactivego/luminance`.** That package already implemented the
sRGB ↔ XYZ(D65) ↔ CIELAB chain correctly and dependency-free, and its
`Lab()`/`RGB()` pair is precisely the tone axis this ADR needs. Its math was
**lifted into `spectrum/color` (today `theme/color`), not imported** — same author, so reuse rather
than a dependency decision. Not imported because the package is MD2-era by
design: its `Lighten`/`Darken` API and `Kn = 18` constant are a chroma.js
port tuned to the retired material.io Color Tool, it declares `go 1.14`,
carries no tests, and its `go.mod` would drag freetype into the foundation's
module graph via shared examples. Lifted: the conversion chain, the D65 white
point, the CIE ϵ/κ constants — with a file header recording the lineage.
Left behind: the lighten/darken API and the per-channel clamp, which is not
gamut mapping and was replaced by chroma-reduction mapping.

**Tailwind's ramps** may survive as an optional palette provider. They must
not appear in the semantic layer.

### ADR-003: The theme owns the typeface

**Decision.** `Typography` is a theme token carrying, per MD3 role, a full
`TextStyle` — typeface, weight, size, line height, tracking — plus the face
collection and a lazily built shaper. Roboto is the default because the
default typography names it. `Props.Shaper` survives only as an explicit
per-call override. No library source file may import `gioui.org/font/gofont`;
a CI lint enforces it.

**Why.** The predecessor `TypeScale` was fifteen `float32` sizes and nothing
else, so the theme had no seam for a typeface at all. The consequence was
mechanical: seventeen `Props` structs and 118 function signatures carried a
`*text.Shaper`, and components, effects, patterns and markdown all constructed a
gofont shaper inside library code — gofont was not merely used by the
examples, it was the compiled-in default of the component library. Meanwhile
`font` and `style`, the repos that package Roboto and a type scale, went
unconsumed by any library source.

**A measurement that sharpened the decision.** A survey mid-plan found
`style` imported by twenty-one files — thirteen demo mains, but also four of
the seven workbench applications. So `style` was not a vestige nobody used;
it was how every Vibrant Gio application that drew its own text got its
shaper. That did not weaken the ADR — the scale still could not vary with the
theme, which is the actual argument — but it made the migration a real
migration rather than a cleanup. It has since completed: as of Phase F1 no
reference app imports `style`, builds a shaper, or touches gofont.

**Consequences.** `style` is frozen rather than deleted: its MD2 scale is
superseded by `Typography`, and it keeps working through the deprecation
window. (`textdraw` is *not* deprecated — it is the low-level text layer and
has no replacement; only the scale froze.) The default Typography grew a
sixteenth style, `Code`, on the Roboto Mono face packaged alongside Roboto.

### ADR-004: The canonical agent guide lives in `workbench`

**Decision (amended 2026-08-23).** `llms.txt` lives at the root of the
`workbench` repository and is the single source. Every repository carries a
short, static `AGENTS.md` that names the module in one sentence and links
that raw URL. The content is never duplicated — only pointed at. `AGENTS.md`
is not generated, does not embed tags or import graphs, and is not rewritten
when `.github` changes. The original decision placed the guide in `.github`;
that is reversed: `workbench` is where the org showcases how to build
applications with Vibrant Gio, and so is the natural home for the guide that
teaches exactly that, while `.github` is where Vibrant Gio itself is built —
the plan and the gates. The version-sync scripts stay in `.github` and reach
across to the workbench clone.

**Why.** The guide was genuinely good — hundreds of accurate lines on the
MVU loop, rx semantics and real pitfalls — but it existed exactly once,
inside `workbench/`, and nothing linked to it: not the org profile, not any
repo README, and no repo had an `AGENTS.md` at all. An assistant pointed at
the organization read the profile README, found a repo list and screenshots,
and stopped. Worse, the guide taught the then-current defect: it omitted
`style` and `font` from its bootstrap skeleton, so an assistant that followed
it perfectly shipped a gofont application.

**Consequences.** The guide was moved, corrected, and has been rewritten
against each phase as it landed; every repo's `AGENTS.md` is a pointer to
it. The generated `AGENTS.md` machinery — `sync-agents.sh`,
`check-agents.sh`, `templates/` — was removed: a second copy of measurable
facts forced a commit in every sibling whenever the template or a nested
tag moved, which made `git log` on a leaf unreadable. The division of
labour that remains is the one this document's preamble states: `llms.txt`
teaches building *with* the system, DESIGN.md records *why* the system is
shaped this way.

### ADR-005: MD3's system, not MD3's look

**Decision.** Take MD3's *system* and reject MD3's *look*. Specifically:

- **From MD3:** the generative token model (ADR-002), the type-role scale,
  state layers, tonal elevation, and the motion semantics.
- **From shadcn/ui:** density, restraint, and the component inventory. Its
  metrics are the target for the `Density` token — copied rather than
  invented.
- **From neither:** the visual identity. That comes from `effects` — glow,
  depth, spring physics — which is what this document names as the point of
  the project.

**Why.** MD3 is touch-first: 48 dp targets, generous spacing, large type, and
a component set shaped for phones — FAB, navigation rail, bottom sheet,
chips, snackbar. Adopting its look would make a Mac app read as an Android
port, which defeats the word "native" in the project's own vision statement.

Patterns had *already* made this choice without recording it. Its inventory is
shadcn's inventory, not MD3's — MD3 has no breadcrumb, no data table and no
pricing section. This ADR ratified a decision the code made a year earlier,
so the next contributor stops trying to reconcile the two. The hardcoded
44 dp minimum height in `components/button` was the same tension showing up as a
magic number; it is now the density-independent pointer-target floor, with
36/28 dp as the measured control heights (§Desktop divergences).

**What shadcn is not adopted for.** Its colour model is flat, hand-authored
semantic pairs written twice, once under `:root` and once under `.dark` —
structurally what the predecessor token package already did, so taking it
would be standing still. shadcn moved to OKLCH values without moving to
generation; MD3 generates without a modern space; ADR-002 does both. Its
distribution model (copying component source into the consumer's repo) is
also not adopted: it has no Go idiom and fights module versioning and golden
tests. The philosophy behind it does carry over — components should be
readable and forkable, not opaque configuration surfaces.

**Consequences.** Phase E implemented the divergences this ADR licensed:
measured density, the tonal elevation ladder with opt-in shadows, the desktop
motion subset, and the blur economics — all recorded with their evidence in
§Desktop divergences from MD3.

### ADR-006: One workspace while developing, tags at the seams

*This is the record an outside contributor cannot infer from the repos: no
member repo contains a `go.work`, a `replace` directive, or any visible trace
of how twenty-odd mutually dependent modules are developed together.*

**Decision.** Development across the org's modules happens under a single
`go.work` at the workspace root — the shared parent directory where every
repository, `.github` among them, is cloned as a sibling. The file is
generated by `.github`'s `clone-all.sh` from what is actually cloned (36
modules in a full checkout) and committed nowhere: a workspace describes one
developer's checkout, and a partial checkout is a smaller workspace rather
than a list of dangling paths. No member repo ever gets a `replace`
directive. *(As originally adopted, the clones hid under a gitignored
`.repos/` inside `.github` and `go.work` was committed there; Phase G0E
moved the checkout to sibling directories and made the workspace
generated.)*

**Green has two meanings, and a change is done only when it has both:**

- **Green under the workspace** — `go build ./... && go test ./...` with the
  workspace active, resolving siblings from the working tree. This is what
  "green before commit" means for every change.
- **Green without it** — the same commands under `GOWORK=off`, resolving each
  `go.mod` against published tags. This is what CI sees, and what a stranger
  running `go get` sees.

**The two diverge at seams.** A seam is any change that creates or widens a
dependency edge — during the rework: spectrum gaining tokens and theme, then
Typography, the derived ramps, density, motion, and the mono face. At each
seam the lower module is tagged and pushed before the upper module's `go.mod`
names it, and a check script (`check-no-workspace.sh`) reports the
outstanding debt between a seam landing and its tags being cut — that gap is
where "green in the tree, broken for the world" lives.

**Why not `replace` directives.** They would have to be committed to be
useful to the next change, and a committed `replace` in a public module
breaks every consumer not sitting in this working tree. `go.work` is what Go
added for exactly this, and it is invisible downstream.

**Why not defer all tagging to the release.** The seams are load-bearing and
there were many: three separate phases each moved a contract down into
spectrum and then migrated four repos onto it. Leaving the modules mutually
unbuildable for months would make "never commit red" unenforceable — and a
rule that cannot be checked is not a rule.

**Sequencing note that generalises:** the workspace was established only
*after* version alignment, not before. A workspace resolves shared
dependencies at the highest version any member requires, so joining 36
modules while their Gio versions were still spread would have broken the ones
on older versions — a self-inflicted failure looking exactly like the drift
the alignment existed to fix. Align, then join.

**The seam procedure is proven, not theoretical.** The version-alignment
baseline ran it end to end: tag and push the bottom layer, bump the layer
above onto those tags, verify with `GOWORK=off`, tag and push it, repeat.
Seven layers, in this order:

```
0  mvu font traer svg seen ivg kiwi noise csg circle gradient backdrop textdraw
1  style  kiwi/gio  seen/context/gio  svg/driver/{pdf,raster}
2  svg/driver/{gio,seen}  ivg/raster/gio  traer/gio
3  prism      4  pulse      5  spectrum      6  cadence markdown
7  workbench/*  mvu/example  prism/gallery
```

(That list is the alignment baseline's order, under the old names and the
pre-inversion stack — spectrum then still sat above prism. The release
protocol's coarser cut under today's stack: mvu, font → theme →
components → effects → patterns, markdown → workbench apps, with the nested demo
modules last. font tags beside mvu, not beside theme, because theme
requires it — ADR-003's tier-0/tier-1 edge. No layer is tagged while the
layer check or the workspace-debt check fails.)

Three lessons that cost time and will cost it again:

1. **A module's newest git tag is authoritative, not the proxy's `@v/list`**
   — that endpoint caches and reports a version behind for a while after a
   push, which reads as false staleness.
2. **Tagging a whole layer in one round leaves each new tag referencing its
   siblings' previous tags**; making the set self-referencing needs a second
   pass, so budget both.
3. **Go keeps two caches, and clearing one is not enough.** Besides the
   module cache at `$GOMODCACHE/cache/download`, Go keeps a bare git clone
   per repository under `$GOMODCACHE/cache/vcs`, and that clone still holds
   tags deleted upstream. Evicting only the download cache and re-resolving
   appears to prove a withdrawn version is still published when it is being
   served from the stale clone. Evict both, or `go clean -modcache`, before
   concluding anything about what a stranger can fetch. (Measured after
   evicting both: two properly withdrawn versions failed with `unknown
   revision` while their replacements resolved normally.)

**Tag rules.**

- **No double-digit component in any tag. Ever.** When a series hits `.9`,
  the next release rolls the component above: `v0.0.9 → v0.1.0`, never
  `v0.0.10`. This is a hard rule, not a preference: a tag is immutable the
  moment the proxy sees it, so a `v0.0.10` cannot be withdrawn, only buried
  under a correction that leaves it in the list forever.
- **Violated once, recorded honestly.** A run of seam tags was cut without
  consulting the rule: spectrum ran v0.0.10 through v0.0.15 and pulse v0.0.10
  through v0.0.12, and all nine are on the remotes, immutable. The remedy was
  the rule's own: bury them. Spectrum's next tag was **v0.1.0** and pulse's
  was **v0.1.0**, cut exactly as the release tasks named them — and neither
  repo has published another v0.0.x since.
- **A nested module's tag mirrors its root's version.** A module in a
  subdirectory is tagged `<subdir>/vX.Y.Z`, requiring the root at exactly
  `vX.Y.Z`; root first (committed, tagged, *pushed* — the submodule cannot
  `go get` an unpushed tag), submodule second. A submodule is not re-tagged
  every time the root moves — the root running ahead of an unchanged
  submodule is the normal state, not drift. What the rule forbids is the
  submodule running on its own counter, which destroys the correspondence;
  if the matching root number is already taken, cut a fresh root tag rather
  than letting the submodule run ahead.

**The retraction window, precisely.** `GOPRIVATE` on the development machines
covers `github.com/vibrantgio/*`, so Go there bypasses `proxy.golang.org` and
`sum.golang.org` entirely and resolves straight from GitHub — which is why
`git ls-remote` is the authority worth consulting locally. But that is a fact
about those machines, not about the modules, and reading it as the latter was
a recorded mistake. These are public repositories: anyone without `GOPRIVATE`
resolves through the proxy and verifies against `sum.golang.org`, and doing
so records the version there *permanently* (one `go install` from an
unconfigured machine did exactly that; measured afterwards, the proxy listed
25 versions of mvu and the sumdb answered 200 for probed versions). A version
recorded in the checksum database is immutable forever: retag it and every
non-`GOPRIVATE` consumer gets a permanent checksum mismatch no origin repair
can fix. So the retraction window closes per version, silently, and nothing
local tells you it has.

**And the check is not free — this is the trap.** The proxy fetches from
origin on a cache miss, so probing a version that exists is one of the ways
it gets mirrored; `sum.golang.org/lookup` will compute and *append* a missing
entry. The observation causes the thing being observed. Therefore: while a
tag is in flux, use `git ls-remote` — authoritative and inert — and probe the
public infrastructure only when a version is final or when the answer changes
the next action.

### ADR-007: Nine functional steps, paired dark ramps, APCA contrast

**Decision.** Tone stops map to roles the functional way, in Claude Design's
vocabulary: every colour role carries a **nine-step ramp, 100–900**, where
the step *is* the meaning — 100–300 tinted fills, hovers and subtle borders,
500 the mid-value reference, 700–900 text over tinted fills and pressed
states — and the role's **base is pinned separately** from the ramp, exactly
as the reference project's `theme.json` pins `accent`. **Dark mode is a
paired ramp, not a second table**: the generator emits light and dark scales
in which the same step keeps the same job — Radix's pairing mechanism under
Claude Design's numbering — so a component asks for neutral‑200 and gets a
light card on a light ground and a dark card on a dark one, with no second
assignment table to drift. **The contrast gate is APCA**: in both ramps, step
900 must reach Lc 90 and step 700 Lc 60 over the step‑100 and step‑200
grounds, and each pinned base's on-colour Lc 60 over the base. WCAG 2 ratios
are still computed and reported — conformance claims cite them — but they do
not gate the palette.

MD3's role→tone tables are retired. A thin semantic layer — background,
surface, text, divider, plus the pinned role bases — sits over the ramps so
call sites read intent; the MD3-named fields survived as deprecated aliases
resolved into ramp steps until the breaking release deleted them.

**The surface mapping.** Identical in both modes, because the dark ramp is
paired rather than re-assigned:

| Surface | Step |
| --- | --- |
| app background | neutral‑100 (or the pinned `bg`) |
| card / raised surface | neutral‑200 |
| hovered element background | one step past its ground (200 → 300) |
| pressed / selected background | two steps past its ground |
| subtle border, separator | neutral‑300 |
| strong border, focusable edge | neutral‑500 |
| solid fill | the pinned role base |
| solid hover / pressed | one / two steps from the pin toward 900 |
| low-contrast text | neutral‑700 (Lc ≥ 60 guaranteed) |
| body / high-contrast text | neutral‑900 (Lc ≥ 90 guaranteed) |

**The evidence.** All three candidate models were generated from the seed
`#6750A4` by a spike script implementing ADR-002's math — CIELAB tone, OKLCh
hue and chroma, chroma-reduction gamut mapping — which reproduces the seed
exactly at tone 40. The Radix columns re-hue its published violet scales to
the seed's hue and chroma; the Claude Design columns use the shared lightness
scale measured from the reference project's own ramps (steps 100–900 ≈ L\*
97, 92, 85, 74, 63, 51, 39, 28, 18). MD3 states are its 8%/12% overlays; the
other two walk ramp steps.

| Surface | MD3 light | Radix light | CD light | MD3 dark | Radix dark | CD dark |
| --- | --- | --- | --- | --- | --- | --- |
| app background | `#faf9ff` | `#fcfcfe` | `#f5f4fc` | `#141318` | `#14121c` | `#18171c` |
| card | `#eeedf4` | `#f9f8fe` | `#e8e7ee` | `#201f24` | `#191622` | `#222126` |
| hover | `#dddce3` | `#e9e5f9` | `#d4d3da` | `#2f2e34` | `#32284f` | `#2e2e33` |
| pressed | `#d5d4db` | `#dfdbf6` | `#b7b6bd` | `#37363c` | `#3c315b` | `#47464c` |
| subtle border | `#c7c5d3` | `#d3ccf2` | `#d4d3da` | `#474551` | `#463b68` | `#2e2e33` |
| strong border | `#787582` | `#c1b8e6` | `#98979e` | `#918f9d` | `#554a7b` | `#9e9da4` |
| solid fill | `#6750a4` | `#735cb1` | `#6750a4` | `#cbbeff` | `#735cb1` | `#a690ea` |
| low-contrast text | `#474551` | `#68559f` | `#5d5c62` | `#c7c5d3` | `#b8abeb` | `#cccbd2` |
| body text | `#1c1b20` | `#332851` | `#2b2a30` | `#e3e2e9` | `#e2def6` | `#eeedf4` |

All three set a competent table — the differences are in who maintains what.
MD3's dark column exists because a second hand-written table says so; the
other two derive it. Radix alone keeps the brand colour identical in dark
mode; MD3 and this ADR lighten the fill and accept that dark mode shifts the
accent, which is also what every Material app already does.

**Where nine cannot say what twelve can.** Hover background and subtle border
collide on step 300 (visible in the CD columns above), and there is no
dedicated solid-hover stop. Neither costs this org anything: components and
patterns carry exactly two border weights — mapped to 500 and 300 — and
derive hover as a state resolution rather than a distinct token; MD3 itself
makes hover an 8% overlay, not a stop. Radix's extra resolution buys
precision nothing in components or patterns consumes.

**The seed sits deep, so bases are pins.** `#6750A4` is L\* 40 — step‑700
depth on the shared scale, where 500 sits at L\* 63. Reading the solid fill
off the ramp would lighten the brand colour to `#a08ae4`; pinning reproduces
it exactly. This is the reference project's own practice, not a deviation:
its `.btn-primary` uses `--color-accent` — the pin — while the ramp supplies
tints and text shades.

**Why APCA and not WCAG 2.** The theme tracks OS dark mode by default, and
WCAG 2's known failure mode is over-rating light-on-dark. Measured on this
seed's own dark palettes: outline-strength text `#918f9d` on an MD3 card
`#201f24` scores WCAG 5.17:1 — a clean AA pass — at APCA Lc −41, unreadable
as body text. The seed's tone 60 `#9983dc` on the dark ground passes AA at
5.85:1 with Lc −42; tone 50 passes AA-large at 4.13:1 with Lc −30. Every one
of those would sail through a ratio gate and fail readers. On pairs that are
genuinely fine the two metrics agree (dark-mode body text lands Lc −87…−96
across all three models), so APCA costs nothing where WCAG was right. Radix
reaches Lc 60/90 by hand-tuning; here the generator meets the same numbers by
test.

**Tunings the gate forced, landed with the generator.** The spike predicted
one small tuning; implementation found it needed two larger ones. APCA's soft
black clamp caps even pure black near Lc 92 over the L\* 92 step‑200 ground,
so the light 900 stop deepened from the measured L\* 18 to L\* 6 — the depth
where all five ramps clear Lc 90 with margin. And the dark pins rose from the
measured L\* 65 — a mid-tone no text of any colour reaches Lc 60 over (black
tops out near 52, white near 57) — to L\* 82, the dark scale's step‑700 depth
beside MD3's dark accent tone 80. The dark fill in the evidence table,
`#a690ea`, survives as the dark primary ramp's step 500; the shipped dark pin
for this seed is `#d0c4ff`. The evidence table itself is the spike's
measurement and stands unchanged.

**Why not MD3's tables.** The design knowledge lives in two hand-written
role→tone tables that must be kept in step — dual authorship of dark mode,
the exact drift this system keeps removing elsewhere — feeding a
twenty-six-name role set whose meaning lives in documentation rather than
structure. Its states are alpha overlays a token sheet cannot address, and
its conformance anchor is the ratio shown above over-rating every dark pair.
Its tone *mathematics* is kept in full (ADR-002); only the assignment tables
retire.

**Why not Radix's twelve.** The strongest model in isolation, and this ADR
takes its two best ideas — paired scales and APCA guarantees. But its step
numbers are its vocabulary, and that vocabulary collides with the surface
this org prototypes on: the token export feeds a Claude Design project whose
entire convention speaks `--color-*-100…900`. Adopting step‑9/step‑11 would
put a translation table between every prototype and the app it prototypes —
the exact incoherence this system exists to remove. Its scales are also
hand-tuned per hue rather than generated, so "adopt Radix" really means
"build a different generator and borrow its numbering" — and the numbering is
the part that costs.

**Consequences, as landed.** The ramp/pin/semantic types and the paired
generator live in `theme/tokens` (the MD3-name aliases are gone since the
breaking release);
interaction states resolve as step walks rather than opacity overlays; the
APCA gates run in the test suite with WCAG AA reported alongside; the token
export emits `--color-<role>-100…900` plus the pinned bases, and the colour
foundation page annotates step purposes and the measured Lc of each text
pair. The elevation ladder (§The generative model) extended the same
mechanism to surfaces. ADR-002 is amended by this record: its mathematics
stands untouched; its role-vocabulary clause is superseded.

---

## Design principles

- **One theme carries the whole look:** colour, typography, density,
  elevation and motion all arrive through the theme observable; a correct
  app wires the theme once and styles nothing per component
- **Derived, not picked:** palettes generate from a seed; contrast is
  guaranteed by construction and gated in APCA terms
- **Observable-native:** components participate in the FRP graph from birth,
  not retrofitted
- **Single-threaded UI:** the events goroutine owns rendering; everything
  else beats to its rhythm
- **Defer for interaction state:** mutable state lives in `rx.Defer`
  closures, never in subjects
- **Frame-driven motion:** animated components self-schedule and idle when
  settled; reduced motion snaps for free
- **Progressive enhancement is explicit:** effects widgets are *variants* of
  components widgets, not silent decorators
- **Desktop-native over touch-translated:** measured density, tonal
  elevation, a faster motion subset, shadows only for what floats
- **Accessibility is non-optional:** keyboard, focus, contrast floors, hit
  targets and reduced motion are contract, not options
- **No string tokens:** all design values are typed Go values, and the lints
  make the old practices fail the build
- **Module boundaries are enforced:** the tier table is checked in CI, and
  green has two meanings — the workspace and the world
- **Performance is measured, not assumed:** density metrics, blur budgets
  and glow verdicts all cite benchmarks, and golden tests pin what renders
