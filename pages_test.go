package design

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

var (
	// A custom property's declaration and a reference to one.
	varDeclRE = regexp.MustCompile(`(--[A-Za-z0-9-]+)\s*:`)
	varRefRE  = regexp.MustCompile(`var\((--[A-Za-z0-9-]+)\)`)
	// The two places a page styles from. Prose naming a variable — the
	// intros do, in <code> — is reader-facing text and styles nothing.
	styleBlockRE = regexp.MustCompile(`(?s)<style>(.*?)</style>`)
	styleAttrRE  = regexp.MustCompile(`style="([^"]*)"`)
)

// TestPageVarClosure asserts that every var() a page styles through resolves
// in the bundle's own sheet: a page can never name a token styles.css does
// not declare.
//
// It covers every page the bundle carries — the three foundation pages
// theme/export generates AND the six hand-authored ones under components/ —
// because the hand-authored pages are the ones nothing else checks. A
// reference that resolves to nothing is silent in a browser: the property
// falls back to its initial value, so a panel loses its fill and still lays
// out, and only a reader who knows what the page should look like sees it.
// theme/export's own TestPageVarClosure covers what the generator emits at
// the moment it emits it; this one covers what is published.
func TestPageVarClosure(t *testing.T) {
	sheet, err := Bundle.ReadFile("styles.css")
	if err != nil {
		t.Fatalf("Bundle.ReadFile(styles.css): %v", err)
	}
	declared := map[string]bool{}
	for _, m := range varDeclRE.FindAllStringSubmatch(string(sheet), -1) {
		declared[m[1]] = true
	}
	if len(declared) == 0 {
		t.Fatal("styles.css declares no custom properties at all")
	}

	pages := bundlePages(t)
	if len(pages) < 9 {
		t.Errorf("the bundle carries %d pages, want the three generated and the six hand-authored ones", len(pages))
	}
	for _, name := range pages {
		src, err := Bundle.ReadFile(name)
		if err != nil {
			t.Fatalf("Bundle.ReadFile(%q): %v", name, err)
		}
		var styled []string
		for _, m := range styleBlockRE.FindAllStringSubmatch(string(src), -1) {
			styled = append(styled, m[1])
		}
		for _, m := range styleAttrRE.FindAllStringSubmatch(string(src), -1) {
			styled = append(styled, m[1])
		}
		refs := 0
		for _, css := range styled {
			for _, m := range varRefRE.FindAllStringSubmatch(css, -1) {
				refs++
				if !declared[m[1]] {
					t.Errorf("%s styles through %s, which styles.css does not declare", name, m[1])
				}
			}
		}
		if refs == 0 {
			t.Errorf("%s styles through no token variable at all", name)
		}
	}
}

// bundlePages lists every page the bundle publishes, generated and
// hand-authored alike, by walking it rather than by naming them: a page added
// to the bundle and left out of a list here is a page nothing checks.
func bundlePages(t *testing.T) []string {
	t.Helper()
	var pages []string
	err := fs.WalkDir(Bundle, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".html") {
			pages = append(pages, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk the bundle: %v", err)
	}
	return pages
}
