package pdf

import (
	"fmt"
	"strings"
)

// ─── Alignment ─────────────────────────────────────────────────────────────

// HAlign specifies the horizontal alignment of text within a cell.
type HAlign int

const (
	// HAlignDefault inherits the alignment from the table's DefaultCellStyle.
	// When the table default is also HAlignDefault, text is left-aligned.
	HAlignDefault HAlign = iota
	// HAlignLeft aligns text to the left edge of the content area.
	HAlignLeft
	// HAlignCenter centres each line of text horizontally.
	HAlignCenter
	// HAlignRight aligns text to the right edge of the content area.
	// Use this for numeric columns or for a visual right-to-left feel.
	HAlignRight
)

// VAlign specifies the vertical alignment of text within a cell.
type VAlign int

const (
	// VAlignDefault inherits the alignment from the table's DefaultCellStyle.
	// When the table default is also VAlignDefault, text is top-aligned.
	VAlignDefault VAlign = iota
	// VAlignTop places text at the top of the content area (default).
	VAlignTop
	// VAlignMiddle centres text vertically within the cell.
	VAlignMiddle
	// VAlignBottom places text at the bottom of the content area.
	VAlignBottom
)

// ─── Cell style ────────────────────────────────────────────────────────────

// CellStyle defines the visual appearance of a table cell.
// All fields are optional; zero values fall back to the table's
// DefaultCellStyle and ultimately to built-in defaults.
type CellStyle struct {
	// Padding is the inner spacing between the cell edge and its text.
	Padding Padding

	// Border draws lines around the cell.  Each side is merged independently
	// with the table's DefaultCellStyle: a non-nil side in the cell's own
	// Border overrides the corresponding default side.
	Border Border

	// Background fills the cell with a solid colour.
	// When nil the row's Background (if any) is used, then the table default.
	Background *Color

	// TextColor is the colour of the cell text.  Defaults to black when nil.
	TextColor *Color

	// FontName selects a font registered with Document.RegisterFont.
	// If empty the table's DefaultCellStyle.FontName is used.
	FontName string

	// FontSize sets the font size in points.
	// If 0 the table's DefaultCellStyle.FontSize (or the document default) is used.
	FontSize float64

	// HAlign sets the horizontal text alignment within the cell.
	// HAlignDefault (zero value) inherits from the table's DefaultCellStyle.
	// When the table default is also HAlignDefault, text is left-aligned.
	HAlign HAlign

	// VAlign sets the vertical text alignment within the cell.
	// VAlignDefault (zero value) inherits from the table's DefaultCellStyle.
	// When the table default is also VAlignDefault, text is top-aligned.
	VAlign VAlign
}

// ─── Cell ──────────────────────────────────────────────────────────────────

// Cell is a single table cell with text content and optional style overrides.
//
// Cells within a row must collectively fill all table columns, accounting for
// ColSpan values.  Column positions already covered by a rowspan from a
// previous row must be omitted.
type Cell struct {
	// Text is the cell content.  Word wrapping is applied automatically.
	// Explicit \n newlines are honoured as paragraph breaks.
	Text string

	// ColSpan is the number of columns this cell occupies.  Defaults to 1.
	ColSpan int

	// RowSpan is the number of rows this cell occupies.  Defaults to 1.
	// Rows joined by a rowspan are kept together during page overflow
	// (they are never split across a page break).
	RowSpan int

	// Style overrides the table DefaultCellStyle for this cell.
	// Zero-valued fields inherit from the default.
	Style CellStyle
}

// ─── Row ───────────────────────────────────────────────────────────────────

// Row is a single table row.
type Row struct {
	// Cells defines the cell content.  See Cell.ColSpan and Cell.RowSpan for
	// details on spanning.
	Cells []Cell

	// Height is the fixed row height in points.  When 0, the height is
	// calculated automatically from the tallest cell content in the row.
	//
	// Note: auto-height calculation requires an active font.  When using
	// Document.Build, set a fixed Height or ensure fonts are registered
	// and the table's DefaultCellStyle.FontName is set before calling Draw.
	Height float64

	// Background fills all cells in this row that have no per-cell background.
	Background *Color
}

// ─── TableConfig ───────────────────────────────────────────────────────────

