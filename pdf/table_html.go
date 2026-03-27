package pdf

import (
	"strconv"
	"strings"

	htmlpkg "github.com/gvanbeck/nautilus/pdf/html"
)

// ─── Options ────────────────────────────────────────────────────────────────

// HtmlTableOptions controls how a parsed HtmlTable is converted to a pdf.Table.
type HtmlTableOptions struct {
	// SpanFontFor maps an inline html.Style to a registered font name.
	// It is called for every text span inside a cell so you can return
	// a bold, italic, or monospace font as appropriate.
	// When nil, all spans fall back to the cell's effective FontName.
	//
	// Example:
	//   opts.SpanFontFor = func(s htmlpkg.Style) string {
	//       switch {
	//       case s.Bold && s.Italic: return "bolditalic"
	//       case s.Bold:             return "bold"
	//       case s.Italic:           return "italic"
	//       case s.Monospace:        return "mono"
	//       default:                 return "regular"
	//       }
	//   }
	SpanFontFor func(style htmlpkg.Style) string

	// HeaderStyle is applied to cells that belong to a header row
	// (<thead> section or a row where every cell is <th>).
	// Fields left at their zero value inherit from TableConfig.DefaultCellStyle.
	HeaderStyle CellStyle

	// FooterStyle is applied to cells that belong to a footer row
	// (<tfoot> section).
	// Fields left at their zero value inherit from TableConfig.DefaultCellStyle.
	FooterStyle CellStyle
}

// ─── Conversion ─────────────────────────────────────────────────────────────

// TableFromHTML converts a parsed HtmlTable into a pdf.Table ready to draw.
//
// cfg must include ColWidths; there is no auto-sizing.  Use HtmlCell.Width
// hints from the parsed table as a guide when setting column widths.
//
// opts.SpanFontFor maps inline styles (bold, italic, monospace) to registered
// font names; all cell content is passed through this function so that
// rich formatting is preserved across word wrapping.
//
// HTML features mapped to pdf.Table capabilities:
//   - <thead> / all-<th> rows       → HeaderStyle applied
//   - <tfoot> rows                  → FooterStyle applied
//   - colspan / rowspan             → Cell.ColSpan / Cell.RowSpan
//   - align / text-align            → CellStyle.HAlign
//   - valign / vertical-align       → CellStyle.VAlign
//   - bgcolor / background-color    → CellStyle.Background / Row.Background
//   - color (style)                 → CellStyle.TextColor (per cell)
//   - <b>, <strong>, <th>           → bold font via SpanFontFor
//   - <i>, <em>, <cite>, <var>, dfn → italic font via SpanFontFor
//   - <code>, <tt>, <kbd>, <samp>   → monospace font via SpanFontFor
//   - <br> in cells                 → newline / paragraph break
//   - Nested <table>                → skipped (not recursed into)
func (d *Document) TableFromHTML(ht *htmlpkg.HtmlTable, cfg TableConfig, opts HtmlTableOptions) *Table {
	tbl := d.NewTable(cfg)

	for _, hr := range ht.Rows {
		row := Row{}

		// Row background colour.
		if bg, ok := parseHTMLColor(hr.BgColor); ok {
			row.Background = bg
		}

		for _, hc := range hr.Cells {
			cell := Cell{
				ColSpan: effectiveSpan(hc.ColSpan),
				RowSpan: effectiveSpan(hc.RowSpan),
			}

			// ── Cell style ──────────────────────────────────────────────
			var style CellStyle

			// Inherit section-level base style.
			switch {
			case hr.IsHeader:
				style = opts.HeaderStyle
			case hr.IsFooter:
				style = opts.FooterStyle
			}

			// Cell background (overrides row background when set).
			if bg, ok := parseHTMLColor(hc.BgColor); ok {
				style.Background = bg
			}

			// Cell text colour.
			if c, ok := parseHTMLColor(hc.Color); ok {
				style.TextColor = c
			}

			// Alignment.
			style.HAlign = htmlHAlign(hc.HAlign)
			style.VAlign = htmlVAlign(hc.VAlign)

			cell.Style = style

			// ── Rich spans ──────────────────────────────────────────────
			if len(hc.Spans) > 0 {
				rich := make([]RichSpan, 0, len(hc.Spans))
				for _, sp := range hc.Spans {
					fn := ""
					if opts.SpanFontFor != nil {
						fn = opts.SpanFontFor(sp.Style)
					}
					rich = append(rich, RichSpan{
						Text:     sp.Text,
						FontName: fn,
					})
				}
				cell.Spans = rich
			}

			row.Cells = append(row.Cells, cell)
		}

		tbl.AddRow(row)
	}

	return tbl
}

