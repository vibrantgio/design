package mirror

// The G2.3 navigation-page verdicts: the .navbar, .tabs, .sidebar and
// .crumbs classes — the vocabulary components/navigation.html composes
// with — captured from the real patterns widgets (patterns/navbar,
// patterns/tabs, patterns/sidebar, patterns/breadcrumb) and compared
// against browser captures of per-specimen fixtures wearing exactly the
// published sheet's classes. Like TestCalibration and the G2.1/G2.2
// verdicts, these only deliver a verdict on the authoritative machine;
// elsewhere one half of the harness skips loudly.

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"

	"github.com/vibrantgio/components/golden"
	"github.com/vibrantgio/patterns/breadcrumb"
	"github.com/vibrantgio/patterns/navbar"
	"github.com/vibrantgio/patterns/sidebar"
	"github.com/vibrantgio/patterns/tabs"
	"github.com/vibrantgio/theme/tokens"
)

// navbarSize is the navbar capture viewport: the navbar goldens' 480 width
// at the density-pinned bar height patterns/shell allocates — ControlHeight
// + 2·PaddingY = 52 comfortable — rather than the goldens' 64, because the
// bar fills whatever it is handed and the shell pin is the height the class
// layer publishes.
var navbarSize = image.Pt(480, 52)

// tabsSize is the tabs capture viewport — TestTabsGolden's 240x128 canvas:
// the ControlHeight strip and, below it, the selected tab's content panel
// in tabs_test.go's fixed specimen colour (contentRect's #ff4040), which
// the fixture pins the same way the sidebar fixture pins its icon squares.
var tabsSize = image.Pt(240, 128)

// Sidebar capture viewports: the pattern's two contractual widths (192
// expanded, 48 collapsed — sidebar.go's expandedDp/collapsedDp) at the
// sidebar goldens' 256 height.
var (
	sidebarSize          = image.Pt(192, 256)
	sidebarCollapsedSize = image.Pt(48, 256)
)

// breadcrumbSize is the breadcrumb goldens' 320x32 canvas.
var breadcrumbSize = image.Pt(320, 32)

// fillRect mirrors tabs_test.go's contentRect: a widget filling its
// constraints with a fixed specimen colour, so the selected tab's content
// panel compares deterministically.
func fillRect(c color.NRGBA) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, c, clip.Rect{Max: gtx.Constraints.Max}.Op())
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}
}

// navIcon mirrors sidebar_test.go's testIcon: a 16x16 filled square in a
// fixed mid-blue, so the icon slot compares deterministically without
// dragging font rasterisation into the icon column.
func navIcon() layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		size := image.Pt(16, 16)
		paint.FillShape(gtx.Ops, color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0xff}, clip.Rect{Max: size}.Op())
		return layout.Dimensions{Size: size}
	}
}

// navItems is the sidebar fixture's item set: the first three of
// sidebar_test.go's labels, the second Active, so the selected row's
// primary-400 fill is in frame in both widths.
func navItems() []sidebar.Item {
	labels := []string{"Overview", "Tokens", "Colour"}
	items := make([]sidebar.Item, len(labels))
	for i, l := range labels {
		items[i] = sidebar.Item{Icon: navIcon(), Label: l, Active: i == 1}
	}
	return items
}

// TestNavigationMirrors scores each G2.3 specimen pair: the patterns widget
// against the browser render of the matching fixture, both at the same
// viewport. Every distance is logged; each must land under Tolerance for
// the page to count as a mirror of the pattern rather than a drawing of
// one.
func TestNavigationMirrors(t *testing.T) {
	srv := Serve(t)
	shaper := tokens.DefaultTypography.DeterministicShaper()

	cases := []struct {
		fixture string
		size    image.Point
		gio     layout.Widget
	}{
		{"navbar.html", navbarSize, navbar.Render(
			shaper,
			navbar.Props{Links: []navbar.Link{
				{Label: "Docs"},
				{Label: "Components", Active: true},
			}},
			tokens.DefaultLight, tokens.Spacing,
			tokens.DefaultTypography.LabelLarge, tokens.Comfortable,
		)},
		{"tabs.html", tabsSize, tabs.Render(
			shaper,
			tabs.Props{Tabs: []tabs.Tab{
				{Label: "Preview", Content: fillRect(color.NRGBA{R: 0xff, G: 0x40, B: 0x40, A: 0xff})},
				{Label: "Code"}, {Label: "Notes"},
			}},
			0,
			tokens.DefaultLight, tokens.Spacing,
			tokens.DefaultTypography.LabelLarge, tokens.Comfortable,
		)},
		{"sidebar.html", sidebarSize, sidebar.Render(
			shaper,
			sidebar.Props{Items: navItems()},
			false,
			tokens.DefaultLight, tokens.Spacing,
			tokens.DefaultTypography.LabelLarge, tokens.Comfortable,
		)},
		{"sidebar-collapsed.html", sidebarCollapsedSize, sidebar.Render(
			shaper,
			sidebar.Props{Items: navItems()},
			true,
			tokens.DefaultLight, tokens.Spacing,
			tokens.DefaultTypography.LabelLarge, tokens.Comfortable,
		)},
		{"breadcrumb.html", breadcrumbSize, breadcrumb.Render(
			shaper,
			breadcrumb.Props{Items: []breadcrumb.Item{
				{Label: "Home"}, {Label: "Design"}, {Label: "Tokens"},
			}},
			tokens.DefaultLight, tokens.Spacing,
			tokens.DefaultTypography.TitleSmall,
		)},
	}

	for _, c := range cases {
		t.Run(c.fixture, func(t *testing.T) {
			gio := golden.Capture(t, c.size, onBackground(c.gio))
			if !FixtureExists("fixtures/" + c.fixture) {
				t.Fatalf("no embedded fixture %q — a typo here would screenshot a 404 page", c.fixture)
			}
			web := CaptureBrowser(t, srv.URL+"/fixtures/"+c.fixture, c.size)
			d := Distance(gio, web)
			t.Logf("distance gio vs %s: %.4f (Tolerance %.4f)", c.fixture, d, Tolerance)
			if d > Tolerance {
				t.Errorf("pair %s scored %.4f > Tolerance %.4f: the mirror does not read as the pattern", c.fixture, d, Tolerance)
			}
		})
	}
}
