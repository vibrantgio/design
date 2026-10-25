# Vibrant Gio Design System — Architecture

> **Status:** This document describes the system as shipped. It is for
> people and agents working **on** the design system. The guide for building
> applications **with** it is
> [`llms.txt`](https://raw.githubusercontent.com/vibrantgio/workbench/master/llms.txt)
> in the `workbench` repository: the bootstrap skeleton, the token
> vocabulary, the component catalogue and the pitfalls. What that guide
> says is referred to here, not repeated.

---

## What Vibrant Gio is

Vibrant Gio is a design system for native desktop applications on macOS,
Windows and Linux, built on [gioui.org](https://gioui.org). An application
built with it looks like a macOS application: its colours, control heights,
text sizes and shapes are the platform's, measured off the platform's own
windows where a capture holds them and taken from its published guidelines
otherwise.

It is one design system across nineteen library repositories, the
`workbench` repository with its example applications, this `design`
repository with the published token bundle, and the `organisation`
repository with the agent guide and the gate scripts. One theme observable carries
the whole look — colour, typography, density, levels and motion — and
every library module sits in an enforced tier.

The application model is Functional Reactive Programming with
`reactivego/rx` and `vibrantgio/mvu`. Components take the theme observable
and need nothing else. Two styles of application use the same components:
pure FRP composition, and an explicit MVU (Model–View–Update) state machine
(§Key architectural patterns).

### Non-goals

- **No web target.** Tokens are exported to CSS for prototyping in Claude
  Design, and those pages show components, not applications.
- **No mobile target.** Gio supports Android and iOS; Vibrant Gio is made
  for desktop input, window chrome and platform integration.
- **No embedded or kiosk target.** No memory or binary-size budgets.
- **No CSS-like dynamic styling.** Tokens are typed Go values, not maps of
  strings.
- **No picked palettes and no derived ones.** Colour is the platform's,
  read off the platform per scheme; a lint fails `go test` on colour
  literals in the component repositories.

---

## The layering

The design-system spine, foundation first:

```
mvu  →  theme  →  components  →  effects  →  patterns, markdown
```

`theme` — the theme runtime and every design token — is the foundation,
below the components it themes. Tokens flow up; nothing flows down.

The tier table judges all nineteen library modules. A module may import
only modules in a strictly lower tier, plus anything in the support row:

| Tier | Modules | May import |
| --- | --- | --- |
| 0 | mvu, font, style, textdraw, backdrop, gradient, circle | support libraries only |
| 1 | theme | tier 0 |
| 2 | components | tiers 0–1 |
| 3 | effects | tiers 0–2 |
| 4 | patterns, markdown | tiers 0–3 |
| — | ivg, svg, seen, csg, kiwi, noise, traer | nothing in this table |

How to read it:

- **`font` is tier 0** because theme takes its default faces, Roboto and
  Roboto Mono, from it.
- **Two edges inside tier 0 are admitted:** `style` imports `font` and
  `textdraw`. `style` is a frozen type scale; the check admits those two
  edges rather than renumbering the tier.
- **The support libraries** never import the design system.
- **Nested demo modules may import upward; their parents may not.**
  `components/gallery` and `mvu/example` are separate modules that may
  import tiers above their parent, and a parent that imports its own nested
  module fails the check. This keeps a demo from closing a dependency
  cycle.
- **Nested adapter modules** (`ivg/raster/gio`, `svg/driver/*`, `kiwi/gio`,
  `traer/gio`, `seen/context/gio`) stand outside the table.
- **Applications are at the top of the stack:** the workbench's
  application modules, its root module and `design` may import anything,
  and no library imports them.

The table is enforced: `organisation/scripts/check-layers.sh` walks
`go list -deps` for every module and fails on any edge the table does not
permit.

---

## The colour model

Colour is **read off the platform, never derived**. `tokens.PlatformColors`
carries AppKit's semantic colours under their own names, one field per
name in Go casing, with the value that name reports for one appearance.
`tokens.PlatformLight` and `tokens.PlatformDark` are the two recorded sets,
and the theme observable carries the current one to every component.

### The names are the platform's

Forty-four fields carry AppKit's names. Forty-three were read off AppKit
with a command-line program under both appearances, converted to sRGB with
alpha kept, and are stored with that program in the reference; a test pins
both sets against the stored catalogue. The forty-fourth, `FindHighlight`,
is measured off Mail's find bar, because AppKit's own value is not what the
platform's applications paint. Twenty-eight more are fills the platform
draws without giving them a name — among them the sidebar material, the
grouped box, the push button's fill, the hover and press overlays, the
floating shadow, the field's hairline, the overlay scrollbar's thumb, the
alternating row, the dim behind a sheet and the toolbar's controls. They
were measured off stored captures of the platform's windows — the
alternating row read off the array AppKit answers with and confirmed
against a capture — and each carries its provenance beside its value. No
colour is generated from another. Where the platform's published guideline and what
it draws differ, what it draws wins.

### A coverage is part of the value

Text, a separator, a disabled control's text, the focus ring, the hover and
press overlays and the shadow are each a colour at a coverage: they
composite over whatever is beneath, which is how one recorded value is
right on every fill. The platform composites in **encoded sRGB**, and Gio's
rasterizer composites in linear light, so a component flattens the colour
onto the fill it is drawn on (`theme/color.Flatten`) and hands Gio an opaque
colour. Which fill that is, is the one thing a component must be told: a
caller that puts it on a card, a selected row or a coloured fill says so
through the component's `Surface` property.

An overlay with no one fill beneath it — a scrim over a page — is
composited against what is already painted: the frame so far is rendered
offscreen and the overlay flattened over every pixel
(`components/composite`). The floating shadow's gradient takes the coverage
fitted to Gio's own blend (`theme/color.LinearCoverage`).

### The accent

`ControlAccent` is the accent the user chose, and seven more fields follow
it: four AppKit names — the emphasized selection, the text selection, the
selected control and the focus ring — and three measured fills — the
sidebar's selection, the sidebar's unemphasized selected text and the
default button's fill. On macOS `theme/system` asks AppKit for the four
names, so the platform's own value arrives. `PlatformColors.WithAccent`
rebuilds all seven: it keeps the accent's hue and saturation and takes
each recorded field's own lightness and alpha. It is the rule on Windows and
Linux, whose desktops publish an accent but no colour set, and on every
platform when a theme colour is chosen; on macOS with no theme colour
chosen, the three measured fills keep their recorded values. The theme colour — the platform's accent unless the user chooses
one in the themer — is used wherever the platform uses its accent, and
nothing else is computed from it.

### States are the platform's

Hover and press are the measured overlays, laid over a control's own fill,
or over the surface beneath where the control has none. Selected is the selection colours, emphasized
in the active window and the inactive grey otherwise. The default action is
the accent under `AlternateSelectedControlText`. Disabled is
`DisabledControlText`.

### Contrast

APCA lightness contrast, Lc, is the one measure; no second ratio is computed
or reported. A pair the platform paints — white on `SystemGreen`, say — is
kept as the platform paints it. A fill the user chose takes as its
foreground whichever of black or white reads better on it by that measure
(`theme/color.BestOn`).

---

## Levels

A window has five levels — backdrop, chrome, content, raised and floating —
and each level's fill is the platform's, read off the platform per scheme
and never derived from another level's. No level is lighter or darker than
another by rule.

| Level | Holds | Fill |
| --- | --- | --- |
| backdrop | the window's own surface, wherever nothing else is | `WindowBackground` |
| chrome | sidebars, toolbars, navbars, inspectors, status bars, panes | the measured sidebar material |
| content | the document being read, lists, tables | `TextBackground` or `ControlBackground` |
| raised | cards, fields and insets on the content | the measured grouped box; a field the platform's field |
| floating | dialogs, toasts, menus, popovers, tooltips | `WindowBackground`, under the platform's shadow |

A raised thing is told by the box's small step of fill, with no hairline
and no shadow. A floating thing is told by its shadow: black at a measured
coverage of 0.075 at the surface's edge, falling linearly to nothing 24 px
out. `effects/depth` draws it, as eight gradient tiles and one interior
fill, nine paint operations per shadow; a step of fill is one. The platform
draws two more measured shadows, around an inset sidebar pane and under a
bordered toolbar control, and both are tokens of `PlatformColors` beside the
floating one.

---

## Density

`tokens.Density` carries the heights and the inner padding a control is
drawn at, set once for the window and never per component. Comfortable is
the platform's regular control and the default; Compact is its small one.
Switching an application to Compact is a theme change.

| Height | Comfortable | Compact |
| --- | --- | --- |
| control | 24 dp, measured | 19 dp, the platform's published small control |
| text field | 27 dp, measured | 21 dp |
| list and table row | 20 dp, measured | 19 dp |
| checkbox row | 22 dp, measured | 17 dp |
| toolbar control | 36 dp, measured | 36 dp |

The control and field heights are the least a control is drawn at: a
control is `max(height, line box + 2 × PaddingY)` tall, so a Comfortable
button draws 24 dp, a Compact button 20 dp, a Comfortable text field 28 dp
and a Compact one 24 dp. Rows and the toolbar control are drawn at exactly
their height. Each number's provenance is in the doc comment of
`theme/tokens/density.go`.

---

## Motion

`tokens.MotionScale` carries five duration stops (50, 150, 250, 400 and
500 ms), the standard and emphasized easing families as cubic-bezier
control points, and three spring presets (mass, stiffness and damping).
Components that animate over a duration take it from the theme's Motion
observable, which is what makes reduced motion free (§Accessibility).
`effects/springbutton` does not: it uses spring constants of its own,
steps its spring once per frame at a fixed step of 1/60 s, so it runs
faster on a faster display, and does not read the theme's Motion.

---

## Typography

**The theme owns the typeface.** `tokens.Typography` carries one
`TextStyle` — typeface, weight, size, line height, tracking — per
typography role: Display, Headline, Title, Label and Body, each in Large,
Medium and Small; plus `Code`, one monospaced size; plus `DocumentHeadings`,
the six heading steps of a prose document, derived from Body. A role's size
and weight are the platform's for the text the role sets. Roboto is the
default typeface because the default `Typography` names it, and Roboto Mono
sets `Code`.

`Typography` also carries the face collection and builds two cached
shapers: `Shaper()`, which falls back to the platform's own fonts for
anything the faces lack and is what applications draw with, and
`DeterministicShaper()`, the faces and nothing else, which golden tests draw
with. A lint in components, effects, patterns and markdown fails `go test`
on any import of `gioui.org/font/gofont`, and a second lint does the same
for `color.NRGBA` literals.

---

## Accessibility

- **Contrast is measured in APCA terms** (§Contrast).
- **Increased contrast follows the platform on macOS:** when the
  preference is on, AppKit reports the colours it paints under it and the
  theme carries them.
- **Reduced motion snaps:** while the preference is on, the Motion
  observable emits zero durations, and components driven by durations
  complete in zero frames.
- **Keyboard:** every interactive component takes part in focus and
  keyboard activation.

Applications do nothing to get any of this; it arrives through the theme.

---

## Blur

Gio has no blur primitive and no custom shaders, but `gioui.org/gpu/headless`
renders an op list offscreen and reads the pixels back. `effects/blur`
builds a backdrop blur on that: three box blurs approximating a Gaussian,
run in parallel across the CPUs, with only small per-call worker state
allocated in steady state. `blur.Backdrop` takes the
look as a sigma in full-resolution pixels and renders at a reduced
resolution chosen from it (`DefaultDivisor`; a divisor of 4 at sigma 8),
which cuts render, readback and blur work by the divisor squared and loses
nothing the blur would keep.

Three rules follow:

1. **Refresh on content change.** The pipeline runs synchronously on the
   events goroutine, driven by a change in what is behind it — a dialog
   opened, a scroll settled — never every frame. Painting the cached result
   is free.
2. **Fallback.** Where headless rendering is not available the caller paints
   a flat scrim of one colour (`blur.FallbackOp`), never a crash.
3. **No animated blur glow.** `effects/glow` draws its glow with gradients:
   an animating blur glow would cost 0.2–0.8 ms of events-goroutine time plus
   an image-sized allocation and upload per glow per frame, against about
   0.5 µs for the gradients.

---

## The export

`theme/export` serialises a theme emission, and `theme/cmd/vg-tokens` writes
it into this repository:

- **`theme.json`** — the parameters the theme is reproduced from: the theme
  colour, the platform's whole set per appearance, the faces, both density
  settings and the motion scale.
- **`styles.css`** — the token sheet: the platform's colour set as
  `--platform-<name>` in both schemes, a coverage written as `#rrggbbaa`,
  beside `--font-*`, `--space-*`, `--radius-*`, `--density-*`,
  `--ease-*` and `--duration-*`; then the component class layer, in which
  every colour is a `--platform-` name, the same mapping the Gio components
  take. The browser composites `#rrggbbaa` in encoded sRGB, as the platform
  does and as `theme/color.Flatten` does, so a rule names the platform's
  colour and the browser flattens it.
- **`readme.md`** — the token families and the class layer, generated from
  the same tables the sheet is.
- **The pages** — `foundations/color.html`, `type.html` and `layout.html`,
  generated, and six hand-authored component pages under `components/`.

A round-trip test parses the sheet back and asserts every value against the
Go token it came from. `design/mirror` renders the component pages in a
headless browser and scores them against the Gio components' golden images.
`organisation/scripts/push-design.sh` regenerates the bundle and uploads it
to Claude Design, where look-and-feel questions are tried in a browser.

---

## The applications

The `workbench` repository holds nine example applications — feeds,
iconbrowser, marketing, mindchat, sitedocs, sk150, themer, todos and
vaultview — each a module of its own in a subdirectory, and at its root the
workbench: one window listing those applications, each with a control that
starts it. The library is proved against them; `llms.txt` describes what
each one shows.

---

## Key architectural patterns

### 1. The events goroutine is the heartbeat

Gio reads events and calls `Frame()` on one goroutine. `mvu.Window.Render`
subscribes the `CombineLatest` of all layers on an rx goroutine, stores each
result as an atomic snapshot and invalidates the window; the events
goroutine reads the snapshot when the frame event arrives and lays out from
it. Heavy upstream work runs freely on rx goroutines, values reach rendering
only at a frame boundary, and everything that touches Gio ops runs on one
goroutine. The frame loop leads the FRP graph rather than taking part in it.

### 2. Interaction state is in `rx.Defer` closures

State allocated inside an `rx.Defer` factory is created once per
subscription and captured by reference in the map functions and
`layout.Widget` closures below it. It is read and written only on the events
goroutine, so it needs no locks, and it is never handed to another
goroutine or stored in a subject. The scopes: the Defer closure (once per
subscription), the map closure (per emission), the `layout.Widget` closure
(per frame). `components/keyed` hands back the same per-key state across
reorder, insertion and deletion, and `components/initial` is the
first-frame sentinel.

### 3. `MessageOp` connects components to MVU

Components emit messages by adding `mvu.MessageOp{Message: …}` to the ops;
the runtime collects them during the frame and delivers them to `Update`.
MVU applications use `MessageOp`; FRP applications pass plain callbacks.
One component serves both.

### 4. Animation schedules itself and idles

Animated components step their simulation inside the frame and request the
next frame only while active (`gtx.Execute(op.InvalidateCmd{})`); when they
settle nothing invalidates and Gio goes idle. Invalidation is window-wide,
so expensive components cache their ops while their inputs are unchanged
(`components/cache`). The packages of `effects` are components of their own
that draw the library's components with their effect —
`effects/springbutton` draws `components/button` with a spring — chosen per
call site, never applied globally.

---

## Threading and lifecycle

Besides the rules above: messages cross goroutines through a buffered
channel, and MVU updates run on the loop's goroutine, not the events
goroutine. A subscription begins at `Render`, and every `Defer` factory runs
once. It ends when the window closes (`app.DestroyEvent`); a layer that
completes or errors does not end it. A UI is restarted by building new
layers for a new `Render` call.

---

## Design principles

- **One theme carries the whole look:** colour, typography, density, levels
  and motion arrive through the theme observable; an application wires the
  theme once and styles nothing per component.
- **The platform's values, measured:** colours, heights and sizes are read
  off the platform's windows where a capture holds them, and what it draws
  beats what it publishes.
- **Observable-native:** components take part in the FRP graph from the
  start.
- **Single-threaded UI:** the events goroutine owns rendering.
- **Effects are chosen:** an effect is a component of its own, never a
  silent decorator.
- **Accessibility is part of the contract:** keyboard, focus, measured
  contrast and reduced motion.
- **Typed values:** every design value is a typed Go value, and the lints
  fail `go test` on colour literals and on gofont.
- **Module boundaries are enforced:** the tier table is checked.
- **Measured, not assumed:** density, the shadows, the blur economics and
  the glow verdict cite measurements, and golden tests pin what renders.