// TableConfig configures the position, column layout, and overall appearance
// of a Table.
type TableConfig struct {
	// X, Y is the top-left corner of the table in page coordinates (points).
	X, Y float64

	// ColWidths defines the width of each column in points.
	// The number of elements determines the column count.
	// Column widths are required; there is no auto-sizing.
	ColWidths []float64

	// Border is drawn around the table on each page segment.
	// First-page segment: top + left + right.
	// Middle segments:    left + right only.
	// Last segment:       bottom + left + right.
	// Set to a zero-value Border for no outer border.
	Border Border

	// DefaultCellStyle is applied to every cell before the cell's own Style
	// overrides are merged in.  Set FontName and FontSize here to avoid
	// repeating them on every cell.
	DefaultCellStyle CellStyle

	// PageBottom is the lowest Y coordinate content may occupy before the
	// table flows to the next page.  Typically doc.PageHeight() - bottomMargin.
	// Defaults to doc.PageHeight() - 60 when 0.
	PageBottom float64

	// ContinuationY is the Y position on continuation pages where the table
	// resumes after a page overflow.  Defaults to cfg.Y when 0.
	ContinuationY float64

	// RepeatRows is the number of leading rows to repeat at the top of each
	// continuation page after a page overflow.  Use this for header rows.
	// Defaults to 0 (no repetition).
	RepeatRows int
}

// tableWidth returns the total width of the table (sum of all column widths).
func (cfg *TableConfig) tableWidth() float64 {
	w := 0.0
	for _, cw := range cfg.ColWidths {
		w += cw
	}
	return w
}

// ─── Table ─────────────────────────────────────────────────────────────────

// Table is a grid-based content layout element for PDF pages, similar to an
// HTML table.
//
// Build the table by calling AddRow, then call Draw to render it.
//
// # Column widths
//
// All column widths must be specified explicitly in TableConfig.ColWidths.
// There is no auto-sizing: the caller is responsible for ensuring widths sum
// to the desired table width.
//
// # Cell merging
//
// Horizontal merging (colspan) and vertical merging (rowspan) are both
// supported.  Rows joined by a rowspan are kept together during page overflow.
//
// # Page overflow
//
// When a row group does not fit in the remaining space on the current page,
// the table automatically calls Document.AddPage and continues at
// TableConfig.ContinuationY.  This triggers any registered header/footer
// callbacks.
//
// # Auto row heights and Build
//
// Auto row heights are measured by temporarily setting the cell font and
// measuring wrapped text.  This requires gopdf to have an active font, which
// is only the case during the rendering pass of Build — not the counting pass.
// To use auto-height rows inside Build, ensure all rows have an explicit
// Height, or accept that page-break counting may be approximate.
//
// # Example
//
//	tbl := doc.NewTable(pdf.TableConfig{
//	    X: 50, Y: 100,
//	    ColWidths: []float64{120, 260, 115},
//	    PageBottom: doc.PageHeight() - 60,
//	    DefaultCellStyle: pdf.CellStyle{
//	        Padding:  pdf.UniformPadding(5),
//	        Border:   pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}),
//	        FontName: "regular",
//	        FontSize: 10,
//	    },
//	})
//	// Header row
//	tbl.AddRow(pdf.Row{
//	    Height: 22,
//	    Cells: []pdf.Cell{
//	        {Text: "Name",        Style: pdf.CellStyle{Background: &pdf.ColorNavy, TextColor: &pdf.ColorWhite, FontName: "bold"}},
//	        {Text: "Description", Style: pdf.CellStyle{Background: &pdf.ColorNavy, TextColor: &pdf.ColorWhite}},
//	        {Text: "Value",       Style: pdf.CellStyle{Background: &pdf.ColorNavy, TextColor: &pdf.ColorWhite}},
//	    },
//	})
//	// Data row with rowspan
//	tbl.AddRow(pdf.Row{Cells: []pdf.Cell{
//	    {Text: "Widget", RowSpan: 2},
//	    {Text: "First variant"},
//	    {Text: "42.00"},
//	}})
//	tbl.AddRow(pdf.Row{Cells: []pdf.Cell{
//	    // first column occupied by rowspan above — omit it
//	    {Text: "Second variant"},
//	    {Text: "99.00"},
//	}})
//	if err := tbl.Draw(); err != nil { log.Fatal(err) }
type Table struct {
	doc      *Document
	cfg      TableConfig
	rows     []Row
	currentY float64 // Y position immediately below the last drawn row
}

