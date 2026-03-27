package html

import (
	"strconv"
	"strings"
)

// ─── Parsed table types ─────────────────────────────────────────────────────

// HtmlTable is the result of parsing an HTML <table> element.
// Use ParseTable to produce one, then pass it to Document.TableFromHTML
// to render it into a PDF.
type HtmlTable struct {
	// Caption is the plain text content of the <caption> element, if present.
	Caption string
	// Rows holds every <tr> row in document order, including rows from
	// <thead>, <tbody>, and <tfoot> sections.
	Rows []HtmlRow
}

// HtmlRow is a single <tr> element.
type HtmlRow struct {
	// Cells holds every <td> and <th> cell in this row.
	Cells []HtmlCell
	// IsHeader is true when the row is inside a <thead> section,
	// or when every cell in the row is a <th>.
	IsHeader bool
	// IsFooter is true when the row is inside a <tfoot> section.
	IsFooter bool
	// BgColor is the row background colour from the bgcolor attribute
	// or the background-color CSS property. May be empty.
	BgColor string
	// Align is the default horizontal alignment inherited by cells:
	// "left", "center", or "right". May be empty.
	Align string
	// VAlign is the default vertical alignment inherited by cells:
	// "top", "middle", or "bottom". May be empty.
	VAlign string
}

// HtmlCell is a single <td> or <th> cell.
type HtmlCell struct {
	// Spans holds the parsed inline HTML content of the cell.
	Spans []Span
	// IsHeader is true when this cell is a <th> element.
	IsHeader bool
	// ColSpan is the colspan attribute value; 0 means 1.
	ColSpan int
	// RowSpan is the rowspan attribute value; 0 means 1.
	RowSpan int
	// HAlign is the resolved horizontal text alignment: "left", "center",
	// "right", or "" (inherit from table default).
	HAlign string
	// VAlign is the resolved vertical text alignment: "top", "middle",
	// "bottom", or "" (inherit from table default).
	VAlign string
	// BgColor is the cell background colour (HTML colour string or #RRGGBB).
	// May be empty.
	BgColor string
	// Color is the cell text colour. May be empty.
	Color string
	// Bold indicates the cell text should be rendered in bold.
	// Always true for <th> cells.
	Bold bool
	// NoWrap indicates the nowrap attribute is present.
	NoWrap bool
	// Width is the width attribute value (a hint only; the caller must
	// translate it to explicit column widths in TableConfig.ColWidths).
	Width string
}

// ─── ParseTableError ────────────────────────────────────────────────────────

// ParseTableError is returned by ParseTable when the input contains no
// recognisable <table> element.
type ParseTableError struct {
	msg string
}

func (e *ParseTableError) Error() string { return "html: parse table: " + e.msg }

// ─── Public API ─────────────────────────────────────────────────────────────

// ParseTable parses an HTML <table>...</table> element and returns an HtmlTable.
//
// The input must contain at least one <table> tag; surrounding whitespace and
// other text before the opening tag are silently skipped.  Only the first
// <table> element is parsed.
//
// classes is forwarded to the inline HTML span parser (html.Parse) that
// processes each cell's content.
//
// Supported structural elements:
//   - <table>, <caption>, <thead>, <tbody>, <tfoot>, <tr>, <th>, <td>
//
// Supported per-row and per-cell attributes:
//   - colspan, rowspan
//   - align, valign, bgcolor
//   - style (background-color, color, text-align, vertical-align, font-weight)
//   - width (returned as hint in HtmlCell.Width)
//   - nowrap
//
// <br> / <br/> tags within cell content are converted to newline characters
// before the inline parser processes the cell HTML.
func ParseTable(input string, classes ClassStyle) (*HtmlTable, error) {
	inner, _, ok := extractBlock(strings.TrimSpace(input), "table", 0)
	if !ok {
		return nil, &ParseTableError{msg: "no <table> element found"}
	}
	return parseTableInner(inner, classes), nil
}

// ─── Inner parser ────────────────────────────────────────────────────────────

func parseTableInner(inner string, classes ClassStyle) *HtmlTable {
	ht := &HtmlTable{}

	// Extract caption text (strip any inline tags).
	if capInner, _, ok := extractBlock(inner, "caption", 0); ok {
		ht.Caption = stripTags(capInner)
	}

	ht.Rows = parseRows(inner, classes)
	return ht
}

// ─── Row parsing ─────────────────────────────────────────────────────────────

