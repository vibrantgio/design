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
- `theme.json` holds the parameters that reproduce the theme (the theme
  colour and the platform's whole set per appearance). `readme.md` documents the token families. Both are generated,
  like `styles.css`, by the Go theme exporter — treat all three as read-only.
- `components/*.html` are the copyable markup reference; `foundations/*.html`
  show every token rendered. The pages are verified pixel-wise against the Gio
  implementation's golden images, so their markup is the idiom to copy.
- Mode and density are class switches: `class="dark"` on the root element,
  `class="compact"` on any subtree. They are orthogonal; never restyle for
  them by hand.

## Class families

- **Buttons** — `.btn` on `<button>`: filled by default (accent under
  on-accent). Emphasis modifiers: `.btn.tonal` (the platform's push button), `.btn.ghost`
  (no fill at rest). A ghost's hover is its host surface's own one-step
  walk — inside `.card`, `.dialog` or `.popover` the hover/press fills
  re-derive from that surface automatically; never restyle them by hand.
  `.btn.icon` is the square icon-only form (inline SVG on `currentColor`).
  `.selected` marks a toggled-on button; ghosts have no selected treatment.
  There is **no size modifier** — density is the size axis.
- **Badges** — `.badge`: label-medium text in one derived colour, no fill and no
  padding. Read, not used — no interaction states. Status badges are
  `.badge.success` / `.badge.warning` / `.badge.error` / `.badge.info`: the
  role's own hue derived against the surface the badge stands on, five
  variants differing in hue alone. Compose these for any status — a build
  state, a health verdict — and never inline-style a status colour.
- **Forms** — native elements, no script: `.input` on a text `<input>`;
  a dropdown is `<select class="input select">` inside a `.select-wrap`
  (which draws the chevron); `.checkbox` and `.radio` on their native input
  types. Disabled is always the native `disabled` attribute.
- **Cards and groups** — `.card` is one thing singled out: the platform's
  grouped-box fill, a small step of fill on the content, no line of its own. `.group` divides
  the page: a hairline at the level of the surface it is in, no fill of its
  own, with an optional `.group-label` top-leading inside it. Which one to
  reach for answers one question — am I dividing the page, or singling
  something out? A row of tiers, a form in sections, a list of articles are
  groups; the one thing that must stand apart is the card. A group may hold
  a card; it never holds another group, and neither wears a role.
- **Table** — `.table` on a real `<table>`: header band, one-control-height
  rows, seam rules, no zebra. Tables are unframed: the Surface fill and
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
  a right-aligned `.dialog-footer`. Two purposes: a **decision** dialog has a
  footer ending in the Return-bound default and **no close X**; a
  dismissable **panel** has a ghost close (`.btn.ghost.icon`) top-right and
  no footer. A corner affordance like that close draws at control height,
  the same as every control — density is the only size knob, and the control
  it draws is the pointer target, never a size of its own. Anchored surfaces: `.popover` with a `.popover-tail` on side
  `.top`/`.bottom`/`.left`/`.right`; `.tooltip` (inverse video); `.toast`
  (purposes `.success`/`.warning`/`.error`) stacked in a `.toast-stack`.
- **State-forcing twins** — `.is-hover`, `.is-active`, `.is-focus`,
  `.is-checked` share the live `:hover`/`:active`/`:focus-visible`/`:checked`
  declarations, for showing a state statically. In real interactive markup
  rely on the pseudo-classes and leave the twins off.

## Token families

- Colour: `--platform-<name>`, one per semantic colour the platform names —
  `--platform-window-background`, `--platform-text-background`,
  `--platform-sidebar-material`, `--platform-card-fill`,
  `--platform-control-accent`, `--platform-label`,
  `--platform-secondary-label`, `--platform-separator`,
  `--platform-selected-content-background`, and the rest. Nothing is derived
  from anything else: reach for the name of what the thing IS. A value the
  platform states at a coverage carries its alpha and composites over
  whatever is beneath, which is how one value reads right on every fill.
- Type: `--font-family`, `--font-family-code`, and per-role
  `--font-<role>-size`/`-line-height`/`-weight`/`-tracking` for roles
  `display-large`…`body-small` plus `code` (e.g.
  `--font-title-medium-size`). Set all four together; never a bare
  `font-size`.
- Density: `--density-control-height`, `--density-chip-height`,
  `--density-field-height`, `--density-row-height`, `--density-padding-x`,
  `--density-padding-y`. Controls are exactly one control height tall and a
  chip is four px under it; `.compact` re-pitches every one of them.
- Space: `--space-0`…`--space-24` on the 4-pt grid (keys 0–6, 8, 10, 12,
  16, 20, 24). Radius: `--radius-none`, `-sm`, `-base`, `-md`, `-lg`,
  `-xl`, `-2xl`, `-3xl`, `-full`.
- Levels: five, counted from the backdrop up — the backdrop (the window's
  own plane, what shows wherever nothing stands), the chrome regions
  (navbar, toolbar, sidebar, inspector, status bar, pane), the content, what
  is raised on it, and what floats. A level is not a fill source: each one
  is filled with the platform name for what stands there, so read the level
  to know which `--platform-*` to use. `--shadow-0`…`--shadow-3` are the
  cue a floating transient carries (dialog, popover, tooltip, toast);
  resting surfaces never cast one.
- Interaction: `--platform-keyboard-focus-indicator` with
  `--focus-ring-width` (the one focus treatment everywhere), and
  `--platform-scrim` (the veil a modal draws over what it covers).
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
  offscreen pass — what lies behind is rendered once, blurred, and repainted
  from cache until the content behind it changes. A design that assumes
  continuous blur under motion (frosted glass over scrolling or animating
  content) will not port. Prefer the system's own overlay grammar:
  `--platform-scrim` behind dialogs, the platform's own fills for everything else.