// CurrentY returns the Y coordinate immediately below the bottom of the table
// after Draw has been called.  Use this to position content that follows the
// table (e.g. a caption).  Returns 0 before Draw is called.
func (t *Table) CurrentY() float64 { return t.currentY }

// NewTable creates a new Table.  Add rows with AddRow, then call Draw.
func (d *Document) NewTable(cfg TableConfig) *Table {
	return &Table{doc: d, cfg: cfg}
}

// AddRow appends row to the table.  Returns the Table to allow chaining.
func (t *Table) AddRow(row Row) *Table {
	t.rows = append(t.rows, row)
	return t
}

// AddRows appends multiple rows at once.  Returns the Table.
func (t *Table) AddRows(rows ...Row) *Table {
	t.rows = append(t.rows, rows...)
	return t
}

// ─── Rendering ─────────────────────────────────────────────────────────────

// Draw renders the table to the document starting at TableConfig.X/Y.
//
// Draw calls Document.AddPage when a row group does not fit on the current
// page, which triggers any registered header/footer callbacks.
//
// During the counting pass of Build, Draw simulates page breaks without
// rendering any content.
func (t *Table) Draw() error {
	if len(t.cfg.ColWidths) == 0 {
		return fmt.Errorf("pdf: table: no column widths defined")
	}
	if len(t.rows) == 0 {
		return nil
	}

	// ── Step 1: resolve the cell grid ───────────────────────────────────
	placed, err := t.resolveGrid()
	if err != nil {
		return err
	}

	// ── Step 2: compute column x positions ──────────────────────────────
	colX := t.colPositions()

	// ── Step 3: resolve row heights ─────────────────────────────────────
	rowH, err := t.resolveRowHeights(placed)
	if err != nil {
		return err
	}

	// ── Step 4: build row groups (must stay on same page) ───────────────
	groups := t.buildRowGroups(placed)

	// ── Step 5: paginate and render ──────────────────────────────────────
	contY := t.continuationY()
	pageBottom := t.bottomLimit()
	curY := t.cfg.Y
	segTopY := curY // top Y of the current page segment (for outer border)
	isFirst := true

	for _, grp := range groups {
		grpH := sumF(rowH[grp.startRow : grp.endRow+1])

		// Page overflow: start new page when group doesn't fit and we're not
		// already at the top of the page.
		if curY+grpH > pageBottom && curY > segTopY {
			// Draw outer border for the segment we're closing.
			if err := t.drawSegmentBorder(segTopY, curY-segTopY, isFirst, false); err != nil {
				return err
			}
			t.doc.AddPage()
			curY = contY
			segTopY = curY
			isFirst = false

			// Re-render header rows on the new page.
			if rr := t.cfg.RepeatRows; rr > 0 && rr <= len(t.rows) && !t.doc.countingMode {
				repeatH := sumF(rowH[0:rr])
				if err := t.renderRowRange(placed, rowH, colX, curY, 0, rr-1); err != nil {
					return err
				}
				curY += repeatH
			}
		}

		if t.doc.countingMode {
			// Counting pass: only simulate page breaks, no drawing.
			curY += grpH
			continue
		}

		// Compute absolute Y per row in this group.
		rowY := make([]float64, grp.endRow-grp.startRow+1)
		y := curY
		for i := range rowY {
			rowY[i] = y
			y += rowH[grp.startRow+i]
		}

		// Cells that start within this group.
		var groupCells []placedCell
		for _, pc := range placed {
			if pc.row >= grp.startRow && pc.row <= grp.endRow {
				groupCells = append(groupCells, pc)
			}
		}

		// ── Pass 1: backgrounds ──────────────────────────────────────
		for _, pc := range groupCells {
			style := t.resolveStyle(pc.cell, t.rows[pc.row].Background)
			if style.Background == nil {
				continue
			}
			cellY := rowY[pc.row-grp.startRow]
			cellH := sumF(rowH[pc.row : pc.row+pc.rowSpan])
			t.doc.pdf.SetFillColor(style.Background.R, style.Background.G, style.Background.B)
			t.doc.pdf.RectFromUpperLeftWithStyle(colX[pc.col], cellY, pc.width, cellH, "F")
		}

		// ── Pass 2: text ─────────────────────────────────────────────
		for _, pc := range groupCells {
			style := t.resolveStyle(pc.cell, t.rows[pc.row].Background)

			cellY := rowY[pc.row-grp.startRow]
			cellH := sumF(rowH[pc.row : pc.row+pc.rowSpan])
			contentX := colX[pc.col] + style.Padding.Left
			contentY := cellY + style.Padding.Top
			contentW := pc.width - style.Padding.Left - style.Padding.Right

			// Apply font.
			if fn, fs := t.effectiveFont(style); fn != "" {
				t.doc.SetFont(fn, fs) //nolint:errcheck
			}

			// Apply text color.
			if style.TextColor != nil {
				t.doc.SetTextColor(style.TextColor.R, style.TextColor.G, style.TextColor.B)
			} else {
				t.doc.SetTextColor(0, 0, 0)
			}

			if pc.cell.Text != "" && contentW > 0 {
				if err := t.renderCellText(pc.cell.Text, contentX, contentY, contentW, cellH, style); err != nil {
					return err
				}
			}
		}

		// ── Pass 3: cell borders ─────────────────────────────────────
		for _, pc := range groupCells {
			style := t.resolveStyle(pc.cell, t.rows[pc.row].Background)
			cellY := rowY[pc.row-grp.startRow]
			cellH := sumF(rowH[pc.row : pc.row+pc.rowSpan])
			if err := t.doc.DrawBorder(colX[pc.col], cellY, pc.width, cellH, style.Border); err != nil {
				return err
			}
		}

		curY += grpH
	}

	// ── Step 6: outer border for last (or only) segment ─────────────────
	if !t.doc.countingMode && curY > segTopY {
		if err := t.drawSegmentBorder(segTopY, curY-segTopY, isFirst, true); err != nil {
			return err
		}
	}

	t.currentY = curY
	return nil
}

