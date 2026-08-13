# Vibrant Gio Design System — Architecture & Strategy

> **SUPERSEDED (2026-08).** This is the original, pre-rework design document,
> kept for history. The current document is [DESIGN.md](DESIGN.md), which
> records how each of this document's plans and open questions resolved.

> **Status:** Working design. Phase names committed: **Prism** (foundation), **Spectrum** (theme), **Pulse** (effects), **Cadence** (patterns) — see *Naming Considerations*. Currently pinned to `gioui.org v0.1.0` (2023); migration to a current Gio release is tracked separately as Phase −1 below.

## Vision

Vibrant Gio is a platform for building beautiful, native desktop applications on macOS, Windows, and Linux, built on [gioui.org](https://gioui.org). The goal is a first-class design system — analogous to Material Design for Google — but unique to Vibrant Gio. The "vibrancy" in the name is both literal (rich colour, depth, motion) and philosophical (alive, reactive, responsive).

The application model is **Functional Reactive Programming** using `reactivego/rx` and `vibrantgio/mvu`. The design system must feel native to this model — not bolted on top of it.

### Non-goals

- **No web target.** This is a desktop-native system; we do not aim for browser deployment.
- **No mobile target initially.** Gio supports Android/iOS but Vibrant Gio is optimised for desktop input modalities, window chrome, and platform integration.
- **No embedded / kiosk target.** No memory- or binary-size budgets shaped around constrained devices.
- **No CSS-like dynamic styling.** Tokens are typed Go values, not stringly-typed maps.

---

## Current Module Inventory

### Foundations (production-ready)

| Module | Role |
|---|---|
| `mvu` | MVU runtime: generic `Loop`/`Run`, `Window.Render(layers...)`, `MessageOp` widget protocol |
| `textdraw` | Low-level text: glyph-level control, alignment, label backgrounds |
| `style` | Typography scale (H1–H6, Body, Button, Caption) wired to Roboto |
| `font/roboto` | Roboto typeface, five weights |
| `circle` | Mathematically precise circle via Bezier approximation |
| `backdrop` | Solid colour fill widget |
| `gradient` | Linear gradient fill |
| `svg` | SVG parse + render; `IconWidget(icon, w, h, opacity)` Gio integration |
| `ivg` | IconVG encode/decode + rasteriser; Material icons bundled |

### Simulation & 3D (complete, partially wired to Gio)

| Module | Role |
|---|---|
| `traer` | Particle physics: springs, attractions, Verlet integration, `FPS` counter |
| `seen` | Hierarchical 3D scene graph: groups, transforms, lights, cameras |
| `csg` | Constructive solid geometry: Union, Subtract, Intersect via BSP |
| `kiwi` | Cassowary constraint solver for adaptive layouts |

### Applications (reference implementations)

| App | Pattern | Notes |
|---|---|---|
| `coinviz` | Pure FRP (RX) | Real-time crypto charts, 16 pane types, live WebSocket data |
| `appviz` | FRP (RX) | Business dashboard, FX rates, fiscal reports |
| `mindchat` | MVU | Claude API chat interface |
| `todos` | MVU | Full CRUD with persist — canonical MVU reference |

---

## Key Architectural Patterns

### 1. FRP Application Structure (coinviz model)

Data flows in one direction through observable transformations. There is no explicit state struct; state emerges from the observable graph.

```
Data source (WebSocket / file)
  → rx.Observable[DataFrame]
  → rx.CombineLatest3(themes, panes, data)
  → rx.Map(…, transformToWidget)
  → Window.Render(background, content)
```

Theme switching is reactive: `theme.AutoLightDark()` returns an `rx.Observable[Theme]` that emits on time-of-day changes. The entire UI re-renders with no imperative wiring.

### 2. MVU Application Structure (todos / mindchat model)

Explicit state machine: `Init → (Model, Command) → Update(Model, Message) → (Model, Command) → View(Model) → layout.Widget`. Async side-effects (API calls, persistence) are modelled as `Command` values executed by the runtime.

MVU uses RX internally for async command execution. Both patterns co-exist in the workspace and are not mutually exclusive.

### 3. The `WithLatestFrom2` Frame Synchronisation Model

**The single most important architectural fact about Vibrant Gio.** It resolves what would otherwise be a thicket of race conditions.

In `mvu/window.go:56`:

```go
pairs := rx.WithLatestFrom2(
    events,
    rx.Map(rx.CombineLatest(layers...), invalidate).SubscribeOn(rx.Goroutine),
)
```

`WithLatestFrom2` makes `events` the **leading observable**. A pair is delivered only when the leading observable emits. The trailing observable (the `CombineLatest` of all layers, mapped through `invalidate`) can recompute on a separate goroutine — its `SubscribeOn(rx.Goroutine)` says exactly that — but its values only reach the observer when a frame event arrives.

**The consequence:**

- Heavy upstream work (data processing, theme computation, pane layout) parallelises freely on RX goroutines.
- Everything that touches Gio code runs on a single thread, gated by the events drumbeat.
- The observer (the subscribe callback) is invoked sequentially. There is no concurrent execution inside `widget(gtx)`.
- `Defer`-scoped state mutations are safe because the only place they happen is inside the observer.

When a layer's value changes, `invalidate` calls `w.window.Invalidate()`, which causes Gio to post a `FrameEvent`. That `FrameEvent` is what unblocks `WithLatestFrom2` to deliver the latest layer values for rendering. **The events observable is the heartbeat; everything else beats to its rhythm.**

This is the answer to "how does FRP coexist with Gio's single-threaded immediate-mode model" — by making the immediate-mode loop the leader of the FRP graph, not a participant in it.

### 4. The `rx.Defer` Subscription-State Pattern

Used throughout coinviz; canonical example at `coinviz/content.go:27-31`. State allocated **inside** an `rx.Defer` closure is created once per subscription and captured by reference in all subsequent map functions and widget closures.

```go
rx.Defer(func() rx.Observable[layout.Widget] {
    // Allocated ONCE when subscribed — survives all emissions
    button    := widget.Clickable{}
    offset    := struct{ X, Y unit.Dp }{X: -1}
    crosshair := f32.Pt(-10, -10)

    return rx.Map(rx.CombineLatest3(themes, panes, data),
        func(next rx.Tuple3[...]) layout.Widget {
            return func(gtx layout.Context) layout.Dimensions {
                // Mutations here survive to the next emission
            }
        })
})
```

**Scope hierarchy:**
- `Defer` closure → runs once per subscription → owns the state lifetime
- `Map` closure → runs per emission → captures state by reference
- `layout.Widget` closure → runs per Gio frame, on the events thread → reads and mutates state

**Why this works without locks:** because of the `WithLatestFrom2` synchronisation model (Pattern 3). Both the `Map` closure (reached via the trailing observable) and the widget closure (executed inside the observer) ultimately serialise through the same observer. The state is only ever read or written from the events thread.

**Why `rx.Defer` and not `rx.Scan`:** Scan is the idiomatic FRP accumulator, but it can only update from upstream values — not from synchronous mutations inside a frame callback. Pointer events, scroll deltas, and physics ticks happen *during* `widget(gtx)`; they need to mutate state right there, then have the next emission see those mutations. Defer-scoped variables provide that escape hatch without leaking interaction state into the observable graph.

**Failure modes to be aware of:**

- **Multi-subscription duplicates state.** `rx.Defer` re-runs its factory on every subscription. If the same observable is subscribed twice (e.g., a pane shared across windows), each subscription gets independent state — usually wrong. Use `rx.Share` / `rx.Publish` upstream of `Defer` if a single state instance must back multiple consumers.
- **Re-subscription resets state silently.** If the upstream completes or errors and something resubscribes, state is reborn. Scroll position, crosshair, animation phase all reset with no warning. Components must declare whether they tolerate re-subscription, and avoid sources that complete unexpectedly.
- **State outlives subscription only via leak.** When a subscription ends (the layer is removed, the window closes), the state is garbage-collected with the closure. There is no "tear down on unsubscribe" hook beyond what RX provides — components needing cleanup (e.g., closing files, releasing OS resources) must use `rx.Finally` upstream.
- **Initial-frame sentinels.** The first emission usually fires before the first frame, so layout-derived defaults are not yet known. coinviz uses `offset.X = -1` as a sentinel meaning "compute on first frame." This is a recurring pattern that Prism components will hit repeatedly; standardise on a `prism.Initial[T]` helper rather than ad-hoc sentinels.

### 5. Frame-Driven Physics Loop (traer/gio pattern)

Physics does not produce observables. It is a stateful simulation that runs *inside* the Gio frame handler, using `op.InvalidateOp` for self-scheduling. Reference: `traer/gio/gravity/main.go:58,99`.

```go
activity := ps.Tick(math.Max(1, fps.Value/30))
// ... render from current positions ...
if activity > 0.01 {
    op.InvalidateOp{}.Add(gtx.Ops)
}
```

Properties:
- **Self-scheduling:** the widget requests its next frame; no external ticker
- **Idle at rest:** when `activity` settles, no `InvalidateOp` is emitted, and Gio goes idle
- **Composable:** multiple animated widgets each contribute `InvalidateOp` independently; one is enough to keep the window animating

**Caveats for multi-widget Pulse use:**

- **`InvalidateOp` is window-global.** A single animated widget triggers a full layout pass — every other widget on the window re-runs its layout function. Components with expensive layouts must internally cache results when their inputs are unchanged. Document this on every Prism component.
- **Independent simulations are not synchronised.** Two widgets ticking their own `ParticleSystem` will not produce a coordinated wave. For coordinated motion (e.g., staggered list reveal), introduce a shared clock / animation conductor at the Pulse layer.
- **Variable dt is hostile to Verlet stability.** `max(1, fps.Value/30)` floors the step size, sacrificing real-time accuracy for stability under load. This is a deliberate trade — document it explicitly. Pulse should optionally support a fixed-timestep mode with an accumulator for cases where real-time sync matters.
- **Reduced motion must short-circuit physics entirely.** macOS exposes `NSWorkspace.shared.accessibilityDisplayShouldReduceMotion`; Windows exposes `SystemParametersInfo SPI_GETCLIENTAREAANIMATION`. Prism reads this preference once and exposes it as `rx.Observable[bool]`. Pulse components subscribe and skip the physics tick when reduced motion is on, snapping directly to the target state.
- **Spring physics is overkill for everything.** Pulse needs a tier below `traer`: a simple `Tween[T]` for fades, slides, and colour interpolations. Reach for the particle system only when the motion needs to feel *physical*.

---

## Threading & Lifecycle

### Threading rules

1. **The events observable is single-threaded and authoritative.** Anything that mutates UI state, allocates Gio ops, or reads from Gio's event queue runs on the events thread.
2. **Upstream observables may be multi-threaded.** Use `SubscribeOn(rx.Goroutine)` to offload heavy work. The `WithLatestFrom2` join ensures values only cross into the events thread at frame time.
3. **`Defer`-scoped state is implicitly serialised** by virtue of being read/written only from the events thread. Do not pass it to goroutines.
4. **MessageOps (the `MessageOp` widget protocol) cross thread boundaries via a buffered channel** (`messageOps` in `mvu/window.go:23`). MVU updates run on the channel-reading goroutine, not the events thread.

### Subscription lifecycle

- A subscription begins when `Window.Render(layers...)` is invoked. At that moment, every `rx.Defer` factory in every layer runs once.
- A subscription ends when the underlying observable completes, errors, or the window closes. State allocated inside `Defer` is collected with the closure.
- If a layer observable errors, the current implementation logs and stops delivery. There is no recovery / retry mechanism. **Components must not produce errors except on unrecoverable failures.**
- Re-subscription is currently not supported by `Window.Render` — once `Subscribe` returns, the same subscription cannot be revived. To "restart" a UI, build a new layer observable and pass it to a new `Render` call. This needs explicit documentation as it is a pitfall.

---

## Accessibility (cross-cutting, owned by Prism)

Accessibility is a first-class Phase 1 concern, not a future add-on. Components without a11y support do not ship in Prism.

| Concern | Approach |
|---|---|
| Focus management | Every interactive component owns a `widget.Clickable`-equivalent and participates in tab order via Gio's focus system; a `prism.FocusGroup` orchestrates groups. |
| Keyboard activation | Space/Enter on any focusable component fires the same callback as click. |
| Reduced motion | OS preference exposed as `rx.Observable[ReduceMotion]` in the theme; Pulse consults it before any animation. |
| Contrast | Colour tokens are paired (Background/OnBackground, Surface/OnSurface, etc.) to guarantee minimum WCAG AA contrast in both light and dark themes. |
| Hit targets | Spacing scale enforces minimum 44 dp interactive targets; documented as a hard rule. |
| Screen reader | Every interactive component takes a `Description string` (passed to `app.Description`); decorative-only components take none. |
| Internationalisation | Text shaping uses the full Gio shaper pipeline; CJK/Arabic/Hebrew/RTL are supported via Gio's text layer. Roboto is the default but not the only font; theme tokens carry full `text.FontFace` slices. |
| BiDi text | Inherited from Gio's shaper; documented but not extended. |

---

## Known Fragilities

These are existing implementation hazards that the documented architecture rests on. They are not patterns to emulate — they are debts to repay.

### MessageOp extraction via `unsafe.Pointer`

`mvu/window.go:69-78` reaches into `op.Ops.Internal` via an `unsafe.Pointer` reinterpret cast in order to find `MessageOp` values that widgets have added to the ops buffer:

```go
type unsafeOps struct {
    version int  // "in gioui v0.8 this has become a uint32"
    data    []byte
    refs    []any
}
for _, op := range (*unsafeOps)(unsafe.Pointer(&ops.Internal)).refs {
    if msgOp, matches := op.(MessageOp); matches {
        w.messageOps <- msgOp
    }
}
```

This depends on Gio's *internal* `op.Ops` layout. The inline comment already records that the `version` field type changed between Gio versions, which means the cast silently produces wrong reads on the wrong version. Production code today is correct only because it is pinned to `gioui.org v0.1.0`.

**Repayment plan:** the Phase −1 Gio migration eliminates this hack. The post-rework Gio event API exposes a proper event/router pattern in which `MessageOp` becomes a first-class event source — read via `gtx.Source.Event(filter)` in the runtime. No `unsafe`, no internal-layout dependency. The migration is therefore a load-bearing prerequisite for declaring the MVU runtime stable.

### Gio version coupling

The architecture's correctness claims (notably the `WithLatestFrom2` synchronisation model) assume specific Gio behaviours: `app.Window.Events()` returns a channel, `system.FrameEvent` semantics, ops-buffer inspection. All of these change in current Gio.

**Repayment plan:** Phase −1 migration validates that the architecture's *abstract* claims (events lead, layers trail, observer serialises everything) survive the Gio API rework. The wiring in `mvu/window.go` is rewritten; the architectural claims should hold unchanged. If they don't, the migration uncovers a deeper fragility worth knowing about.

### `reactivego/rx` semantic coupling

The safety story relies on `WithLatestFrom2` having specific pairing semantics (leader-driven, latest-trailer-cached). If `reactivego/rx` ever changes those semantics, every claim in §3 needs re-verification.

**Mitigation:** pin the rx version explicitly; gate any rx upgrade on a re-run of the architectural review. This is documentation-only, but should be flagged on every rx version bump.

---

## Performance

The numbers below are **aspirational targets pending baseline measurement.** A Phase −1 / Phase 00 deliverable is to profile current coinviz on its target hardware and either confirm the budgets or revise them. Until that baseline lands, treat these as direction-setting, not contract.

- **Frame budget:** 16.6 ms target (60 FPS), 8.3 ms ceiling for layout work alone (leaving headroom for paint and present).
- **Allocation policy in the hot path:** zero allocations per frame inside widget closures, except where Gio's API requires them. Pre-allocate paths, slices, ops buffers in `Defer` scope.
- **Layer recomputation cost:** a layer's `Map` closure runs on every emission of its inputs. Expensive transformations (e.g., re-laying-out 16 panes) should be memoised inside the `Defer` scope and re-run only when meaningful inputs change.
- **Profiling:** every Prism component ships with a benchmark in `*_bench_test.go` that exercises `widget(gtx)` for 1000 frames; regression checks are run locally via `go test -bench` against the numbers stored in `BASELINE.md` (no CI gate — this is a solo-dev project).

### Methodology

- **Reference platform:** define one (likely current Apple Silicon macOS — coinviz's primary host) as the baseline. Cross-platform regression numbers are reported relative to it.
- **Baseline measurement:** before Prism work begins, capture per-pane `widget(gtx)` cost and allocation counts in coinviz under representative load (one symbol, 1h candles, all 16 panes visible, mouse moving). Numbers anchor every later comparison.
- **Benchmark harness:** a shared `prism/bench/` package providing a `BenchFrame(b, widget)` helper that drives `widget(gtx)` with synthesized constraints, captures `b.ReportAllocs()`, and standardises measurement across components.
- **Regression checks:** the >5% rule applies per-component to wall-clock and to `B/op`; the harness is responsible for measuring both. Run locally via `go test -bench` and compared by hand against `BASELINE.md`. No CI gate.

---

## Testing

- **Token tests:** `prism/tokens` exposes pure data; trivial to unit-test (contrast ratios, scale monotonicity, etc.).
- **Component golden-image tests:** each component has a `golden_test.go` that renders to a PNG via Gio's headless backend and diffs against a stored image. Failures land alongside the diff for review.
- **Observable tests:** RX observables are tested via `rx.TestScheduler` (verify which emissions happen at which virtual times). Theme switching, animation settling, and command execution are all `TestScheduler`-friendly.
- **Physics convergence tests:** for any spring/animation, a deterministic test runs the simulation with a fixed seed and asserts settling time and final position within tolerance.
- **Accessibility tests:** focus traversal, keyboard activation, reduced-motion bypass — each verified with synthetic events through the `app.Window` test harness.

---

## Bridging FRP and MVU

Components in the design system must be usable by both pattern families. The bridge is event-shaped, not state-shaped.

### Component contract

Every interactive Prism component takes:
1. An `rx.Observable[Theme]` for visual configuration.
2. Component-specific `rx.Observable[State]` inputs (e.g., a `Button`'s `disabled` flag).
3. A callback or message-emitting handle that produces events.

The event handle has two flavours:
- **Direct callback** — `OnClick func()` for simple cases.
- **MVU `Message`** — for MVU consumers, the component emits via `MessageOp` (Gio op embedded via the widget protocol). The MVU runtime reads them out of the ops buffer (`mvu/window.go:74-78`) and delivers them to `Update`.

FRP consumers who want messages-as-stream wrap the callback themselves with `rx.NewSubject[T]()`. MVU consumers use the `MessageOp` flavour. **One component, two integration paths, no duplication.**

---

## Markdown

Document-grade markdown rendering (decision recorded 2026-07-20). The renderer is split across the layers the phase model already defines, rather than landing as one package:

1. **`prism/richtext`** — the inline styled-text primitive: a span model (font, weight, style, size, colour, link URL metadata) with wrapped paragraph layout and interactive link spans. Zero third-party dependencies, themed via `tokens`, full a11y (keyboard focus traversal, visible focus ring, hover pointer cursor).
2. **`github.com/vibrantgio/markdown`** — a standalone module that carries the goldmark dependency (chroma syntax highlighting lives in a `markdown/highlight` subpackage so the core package never imports chroma). It walks the goldmark AST into a block model and maps it onto prism block widgets: type-scale headings, richtext paragraphs, nested lists, blockquotes, rules, code blocks, GFM tables / strikethrough / task lists.
3. **No cadence wrapper** — a deliberate non-goal for now. Cadence patterns are dependency-free compositions of prism primitives; a docs-page wrapper only earns its place once sitedocs proves the shape.

**Rationale:**

- **Dependency hygiene.** Prism's require list stays first-party + rx + Gio. The parser (goldmark) and highlighter (chroma) are quarantined in their own module, and the highlighter is further quarantined in a subpackage.
- **Layer fit.** A document renderer is a composite of prism primitives — the shape the phase model assigns to a downstream module, not to the component foundation.
- **Tag churn.** Early iteration on the renderer bumps only its consumers (sitedocs, mindchat), not every prism consumer.

**Consumers:** sitedocs docs pages (full documents from embedded `.md` sources) and mindchat message bodies (the inline subset plus fenced code blocks).

**Evidence — `gioui.org/x/markdown` evaluation (2026-07-20):** the existing community renderer was evaluated and rejected as a dependency. It flattens the whole document into a single richtext flow and drops blockquotes, thematic rules, images, tables, list nesting, and all of GFM; headings are distinguished by size only; tabs render as tofu; there is no text selection. It remains usable as a span-model reference only — which is what motivates both the prism-owned span primitive and the standalone renderer module.

---

## Implementation Plan

### Phase −1 — Gio Migration *(gates Phase 00)*

The workspace is pinned to `gioui.org v0.1.0` (mid-2023). Current Gio is v0.9.x. The intervening period crosses the major event-API rework (the `Events()` channel was replaced by `app.Window.Event()`, `pointer.InputOp` was replaced by `event.Op` + `gtx.Source.Event(filter)`, `op.InvalidateOp{}` was replaced by `gtx.Execute(op.InvalidateCmd{})`) and improved the accessibility, text shaping, and GPU paths.

Migrating now — **before** any Phase 00 experiment — is correct because:

1. **The Phase 00 experiments must run on the target Gio.** If `KeyedDefer`, many-entity animation, and coordination patterns are validated on v0.1.0 and Phase 1 is built on v0.9, the experiments' findings may not transfer. The whole point of the experiments is to commit only after evidence; that evidence has to be on the same substrate.
2. **The `unsafe.Pointer` MessageOp hack repays itself here.** The new event API exposes message tags as a first-class event source. The hack disappears, not because we work around it but because the new API makes it unnecessary.
3. **Building Prism on v0.1.0 then migrating later means rewriting Prism.** Same reasoning that justifies Phase 0 over rewriting components for observables — pay the migration cost once, on the smaller surface, before the surface grows.
4. **Accessibility commits in Phase 1 require the modern Gio.** The a11y improvements landed in the rework. v0.1.0 cannot honour Phase 1's a11y promises.
5. **The migration validates architectural portability.** If `WithLatestFrom2`, `Defer`-scoped state, and frame-driven physics survive the API rework, the architecture is confirmed independent of any one Gio version. If they don't, we've found a deeper fragility worth knowing about *before* committing to it.

**Deliverables:**
- Update every `go.mod` in the workspace to current Gio.
- Rewrite `mvu/window.go` against the new event loop. The `WithLatestFrom2` synchronisation model is preserved; the wiring around it is rebuilt. The events observable now sources from `window.Event()` instead of `window.Events()`.
- Replace the `unsafe.Pointer` MessageOp extraction with a router-based message tag and `gtx.Source.Event(filter)` read.
- Update every example and reference app — coinviz, appviz, todos, mindchat, traer's Gio examples, seen's Gio examples — to compile and run on the new API.
- Capture the **performance baseline** on the migrated coinviz (per-pane `widget(gtx)` cost, allocation rate, frame timing) — this becomes the reference for the Performance methodology section.
- Document any architectural surprises in a `MIGRATION.md` postmortem alongside this design doc.

**Risk gate:** if migration uncovers that any of the five core architectural patterns no longer holds cleanly on the new API, the design document is revised before Phase 00 begins. We do not paper over a broken pattern with workarounds.

**Estimate:** the migration is the largest contiguous block of mechanical work in the entire plan. Probably days to a couple of weeks for one person, depending on how many examples need touching. The gain is foundational; this is the right place to spend it.

---

### Phase 00 — Validation Experiments *(gates Phase 0)*

Before any Phase 0 contract is written, three prototypes must succeed. They probe the fundamental limits documented in *Architectural Limits & Required Experiments* below. The shape of Phase 0's API depends on what these prototypes find.

**Experiment A — Keyed identity ("`KeyedDefer`")**
- *Probes:* concern #1 (component identity / reconciliation)
- *Build:* a reorderable todo list where each row holds `Defer`-scoped state (an inline edit field, an animation phase). Reordering must preserve per-row state; insertion/deletion must animate.
- *Decide:* whether `rx.KeyedDefer[K, V](key K, factory ...)` (or equivalent) feels natural in the FRP style. If yes, it becomes a Phase 0 primitive. If no, dynamic-list use cases are steered to MVU and the FRP pattern is documented as appropriate only for static composition.

**Experiment B — Many-entity animation**
- *Probes:* concern #2 (window-global invalidation cost)
- *Build:* a force-directed graph with 200+ nodes using `traer`, animating continuously at 60 FPS. Measure layout-pass cost and allocation rate.
- *Decide:* whether per-widget op caching (recording ops once, replaying when inputs unchanged) is sufficient, or whether Phase 3 needs a dedicated "scene" abstraction that bypasses the standard layout path. Output: a documented pattern for animation-heavy widgets.

**Experiment C — Coordination context**
- *Probes:* concern #3 (cross-widget coordination)
- *Build:* a kanban-style board with drag-and-drop between columns, plus modal stacking, plus tooltip arbitration. The drag must communicate hover state to drop targets in real time.
- *Decide:* the general-purpose coordination primitive. Probably an `rx.Subject` injected via layer or context, but the exact shape needs to be discovered. Output: a Phase 1 `prism.Coordination` package.

**Pre-Phase 0 commitments**
- Decide concern #4 (undo/redo): commit to "FRP apps don't undo, MVU apps do" *or* design a `MessageOp`-replay buffer. Document the choice.
- Decide concern #5 (plugin systems): commit to "no extensibility" as a non-goal *or* plan a scripting layer. Document the choice.

Phase 0 begins only after these experiments produce documented decisions. Their findings retroactively shape the token contract — for example, if `KeyedDefer` is adopted, the theme observable contract may need keyed variants for per-instance theming.

---

### Phase 0 — Token & Theme Contract

**Goal:** define the data shapes and observable interfaces that every later phase consumes. No components, no rendering — just types.

**Deliverables:**
- `prism/tokens/{colors,spacing,radius,elevation,motion}.go` — typed Go constants and structs.
- `prism/theme/theme.go` — `type Theme struct { Color rx.Observable[ColorTokens]; Type rx.Observable[TypeScale]; Motion rx.Observable[MotionTokens]; … }`.
- `prism/theme/auto.go` — `AutoLightDark` (extracted from coinviz) producing `rx.Observable[Theme]`.
- `prism/a11y/preferences.go` — `rx.Observable[A11yPrefs]` (reduced motion, high contrast, increased text size) backed by OS APIs.

**Why this exists:** building components against static tokens then wrapping in observables later forces a rewrite. Defining the observable contract first means components are observable-aware from birth, with no v1 → v2 migration. This is the single biggest sequencing decision in the plan.

**Conditional deliverables (set by Phase 00 outcomes):**
- If Experiment A succeeds: add `prism/keyed/` (`KeyedDefer[K, V]`) as a token-contract primitive, and extend the theme contract with optional keyed variants for per-instance theming.
- If Experiment B yields an op-cache pattern: add `prism/cache/` describing the standard frame-cache contract that animation-heavy widgets implement.
- If Experiment C produces a coordination primitive: its types live in `prism/coordination/` (Phase 1 module) but its observable shape is fixed here in the contract.

These are explicit hooks rather than open-ended scope creep: each deliverable lands only if its experiment succeeds. If an experiment finds its premise is wrong, the corresponding deliverable is dropped and the limit is documented as architectural.

---

### Phase 1 — Prism (component foundation)

**Goal:** a useable widget catalogue against the Phase 0 contract. Apps stop reinventing buttons, theming, and layout spacing.

**Module:** `vibrantgio/prism`

```
prism/
  tokens/      — from Phase 0
  theme/       — from Phase 0
  a11y/        — from Phase 0
  button/      — Button (hover, focus, press; Defer-scoped state; a11y; MessageOp + callback)
  input/       — TextField, Checkbox, Radio, Dropdown
  list/        — virtual scrolling list (generalised from todos/list.go)
  icon/        — unified SVG + IVG icon registry
  layout/      — spacing helpers, FocusGroup, basic flex/grid wrappers
  initial/     — `Initial[T]` helper to replace ad-hoc first-frame sentinels
  coordination/ — output of Experiment C: scoped subjects, drop-zone registry,
                  modal stack, tooltip arbitration, gesture disambiguation
  gallery/     — demonstration app exercising every Prism primitive (Material
                  Catalog / TailwindUI equivalent), used as a manual test bed and
                  as the canonical reference for component usage
```

**Migration path:** coinviz's `theme/theme.go` is the source for token values; its existing struct is sliced into the typed token modules. The four reference apps migrate one at a time and on different tracks because their patterns differ:
- **coinviz** (pure FRP) — replace its bespoke `theme.Theme` with consumption of `rx.Observable[prism.Theme]`. Most of the work is mechanical renaming; the architecture is already shaped for it.
- **appviz** (FRP) — same shape as coinviz.
- **todos** and **mindchat** (MVU) — adopt Prism components via the `MessageOp` callback path. Each component swap replaces a bespoke widget and exercises the FRP/MVU bridge.

The `gallery/` app is the canonical *forward-looking* reference: every primitive, every variant, every a11y mode demonstrated in one place. Migrated apps demonstrate that Prism survives contact with real codebases; the gallery demonstrates Prism's intended API.

---

### Phase 2 — Spectrum (reactive theme runtime)

**Goal:** make the theme stream a deeply integrated runtime concept, not just a contract. Includes user-preference persistence, system-event bridging (dark mode, reduce motion, accent colour), animated theme transitions, and per-window theme overrides.

This phase is where the *runtime* of reactive theming lives. Phase 0 defined the contract; Phase 2 (Spectrum) implements the behaviours that contract enables. The name reflects the prism→spectrum metaphor: Prism refracts into a Spectrum of theme tokens.

---

### Phase 3 — Pulse (visual effects layer)

**Goal:** the vibrancy. Spring physics, glow, depth, motion — layered on top of Prism components.

```
pulse/
  tween/        — simple `Tween[T]` for non-physical motion (fades, slides, colour)
  spring/       — physics-based motion via traer; Defer-scoped ParticleSystem
  glow/         — luminance halos via gradient composition
  depth/        — elevation-driven shadow layers
  motion/       — enter/exit/transition primitives
  conductor/    — shared clock for coordinated animation across widgets
```

**Composition mechanism (concrete):** Phase 3 widgets are *variants* exported alongside their Prism counterparts, not magic decorators.

```go
// Without Pulse:
prism.Button(theme, ButtonProps{Label: "Save", OnClick: save})

// With Pulse:
pulse.SpringButton(theme, ButtonProps{Label: "Save", OnClick: save},
    pulse.SpringOptions{Stiffness: 0.4, Damping: 0.7})
```

`SpringButton` internally embeds a `prism.Button` and wraps it with physics-driven press/release. This keeps the API explicit, the implementation reusable, and the dependency direction clean (Phase 3 → Phase 1, never the reverse).

---

### Phase 4 — Cadence (pattern library)

**Goal:** the Vibrant Gio equivalent of TailwindPlus or Bootstrap — a curated library of prebuilt application patterns composed from Prism primitives. Reduces time-to-build for common desktop UI shapes.

**Why this exists:** Bootstrap and Tailwind succeeded because they reduced time-to-build for common patterns. Material Web, Fluent, and similar exist for web/Microsoft platforms; no equivalent exists for native cross-platform desktop on Go. This phase fills that gap.

**Module:** `vibrantgio/cadence`

```
cadence/
  card/         — content cards with header/body/footer slots
  alert/        — info / success / warning / error banners
  modal/        — dialog with backdrop, focus trap, escape handling
  popover/      — anchored floating content
  tooltip/      — small hover/focus annotations
  navbar/       — top navigation with branding, links, actions
  sidebar/      — collapsible side navigation
  shell/        — application shells (sidebar+header+main, split-pane, etc.)
  table/        — sortable, filterable, virtualised tables
  pagination/   — page navigation controls
  breadcrumb/   — hierarchical location indicators
  tabs/         — tab strip + content panel
  accordion/    — collapsible section groups
  toast/        — transient notification stack
  marketing/    — hero, pricing, feature, testimonial sections (for app landing/onboarding)
```

**Composition contract:** every pattern is exported as a callable Go function consuming a Prism theme observable. The same source is also documented as copy-paste-friendly — users can call it as a library *or* copy the source into their own app and modify it. This is the shadcn/ui model adapted to Go: the code is the spec, no opaque runtime configuration. Gio is a type-safe environment; a fork-and-modify path costs nothing and respects the user's autonomy.

**Token alignment:** Phase 0's token scale should align with Tailwind's well-considered values (4-pt spacing, 50–950 colour shades, harmonic type ramp). Adopting Tailwind's scale wholesale saves months of bikeshedding and gives the patterns module a familiar foundation for any developer who has touched Tailwind.

**Dependencies:**
- Requires Phase 1 (Prism primitives).
- Modal, popover, tooltip, toast all depend on **Experiment C's coordination primitive** (concern #3). Until Experiment C lands, these patterns can be built but will not compose well across each other.
- Table and virtualised list patterns may benefit from Experiment A's `KeyedDefer` for per-row state preservation under sort/filter.

**Sequencing:** Phases 2, 3, and 4 are parallelisable once Phase 1 lands. Phase 4 has the highest external visibility (it is what most developers will see first when evaluating the design system) and likely deserves priority once Prism is stable.

---

## Architectural Limits & Required Experiments

This section catalogues the limits of the architecture as currently designed. It is honest about where the FRP pattern fights certain app shapes. The Phase 00 experiments above are derived from this catalogue; the entries here exist so the reasoning survives the experiments.

### Fundamental limits *(constrain which apps are practical)*

#### 1. No component identity / reconciliation
`rx.Map(items, ...)` produces fresh `layout.Widget` closures on every emission. There is no notion of "this widget is the same instance as last time." Per-item `Defer` state effectively re-binds when the list re-emits.

**App types impacted:**
- Reorderable lists with per-row state (drag-to-reorder, inline edit)
- Animated insertion / deletion / reordering
- Tree views with expansion state
- Tab systems where each tab has its own scroll position

**Mitigation:** Experiment A (`KeyedDefer`). React's `key` prop and Flutter's element-tree diffing solve this elsewhere; we need an FRP-shaped equivalent.

#### 2. Window-global invalidation
`op.InvalidateOp` invalidates the whole window. One animated widget → every widget's layout closure re-runs. Linear cost in widget count per frame.

**App types impacted:**
- Games / game-like UIs
- Real-time scientific visualisation (particle simulations, large 3D scenes)
- Network / graph visualisation with continuous physics
- Live audio waveform displays with many channels

**Mitigation:** Experiment B. Likely an op-caching pattern at the widget level: record ops once, replay when inputs unchanged. Possibly a Phase 3 scene primitive that bypasses the standard layout path entirely.

#### 3. Coordination ceiling
`Defer`-scoped state is great in isolation but offers no mechanism for widgets to coordinate. Every cross-widget concern (focus traversal, drop zones, modal stacking, shared scroll, tooltip arbitration, gesture disambiguation) is a one-off design problem.

**App types impacted:**
- Design tools (Figma-style)
- IDEs with linked editor / minimap / outline
- Multi-pane layouts with synchronised scroll
- Drag-and-drop-heavy apps (file managers, kanban, node editors)

**Mitigation:** Experiment C. Likely a `prism.Coordination` package providing scoped subjects and context-passed coordination state.

#### 4. No undo/redo story for FRP apps
MVU apps get undo for free (snapshot the `Model`). Pure FRP apps in coinviz style scatter state across many `Defer` closures and observable accumulators — there is no single snapshottable "current state."

**App types impacted:** code editors, document editors, drawing tools, CAD/modelling apps — anything where users iterate experimentally.

**Mitigation:** Pre-Phase 0 decision. Either commit to "apps needing undo use MVU" (clear, simple) or design a `MessageOp`-replay buffer that captures all interaction events for later replay. **Document the choice; do not leave it implicit.**

#### 5. Plugin / extension systems
Go's plugin support is poor cross-platform. The architecture compiles all components statically. Apps with third-party extensions (VS Code, Obsidian, DAWs with VSTs) need a scripting layer — Tengo, Starlark, WASM, JS via goja.

**App types ruled out without a scripting decision:** IDEs with extensions, DAWs, apps with user macros/automation.

**Mitigation:** Pre-Phase 0 decision. Either commit to "no extensibility" as a non-goal, or plan a scripting integration as a future module.

### Solvable gaps *(extendable within the philosophy)*

These do not change the architecture — they are documentation or library work that can land when needed.

#### 6. Multi-window shared state
Pattern: an upstream `rx.Subject` shared across multiple `Window.Render` calls. Works today; needs a documented recipe. Photoshop-style floating palettes, multi-document editors fit here.

#### 7. OS integration beyond the window
System tray, native menu bar (especially macOS), notifications, dock badges, file watchers, system shortcuts, deep linking. Gio support varies by platform; the design system has no abstractions yet. Apps that need to feel deeply native to one platform will hit this wall — flag per-platform what is feasible.

#### 8. Async work with progress UI
Long-running CPU work (image decode, compilation, ML inference) with cancellation and progress reporting. MVU's command pattern handles it; FRP needs a documented pattern (likely `rx.NewSubject[Progress]` per task).

#### 9. Forms with cross-field validation
Cross-field rules, async validation against a server, multi-step wizards, dirty tracking. Combinator patterns over input observables; needs a Phase 1 forms guide.

#### 10. Streaming media
Video playback, audio editing, real-time camera. Gio renders frames, platform APIs decode/capture. Not impossible, not addressed. Acknowledge as a future module or non-goal.

### Inherited Gio limits *(out of Vibrant Gio's hands)*

- **Custom-drawn chrome looks "almost native" but never quite.** macOS users notice. Apps requiring full Cocoa fidelity need a different substrate.
- **Accessibility on macOS lags VoiceOver expectations.** Compliance-sensitive apps (government, education, enterprise procurement) may not pass.
- **Printing.** Effectively absent.
- **Native input methods (CJK IME).** Works but with rough edges.

### Where the architecture is excellent

To balance the above — the architecture is genuinely strong for:
- Real-time data visualisation (proven by coinviz)
- Dashboards, monitoring, observability UIs
- Chat and messaging (mindchat)
- Productivity tools with bounded UI complexity (todos)
- Custom finance, trading, scientific instruments
- Anything where reactive theming and visual freshness are differentiators

These app types should be the focus of demonstration projects in Phase 1. Apps from the "limits" list become viable as the experiments produce mitigations.

---

## Naming Considerations

**Decided.** All four phase names are committed: **Prism** (foundation), **Spectrum** (theme), **Pulse** (effects), **Cadence** (patterns). The progression is thematically tight — Prism refracts into a Spectrum, which Pulses with motion, composed in Cadence. Module rename cost was zero today, infinite once published; names locked before Phase 2 implementation begins.

| Phase | Name | Rejected working name + original candidates | Why this name |
|---|---|---|---|
| 1 | **Prism** | — | Refraction (one source → many components) is on-thesis, memorable, not crowded. |
| 2 | **Spectrum** | Working name "Aura" rejected: crowded namespace (AuraJS, Project Aura, wellness/spiritual associations); doesn't suggest "reactive observable theme stream". Original candidates: Pulse, Cadence, Resonance. | Outside the original set: makes the prism→spectrum metaphor literal — Prism refracts into a Spectrum of theme tokens. Tradeoff: Adobe Spectrum is a known web design system, accepted because it sits in a different ecosystem (web vs native Go/Gio). |
| 3 | **Pulse** | Working name "Nimbus" rejected: a nimbus is a *static* halo/cloud — opposite of motion; crowded namespace (NimbusJS, Nimbus Note, Nimbus Sans). Original candidates: Verve, Vivace, Kinesis. | Repurposed from a Phase 2 candidate: the rhythm/oscillation metaphor fits motion (animations, springs, glows pulse) more naturally than theme emissions. |
| 4 | **Cadence** | No working name. Original candidates: Folio, Atelier, Suite. | Reassigned from a Phase 2 candidate: "rhythm of composition" reads more naturally for a pattern library than for theme emission. Patterns are how components flow together; cadence in music is the compositional building block. |

---

## Design Principles

- **Observable-native:** components participate in the FRP graph from birth, not retrofitted
- **Single-threaded UI:** the events observable owns the UI thread; everything else beats to its rhythm
- **Defer for interaction state:** mutable state lives in `rx.Defer` closures, never in subjects
- **Frame-driven motion:** animated components self-schedule via `op.InvalidateOp`; idle when settled
- **Progressive enhancement is explicit:** Phase 3 widgets are *variants* of Phase 1 widgets, not silent decorators
- **Accessibility is non-optional:** every interactive Prism component supports keyboard, focus, screen reader, reduced motion, contrast
- **No string tokens:** all design values are typed Go values
- **Module boundaries:** utilities never depend on applications; later phases never depend on later phases (only earlier)
- **Performance is measured, not assumed:** every component has a benchmark; CI rejects regressions
