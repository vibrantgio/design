# Vibrant Gio — conventions

Vibrant Gio is a CSS-class design system mirroring a Go/Gio desktop component
library. Compose screens from the classes and tokens below, exactly as named —
never invent a class, a token, or a size variant.

## Where the truth lives

- `styles.css` is the only stylesheet: token blocks (`:root` light + scales,
  `.dark` colour override, `.compact` density override) followed by the
  component class layer. Every styled value in every page is a `var(--…)`
  reference; write new markup the same way — no literal colours, sizes or
  radii.
- `theme.json` holds the generative parameters (everything derives from one
  brand seed). `readme.md` documents the token families. Both are generated,
  like `styles.css`, by the Go theme exporter — treat all three as read-only.
- `components/*.html` are the copyable markup reference; `foundations/*.html`
  show every token rendered. The pages are verified pixel-wise against the Gio
  implementation's golden images, so their markup is the idiom to copy.
- Mode and density are class switches: `class="dark"` on the root element,
  `class="compact"` on any subtree. They are orthogonal; never restyle for
  them by hand.

## Class families

- **Buttons** — `.btn` on `<button>`: filled by default (accent under
  on-accent). Emphasis modifiers: `.btn.tonal` (tinted fill), `.btn.ghost`
  (no ground at rest). `.btn.icon` is the square icon-only form (inline SVG
  on `currentColor`). `.selected` marks a toggled-on button; ghosts have no
  selected treatment. There is **no size modifier** — density is the size
  axis.
- **Tags** — `.tag`, `.tag.tonal`: label-small pills. Labels, not controls —
  no interaction states. Status chips are `.tag.success` / `.tag.warning` /
  `.tag.error`: the same level colour the toast carries (the level pin
  tinted over the Surface ground, ringed by the level outline). Compose
  these for any status — a build state, a health badge — and never
  inline-style a status colour.
- **Forms** — native elements, no script: `.input` on a text `<input>`;
  a dropdown is `<select class="input select">` inside a `.select-wrap`
  (which draws the chevron); `.checkbox` and `.radio` on their native input
  types. Disabled is always the native `disabled` attribute.
- **Cards** — `.card` (outlined, elevation-1 fill), `.card.elevated`
  (borderless, elevation-2 fill).
- **Table** — `.table` on a real `<table>`: header band, one-control-height
  rows, divider rules, no zebra. Tables are unframed: the Surface ground and
  the header band *are* the frame, so never wrap a table in a card, an
  outline or a border of your own. A framed table is not in the vocabulary;
  if one is ever wanted it enters the Gio library first, never these pages.
  Sortable headers take `th.sortable` plus `.sort-asc` or `.sort-desc` on
  the active column only.
- **Navigation** — `.navbar` bar with a centred `.navbar-links` row of
  `.navbar-link`s; `.tabs` strip of `.tab`s (`role="tablist"`/`role="tab"`);
  `.sidebar` (modifier `.collapsed`) with `.sidebar-toggle`, `.sidebar-item`
  and `.sidebar-item-icon`; breadcrumbs are `.crumbs` holding `.crumb`s
  separated by `.crumb-sep`, the last (or `.current`) crumb in text colour.
  `.selected` marks the active link, tab or sidebar item.
- **Overlays** — a modal is `.dialog` inside a full-viewport `.scrim`, with
  `.dialog-header` (holding the `.dialog-title`) and, for decision dialogs,
  a right-aligned `.dialog-footer`. Two intents: a **decision** dialog has a
  footer ending in the Return-bound default and **no close X**; a
  dismissable **panel** has a ghost close (`.btn.ghost.icon`) top-right and
  no footer. A corner affordance like that close draws at control height,
  the same as every control — density is the only size knob, and the 44 dp
  accessibility floor is an invisible hit target, never a painted size. Anchored surfaces: `.popover` with a `.popover-tail` on side
  `.top`/`.bottom`/`.left`/`.right`; `.tooltip` (inverse video); `.toast`
  (intents `.success`/`.warning`/`.error`) stacked in a `.toast-stack`.
- **State-forcing twins** — `.is-hover`, `.is-active`, `.is-focus`,
  `.is-checked` share the live `:hover`/`:active`/`:focus-visible`/`:checked`
  declarations, for showing a state statically. In real interactive markup
  rely on the pseudo-classes and leave the twins off.