func parseRows(html string, classes ClassStyle) []HtmlRow {
	var rows []HtmlRow
	section := "" // "thead", "tbody", "tfoot", or ""
	pos := 0

	for pos < len(html) {
		name, attrsStr, isClose, isSelf, afterTag := nextHTMLTag(html, pos)
		if afterTag == -1 {
			break
		}
		if isSelf || name == "" {
			pos = afterTag
			continue
		}
		if isClose {
			switch name {
			case "thead", "tbody", "tfoot":
				section = ""
			}
			pos = afterTag
			continue
		}

		switch name {
		case "thead":
			section = "thead"
			pos = afterTag
		case "tbody":
			section = "tbody"
			pos = afterTag
		case "tfoot":
			section = "tfoot"
			pos = afterTag
		case "caption":
			// Skip: already extracted above.
			closeOff := findMatchingClose(html[afterTag:], "caption")
			if closeOff == -1 {
				pos = len(html)
			} else {
				pos = skipPastCloseTag(html, afterTag+closeOff)
			}
		case "tr":
			closeOff := findMatchingClose(html[afterTag:], "tr")
			var rowInner string
			var nextPos int
			if closeOff == -1 {
				rowInner = html[afterTag:]
				nextPos = len(html)
			} else {
				rowInner = html[afterTag : afterTag+closeOff]
				nextPos = skipPastCloseTag(html, afterTag+closeOff)
			}
			row := parseRow(rowInner, attrsStr, section, classes)
			rows = append(rows, row)
			pos = nextPos
		case "table":
			// Skip nested tables entirely.
			closeOff := findMatchingClose(html[afterTag:], "table")
			if closeOff == -1 {
				pos = len(html)
			} else {
				pos = skipPastCloseTag(html, afterTag+closeOff)
			}
		default:
			pos = afterTag
		}
	}

	return rows
}

// ─── Row ─────────────────────────────────────────────────────────────────────

func parseRow(inner, attrsStr, section string, classes ClassStyle) HtmlRow {
	attrs := parseAttrs(attrsStr)
	style := parseStyle(attrs["style"])

	row := HtmlRow{
		IsHeader: section == "thead",
		IsFooter: section == "tfoot",
		BgColor:  firstNonEmpty(attrs["bgcolor"], style["background-color"]),
		Align:    normalizeAlign(firstNonEmpty(attrs["align"], style["text-align"])),
		VAlign:   normalizeVAlign(firstNonEmpty(attrs["valign"], style["vertical-align"])),
	}

	pos := 0
	for pos < len(inner) {
		name, cellAttrsStr, isClose, isSelf, afterTag := nextHTMLTag(inner, pos)
		if afterTag == -1 {
			break
		}
		if isSelf || name == "" || isClose {
			pos = afterTag
			continue
		}

		switch name {
		case "td", "th":
			isHeader := name == "th" || section == "thead"
			closeOff := findMatchingClose(inner[afterTag:], name)
			var cellInner string
			var nextPos int
			if closeOff == -1 {
				cellInner = inner[afterTag:]
				nextPos = len(inner)
			} else {
				cellInner = inner[afterTag : afterTag+closeOff]
				nextPos = skipPastCloseTag(inner, afterTag+closeOff)
			}
			cell := parseCell(cellInner, cellAttrsStr, isHeader, row.Align, row.VAlign, row.BgColor, classes)
			row.Cells = append(row.Cells, cell)
			pos = nextPos
		default:
			pos = afterTag
		}
	}

	// Mark the row as a header when every cell is <th>.
	if !row.IsHeader && len(row.Cells) > 0 {
		allTh := true
		for _, c := range row.Cells {
			if !c.IsHeader {
				allTh = false
				break
			}
		}
		if allTh {
			row.IsHeader = true
		}
	}

	return row
}

// ─── Cell ─────────────────────────────────────────────────────────────────────

func parseCell(inner, attrsStr string, isHeader bool, rowAlign, rowVAlign, rowBg string, classes ClassStyle) HtmlCell {
	attrs := parseAttrs(attrsStr)
	style := parseStyle(attrs["style"])

	colspan, _ := strconv.Atoi(attrs["colspan"])
	rowspan, _ := strconv.Atoi(attrs["rowspan"])
	_, noWrap := attrs["nowrap"]

	bgColor := firstNonEmpty(attrs["bgcolor"], style["background-color"], rowBg)
	color := firstNonEmpty(style["color"], attrs["color"])

	hAlign := normalizeAlign(firstNonEmpty(attrs["align"], style["text-align"], rowAlign))
	vAlign := normalizeVAlign(firstNonEmpty(attrs["valign"], style["vertical-align"], rowVAlign))

	bold := isHeader || strings.Contains(strings.ToLower(style["font-weight"]), "bold")

	// Preprocess: <br> → newline before inline parsing.
	cellHTML := replaceBR(inner)

	spans, _ := Parse(cellHTML, classes)

	return HtmlCell{
		Spans:    spans,
		IsHeader: isHeader,
		ColSpan:  colspan,
		RowSpan:  rowspan,
		HAlign:   hAlign,
		VAlign:   vAlign,
		BgColor:  bgColor,
		Color:    color,
		Bold:     bold,
		NoWrap:   noWrap,
		Width:    firstNonEmpty(attrs["width"], style["width"]),
	}
}

