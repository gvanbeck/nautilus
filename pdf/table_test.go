package pdf

import (
	"testing"
)

// ─── helpers ───────────────────────────────────────────────────────────────

// newTestDoc creates a Document with the system font registered as "regular"
// and "bold". The test is skipped when no font is found.
func newTestTableDoc(t *testing.T) *Document {
	t.Helper()
	doc, err := New(Config{
		PageSize:         PageSizeA4,
		DefaultFontSize:  12,
		LineHeightFactor: 1.2,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	font := systemFont(t) // defined in document_test.go
	if err := doc.RegisterFont("regular", font); err != nil {
		t.Fatalf("RegisterFont regular: %v", err)
	}
	if err := doc.RegisterFont("bold", font); err != nil {
		t.Fatalf("RegisterFont bold: %v", err)
	}
	doc.AddPage()
	if err := doc.SetFont("regular", 12); err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	return doc
}

// simpleTable builds a 3-column, 3-row table with minimal config.
func simpleTableCfg() TableConfig {
	return TableConfig{
		X: 50, Y: 60,
		ColWidths: []float64{100, 200, 100},
		DefaultCellStyle: CellStyle{
			Padding: UniformPadding(4),
			Border: NewUniformBorder(BorderSpec{
				Thickness: 0.5,
				Color:     ColorLightGray,
			}),
		},
		PageBottom: PageSizeA4.Height - 60,
	}
}

// ─── resolveGrid ───────────────────────────────────────────────────────────

func TestResolveGrid_simple(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	tbl.AddRow(Row{Cells: []Cell{{Text: "A"}, {Text: "B"}, {Text: "C"}}})
	tbl.AddRow(Row{Cells: []Cell{{Text: "D"}, {Text: "E"}, {Text: "F"}}})

	placed, err := tbl.resolveGrid()
	if err != nil {
		t.Fatalf("resolveGrid: %v", err)
	}
	if len(placed) != 6 {
		t.Errorf("expected 6 placed cells, got %d", len(placed))
	}
}

func TestResolveGrid_colspan(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	// Row 0: one cell spanning all 3 columns.
	tbl.AddRow(Row{Cells: []Cell{{Text: "Merged", ColSpan: 3}}})
	// Row 1: three normal cells.
	tbl.AddRow(Row{Cells: []Cell{{Text: "A"}, {Text: "B"}, {Text: "C"}}})

	placed, err := tbl.resolveGrid()
	if err != nil {
		t.Fatalf("resolveGrid: %v", err)
	}
	if len(placed) != 4 {
		t.Errorf("expected 4 placed cells, got %d", len(placed))
	}
	// The first placed cell must cover all 3 columns.
	if placed[0].colSpan != 3 {
		t.Errorf("expected colSpan 3, got %d", placed[0].colSpan)
	}
	// Width must equal sum of all three column widths.
	want := 100.0 + 200.0 + 100.0
	if placed[0].width != want {
		t.Errorf("expected width %.0f, got %.0f", want, placed[0].width)
	}
}

func TestResolveGrid_rowspan(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	// Row 0: cell in column 0 spans 2 rows.
	tbl.AddRow(Row{Cells: []Cell{{Text: "Tall", RowSpan: 2}, {Text: "R0C1"}, {Text: "R0C2"}}})
	// Row 1: column 0 is occupied by the rowspan — only supply columns 1 and 2.
	tbl.AddRow(Row{Cells: []Cell{{Text: "R1C1"}, {Text: "R1C2"}}})

	placed, err := tbl.resolveGrid()
	if err != nil {
		t.Fatalf("resolveGrid: %v", err)
	}
	if len(placed) != 5 {
		t.Errorf("expected 5 placed cells, got %d", len(placed))
	}
	// First cell starts at row 0, col 0 with rowSpan 2.
	if placed[0].row != 0 || placed[0].col != 0 || placed[0].rowSpan != 2 {
		t.Errorf("first cell: want row=0 col=0 rowSpan=2, got row=%d col=%d rowSpan=%d",
			placed[0].row, placed[0].col, placed[0].rowSpan)
	}
}

func TestResolveGrid_colspanExceedsColumns_error(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	tbl.AddRow(Row{Cells: []Cell{{Text: "X", ColSpan: 5}}}) // only 3 columns exist

	_, err := tbl.resolveGrid()
	if err == nil {
		t.Error("expected error for colspan exceeding column count, got nil")
	}
}

func TestResolveGrid_rowspanExceedsRows_error(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	tbl.AddRow(Row{Cells: []Cell{{Text: "X", RowSpan: 5}, {Text: "B"}, {Text: "C"}}})

	_, err := tbl.resolveGrid()
	if err == nil {
		t.Error("expected error for rowspan exceeding row count, got nil")
	}
}

// ─── buildRowGroups ────────────────────────────────────────────────────────

func TestBuildRowGroups_noSpan(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	tbl.AddRow(Row{Cells: []Cell{{Text: "A"}, {Text: "B"}, {Text: "C"}}})
	tbl.AddRow(Row{Cells: []Cell{{Text: "D"}, {Text: "E"}, {Text: "F"}}})
	tbl.AddRow(Row{Cells: []Cell{{Text: "G"}, {Text: "H"}, {Text: "I"}}})

	placed, err := tbl.resolveGrid()
	if err != nil {
		t.Fatalf("resolveGrid: %v", err)
	}
	groups := tbl.buildRowGroups(placed)
	if len(groups) != 3 {
		t.Errorf("expected 3 groups (one per row), got %d", len(groups))
	}
	for i, g := range groups {
		if g.startRow != i || g.endRow != i {
			t.Errorf("group %d: want start=%d end=%d, got start=%d end=%d",
				i, i, i, g.startRow, g.endRow)
		}
	}
}

func TestBuildRowGroups_withRowspan(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	tbl.AddRow(Row{Cells: []Cell{{Text: "Tall", RowSpan: 2}, {Text: "R0C1"}, {Text: "R0C2"}}})
	tbl.AddRow(Row{Cells: []Cell{{Text: "R1C1"}, {Text: "R1C2"}}})
	tbl.AddRow(Row{Cells: []Cell{{Text: "G"}, {Text: "H"}, {Text: "I"}}})

	placed, err := tbl.resolveGrid()
	if err != nil {
		t.Fatalf("resolveGrid: %v", err)
	}
	groups := tbl.buildRowGroups(placed)
	// Rows 0 and 1 are linked by the rowspan; row 2 is independent.
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].startRow != 0 || groups[0].endRow != 1 {
		t.Errorf("first group: want 0-1, got %d-%d", groups[0].startRow, groups[0].endRow)
	}
	if groups[1].startRow != 2 || groups[1].endRow != 2 {
		t.Errorf("second group: want 2-2, got %d-%d", groups[1].startRow, groups[1].endRow)
	}
}

// ─── resolveRowHeights ─────────────────────────────────────────────────────

func TestResolveRowHeights_fixed(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	tbl.AddRow(Row{Height: 30, Cells: []Cell{{Text: "A"}, {Text: "B"}, {Text: "C"}}})
	tbl.AddRow(Row{Height: 20, Cells: []Cell{{Text: "D"}, {Text: "E"}, {Text: "F"}}})

	placed, err := tbl.resolveGrid()
	if err != nil {
		t.Fatalf("resolveGrid: %v", err)
	}
	heights, err := tbl.resolveRowHeights(placed)
	if err != nil {
		t.Fatalf("resolveRowHeights: %v", err)
	}
	if heights[0] != 30 {
		t.Errorf("row 0: want 30, got %.1f", heights[0])
	}
	if heights[1] != 20 {
		t.Errorf("row 1: want 20, got %.1f", heights[1])
	}
}

func TestResolveRowHeights_auto(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	// Short single-line text — auto height should be at least lineHeight + padding.
	tbl.AddRow(Row{Cells: []Cell{{Text: "Hello"}, {Text: "World"}, {Text: "!"}}})

	placed, err := tbl.resolveGrid()
	if err != nil {
		t.Fatalf("resolveGrid: %v", err)
	}
	heights, err := tbl.resolveRowHeights(placed)
	if err != nil {
		t.Fatalf("resolveRowHeights: %v", err)
	}
	minExpected := doc.lineHeight() + 4 // padding minimum
	if heights[0] < minExpected {
		t.Errorf("auto row height %.1f below expected minimum %.1f", heights[0], minExpected)
	}
}

// ─── resolveStyle ──────────────────────────────────────────────────────────

func TestResolveStyle_cellOverridesDefault(t *testing.T) {
	doc := newTestTableDoc(t)
	red := Color{R: 255}
	blue := Color{B: 255}
	tbl := doc.NewTable(TableConfig{
		ColWidths: []float64{100, 100, 100},
		DefaultCellStyle: CellStyle{
			FontName:   "regular",
			FontSize:   10,
			Background: &red,
		},
	})

	// Cell overrides background with blue.
	cell := Cell{Style: CellStyle{Background: &blue}}
	s := tbl.resolveStyle(cell, nil)
	if s.Background == nil || s.Background.B != 255 {
		t.Error("cell background should override default background")
	}
	// Font inherited from default.
	if s.FontName != "regular" {
		t.Errorf("expected FontName 'regular', got %q", s.FontName)
	}
}

func TestResolveStyle_rowBgFallback(t *testing.T) {
	doc := newTestTableDoc(t)
	rowBg := Color{G: 200}
	tbl := doc.NewTable(TableConfig{ColWidths: []float64{100}})

	// Cell has no background; row background should apply.
	cell := Cell{}
	s := tbl.resolveStyle(cell, &rowBg)
	if s.Background == nil || s.Background.G != 200 {
		t.Error("row background should be used when neither cell nor default has one")
	}
}

func TestResolveStyle_borderMerge(t *testing.T) {
	doc := newTestTableDoc(t)
	defaultBorder := NewUniformBorder(BorderSpec{Thickness: 1, Color: ColorGray})
	topOverride := &BorderSpec{Thickness: 3, Color: ColorRed}

	tbl := doc.NewTable(TableConfig{
		ColWidths:        []float64{100},
		DefaultCellStyle: CellStyle{Border: defaultBorder},
	})

	// Cell overrides only the top border.
	cell := Cell{Style: CellStyle{Border: Border{Top: topOverride}}}
	s := tbl.resolveStyle(cell, nil)

	if s.Border.Top != topOverride {
		t.Error("cell top border should override default top border")
	}
	// Other sides remain from the default.
	if s.Border.Right != defaultBorder.Right {
		t.Error("right border should remain from default")
	}
}

// ─── Draw ──────────────────────────────────────────────────────────────────

func TestDraw_noColumns_error(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(TableConfig{}) // no ColWidths
	tbl.AddRow(Row{Cells: []Cell{{Text: "X"}}})
	if err := tbl.Draw(); err == nil {
		t.Error("expected error when no column widths defined")
	}
}

func TestDraw_emptyTable(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	// No rows — Draw should succeed silently.
	if err := tbl.Draw(); err != nil {
		t.Errorf("Draw on empty table: %v", err)
	}
}

func TestDraw_simple(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	tbl.AddRow(Row{
		Height: 24,
		Cells: []Cell{
			{Text: "Name"},
			{Text: "Description"},
			{Text: "Value"},
		},
	})
	tbl.AddRow(Row{
		Height: 20,
		Cells: []Cell{
			{Text: "Widget"},
			{Text: "A small thing"},
			{Text: "42"},
		},
	})
	if err := tbl.Draw(); err != nil {
		t.Errorf("Draw simple table: %v", err)
	}
}

func TestDraw_withColspan(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	tbl.AddRow(Row{
		Height: 24,
		Cells:  []Cell{{Text: "Header spanning all columns", ColSpan: 3}},
	})
	tbl.AddRow(Row{
		Height: 20,
		Cells:  []Cell{{Text: "A"}, {Text: "B"}, {Text: "C"}},
	})
	if err := tbl.Draw(); err != nil {
		t.Errorf("Draw with colspan: %v", err)
	}
}

func TestDraw_withRowspan(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	tbl.AddRow(Row{
		Height: 24,
		Cells:  []Cell{{Text: "Tall", RowSpan: 2}, {Text: "R0C1"}, {Text: "R0C2"}},
	})
	tbl.AddRow(Row{
		Height: 24,
		Cells:  []Cell{{Text: "R1C1"}, {Text: "R1C2"}},
	})
	if err := tbl.Draw(); err != nil {
		t.Errorf("Draw with rowspan: %v", err)
	}
}

func TestDraw_withBackground(t *testing.T) {
	doc := newTestTableDoc(t)
	navy := ColorNavy
	white := ColorWhite
	tbl := doc.NewTable(simpleTableCfg())
	tbl.AddRow(Row{
		Height: 24,
		Cells: []Cell{
			{Text: "Name", Style: CellStyle{Background: &navy, TextColor: &white}},
			{Text: "Description", Style: CellStyle{Background: &navy, TextColor: &white}},
			{Text: "Value", Style: CellStyle{Background: &navy, TextColor: &white}},
		},
	})
	tbl.AddRow(Row{
		Height: 20,
		Cells:  []Cell{{Text: "A"}, {Text: "B"}, {Text: "C"}},
	})
	if err := tbl.Draw(); err != nil {
		t.Errorf("Draw with background: %v", err)
	}
}

func TestDraw_pageOverflow(t *testing.T) {
	doc := newTestTableDoc(t)
	cfg := simpleTableCfg()
	cfg.PageBottom = 150 // very small page bottom forces overflow quickly

	tbl := doc.NewTable(cfg)
	for i := range 10 {
		tbl.AddRow(Row{
			Height: 24,
			Cells:  []Cell{{Text: "Row"}, {Text: "Number"}, {Text: string(rune('0' + i))}},
		})
	}
	if err := tbl.Draw(); err != nil {
		t.Errorf("Draw with page overflow: %v", err)
	}
	// Table must have added at least one extra page.
	if doc.PageCount() < 2 {
		t.Errorf("expected multiple pages after overflow, got %d", doc.PageCount())
	}
}

func TestDraw_countingMode(t *testing.T) {
	doc := newTestTableDoc(t)
	doc.countingMode = true

	tbl := doc.NewTable(simpleTableCfg())
	for range 5 {
		tbl.AddRow(Row{
			Height: 20,
			Cells:  []Cell{{Text: "A"}, {Text: "B"}, {Text: "C"}},
		})
	}
	// In counting mode Draw should succeed without generating any PDF output.
	if err := tbl.Draw(); err != nil {
		t.Errorf("Draw in counting mode: %v", err)
	}
}

func TestDraw_outerBorder(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(TableConfig{
		X: 50, Y: 60,
		ColWidths: []float64{100, 200},
		Border: NewUniformBorder(BorderSpec{
			Thickness: 1.5,
			Color:     ColorNavy,
		}),
		DefaultCellStyle: CellStyle{Padding: UniformPadding(5)},
		PageBottom:       PageSizeA4.Height - 60,
	})
	tbl.AddRow(Row{Height: 24, Cells: []Cell{{Text: "Left"}, {Text: "Right"}}})
	if err := tbl.Draw(); err != nil {
		t.Errorf("Draw with outer border: %v", err)
	}
}

func TestDraw_autoHeight(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(TableConfig{
		X: 50, Y: 60,
		ColWidths: []float64{200, 200},
		DefaultCellStyle: CellStyle{
			Padding:  UniformPadding(4),
			FontName: "regular",
			FontSize: 10,
		},
		PageBottom: PageSizeA4.Height - 60,
	})
	// Auto-height row: height should be computed from content.
	tbl.AddRow(Row{Cells: []Cell{
		{Text: "Short"},
		{Text: "A longer piece of text that may wrap to multiple lines when the column is narrow"},
	}})
	if err := tbl.Draw(); err != nil {
		t.Errorf("Draw with auto height: %v", err)
	}
}

// ─── colPositions ──────────────────────────────────────────────────────────

func TestColPositions(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(TableConfig{
		X:         100,
		ColWidths: []float64{50, 100, 75},
	})
	pos := tbl.colPositions()
	want := []float64{100, 150, 250}
	for i, w := range want {
		if pos[i] != w {
			t.Errorf("col %d: want %.0f, got %.0f", i, w, pos[i])
		}
	}
}

// ─── sumF helper ───────────────────────────────────────────────────────────

func TestSumF(t *testing.T) {
	if v := sumF([]float64{1, 2, 3}); v != 6 {
		t.Errorf("sumF: want 6, got %.1f", v)
	}
	if v := sumF(nil); v != 0 {
		t.Errorf("sumF nil: want 0, got %.1f", v)
	}
}

// ─── effectiveFont ─────────────────────────────────────────────────────────

func TestEffectiveFont_fallback(t *testing.T) {
	doc := newTestTableDoc(t)
	doc.currentFont = "regular"
	doc.fontSize = 12

	tbl := doc.NewTable(TableConfig{ColWidths: []float64{100}})
	// Style with no font info — should fall back to doc state.
	name, size := tbl.effectiveFont(CellStyle{})
	if name != "regular" {
		t.Errorf("expected 'regular', got %q", name)
	}
	if size != 12 {
		t.Errorf("expected 12, got %.1f", size)
	}
}

func TestEffectiveFont_cellOverride(t *testing.T) {
	doc := newTestTableDoc(t)
	doc.currentFont = "regular"
	doc.fontSize = 12

	tbl := doc.NewTable(TableConfig{ColWidths: []float64{100}})
	name, size := tbl.effectiveFont(CellStyle{FontName: "bold", FontSize: 16})
	if name != "bold" {
		t.Errorf("expected 'bold', got %q", name)
	}
	if size != 16 {
		t.Errorf("expected 16, got %.1f", size)
	}
}

// ─── Alignment ─────────────────────────────────────────────────────────────

func TestDraw_hAlignRight(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	tbl.AddRow(Row{
		Height: 24,
		Cells: []Cell{
			{Text: "Left"},
			{Text: "Centre", Style: CellStyle{HAlign: HAlignCenter}},
			{Text: "Right", Style: CellStyle{HAlign: HAlignRight}},
		},
	})
	if err := tbl.Draw(); err != nil {
		t.Errorf("Draw with HAlign: %v", err)
	}
}

func TestDraw_vAlignMiddleBottom(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	tbl.AddRow(Row{
		Height: 60,
		Cells: []Cell{
			{Text: "Top"},
			{Text: "Middle", Style: CellStyle{VAlign: VAlignMiddle}},
			{Text: "Bottom", Style: CellStyle{VAlign: VAlignBottom}},
		},
	})
	if err := tbl.Draw(); err != nil {
		t.Errorf("Draw with VAlign: %v", err)
	}
}

func TestDraw_tableDefaultAlignment(t *testing.T) {
	doc := newTestTableDoc(t)
	// Table default: centre + middle; one cell overrides to right + bottom.
	tbl := doc.NewTable(TableConfig{
		X: 50, Y: 60,
		ColWidths: []float64{150, 150, 150},
		DefaultCellStyle: CellStyle{
			Padding:  UniformPadding(4),
			FontName: "regular",
			FontSize: 10,
			HAlign:   HAlignCenter,
			VAlign:   VAlignMiddle,
		},
		PageBottom: PageSizeA4.Height - 60,
	})
	tbl.AddRow(Row{
		Height: 48,
		Cells: []Cell{
			{Text: "Centre/Middle"},  // inherits table default
			{Text: "Centre/Middle"}, // same
			{Text: "Right/Bottom", Style: CellStyle{
				HAlign: HAlignRight,
				VAlign: VAlignBottom,
			}},
		},
	})
	if err := tbl.Draw(); err != nil {
		t.Errorf("Draw with table default alignment: %v", err)
	}
}

func TestResolveStyle_hAlignInherit(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(TableConfig{
		ColWidths:        []float64{100},
		DefaultCellStyle: CellStyle{HAlign: HAlignRight},
	})
	// Cell with HAlignDefault should inherit HAlignRight from table.
	s := tbl.resolveStyle(Cell{}, nil)
	if s.HAlign != HAlignRight {
		t.Errorf("expected HAlignRight inherited from default, got %v", s.HAlign)
	}
}

func TestResolveStyle_hAlignOverride(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(TableConfig{
		ColWidths:        []float64{100},
		DefaultCellStyle: CellStyle{HAlign: HAlignRight},
	})
	// Cell explicitly sets HAlignCenter — should override default.
	cell := Cell{Style: CellStyle{HAlign: HAlignCenter}}
	s := tbl.resolveStyle(cell, nil)
	if s.HAlign != HAlignCenter {
		t.Errorf("expected HAlignCenter from cell override, got %v", s.HAlign)
	}
}

func TestResolveStyle_vAlignInherit(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(TableConfig{
		ColWidths:        []float64{100},
		DefaultCellStyle: CellStyle{VAlign: VAlignBottom},
	})
	s := tbl.resolveStyle(Cell{}, nil)
	if s.VAlign != VAlignBottom {
		t.Errorf("expected VAlignBottom inherited, got %v", s.VAlign)
	}
}

// ─── AddRow / AddRows chaining ─────────────────────────────────────────────

func TestAddRowChaining(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	result := tbl.
		AddRow(Row{Cells: []Cell{{Text: "A"}, {Text: "B"}, {Text: "C"}}}).
		AddRow(Row{Cells: []Cell{{Text: "D"}, {Text: "E"}, {Text: "F"}}})
	if result != tbl {
		t.Error("AddRow should return the same Table pointer")
	}
	if len(tbl.rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(tbl.rows))
	}
}

func TestAddRows(t *testing.T) {
	doc := newTestTableDoc(t)
	tbl := doc.NewTable(simpleTableCfg())
	tbl.AddRows(
		Row{Cells: []Cell{{Text: "A"}, {Text: "B"}, {Text: "C"}}},
		Row{Cells: []Cell{{Text: "D"}, {Text: "E"}, {Text: "F"}}},
		Row{Cells: []Cell{{Text: "G"}, {Text: "H"}, {Text: "I"}}},
	)
	if len(tbl.rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(tbl.rows))
	}
}
