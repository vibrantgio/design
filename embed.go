package design

import "embed"

// Bundle is the published design bundle, byte-identical to the working
// tree: readme.md, theme.json, styles.css, foundations/*.html and the
// hand-authored components/*.html pages — the same paths
// scripts/push-design.sh uploads to claude.ai/design.
//
//go:embed readme.md theme.json styles.css foundations components fonts
var Bundle embed.FS
