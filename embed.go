package design

import "embed"

// Bundle is the published design bundle, byte-identical to the working
// tree: readme.md, theme.json, styles.css and foundations/*.html — the
// same six paths scripts/push-design.sh uploads to claude.ai/design.
//
//go:embed readme.md theme.json styles.css foundations fonts
var Bundle embed.FS