// ─── Internal helpers ──────────────────────────────────────────────────────

// placedCell is a cell with its resolved grid position and geometry.
type placedCell struct {
	row, col int
	rowSpan  int
	colSpan  int
	cell     Cell
	width    float64 // sum of column widths for colSpan columns
}

// rowGroup is a set of consecutive rows that must land on the same page.
type rowGroup struct {
	startRow, endRow int // inclusive
}

// resolveGrid builds the occupancy grid and returns all placed cells.
func (t *Table) resolveGrid() ([]placedCell, error) {
	numCols := len(t.cfg.ColWidths)
	numRows := len(t.rows)

	// occupied[r][c] = true when that position is claimed by a spanning cell.
	occupied := make([][]bool, numRows)
	for i := range occupied {
		occupied[i] = make([]bool, numCols)
	}

	var placed []placedCell

	for rowIdx, row := range t.rows {
		col := 0
		for _, cell := range row.Cells {
			// Advance past positions occupied by rowspan cells from above.
			for col < numCols && occupied[rowIdx][col] {
				col++
			}
			if col >= numCols {
				return nil, fmt.Errorf("pdf: table row %d: more cells than columns (%d)", rowIdx, numCols)
			}

			cs := cell.ColSpan
			rs := cell.RowSpan
			if cs <= 0 {
				cs = 1
			}
			if rs <= 0 {
				rs = 1
			}
			if col+cs > numCols {
				return nil, fmt.Errorf("pdf: table row %d, col %d: colspan %d exceeds column count %d",
					rowIdx, col, cs, numCols)
			}
			if rowIdx+rs > numRows {
				return nil, fmt.Errorf("pdf: table row %d, col %d: rowspan %d exceeds row count %d",
					rowIdx, col, rs, numRows)
			}

			// Mark all spanned positions as occupied.
			for dr := 0; dr < rs; dr++ {
				for dc := 0; dc < cs; dc++ {
					occupied[rowIdx+dr][col+dc] = true
				}
			}

			// Calculate cell width.
			w := 0.0
			for dc := 0; dc < cs; dc++ {
				w += t.cfg.ColWidths[col+dc]
			}

			placed = append(placed, placedCell{
				row: rowIdx, col: col,
				rowSpan: rs, colSpan: cs,
				cell:  cell,
				width: w,
			})
			col += cs
		}
	}

	return placed, nil
}