// ─── HTML scanning helpers ────────────────────────────────────────────────────

// nextHTMLTag scans html from pos and returns information about the next tag.
// Returns afterTag = -1 when no tag is found.
func nextHTMLTag(html string, pos int) (name, attrsStr string, isClose, isSelf bool, afterTag int) {
	tagStart := strings.IndexByte(html[pos:], '<')
	if tagStart == -1 {
		return "", "", false, false, -1
	}
	tagStart += pos

	tagEnd := strings.IndexByte(html[tagStart:], '>')
	if tagEnd == -1 {
		return "", "", false, false, -1
	}

	rawTag := html[tagStart+1 : tagStart+tagEnd]
	afterTag = tagStart + tagEnd + 1

	trimmed := strings.TrimSpace(rawTag)

	// HTML comment or doctype — skip silently.
	if strings.HasPrefix(trimmed, "!") {
		return "", "", false, false, afterTag
	}

	// Closing tag.
	if strings.HasPrefix(trimmed, "/") {
		isClose = true
		name = strings.ToLower(strings.TrimSpace(trimmed[1:]))
		return
	}

	// Self-closing tag.
	if strings.HasSuffix(trimmed, "/") {
		isSelf = true
		trimmed = strings.TrimRight(trimmed[:len(trimmed)-1], " \t\n\r")
	}

	// Extract name (everything up to first whitespace).
	nameEnd := strings.IndexAny(trimmed, " \t\n\r")
	if nameEnd == -1 {
		name = strings.ToLower(trimmed)
	} else {
		name = strings.ToLower(trimmed[:nameEnd])
		attrsStr = strings.TrimSpace(trimmed[nameEnd:])
	}

	return
}

// findMatchingClose finds the byte offset within s of the closing </tagName>
// that matches depth 1 (i.e., s starts right after an already-opened tag).
// Same-name nesting is tracked correctly (e.g., nested <table> within <table>).
// Returns -1 when not found.
func findMatchingClose(s, tagName string) int {
	lname := strings.ToLower(tagName)
	openPat := "<" + lname
	closePat := "</" + lname
	ls := strings.ToLower(s)
	depth := 1
	pos := 0

	for pos < len(s) {
		oi := indexOfTag(ls, openPat, pos)
		ci := indexOfTag(ls, closePat, pos)

		if ci == -1 {
			return -1
		}
		if oi != -1 && oi < ci {
			// Found a nested opening tag — check whether it is self-closing.
			tagEndOff := strings.IndexByte(s[oi:], '>')
			if tagEndOff == -1 {
				return -1
			}
			raw := strings.TrimSpace(s[oi+1 : oi+tagEndOff])
			if !strings.HasSuffix(raw, "/") {
				depth++
			}
			pos = oi + tagEndOff + 1
		} else {
			depth--
			if depth == 0 {
				return ci
			}
			tagEndOff := strings.IndexByte(s[ci:], '>')
			if tagEndOff == -1 {
				pos = ci + len(closePat)
			} else {
				pos = ci + tagEndOff + 1
			}
		}
	}
	return -1
}

// indexOfTag searches for the first occurrence of token in s[from:] where the
// character immediately following token is a tag boundary (space, >, /, etc.).
// Returns -1 when not found.
func indexOfTag(s, token string, from int) int {
	for {
		idx := strings.Index(s[from:], token)
		if idx == -1 {
			return -1
		}
		pos := from + idx
		after := pos + len(token)
		if after >= len(s) || isHTMLTagBoundary(s[after]) {
			return pos
		}
		from = pos + len(token)
	}
}

func isHTMLTagBoundary(ch byte) bool {
	return ch == ' ' || ch == '>' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '/'
}