// effectiveSpan converts a raw HTML span value (0 means unset) to a valid
// table span (minimum 1).
func effectiveSpan(n int) int {
	if n <= 0 {
		return 1
	}
	return n
}

// htmlHAlign converts a normalised HTML alignment string to HAlign.
func htmlHAlign(s string) HAlign {
	switch s {
	case "left":
		return HAlignLeft
	case "center":
		return HAlignCenter
	case "right":
		return HAlignRight
	default:
		return HAlignDefault
	}
}

// htmlVAlign converts a normalised HTML vertical alignment string to VAlign.
func htmlVAlign(s string) VAlign {
	switch s {
	case "top":
		return VAlignTop
	case "middle":
		return VAlignMiddle
	case "bottom":
		return VAlignBottom
	default:
		return VAlignDefault
	}
}

// ─── HTML colour parsing ────────────────────────────────────────────────────

// parseHTMLColor converts an HTML colour string to a *Color.
// Supported formats:
//   - #RRGGBB   (six hex digits)
//   - #RGB      (three hex digits, expanded to six)
//   - rgb(r,g,b) (decimal 0-255 components)
//   - CSS named colours (a subset of the full list)
//
// Returns nil, false when the string is empty, malformed, or unrecognised.
func parseHTMLColor(s string) (*Color, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return nil, false
	}

	// ── Hex: #RRGGBB or #RGB ─────────────────────────────────────────────
	if strings.HasPrefix(s, "#") {
		hex := s[1:]
		if len(hex) == 3 {
			hex = string([]byte{
				hex[0], hex[0],
				hex[1], hex[1],
				hex[2], hex[2],
			})
		}
		if len(hex) == 6 {
			r, e1 := strconv.ParseUint(hex[0:2], 16, 8)
			g, e2 := strconv.ParseUint(hex[2:4], 16, 8)
			b, e3 := strconv.ParseUint(hex[4:6], 16, 8)
			if e1 == nil && e2 == nil && e3 == nil {
				c := Color{R: uint8(r), G: uint8(g), B: uint8(b)}
				return &c, true
			}
		}
		return nil, false
	}

	// ── rgb(r, g, b) ─────────────────────────────────────────────────────
	if strings.HasPrefix(s, "rgb(") && strings.HasSuffix(s, ")") {
		inner := s[4 : len(s)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 3 {
			rv, e1 := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 8)
			gv, e2 := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 8)
			bv, e3 := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 8)
			if e1 == nil && e2 == nil && e3 == nil {
				c := Color{R: uint8(rv), G: uint8(gv), B: uint8(bv)}
				return &c, true
			}
		}
		return nil, false
	}

	// ── Named colours ─────────────────────────────────────────────────────
	if c, ok := htmlNamedColors[s]; ok {
		cc := c
		return &cc, true
	}
	return nil, false
}

// htmlNamedColors maps a subset of CSS named colours to Color values.
var htmlNamedColors = map[string]Color{
	"black":     {0, 0, 0},
	"white":     {255, 255, 255},
	"red":       {255, 0, 0},
	"green":     {0, 128, 0},
	"blue":      {0, 0, 255},
	"yellow":    {255, 255, 0},
	"orange":    {255, 165, 0},
	"gray":      {128, 128, 128},
	"grey":      {128, 128, 128},
	"silver":    {192, 192, 192},
	"navy":      {0, 0, 128},
	"teal":      {0, 128, 128},
	"purple":    {128, 0, 128},
	"maroon":    {128, 0, 0},
	"lime":      {0, 255, 0},
	"aqua":      {0, 255, 255},
	"fuchsia":   {255, 0, 255},
	"cyan":      {0, 255, 255},
	"magenta":   {255, 0, 255},
	"pink":      {255, 192, 203},
	"brown":     {165, 42, 42},
	"coral":     {255, 127, 80},
	"gold":      {255, 215, 0},
	"khaki":     {240, 230, 140},
	"lavender":  {230, 230, 250},
	"wheat":     {245, 222, 179},
	"salmon":    {250, 128, 114},
	"tan":       {210, 180, 140},
	"violet":    {238, 130, 238},
	"indigo":    {75, 0, 130},
	"olive":     {128, 128, 0},
	"crimson":   {220, 20, 60},
	"orchid":    {218, 112, 214},
	"peru":      {205, 133, 63},
	"plum":      {221, 160, 221},
	"sienna":    {160, 82, 45},
	"turquoise": {64, 224, 208},
}