// buildRowGroups returns the minimal set of row groups that must not be split
// across page breaks (due to rowspan constraints).
func (t *Table) buildRowGroups(placed []placedCell) []rowGroup {
	n := len(t.rows)
	if n == 0 {
		return nil
	}

	// groupEnd[i] = the latest row that row i must stay with on the same page.
	groupEnd := make([]int, n)
	for i := range groupEnd {
		groupEnd[i] = i
	}
	for _, pc := range placed {
		if pc.rowSpan > 1 {
			end := pc.row + pc.rowSpan - 1
			for r := pc.row; r <= end; r++ {
				if end > groupEnd[r] {
					groupEnd[r] = end
				}
			}
		}
	}
	// Propagate transitively.
	for changed := true; changed; {
		changed = false
		for i := range groupEnd {
			if groupEnd[groupEnd[i]] > groupEnd[i] {
				groupEnd[i] = groupEnd[groupEnd[i]]
				changed = true
			}
		}
	}

	var groups []rowGroup
	for i := 0; i < n; {
		end := groupEnd[i]
		groups = append(groups, rowGroup{startRow: i, endRow: end})
		i = end + 1
	}
	return groups
}

// resolveRowHeights computes the height of every row.
// Fixed rows use Row.Height; auto rows are measured from cell content.
func (t *Table) resolveRowHeights(placed []placedCell) ([]float64, error) {
	heights := make([]float64, len(t.rows))

	// Initialize with fixed heights.
	for i, row := range t.rows {
		heights[i] = row.Height
	}

	// In counting mode, gopdf has no active font so MeasureTextWidth would
	// fail.  Use a one-line estimate so pagination is approximate but valid.
	if t.doc.countingMode {
		for i := range heights {
			if heights[i] == 0 {
				style := t.resolveStyle(Cell{}, t.rows[i].Background)
				lh := t.doc.lineHeight()
				heights[i] = lh + style.Padding.Top + style.Padding.Bottom
			}
		}
		return heights, nil
	}

	// Pass 1: cells with rowSpan == 1 drive individual row heights.
	// Fixed-height rows (Row.Height > 0) are left unchanged; only auto rows
	// are sized by their content.
	for _, pc := range placed {
		if pc.rowSpan != 1 {
			continue
		}
		if t.rows[pc.row].Height > 0 {
			continue // row has a fixed height — honour it exactly
		}
		h := t.measureCellHeight(pc)
		if h > heights[pc.row] {
			heights[pc.row] = h
		}
	}

	// Pass 2: cells with rowSpan > 1 may require extra height distributed
	// over the last of the spanned rows — only when all spanned rows are auto.
	for _, pc := range placed {
		if pc.rowSpan <= 1 {
			continue
		}
		// If any of the spanned rows has a fixed height, skip expansion.
		allAuto := true
		for r := pc.row; r < pc.row+pc.rowSpan; r++ {
			if t.rows[r].Height > 0 {
				allAuto = false
				break
			}
		}
		if !allAuto {
			continue
		}
		needed := t.measureCellHeight(pc)
		current := sumF(heights[pc.row : pc.row+pc.rowSpan])
		if needed > current {
			// Add the deficit to the last spanned row (deterministic).
			heights[pc.row+pc.rowSpan-1] += needed - current
		}
	}

	// Guarantee a sensible minimum for every auto-height row.
	minH := t.doc.lineHeight() + 4
	for i := range heights {
		if t.rows[i].Height > 0 {
			continue // fixed — do not alter
		}
		if heights[i] < minH {
			heights[i] = minH
		}
	}

	return heights, nil
}

// measureCellHeight returns the required outer height (including padding) for pc.
func (t *Table) measureCellHeight(pc placedCell) float64 {
	style := t.resolveStyle(pc.cell, t.rows[pc.row].Background)
	contentW := pc.width - style.Padding.Left - style.Padding.Right
	vPad := style.Padding.Top + style.Padding.Bottom

	if pc.cell.Text == "" || contentW <= 0 {
		return t.doc.lineHeight() + vPad
	}

	// Temporarily set the cell font so MeasureTextWidth has valid metrics.
	fn, fs := t.effectiveFont(style)
	if fn != "" {
		t.doc.SetFont(fn, fs) //nolint:errcheck
	}

	n, err := t.doc.measureLines(pc.cell.Text, contentW)
	if err != nil || n == 0 {
		n = strings.Count(pc.cell.Text, "\n") + 1
	}
	return float64(n)*t.doc.lineHeight() + vPad
}

