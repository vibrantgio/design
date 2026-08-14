package design

import "testing"

// The uploaded path set is a contract: finalize_plan writes exactly these
// uploaded paths, so the embedded bundle must carry each of them,
// non-empty: the six document paths plus the self-hosted faces behind
// --font-family and --font-family-code, with their font licences.
func TestBundleCarriesTheSixUploadedPaths(t *testing.T) {
	paths := []string{
		"readme.md",
		"theme.json",
		"styles.css",
		"foundations/color.html",
		"foundations/type.html",
		"foundations/layout.html",
		"components/buttons.html",
		"components/cards.html",
		"components/forms.html",
		"components/table.html",
		"fonts/roboto-regular.ttf",
		"fonts/roboto-medium.ttf",
		"fonts/robotomono-regular.ttf",
		"fonts/LICENSE-Roboto-Apache-2.0.txt",
		"fonts/LICENSE-RobotoMono-OFL.txt",
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