## Token families

- Ramps: `--color-neutral-100`…`900`, and likewise `--color-primary-*`,
  `--color-secondary-*`, `--color-tertiary-*`, `--color-error-*`,
  `--color-success-*`, `--color-warning-*` (steps in hundreds; 100/200
  grounds, 300 hover/subtle border, 500 strong border, 700 low-contrast
  text, 900 body text).
- Pins and semantics: `--color-bg`, `--color-surface`, `--color-text`,
  `--color-divider`, `--color-accent`/`--color-on-accent`,
  `--color-secondary`/`--color-on-secondary`,
  `--color-tertiary`/`--color-on-tertiary`,
  `--color-error`/`--color-on-error`,
  `--color-success`/`--color-on-success`,
  `--color-warning`/`--color-on-warning`. Solid fills use the pin; their
  hover/pressed stops are emitted as `--color-accent-hover` and
  `--color-accent-pressed` — use those, never a `color-mix()` of your own.
  For grounds, prefer the semantic pins `--color-bg` and `--color-surface`
  over `--color-neutral-*` ramp steps: they render the same today, but the
  semantic survives a theme remap.
- Type: `--font-family`, `--font-family-code`, and per-role
  `--font-<role>-size`/`-line-height`/`-weight`/`-tracking` for roles
  `display-large`…`body-small` plus `code` (e.g.
  `--font-title-medium-size`). Set all four together; never a bare
  `font-size`.
- Density: `--density-control-height`, `--density-padding-x`,
  `--density-padding-y`, `--density-min-hit-target`. Controls are exactly
  one control height tall; `.compact` re-pitches everything but the
  hit-target floor.
- Space: `--space-0`…`--space-24` on the 4-pt grid (keys 0–6, 8, 10, 12,
  16, 20, 24). Radius: `--radius-none`, `-sm`, `-base`, `-md`, `-lg`,
  `-xl`, `-2xl`, `-3xl`, `-full`.
- Elevation: `--elevation-0`…`--elevation-3` are tonal surface **fills** —
  the default cue; use them as `background`. `--shadow-0`…`--shadow-3` are
  the opt-in cue for floating transients only (dialog, popover, tooltip,
  toast); resting surfaces never cast one.
- Interaction: `--color-focus-ring` with `--focus-ring-width` (the one focus
  treatment everywhere), `--state-disabled-opacity` (the disabled fade),
  `--color-scrim` (the modal backdrop; identical in both modes).
- Motion: `--ease-standard`, `--ease-standard-accelerate`,
  `--ease-standard-decelerate`, `--ease-emphasized`,
  `--ease-emphasized-accelerate`, `--ease-emphasized-decelerate`;
  durations `--duration-x-fast`, `--duration-fast`, `--duration-normal`,
  `--duration-slow`, `--duration-x-slow`.

## Idiomatic build snippet

A decision dialog, verbatim from `components/dialog.html` (a page that passes
the golden-mirror harness). Note the composition: scrim wraps dialog, the
footer ends in the default, the actions are plain `.btn`s:

```html
<div class="scrim">
  <div class="dialog">
    <div class="dialog-header"><div class="dialog-title">Discard changes?</div></div>
    <p class="dialog-body">Your edits to this document have not been saved.</p>
    <div class="dialog-footer">
      <button class="btn ghost">Cancel</button>
      <button class="btn">Discard</button>
    </div>
  </div>
</div>
```

## Gio-specific caveats

These pages mirror a Gio (Go) desktop implementation, and two things do not
port one-to-one:

- **Text shaping differs.** The browser and Gio use different shapers, so
  glyph advances, line breaks and exact text widths diverge slightly (the
  mirror is verified perceptually, not byte-wise). Never design a layout
  that depends on exact text measurement — a label fitting to the pixel, a
  wrap happening at a particular word.
- **Blur is not live.** Gio has no `backdrop-filter`: blur there is a cached
  offscreen pass — the backdrop is rendered once, blurred, and repainted
  from cache until the content behind it changes. A design that assumes
  continuous blur under motion (frosted glass over scrolling or animating
  content) will not port. Prefer the system's own overlay grammar:
  `--color-scrim` behind dialogs, tonal elevation fills for everything else.