// extractBlock finds the first <tagName ...>...</tagName> block in html
// starting at pos.  Returns inner content, the position after the closing tag,
// and whether the block was found.
func extractBlock(html, tagName string, pos int) (inner string, afterBlock int, found bool) {
	ls := strings.ToLower(html)
	openPat := "<" + strings.ToLower(tagName)

	idx := indexOfTag(ls, openPat, pos)
	if idx == -1 {
		return "", 0, false
	}

	openEnd := strings.IndexByte(html[idx:], '>')
	if openEnd == -1 {
		return "", 0, false
	}
	contentStart := idx + openEnd + 1

	closeOff := findMatchingClose(html[contentStart:], tagName)
	if closeOff == -1 {
		return html[contentStart:], len(html), true
	}

	inner = html[contentStart : contentStart+closeOff]
	afterBlock = skipPastCloseTag(html, contentStart+closeOff)
	return inner, afterBlock, true
}

// skipPastCloseTag advances past the closing tag starting at pos (which points
// at the '<' of '</tagname>').
func skipPastCloseTag(html string, pos int) int {
	end := strings.IndexByte(html[pos:], '>')
	if end == -1 {
		return pos
	}
	return pos + end + 1
}

// ─── Attribute / style helpers ───────────────────────────────────────────────

// parseAttrs parses an HTML attribute string like:
//
//	colspan="2" align="center" style="color: red" nowrap
func parseAttrs(s string) map[string]string {
	result := make(map[string]string)
	i := 0
	for i < len(s) {
		// Skip whitespace.
		for i < len(s) && isSpace(s[i]) {
			i++
		}
		if i >= len(s) {
			break
		}

		// Read attribute name.
		nameStart := i
		for i < len(s) && s[i] != '=' && !isSpace(s[i]) {
			i++
		}
		if i == nameStart {
			i++
			continue
		}
		name := strings.ToLower(s[nameStart:i])

		// Skip whitespace.
		for i < len(s) && isSpace(s[i]) {
			i++
		}

		if i >= len(s) || s[i] != '=' {
			result[name] = "" // boolean attribute
			continue
		}
		i++ // skip '='

		// Skip whitespace.
		for i < len(s) && isSpace(s[i]) {
			i++
		}

		if i >= len(s) {
			result[name] = ""
			break
		}

		// Read value.
		var value string
		if s[i] == '"' || s[i] == '\'' {
			quote := s[i]
			i++
			start := i
			for i < len(s) && s[i] != quote {
				i++
			}
			value = s[start:i]
			if i < len(s) {
				i++ // closing quote
			}
		} else {
			start := i
			for i < len(s) && !isSpace(s[i]) && s[i] != '>' {
				i++
			}
			value = s[start:i]
		}
		result[name] = value
	}
	return result
}

// parseStyle parses a CSS style attribute value such as
// "background-color: #fff; font-weight: bold".
func parseStyle(s string) map[string]string {
	result := make(map[string]string)
	for _, decl := range strings.Split(s, ";") {
		decl = strings.TrimSpace(decl)
		if decl == "" {
			continue
		}
		colon := strings.IndexByte(decl, ':')
		if colon == -1 {
			continue
		}
		prop := strings.ToLower(strings.TrimSpace(decl[:colon]))
		val := strings.TrimSpace(decl[colon+1:])
		if prop != "" {
			result[prop] = val
		}
	}
	return result
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// ─── Content helpers ─────────────────────────────────────────────────────────

// replaceBR replaces <br>, <br/>, <br /> (case-insensitive) with newline characters.
func replaceBR(s string) string {
	var b strings.Builder
	ls := strings.ToLower(s)
	i := 0
	for i < len(s) {
		if ls[i] == '<' && i+3 <= len(s) && ls[i:i+3] == "<br" {
			after := i + 3
			if after >= len(s) || isHTMLTagBoundary(ls[after]) {
				b.WriteByte('\n')
				end := strings.IndexByte(s[i:], '>')
				if end != -1 {
					i += end + 1
				} else {
					i++
				}
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// stripTags removes all HTML tags from s, returning plain text.
func stripTags(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '<' {
			end := strings.IndexByte(s[i:], '>')
			if end != -1 {
				i += end + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return strings.TrimSpace(b.String())
}

// ─── Normalisation helpers ───────────────────────────────────────────────────

// normalizeAlign normalises an HTML alignment value to one of the canonical
// strings "left", "center", "right", or "" (unknown / unset).
func normalizeAlign(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "left":
		return "left"
	case "center", "middle":
		return "center"
	case "right":
		return "right"
	default:
		return ""
	}
}

// normalizeVAlign normalises an HTML vertical alignment value to one of
// "top", "middle", "bottom", or "" (unknown / unset).
func normalizeVAlign(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "top":
		return "top"
	case "middle", "center":
		return "middle"
	case "bottom":
		return "bottom"
	default:
		return ""
	}
}

// firstNonEmpty returns the first non-empty string from vals.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
