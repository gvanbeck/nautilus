package html

import (
	"strings"
	"testing"
)

// ─── ParseTable structural tests ─────────────────────────────────────────────

func TestParseTable_Basic(t *testing.T) {
	input := `<table>
		<tr><td>A</td><td>B</td></tr>
		<tr><td>C</td><td>D</td></tr>
	</table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if len(ht.Rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(ht.Rows))
	}
	if len(ht.Rows[0].Cells) != 2 {
		t.Fatalf("row 0: want 2 cells, got %d", len(ht.Rows[0].Cells))
	}
	if text(ht.Rows[0].Cells[0]) != "A" {
		t.Errorf("row 0 cell 0: want %q, got %q", "A", text(ht.Rows[0].Cells[0]))
	}
	if text(ht.Rows[1].Cells[1]) != "D" {
		t.Errorf("row 1 cell 1: want %q, got %q", "D", text(ht.Rows[1].Cells[1]))
	}
}

func TestParseTable_TheadTbody(t *testing.T) {
	input := `<table>
		<thead><tr><th>Name</th><th>Value</th></tr></thead>
		<tbody>
			<tr><td>Alpha</td><td>1</td></tr>
			<tr><td>Beta</td><td>2</td></tr>
		</tbody>
	</table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if len(ht.Rows) != 3 {
		t.Fatalf("want 3 rows, got %d", len(ht.Rows))
	}
	if !ht.Rows[0].IsHeader {
		t.Error("row 0 should be IsHeader (inside <thead>)")
	}
	if ht.Rows[1].IsHeader {
		t.Error("row 1 should not be IsHeader (inside <tbody>)")
	}
	if !ht.Rows[0].Cells[0].IsHeader {
		t.Error("cell 0,0 should be IsHeader (<th>)")
	}
}

func TestParseTable_Tfoot(t *testing.T) {
	input := `<table>
		<tbody><tr><td>Body</td></tr></tbody>
		<tfoot><tr><td>Total</td></tr></tfoot>
	</table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if len(ht.Rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(ht.Rows))
	}
	if ht.Rows[0].IsFooter {
		t.Error("row 0 should not be IsFooter")
	}
	if !ht.Rows[1].IsFooter {
		t.Error("row 1 should be IsFooter (inside <tfoot>)")
	}
}

func TestParseTable_AllThMarkAsHeader(t *testing.T) {
	// A <tr> with all <th> cells but no <thead> parent should be IsHeader.
	input := `<table>
		<tr><th>Name</th><th>Age</th></tr>
		<tr><td>Alice</td><td>30</td></tr>
	</table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if !ht.Rows[0].IsHeader {
		t.Error("row 0 (all <th>) should be IsHeader")
	}
	if ht.Rows[1].IsHeader {
		t.Error("row 1 (<td> only) should not be IsHeader")
	}
}

// ─── Spanning ────────────────────────────────────────────────────────────────

func TestParseTable_ColSpan(t *testing.T) {
	input := `<table>
		<tr><td colspan="3">Wide</td></tr>
		<tr><td>A</td><td>B</td><td>C</td></tr>
	</table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if ht.Rows[0].Cells[0].ColSpan != 3 {
		t.Errorf("colspan: want 3, got %d", ht.Rows[0].Cells[0].ColSpan)
	}
}

func TestParseTable_RowSpan(t *testing.T) {
	input := `<table>
		<tr><td rowspan="2">Tall</td><td>Row 1</td></tr>
		<tr><td>Row 2</td></tr>
	</table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if ht.Rows[0].Cells[0].RowSpan != 2 {
		t.Errorf("rowspan: want 2, got %d", ht.Rows[0].Cells[0].RowSpan)
	}
}

// ─── Alignment ───────────────────────────────────────────────────────────────