// resolveStyle merges the cell's own CellStyle with the table DefaultCellStyle.
// The row background is considered after the cell's own background.
func (t *Table) resolveStyle(cell Cell, rowBg *Color) CellStyle {
	s := t.cfg.DefaultCellStyle // start from table defaults

	cs := cell.Style

	// Merge scalar fields: cell wins over default when non-zero/non-nil.
	if cs.FontName != "" {
		s.FontName = cs.FontName
	}
	if cs.FontSize > 0 {
		s.FontSize = cs.FontSize
	}
	if cs.TextColor != nil {
		s.TextColor = cs.TextColor
	}
	if cs.Background != nil {
		s.Background = cs.Background
	} else if s.Background == nil {
		s.Background = rowBg // row background as fallback
	}
	if cs.Padding != (Padding{}) {
		s.Padding = cs.Padding
	}

	// Merge border per side so a cell can override individual sides.
	if cs.Border.Top != nil {
		s.Border.Top = cs.Border.Top
	}
	if cs.Border.Right != nil {
		s.Border.Right = cs.Border.Right
	}
	if cs.Border.Bottom != nil {
		s.Border.Bottom = cs.Border.Bottom
	}
	if cs.Border.Left != nil {
		s.Border.Left = cs.Border.Left
	}

	// Alignment: HAlignDefault / VAlignDefault (0) means "inherit from table
	// default"; any other value explicitly overrides it.
	if cs.HAlign != HAlignDefault {
		s.HAlign = cs.HAlign
	}
	if cs.VAlign != VAlignDefault {
		s.VAlign = cs.VAlign
	}

	return s
}

// effectiveFont returns the font name and size to use for style, falling back
// to the document's currently active font when style fields are zero.
func (t *Table) effectiveFont(style CellStyle) (name string, size float64) {
	name = style.FontName
	size = style.FontSize
	if name == "" {
		name = t.doc.currentFont
	}
	if size == 0 {
		size = t.doc.fontSize
	}
	return
}

// colPositions returns the absolute x position of each column's left edge.
func (t *Table) colPositions() []float64 {
	pos := make([]float64, len(t.cfg.ColWidths))
	x := t.cfg.X
	for i, w := range t.cfg.ColWidths {
		pos[i] = x
		x += w
	}
	return pos
}

// bottomLimit returns the Y coordinate below which no content may be placed
// on a single page before overflowing to the next.
func (t *Table) bottomLimit() float64 {
	if t.cfg.PageBottom > 0 {
		return t.cfg.PageBottom
	}
	return t.doc.PageHeight() - 60
}

// continuationY returns the Y position on subsequent pages.
func (t *Table) continuationY() float64 {
	if t.cfg.ContinuationY > 0 {
		return t.cfg.ContinuationY
	}
	return t.cfg.Y
}

// drawSegmentBorder draws the outer border rectangle for one page segment.
// isFirst and isLast control which sides of the table Border are drawn:
//   - First segment only: Top side.
//   - Last segment only:  Bottom side.
//   - Every segment:      Left and Right sides.
func (t *Table) drawSegmentBorder(topY, height float64, isFirst, isLast bool) error {
	b := t.cfg.Border
	seg := Border{}
	if isFirst {
		seg.Top = b.Top
	}
	if isLast {
		seg.Bottom = b.Bottom
	}
	seg.Left = b.Left
	seg.Right = b.Right

	return t.doc.DrawBorder(t.cfg.X, topY, t.cfg.tableWidth(), height, seg)
}

