package pdf

import (
	"testing"

	htmlpkg "github.com/gvanbeck/nautilus/pdf/html"
)

// ─── parseHTMLColor ──────────────────────────────────────────────────────────

func TestParseHTMLColor_Hex6(t *testing.T) {
	cases := []struct {
		input string
		r, g, b uint8
	}{
		{"#ff0000", 255, 0, 0},
		{"#00ff00", 0, 255, 0},
		{"#0000ff", 0, 0, 255},
		{"#aabbcc", 0xaa, 0xbb, 0xcc},
		{"#AABBCC", 0xaa, 0xbb, 0xcc}, // uppercase
	}
	for _, tc := range cases {
		c, ok := parseHTMLColor(tc.input)
		if !ok || c == nil {
			t.Errorf("%s: expected color, got none", tc.input)
			continue
		}
		if c.R != tc.r || c.G != tc.g || c.B != tc.b {
			t.Errorf("%s: want (%d,%d,%d), got (%d,%d,%d)", tc.input, tc.r, tc.g, tc.b, c.R, c.G, c.B)
		}
	}
}

func TestParseHTMLColor_Hex3(t *testing.T) {
	c, ok := parseHTMLColor("#f0a")
	if !ok || c == nil {
		t.Fatal("expected color for #f0a")
	}
	if c.R != 0xff || c.G != 0x00 || c.B != 0xaa {
		t.Errorf("want (255,0,170), got (%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestParseHTMLColor_RGB(t *testing.T) {
	c, ok := parseHTMLColor("rgb(10, 20, 30)")
	if !ok || c == nil {
		t.Fatal("expected color for rgb()")
	}
	if c.R != 10 || c.G != 20 || c.B != 30 {
		t.Errorf("want (10,20,30), got (%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestParseHTMLColor_Named(t *testing.T) {
	cases := map[string][3]uint8{
		"black": {0, 0, 0},
		"white": {255, 255, 255},
		"red":   {255, 0, 0},
		"navy":  {0, 0, 128},
	}
	for name, want := range cases {
		c, ok := parseHTMLColor(name)
		if !ok || c == nil {
			t.Errorf("named color %q: expected color", name)
			continue
		}
		if c.R != want[0] || c.G != want[1] || c.B != want[2] {
			t.Errorf("named color %q: want (%d,%d,%d), got (%d,%d,%d)",
				name, want[0], want[1], want[2], c.R, c.G, c.B)
		}
	}
}

func TestParseHTMLColor_Empty(t *testing.T) {
	if c, ok := parseHTMLColor(""); ok || c != nil {
		t.Error("empty string should return nil, false")
	}
}

func TestParseHTMLColor_Invalid(t *testing.T) {
	invalids := []string{"#gg0000", "notacolor", "#12345", "rgb(256,0,0)"}
	for _, s := range invalids {
		_, ok := parseHTMLColor(s)
		if ok {
			t.Errorf("expected invalid for %q, got ok", s)
		}
	}
}

// ─── htmlHAlign / htmlVAlign ──────────────────────────────────────────────────

func TestHtmlHAlign(t *testing.T) {
	cases := map[string]HAlign{
		"left":    HAlignLeft,
		"center":  HAlignCenter,
		"right":   HAlignRight,
		"":        HAlignDefault,
		"unknown": HAlignDefault,
	}
	for input, want := range cases {
		if got := htmlHAlign(input); got != want {
			t.Errorf("htmlHAlign(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestHtmlVAlign(t *testing.T) {
	cases := map[string]VAlign{
		"top":    VAlignTop,
		"middle": VAlignMiddle,
		"bottom": VAlignBottom,
		"":       VAlignDefault,
	}
	for input, want := range cases {
		if got := htmlVAlign(input); got != want {
			t.Errorf("htmlVAlign(%q) = %v, want %v", input, got, want)
		}
	}
}

// ─── TableFromHTML structural ─────────────────────────────────────────────────

func TestTableFromHTML_RowCount(t *testing.T) {
	input := `<table>
		<thead><tr><th>A</th><th>B</th></tr></thead>
		<tbody>
			<tr><td>1</td><td>2</td></tr>
			<tr><td>3</td><td>4</td></tr>
		</tbody>
	</table>`

	ht, err := htmlpkg.ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}

	doc, _ := New(Config{})
	tbl := doc.TableFromHTML(ht, TableConfig{
		ColWidths: []float64{100, 100},
	}, HtmlTableOptions{})

	if len(tbl.rows) != 3 {
		t.Errorf("want 3 rows, got %d", len(tbl.rows))
	}
}

func TestTableFromHTML_HeaderStyle(t *testing.T) {
	input := `<table>
		<thead><tr><th>Header</th></tr></thead>
		<tbody><tr><td>Body</td></tr></tbody>
	</table>`

	ht, err := htmlpkg.ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}

	headerBg := Color{R: 20, G: 20, B: 100}
	opts := HtmlTableOptions{
		HeaderStyle: CellStyle{Background: &headerBg},
	}

	doc, _ := New(Config{})
	tbl := doc.TableFromHTML(ht, TableConfig{ColWidths: []float64{200}}, opts)

	if len(tbl.rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(tbl.rows))
	}

	// Header row cell should have the HeaderStyle background.
	headerCell := tbl.rows[0].Cells[0]
	if headerCell.Style.Background == nil {
		t.Error("header cell should have background from HeaderStyle")
	} else if *headerCell.Style.Background != headerBg {
		t.Errorf("header bg: want %v, got %v", headerBg, *headerCell.Style.Background)
	}

	// Body cell should not have the header background.
	bodyCell := tbl.rows[1].Cells[0]
	if bodyCell.Style.Background != nil {
		t.Errorf("body cell should have no background, got %v", *bodyCell.Style.Background)
	}
}

func TestTableFromHTML_ColAndRowSpan(t *testing.T) {
	input := `<table>
		<tr><td colspan="2">Wide</td></tr>
		<tr><td rowspan="2">Tall</td><td>R2C2</td></tr>
		<tr><td>R3C2</td></tr>
	</table>`

	ht, err := htmlpkg.ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}

	doc, _ := New(Config{})
	tbl := doc.TableFromHTML(ht, TableConfig{ColWidths: []float64{100, 100}}, HtmlTableOptions{})

	if tbl.rows[0].Cells[0].ColSpan != 2 {
		t.Errorf("colspan: want 2, got %d", tbl.rows[0].Cells[0].ColSpan)
	}
	if tbl.rows[1].Cells[0].RowSpan != 2 {
		t.Errorf("rowspan: want 2, got %d", tbl.rows[1].Cells[0].RowSpan)
	}
}

func TestTableFromHTML_CellBackground(t *testing.T) {
	input := `<table>
		<tr><td bgcolor="#112233">Colored</td></tr>
	</table>`

	ht, err := htmlpkg.ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}

	doc, _ := New(Config{})
	tbl := doc.TableFromHTML(ht, TableConfig{ColWidths: []float64{200}}, HtmlTableOptions{})

	bg := tbl.rows[0].Cells[0].Style.Background
	if bg == nil {
		t.Fatal("expected background color on cell")
	}
	if bg.R != 0x11 || bg.G != 0x22 || bg.B != 0x33 {
		t.Errorf("bg color: want (17,34,51), got (%d,%d,%d)", bg.R, bg.G, bg.B)
	}
}

func TestTableFromHTML_RowBackground(t *testing.T) {
	input := `<table>
		<tr bgcolor="#eeeeee"><td>Row bg</td></tr>
	</table>`

	ht, err := htmlpkg.ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}

	doc, _ := New(Config{})
	tbl := doc.TableFromHTML(ht, TableConfig{ColWidths: []float64{200}}, HtmlTableOptions{})

	bg := tbl.rows[0].Background
	if bg == nil {
		t.Fatal("expected row Background")
	}
	if bg.R != 0xee || bg.G != 0xee || bg.B != 0xee {
		t.Errorf("row bg: want (238,238,238), got (%d,%d,%d)", bg.R, bg.G, bg.B)
	}
}

func TestTableFromHTML_Alignment(t *testing.T) {
	input := `<table>
		<tr>
			<td align="left">L</td>
			<td align="center">C</td>
			<td align="right">R</td>
		</tr>
	</table>`

	ht, err := htmlpkg.ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}

	doc, _ := New(Config{})
	tbl := doc.TableFromHTML(ht, TableConfig{ColWidths: []float64{60, 60, 60}}, HtmlTableOptions{})

	cells := tbl.rows[0].Cells
	if cells[0].Style.HAlign != HAlignLeft {
		t.Errorf("cell 0: want HAlignLeft, got %v", cells[0].Style.HAlign)
	}
	if cells[1].Style.HAlign != HAlignCenter {
		t.Errorf("cell 1: want HAlignCenter, got %v", cells[1].Style.HAlign)
	}
	if cells[2].Style.HAlign != HAlignRight {
		t.Errorf("cell 2: want HAlignRight, got %v", cells[2].Style.HAlign)
	}
}

