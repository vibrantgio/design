package mirror

// The pattern-page verdicts: .card, .group and .table — the classes the
// cards.html and table.html component pages compose with — captured from
// the real patterns widgets (patterns/card, patterns/group,
// patterns/table) and compared
// against browser captures of per-specimen fixtures wearing exactly the
// published sheet's classes. Like TestCalibration and TestComponentMirrors,
// these only deliver a verdict on the authoritative machine; elsewhere one
// half of the harness skips loudly.

import (
	"image"
	"image/color"
	"strconv"
	"testing"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"

	"github.com/vibrantgio/components/golden"
	"github.com/vibrantgio/patterns/card"
	"github.com/vibrantgio/patterns/group"
	"github.com/vibrantgio/patterns/table"
	"github.com/vibrantgio/theme/tokens"
	"github.com/vibrantgio/theme/typeset"
)

// cardSize is the card capture viewport — patterns/card's golden canvas
// (280x200); the card fills the constraints it is given, so the fixture
// pins the same 280x200 on the .card box. The group's golden is the same
// size and its fixture pins the same box.
var cardSize = image.Pt(280, 200)

// tableSize is the table capture viewport — TestTableGolden's 360x200: a
// 36 px header band, four 36 px rows, and 20 px of bare Surface below,
// because drawTable grounds its whole constraints on the Surface pin
// before drawing the grid.
var tableSize = image.Pt(360, 200)

// textSlot mirrors card_test.go's slot helper: the card draws no text of
// its own, so the slots are caller-built widgets drawn through
// theme/typeset — a role's LineHeight is the CSS line box, which is
// exactly what makes the browser's line boxes comparable.
func textSlot(shaper *text.Shaper, style tokens.TextStyle, c color.NRGBA, s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		m := op.Record(gtx.Ops)
		paint.ColorOp{Color: c}.Add(gtx.Ops)
		material := m.Stop()
		f := typeset.Font(style, font.Normal)
		lbl := typeset.Label(style, 1)
		// Min is dropped so the slot reports the text it drew rather than
		// the card's own minimum — the same trick the card goldens use to
		// make the slot stack visible.
		gtx.Constraints.Min = image.Point{}
		return typeset.Layout(gtx, shaper, lbl, f, unit.Sp(style.Size), s, material)
	}
}

// cardSlots is the header / body / footer trio both card fixtures carry:
// title-medium at the Text pin, a single non-wrapping body-medium line at
// neutral 700, label-medium at the accent pin — card_test.go's slots(),
// shortened so neither renderer has a line-break decision to disagree on.
func cardSlots(shaper *text.Shaper) (header, body, footer layout.Widget) {
	c := tokens.DefaultLight
	typo := tokens.DefaultTypography
	return textSlot(shaper, typo.TitleMedium, c.Text, "Density"),
		textSlot(shaper, typo.BodyMedium, c.Ramps.Neutral.Step(700), "Comfortable and Compact"),
		textSlot(shaper, typo.LabelMedium, c.Primary, "Read the token")
}

// groupContent is what the group fixture holds: body-medium at the Text pin
// over the same role at neutral 700, both single non-wrapping lines so
// neither renderer has a line-break decision to disagree on. The group's own
// label is not here — the pattern draws that itself, from Props.Label.
func groupContent(shaper *text.Shaper) []layout.Widget {
	c := tokens.DefaultLight
	typo := tokens.DefaultTypography
	return []layout.Widget{
		textSlot(shaper, typo.BodyMedium, c.Text, "Comfortable"),
		textSlot(shaper, typo.BodyMedium, c.Ramps.Neutral.Step(700), "Compact re-pitches"),
	}
}

// tableWidget is TestTableGolden's light-comfortable configuration: ID
// pinned 80 dp, sortable and actively sorted ascending (so the chevron is
// in frame), Name flexed, Steps pinned 96 dp; four rows of RenderTextCell
// body text.
func tableWidget(shaper *text.Shaper) layout.Widget {
	cell := func(f func(int) string) func(int) layout.Widget {
		return func(i int) layout.Widget {
			return table.RenderTextCell(shaper, tokens.DefaultLight, tokens.DefaultTypography.BodyMedium, f(i))
		}
	}
	names := []string{"Tokens", "Density", "Elevation", "Motion"}
	cols := []table.Column[int]{
		{Header: "ID", Width: unit.Dp(80), Sortable: true, Cell: cell(func(i int) string { return strconv.Itoa(i + 1) })},
		{Header: "Name", Sortable: true, Cell: cell(func(i int) string { return names[i] })},
		{Header: "Steps", Width: unit.Dp(96), Cell: cell(func(i int) string { return strconv.Itoa(4 * (i + 1)) })},
	}
	return table.Render(shaper, cols, []int{0, 1, 2, 3}, table.Sort{Column: 0, Asc: true},
		tokens.DefaultLight, tokens.Spacing, tokens.DefaultTypography.LabelLarge, tokens.Comfortable)
}

// TestPatternMirrors scores each pattern specimen pair: the patterns widget
// against the browser render of the matching fixture, both at the same
// viewport. Every distance is logged; each must land under Tolerance for
// the page to count as a mirror of the pattern rather than a drawing of
// one.
func TestPatternMirrors(t *testing.T) {
	srv := Serve(t)
	shaper := tokens.DefaultTypography.DeterministicShaper()
	header, body, footer := cardSlots(shaper)

	cases := []struct {
		fixture string
		size    image.Point
		gio     layout.Widget
	}{
		{"card.html", cardSize, card.Render(
			card.Props{Header: header, Body: body, Footer: footer},
			tokens.DefaultLight, tokens.Spacing, tokens.Radius,
		)},
		{"group.html", cardSize, group.Render(shaper,
			group.Props{Label: "Density", Content: groupContent(shaper)},
			tokens.DefaultLight, tokens.Spacing, tokens.Radius,
			tokens.DefaultTypography.LabelLarge,
		)},
		{"table.html", tableSize, tableWidget(shaper)},
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