func TestParseTable_Align_Attr(t *testing.T) {
	input := `<table>
		<tr>
			<td align="left">L</td>
			<td align="center">C</td>
			<td align="right">R</td>
		</tr>
	</table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	cells := ht.Rows[0].Cells
	if cells[0].HAlign != "left" {
		t.Errorf("want left, got %q", cells[0].HAlign)
	}
	if cells[1].HAlign != "center" {
		t.Errorf("want center, got %q", cells[1].HAlign)
	}
	if cells[2].HAlign != "right" {
		t.Errorf("want right, got %q", cells[2].HAlign)
	}
}

func TestParseTable_VAlign_Attr(t *testing.T) {
	input := `<table><tr>
		<td valign="top">T</td>
		<td valign="middle">M</td>
		<td valign="bottom">B</td>
	</tr></table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	cells := ht.Rows[0].Cells
	if cells[0].VAlign != "top" {
		t.Errorf("want top, got %q", cells[0].VAlign)
	}
	if cells[1].VAlign != "middle" {
		t.Errorf("want middle, got %q", cells[1].VAlign)
	}
	if cells[2].VAlign != "bottom" {
		t.Errorf("want bottom, got %q", cells[2].VAlign)
	}
}

func TestParseTable_RowAlign_Inherited(t *testing.T) {
	// Cell without align should inherit from the row.
	input := `<table>
		<tr align="right">
			<td>No own align</td>
		</tr>
	</table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if ht.Rows[0].Cells[0].HAlign != "right" {
		t.Errorf("want right (inherited), got %q", ht.Rows[0].Cells[0].HAlign)
	}
}

func TestParseTable_StyleTextAlign(t *testing.T) {
	input := `<table><tr>
		<td style="text-align: center">C</td>
	</tr></table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if ht.Rows[0].Cells[0].HAlign != "center" {
		t.Errorf("style text-align: want center, got %q", ht.Rows[0].Cells[0].HAlign)
	}
}

// ─── Colours ─────────────────────────────────────────────────────────────────

func TestParseTable_BgColor_Attr(t *testing.T) {
	input := `<table>
		<tr bgcolor="#aabbcc"><td>row bg</td></tr>
		<tr><td bgcolor="#ff0000">cell bg</td></tr>
	</table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if ht.Rows[0].BgColor != "#aabbcc" {
		t.Errorf("row bgcolor: want #aabbcc, got %q", ht.Rows[0].BgColor)
	}
	if ht.Rows[1].Cells[0].BgColor != "#ff0000" {
		t.Errorf("cell bgcolor: want #ff0000, got %q", ht.Rows[1].Cells[0].BgColor)
	}
}

func TestParseTable_StyleBackgroundColor(t *testing.T) {
	input := `<table><tr>
		<td style="background-color: #123456">X</td>
	</tr></table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if ht.Rows[0].Cells[0].BgColor != "#123456" {
		t.Errorf("want #123456, got %q", ht.Rows[0].Cells[0].BgColor)
	}
}

func TestParseTable_StyleColor(t *testing.T) {
	input := `<table><tr>
		<td style="color: red">X</td>
	</tr></table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if ht.Rows[0].Cells[0].Color != "red" {
		t.Errorf("want red, got %q", ht.Rows[0].Cells[0].Color)
	}
}

// ─── Bold ─────────────────────────────────────────────────────────────────────

func TestParseTable_ThIsBold(t *testing.T) {
	input := `<table><tr><th>Header</th></tr></table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if !ht.Rows[0].Cells[0].Bold {
		t.Error("<th> cell should be Bold")
	}
}

func TestParseTable_StyleFontWeight(t *testing.T) {
	input := `<table><tr>
		<td style="font-weight: bold">Bold</td>
		<td>Normal</td>
	</tr></table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if !ht.Rows[0].Cells[0].Bold {
		t.Error("font-weight:bold cell should be Bold")
	}
	if ht.Rows[0].Cells[1].Bold {
		t.Error("normal cell should not be Bold")
	}
}

// ─── nowrap / width ───────────────────────────────────────────────────────────

func TestParseTable_NoWrap(t *testing.T) {
	input := `<table><tr><td nowrap>No wrap</td><td>Wrap</td></tr></table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if !ht.Rows[0].Cells[0].NoWrap {
		t.Error("cell with nowrap should have NoWrap=true")
	}
	if ht.Rows[0].Cells[1].NoWrap {
		t.Error("cell without nowrap should have NoWrap=false")
	}
}