func TestTableFromHTML_SpanFontFor(t *testing.T) {
	input := `<table><tr><td><b>Bold</b> normal</td></tr></table>`

	ht, err := htmlpkg.ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}

	opts := HtmlTableOptions{
		SpanFontFor: func(s htmlpkg.Style) string {
			if s.Bold {
				return "bold"
			}
			return "regular"
		},
	}

	doc, _ := New(Config{})
	tbl := doc.TableFromHTML(ht, TableConfig{ColWidths: []float64{200}}, opts)

	spans := tbl.rows[0].Cells[0].Spans
	if len(spans) == 0 {
		t.Fatal("expected spans in cell")
	}

	boldFound := false
	regularFound := false
	for _, sp := range spans {
		switch sp.FontName {
		case "bold":
			boldFound = true
		case "regular":
			regularFound = true
		}
	}
	if !boldFound {
		t.Error("expected a span with FontName 'bold'")
	}
	if !regularFound {
		t.Error("expected a span with FontName 'regular'")
	}
}

func TestTableFromHTML_FooterStyle(t *testing.T) {
	input := `<table>
		<tbody><tr><td>Body</td></tr></tbody>
		<tfoot><tr><td>Total</td></tr></tfoot>
	</table>`

	ht, err := htmlpkg.ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}

	footerBg := Color{R: 240, G: 240, B: 240}
	opts := HtmlTableOptions{
		FooterStyle: CellStyle{Background: &footerBg},
	}

	doc, _ := New(Config{})
	tbl := doc.TableFromHTML(ht, TableConfig{ColWidths: []float64{200}}, opts)

	footerCell := tbl.rows[1].Cells[0]
	if footerCell.Style.Background == nil {
		t.Fatal("footer cell should have background")
	}
	if *footerCell.Style.Background != footerBg {
		t.Errorf("footer bg: want %v, got %v", footerBg, *footerCell.Style.Background)
	}
}

func TestTableFromHTML_Caption(t *testing.T) {
	// Caption is parsed and accessible; it is not automatically rendered into
	// the table (callers must render it as text above the table).
	input := `<table>
		<caption>Quarterly Results</caption>
		<tr><td>Q1</td><td>100</td></tr>
	</table>`

	ht, err := htmlpkg.ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if ht.Caption != "Quarterly Results" {
		t.Errorf("caption: want %q, got %q", "Quarterly Results", ht.Caption)
	}
}

// ─── effectiveSpan ───────────────────────────────────────────────────────────

func TestEffectiveSpan(t *testing.T) {
	if effectiveSpan(0) != 1 {
		t.Error("0 should become 1")
	}
	if effectiveSpan(-1) != 1 {
		t.Error("-1 should become 1")
	}
	if effectiveSpan(3) != 3 {
		t.Error("3 should stay 3")
	}
}
