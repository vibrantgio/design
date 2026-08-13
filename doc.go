// Package design carries the published Vibrant Gio design bundle — the
// project layout that lives at claude.ai/design: readme.md, theme.json,
// styles.css and the foundation pages under foundations/.
//
// The bundle is generated, never edited here: theme's cmd/vg-tokens writes
// it (scripts/push-design.sh in the org's .github repository runs that and
// then pushes the result), so a change to a token or a ramp lands in theme
// first and is re-emitted into this repository. Editing one of these files
// directly is the same mistake as editing any other generated file — the
// next emit puts the old words back.
//
// Bundle exposes the six files as an embed.FS so that a consumer — the
// fidelity harness above all — reads a page by name instead of by a
// relative path that breaks the moment a test runs from somewhere else.
package design