// renderRowRange renders the cells in rows [startRow, endRow] (inclusive) at
// the given topY position.  Used to re-draw header rows after page overflow.
func (t *Table) renderRowRange(placed []placedCell, rowH []float64, colX []float64, topY float64, startRow, endRow int) error {
	// Collect cells that start within the range.
	var cells []placedCell
	for _, pc := range placed {
		if pc.row >= startRow && pc.row <= endRow {
			cells = append(cells, pc)
		}
	}

	// Build per-row Y positions relative to topY.
	rowY := make(map[int]float64, endRow-startRow+1)
	y := topY
	for r := startRow; r <= endRow; r++ {
		rowY[r] = y
		y += rowH[r]
	}

	// Pass 1: backgrounds.
	for _, pc := range cells {
		style := t.resolveStyle(pc.cell, t.rows[pc.row].Background)
		if style.Background == nil {
			continue
		}
		cellH := sumF(rowH[pc.row : pc.row+pc.rowSpan])
		t.doc.pdf.SetFillColor(style.Background.R, style.Background.G, style.Background.B)
		t.doc.pdf.RectFromUpperLeftWithStyle(colX[pc.col], rowY[pc.row], pc.width, cellH, "F")
	}
	// Pass 2: text.
	for _, pc := range cells {
		style := t.resolveStyle(pc.cell, t.rows[pc.row].Background)
		cellH := sumF(rowH[pc.row : pc.row+pc.rowSpan])
		contentX := colX[pc.col] + style.Padding.Left
		contentY := rowY[pc.row] + style.Padding.Top
		contentW := pc.width - style.Padding.Left - style.Padding.Right
		if fn, fs := t.effectiveFont(style); fn != "" {
			t.doc.SetFont(fn, fs) //nolint:errcheck
		}
		if style.TextColor != nil {
			t.doc.SetTextColor(style.TextColor.R, style.TextColor.G, style.TextColor.B)
		} else {
			t.doc.SetTextColor(0, 0, 0)
		}
		if pc.cell.Text != "" && contentW > 0 {
			if err := t.renderCellText(pc.cell.Text, contentX, contentY, contentW, cellH, style); err != nil {
				return err
			}
		}
	}
	// Pass 3: borders.
	for _, pc := range cells {
		style := t.resolveStyle(pc.cell, t.rows[pc.row].Background)
		cellH := sumF(rowH[pc.row : pc.row+pc.rowSpan])
		if err := t.doc.DrawBorder(colX[pc.col], rowY[pc.row], pc.width, cellH, style.Border); err != nil {
			return err
		}
	}
	return nil
}

// renderCellText renders text within a cell's content area, applying both
// horizontal and vertical alignment.
//
//   - contentX, contentY: top-left corner of the content area (after padding).
//   - contentW: available horizontal space for text (cell width minus h-padding).
//   - cellH: total outer cell height (including padding), used for VAlign.
//   - style: resolved cell style, used for alignment and padding values.
func (t *Table) renderCellText(text string, contentX, contentY, contentW, cellH float64, style CellStyle) error {
	// Word-wrap: split explicit newlines first, then wrap each paragraph.
	var lines []string
	for _, para := range strings.Split(text, "\n") {
		wrapped, err := t.doc.wrapLine(para, contentW)
		if err != nil {
			return err
		}
		lines = append(lines, wrapped...)
	}
	if len(lines) == 0 {
		return nil
	}

	lh := t.doc.lineHeight()
	totalH := float64(len(lines)) * lh

	// ── Vertical start position ───────────────────────────────────────────
	// contentY is already offset by padding.Top; available space is the inner
	// height (cellH minus both vertical padding values).
	startY := contentY // VAlignTop / VAlignDefault
	switch style.VAlign {
	case VAlignMiddle:
		innerH := cellH - style.Padding.Top - style.Padding.Bottom
		startY = contentY + (innerH-totalH)/2
	case VAlignBottom:
		innerH := cellH - style.Padding.Top - style.Padding.Bottom
		startY = contentY + innerH - totalH
	}

	// ── Render each line with horizontal alignment ────────────────────────
	for _, line := range lines {
		lineX := contentX // HAlignLeft / HAlignDefault
		switch style.HAlign {
		case HAlignCenter, HAlignRight:
			w, err := t.doc.measureWord(line)
			if err != nil {
				return fmt.Errorf("pdf: table cell measure line: %w", err)
			}
			if style.HAlign == HAlignCenter {
				lineX = contentX + (contentW-w)/2
			} else {
				lineX = contentX + contentW - w
			}
		}
		if _, err := t.doc.WriteLine(line, lineX, startY); err != nil {
			return err
		}
		startY += lh
	}
	return nil
}

// sumF returns the sum of a float64 slice.
func sumF(values []float64) float64 {
	s := 0.0
	for _, v := range values {
		s += v
	}
	return s
}