func TestParseTable_Width(t *testing.T) {
	input := `<table><tr><td width="120">W</td></tr></table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if ht.Rows[0].Cells[0].Width != "120" {
		t.Errorf("want width 120, got %q", ht.Rows[0].Cells[0].Width)
	}
}

// ─── Caption ─────────────────────────────────────────────────────────────────

func TestParseTable_Caption(t *testing.T) {
	input := `<table>
		<caption>My <b>Table</b></caption>
		<tr><td>X</td></tr>
	</table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if ht.Caption != "My Table" {
		t.Errorf("caption: want %q, got %q", "My Table", ht.Caption)
	}
}

// ─── Inline HTML in cells ────────────────────────────────────────────────────

func TestParseTable_InlineHTML(t *testing.T) {
	input := `<table><tr>
		<td><b>Bold</b> and <i>italic</i></td>
	</tr></table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	spans := ht.Rows[0].Cells[0].Spans
	if len(spans) < 3 {
		t.Fatalf("want ≥3 spans (bold, and_italic), got %d", len(spans))
	}
	boldFound := false
	italicFound := false
	for _, sp := range spans {
		if sp.Style.Bold {
			boldFound = true
		}
		if sp.Style.Italic {
			italicFound = true
		}
	}
	if !boldFound {
		t.Error("no bold span found in cell content")
	}
	if !italicFound {
		t.Error("no italic span found in cell content")
	}
}

func TestParseTable_BrInCell(t *testing.T) {
	input := `<table><tr><td>Line 1<br>Line 2<br/>Line 3</td></tr></table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	full := spansText(ht.Rows[0].Cells[0].Spans)
	if !strings.Contains(full, "\n") {
		t.Errorf("expected newlines from <br>, got: %q", full)
	}
	lines := strings.Split(full, "\n")
	if len(lines) < 3 {
		t.Errorf("want ≥3 lines, got %d: %q", len(lines), full)
	}
}

// ─── Class forwarding ─────────────────────────────────────────────────────────

func TestParseTable_ClassForwarding(t *testing.T) {
	cs := ClassStyle{
		"hi": {Bold: true},
	}
	input := `<table><tr><td><span class="hi">text</span></td></tr></table>`

	ht, err := ParseTable(input, cs)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	spans := ht.Rows[0].Cells[0].Spans
	if len(spans) == 0 {
		t.Fatal("no spans")
	}
	if !spans[0].Style.Bold {
		t.Error("class hi should make span bold")
	}
}

// ─── Error cases ──────────────────────────────────────────────────────────────

func TestParseTable_NoTable(t *testing.T) {
	_, err := ParseTable("<div>not a table</div>", nil)
	if err == nil {
		t.Error("expected error for input without <table>")
	}
}

func TestParseTable_Empty(t *testing.T) {
	ht, err := ParseTable("<table></table>", nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if len(ht.Rows) != 0 {
		t.Errorf("want 0 rows, got %d", len(ht.Rows))
	}
}

// ─── Nesting ─────────────────────────────────────────────────────────────────

func TestParseTable_NestedTableSkipped(t *testing.T) {
	// A nested <table> inside a cell should not add extra rows to the outer table.
	input := `<table>
		<tr><td>outer cell with <table><tr><td>inner</td></tr></table> more</td></tr>
		<tr><td>second outer row</td></tr>
	</table>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	// Outer table should have exactly 2 rows.
	if len(ht.Rows) != 2 {
		t.Errorf("want 2 rows, got %d", len(ht.Rows))
	}
}

// ─── Case insensitivity ──────────────────────────────────────────────────────

func TestParseTable_CaseInsensitive(t *testing.T) {
	input := `<TABLE>
		<TR><TH>Col</TH></TR>
		<TR><TD>Val</TD></TR>
	</TABLE>`

	ht, err := ParseTable(input, nil)
	if err != nil {
		t.Fatalf("ParseTable: %v", err)
	}
	if len(ht.Rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(ht.Rows))
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// text returns the concatenated plain text from all spans in a cell.
func text(c HtmlCell) string {
	return strings.TrimSpace(spansText(c.Spans))
}

func spansText(spans []Span) string {
	var b strings.Builder
	for _, sp := range spans {
		b.WriteString(sp.Text)
	}
	return b.String()
}
