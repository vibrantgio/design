package design

import "testing"

// The uploaded path set is a contract: finalize_plan writes exactly these
// six paths, so the embedded bundle must carry each of them, non-empty.
func TestBundleCarriesTheSixUploadedPaths(t *testing.T) {
	paths := []string{
		"readme.md",
		"theme.json",
		"styles.css",
		"foundations/color.html",
		"foundations/type.html",
		"foundations/layout.html",
	}
	for _, p := range paths {
		b, err := Bundle.ReadFile(p)
		if err != nil {
			t.Errorf("Bundle.ReadFile(%q): %v", p, err)
			continue
		}
		if len(b) == 0 {
			t.Errorf("Bundle.ReadFile(%q): empty file", p)
		}
	}
}
