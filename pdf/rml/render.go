package rml

import (
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/layout"
)

// fontFamily maps the four variants of a font family.
type fontFamily struct {
	regular    string
	bold       string
	italic     string
	boldItalic string
}

// ─── Entry point ─────────────────────────────────────────────────────────────

func render(doc *rmlDoc, opts Options) (*pdf.Document, error) {
	fontDir := opts.FontDir
	if fontDir == "" {
		fontDir = "."
	}

	// ── Page size & margins ───────────────────────────────────────────────
	ps, err := parsePageSize(doc.template.pageSize)
	if err != nil {
		return nil, err
	}
	mLeft := ptDef(doc.template.leftMargin, 55)
	mRight := ptDef(doc.template.rightMargin, 55)
	mTop := ptDef(doc.template.topMargin, 55)
	mBottom := ptDef(doc.template.bottomMargin, 55)

	// ── Create pdf.Document ───────────────────────────────────────────────
	pdfdoc, err := pdf.New(pdf.Config{
		PageSize:         ps,
		Margins:          pdf.Margins{Top: mTop, Right: mRight, Bottom: mBottom, Left: mLeft},
		DefaultFontSize:  11,
		LineHeightFactor: 1.3,
	})
	if err != nil {
		return nil, err
	}

	// ── Metadata ──────────────────────────────────────────────────────────
	if doc.template.title != "" || doc.template.author != "" ||
		doc.template.subject != "" || doc.template.creator != "" {
		pdfdoc.SetInfo(doc.template.title, doc.template.author,
			doc.template.subject, doc.template.creator)
	}

	// ── Register fonts ────────────────────────────────────────────────────
	for _, fr := range doc.docinit.fonts {
		path := fr.file
		if !filepath.IsAbs(path) {
			path = filepath.Join(fontDir, path)
		}
		if err := pdfdoc.RegisterFont(fr.name, path); err != nil {
			return nil, fmt.Errorf("registerFont %q: %w", fr.name, err)
		}
	}

	// ── Build font family map ─────────────────────────────────────────────
	families := make(map[string]fontFamily)
	for _, ff := range doc.docinit.families {
		families[ff.name] = fontFamily{
			regular:    ff.fontName,
			bold:       ff.bold,
			italic:     ff.italic,
			boldItalic: ff.boldItalic,
		}
		// Also index by fontName (regular variant).
		if ff.fontName != "" && ff.fontName != ff.name {
			families[ff.fontName] = families[ff.name]
		}
	}

	// ── Stylesheet ────────────────────────────────────────────────────────
	ss := buildStylesheet(doc.stylesheet)

	// ── Build layout.DocTemplate ─────────────────────────────────────────
	dt := layout.NewDocTemplate(pdfdoc)
	if len(doc.template.templates) == 0 {
		doc.template.templates = []pageTmpl{{
			id: "default",
			frames: []frameDef{{
				id:     "body",
				x1:     fmt.Sprintf("%g", mLeft),
				y1:     fmt.Sprintf("%g", mBottom),
				width:  fmt.Sprintf("%g", ps.Width-mLeft-mRight),
				height: fmt.Sprintf("%g", ps.Height-mTop-mBottom),
			}},
		}}
	}

	for _, pt := range doc.template.templates {
		lpt := &layout.PageTemplate{ID: pt.id}

		// Wire up pageGraphics as OnPage decorator.
		if pt.pageGraphics != nil {
			gb := *pt.pageGraphics
			pageH := ps.Height
			lpt.OnPage = func(d *pdf.Document, pageNum int) {
				execGraphicsWithPage(d, gb, pageH, pageNum)
			}
		}

		for _, fd := range pt.frames {
			x1 := ptDef(fd.x1, mLeft)
			y1 := ptDef(fd.y1, mBottom)
			w := ptDef(fd.width, ps.Width-mLeft-mRight)
			h := ptDef(fd.height, ps.Height-mTop-mBottom)
			nautilusY := ps.Height - y1 - h
			lpt.Frames = append(lpt.Frames, &layout.LayoutFrame{
				ID:     fd.id,
				X:      x1,
				Y:      nautilusY,
				Width:  w,
				Height: h,
			})
		}
		dt.AddPageTemplate(lpt)
	}

	// ── Build story ───────────────────────────────────────────────────────
	story, err := buildStory(doc.story, ss, pdfdoc, families)
	if err != nil {
		return nil, err
	}

	if err := dt.Build(story); err != nil {
		return nil, fmt.Errorf("layout: %w", err)
	}

	return pdfdoc, nil
}

// ─── Stylesheet ──────────────────────────────────────────────────────────────

type resolvedStylesheet struct {
	para  map[string]layout.ParagraphStyle
	table map[string]blockTableStyle
}

func buildStylesheet(ss stylesheet) *resolvedStylesheet {
	rs := &resolvedStylesheet{
		para:  make(map[string]layout.ParagraphStyle),
		table: make(map[string]blockTableStyle),
	}
	for _, ps := range ss.paraStyles {
		rs.para[ps.name] = toParagraphStyle(ps)
	}
	for _, bts := range ss.tableStyles {
		rs.table[bts.id] = bts
	}
	return rs
}

func toParagraphStyle(ps paraStyle) layout.ParagraphStyle {
	var s layout.ParagraphStyle
	s.FontName = ps.fontName
	s.FontSize = ptDef(ps.fontSize, 0)
	s.Leading = ptDef(ps.leading, 0)
	s.SpaceBefore = ptDef(ps.spaceBefore, 0)
	s.SpaceAfter = ptDef(ps.spaceAfter, 0)
	s.LeftIndent = ptDef(ps.leftIndent, 0)
	s.RightIndent = ptDef(ps.rightIndent, 0)
	s.KeepWithNextPara = ps.keepWithNext == "1" || strings.EqualFold(ps.keepWithNext, "true")
	switch strings.ToLower(strings.TrimSpace(ps.alignment)) {
	case "center", "centre":
		s.Alignment = layout.AlignCenter
	case "right":
		s.Alignment = layout.AlignRight
	default:
		s.Alignment = layout.AlignLeft
	}
	if c, ok := parseColor(ps.textColor); ok {
		s.TextColor = &c
	}
	// backColor, firstLineIndent, underline, strike are stored but layout.Paragraph
	// does not yet support them; they are preserved here for future use.
	return s
}

// ─── Story ───────────────────────────────────────────────────────────────────

func buildStory(nodes []storyNode, ss *resolvedStylesheet, pdfdoc *pdf.Document, families map[string]fontFamily) ([]layout.Flowable, error) {
	var flowables []layout.Flowable
	for _, n := range nodes {
		f, err := nodeToFlowable(n, ss, pdfdoc, families)
		if err != nil {
			return nil, err
		}
		if f != nil {
			flowables = append(flowables, f)
		}
	}
	return flowables, nil
}

func nodeToFlowable(n storyNode, ss *resolvedStylesheet, pdfdoc *pdf.Document, families map[string]fontFamily) (layout.Flowable, error) {
	switch node := n.(type) {
	case *paraNode:
		style := ss.para[node.style]
		// Use htmlParagraph when text contains inline markup, plain Paragraph otherwise.
		if strings.ContainsAny(node.text, "<>") {
			monoFont := ""
			if fam, ok := families["mono"]; ok && fam.regular != "" {
				monoFont = fam.regular
			}
			return &htmlParagraph{text: node.text, style: style, families: families, monoFont: monoFont}, nil
		}
		return &layout.Paragraph{Text: node.text, Style: style}, nil

	case *spacerNode:
		h := ptDef(node.length, 12)
		w := ptDef(node.width, 0)
		return &layout.Spacer{Width: w, Height: h}, nil

	case *pageBreakNode:
		return &layout.PageBreak{}, nil

	case *frameBreakNode:
		return &layout.FrameBreak{}, nil

	case *nextPageTemplateNode:
		return &layout.NextPageTemplate{TemplateID: node.id}, nil

	case *condPageBreakNode:
		return &layout.CondPageBreak{MinHeight: ptDef(node.height, 0)}, nil

	case *hrNode:
		th := ptDef(node.thickness, 1)
		c, _ := parseColor(node.colorName)
		hr := &layout.HRFlowable{Thickness: th, Color: c}
		if !strings.HasSuffix(node.width, "%") {
			hr.Width = ptDef(node.width, 0)
		}
		return hr, nil

	case *imageNode:
		im := &imageFlowable{
			path:        node.file,
			width:       ptDef(node.width, 0),
			height:      ptDef(node.height, 0),
			spaceBefore: ptDef(node.spaceBefore, 0),
			spaceAfter:  ptDef(node.spaceAfter, 6),
		}
		switch strings.ToLower(node.align) {
		case "center", "centre":
			im.align = layout.AlignCenter
		case "right":
			im.align = layout.AlignRight
		}
		return im, nil

	case *keepTogetherNode:
		inner, err := buildStory(node.children, ss, pdfdoc, families)
		if err != nil {
			return nil, err
		}
		return &layout.KeepTogether{Flowables: inner}, nil

	case *indentNode:
		inner, err := buildStory(node.children, ss, pdfdoc, families)
		if err != nil {
			return nil, err
		}
		return &indentFlowable{
			left:  ptDef(node.left, 0),
			right: ptDef(node.right, 0),
			inner: inner,
		}, nil

	case *listNode:
		style := ss.para[node.style]
		items := make([]listItem, len(node.items))
		for i, li := range node.items {
			s := ss.para[li.style]
			if s.FontName == "" {
				s = style
			}
			items[i] = listItem{text: li.text, style: s}
		}
		start := 1
		if node.start != "" {
			if v, err := strconv.Atoi(node.start); err == nil {
				start = v
			}
		}
		return &listFlowable{
			items:        items,
			ordered:      node.ordered,
			start:        start,
			style:        style,
			bulletIndent: ptDef(node.bulletIndent, 0),
			spaceAfter:   6,
		}, nil

	case *blockTableNode:
		return buildTableFlowable(node, ss, pdfdoc)

	default:
		return nil, nil
	}
}

// ─── Table flowable ──────────────────────────────────────────────────────────

// tableFlowable wraps a blockTableNode as a layout.Flowable so it can be
// flowed through a layout.DocTemplate frame.
type tableFlowable struct {
	node        *blockTableNode
	ss          *resolvedStylesheet
	pdfdoc      *pdf.Document
	colWidths   []float64
	rowHeights  []float64
	repeatRows  int
	defaultRowH float64
	sBefore     float64
	sAfter      float64
	startRow, endRow int
}

const defaultRowHeight = 22.0

func buildTableFlowable(node *blockTableNode, ss *resolvedStylesheet, pdfdoc *pdf.Document) (*tableFlowable, error) {
	cw, err := splitWidths(node.colWidths)
	if err != nil {
		return nil, fmt.Errorf("blockTable colWidths: %w", err)
	}

	var rh []float64
	if node.rowHeights != "" {
		rh, err = splitWidths(node.rowHeights)
		if err != nil {
			return nil, fmt.Errorf("blockTable rowHeights: %w", err)
		}
	}
	for len(rh) < len(node.rows) {
		rh = append(rh, 0)
	}

	repeatRows := 0
	if node.repeatRows != "" {
		if v, e := strconv.Atoi(node.repeatRows); e == nil {
			repeatRows = v
		}
	}

	return &tableFlowable{
		node:        node,
		ss:          ss,
		pdfdoc:      pdfdoc,
		colWidths:   cw,
		rowHeights:  rh,
		repeatRows:  repeatRows,
		defaultRowH: defaultRowHeight,
		sBefore:     ptDef(node.spaceBefore, 0),
		sAfter:      ptDef(node.spaceAfter, 6),
		startRow:    0,
		endRow:      len(node.rows),
	}, nil
}

func (tf *tableFlowable) SpaceBefore() float64 { return tf.sBefore }
func (tf *tableFlowable) SpaceAfter() float64  { return tf.sAfter }
func (tf *tableFlowable) KeepWithNext() bool   { return false }
func (tf *tableFlowable) MinWidth() float64 {
	w := 0.0
	for _, cw := range tf.colWidths {
		w += cw
	}
	return w
}

func (tf *tableFlowable) rowH(i int) float64 {
	if i < len(tf.rowHeights) && tf.rowHeights[i] > 0 {
		return tf.rowHeights[i]
	}
	return tf.defaultRowH
}

func (tf *tableFlowable) totalHeight() float64 {
	h := 0.0
	for i := tf.startRow; i < tf.endRow; i++ {
		h += tf.rowH(i)
	}
	return h
}

func (tf *tableFlowable) Wrap(_ *pdf.Document, availW, _ float64) (float64, float64) {
	w := tf.MinWidth()
	if w > availW {
		w = availW
	}
	return w, tf.totalHeight()
}

func (tf *tableFlowable) Split(_ *pdf.Document, availW, availH float64) []layout.Flowable {
	h := 0.0
	splitAt := tf.startRow
	for i := tf.startRow; i < tf.endRow; i++ {
		rh := tf.rowH(i)
		if h+rh > availH {
			break
		}
		h += rh
		splitAt = i + 1
	}
	if splitAt <= tf.startRow || splitAt >= tf.endRow {
		return nil
	}
	t1 := *tf
	t1.endRow = splitAt
	t1.sAfter = 0
	t2 := *tf
	t2.startRow = splitAt
	t2.sBefore = 0
	return []layout.Flowable{&t1, &t2}
}

func (tf *tableFlowable) Draw(doc *pdf.Document, x, y float64) error {
	numRows := tf.endRow - tf.startRow
	numCols := len(tf.colWidths)

	// Resolve blockTableStyle commands into a cell-style matrix.
	// matrix[row][col] holds accumulated CellStyle overrides.
	matrix := make([][]cellOverride, numRows)
	for i := range matrix {
		matrix[i] = make([]cellOverride, numCols)
	}

	// Outer border spec (from OUTLINE / BOX lineStyle commands)
	var outerBorder *pdf.BorderSpec
	var innerSpec *pdf.BorderSpec

	if bts, ok := tf.ss.table[tf.node.style]; ok {
		for _, cmd := range bts.commands {
			// Resolve absolute row/col ranges against the window [startRow, endRow)
			r0 := resolveIdx(cmd.startRow, numRows)
			r1 := resolveIdx(cmd.stopRow, numRows)
			c0 := resolveIdx(cmd.startCol, numCols)
			c1 := resolveIdx(cmd.stopCol, numCols)

			switch cmd.kind {
			case "lineStyle":
				spec := &pdf.BorderSpec{
					Thickness: cmd.thickness,
				}
				if c, ok := parseColor(cmd.colorName); ok {
					spec.Color = c
				}
				switch strings.ToUpper(cmd.lineKind) {
				case "GRID":
					// apply to all cells
					for r := r0; r <= r1 && r < numRows; r++ {
						for c := c0; c <= c1 && c < numCols; c++ {
							s := spec
							s2 := *s
							matrix[r][c].border = pdf.NewUniformBorder(s2)
						}
					}
				case "OUTLINE", "BOX":
					outerBorder = spec
				case "INNERGRID":
					innerSpec = spec
				}

			case "blockBackground":
				if c, ok := parseColor(cmd.colorName); ok {
					cc := c
					for r := r0; r <= r1 && r < numRows; r++ {
						for col := c0; col <= c1 && col < numCols; col++ {
							matrix[r][col].bg = &cc
						}
					}
				}

			case "blockFont":
				applyToRange(matrix, r0, r1, c0, c1, numRows, numCols, func(cell *cellOverride) {
					cell.fontName = cmd.fontName
					cell.fontSize = cmd.fontSize
				})

			case "blockTextColor":
				if c, ok := parseColor(cmd.colorName); ok {
					cc := c
					applyToRange(matrix, r0, r1, c0, c1, numRows, numCols, func(cell *cellOverride) {
						cell.textColor = &cc
					})
				}

			case "blockAlignment":
				applyToRange(matrix, r0, r1, c0, c1, numRows, numCols, func(cell *cellOverride) {
					cell.halign = cmd.alignment
				})

			case "blockValign":
				applyToRange(matrix, r0, r1, c0, c1, numRows, numCols, func(cell *cellOverride) {
					cell.valign = cmd.valign
				})

			case "blockLeftPadding":
				applyToRange(matrix, r0, r1, c0, c1, numRows, numCols, func(cell *cellOverride) {
					cell.leftPadding = cmd.padding
				})
			case "blockRightPadding":
				applyToRange(matrix, r0, r1, c0, c1, numRows, numCols, func(cell *cellOverride) {
					cell.rightPadding = cmd.padding
				})
			case "blockTopPadding":
				applyToRange(matrix, r0, r1, c0, c1, numRows, numCols, func(cell *cellOverride) {
					cell.topPadding = cmd.padding
				})
			case "blockBottomPadding":
				applyToRange(matrix, r0, r1, c0, c1, numRows, numCols, func(cell *cellOverride) {
					cell.bottomPadding = cmd.padding
				})
			case "blockPadding":
				applyToRange(matrix, r0, r1, c0, c1, numRows, numCols, func(cell *cellOverride) {
					cell.leftPadding = cmd.padding
					cell.rightPadding = cmd.padding
					cell.topPadding = cmd.padding
					cell.bottomPadding = cmd.padding
				})
			}
		}
	}

	// Apply INNERGRID to interior borders
	if innerSpec != nil {
		for r := 0; r < numRows; r++ {
			for c := 0; c < numCols; c++ {
				b := &matrix[r][c].border
				if c > 0 {
					s := *innerSpec
					b.Left = &s
				}
				if c < numCols-1 {
					s := *innerSpec
					b.Right = &s
				}
				if r > 0 {
					s := *innerSpec
					b.Top = &s
				}
				if r < numRows-1 {
					s := *innerSpec
					b.Bottom = &s
				}
			}
		}
	}

	// Default cell style: 5pt padding, 10pt font, regular font (first registered)
	defaultCS := pdf.CellStyle{
		Padding:  pdf.Padding{Top: 5, Right: 8, Bottom: 5, Left: 8},
		FontSize: 10,
	}

	// Outer border on table config
	var tableBorder pdf.Border
	if outerBorder != nil {
		tableBorder = pdf.NewUniformBorder(*outerBorder)
	}

	hugePB := math.MaxFloat64

	tbl := doc.NewTable(pdf.TableConfig{
		X:                x,
		Y:                y,
		ColWidths:        tf.colWidths,
		Border:           tableBorder,
		DefaultCellStyle: defaultCS,
		PageBottom:       hugePB,
		RepeatRows:       tf.repeatRows,
	})

	for ri, trn := range tf.node.rows[tf.startRow:tf.endRow] {
		globalRow := tf.startRow + ri

		var rowBg *pdf.Color
		if c, ok := parseColor(trn.bg); ok {
			rowBg = &c
		}

		cells := make([]pdf.Cell, 0, len(trn.cells))
		for ci, tdn := range trn.cells {
			cs := pdf.CellStyle{}

			// Apply matrix overrides first (from blockTableStyle)
			if ci < numCols {
				ov := matrix[ri][ci]
				cs.Border = ov.border
				cs.Background = ov.bg
				cs.TextColor = ov.textColor
				if ov.fontName != "" {
					cs.FontName = ov.fontName
				}
				if ov.fontSize != "" {
					cs.FontSize = ptDef(ov.fontSize, 0)
				}
				if ov.halign != "" {
					cs.HAlign = parseHAlignPDF(ov.halign)
				}
				if ov.valign != "" {
					cs.VAlign = parseVAlignPDF(ov.valign)
				}
				// Padding from blockTableStyle (only overrides when set).
				applyPaddingOverride(&cs.Padding, ov.topPadding, ov.rightPadding, ov.bottomPadding, ov.leftPadding)
			}

			// Per-cell <td> attributes override matrix
			if tdn.bg != "" {
				if c, ok := parseColor(tdn.bg); ok {
					cc := c
					cs.Background = &cc
				}
			}
			if tdn.textColor != "" {
				if c, ok := parseColor(tdn.textColor); ok {
					cc := c
					cs.TextColor = &cc
				}
			}
			if tdn.fontName != "" {
				cs.FontName = tdn.fontName
			}
			if tdn.bold == "1" || strings.EqualFold(tdn.bold, "true") {
				cs.FontName = cs.FontName + "Bold"
			}
			if tdn.fontSize != "" {
				cs.FontSize = ptDef(tdn.fontSize, 0)
			}
			if tdn.halign != "" {
				cs.HAlign = parseHAlignPDF(tdn.halign)
			}
			if tdn.valign != "" {
				cs.VAlign = parseVAlignPDF(tdn.valign)
			}
			// Per-cell padding from <td> attributes.
			applyPaddingOverride(&cs.Padding, tdn.topPadding, tdn.rightPadding, tdn.bottomPadding, tdn.leftPadding)

			cs, _ = applyOuterBorderToCell(cs, outerBorder, ri, ci, numRows-1, numCols-1, globalRow == tf.startRow)

			colspan, _ := strconv.Atoi(tdn.colSpan)
			rowspan, _ := strconv.Atoi(tdn.rowSpan)
			if colspan < 1 {
				colspan = 1
			}
			if rowspan < 1 {
				rowspan = 1
			}

			cells = append(cells, pdf.Cell{
				Text:    tdn.text,
				ColSpan: colspan,
				RowSpan: rowspan,
				Style:   cs,
			})
		}

		rowH := 0.0
		if trn.height != "" {
			rowH = ptDef(trn.height, 0)
		} else {
			rowH = tf.rowH(globalRow)
		}

		tbl.AddRow(pdf.Row{
			Height:     rowH,
			Background: rowBg,
			Cells:      cells,
		})
	}

	return tbl.Draw()
}

// applyPaddingOverride sets individual padding sides when the string value is non-empty.
func applyPaddingOverride(p *pdf.Padding, top, right, bottom, left string) {
	if top != "" {
		p.Top = ptDef(top, p.Top)
	}
	if right != "" {
		p.Right = ptDef(right, p.Right)
	}
	if bottom != "" {
		p.Bottom = ptDef(bottom, p.Bottom)
	}
	if left != "" {
		p.Left = ptDef(left, p.Left)
	}
}

// applyOuterBorderToCell adds outer border sides to edge cells when an
// outerBorder spec is present.
func applyOuterBorderToCell(cs pdf.CellStyle, outer *pdf.BorderSpec,
	row, col, lastRow, lastCol int, isFirstRenderedRow bool) (pdf.CellStyle, bool) {
	if outer == nil {
		return cs, false
	}
	s := *outer
	if row == 0 && isFirstRenderedRow {
		cs.Border.Top = &s
	}
	if row == lastRow {
		cs.Border.Bottom = &s
	}
	if col == 0 {
		cs.Border.Left = &s
	}
	if col == lastCol {
		cs.Border.Right = &s
	}
	return cs, true
}

// cellOverride holds the accumulated style overrides from a blockTableStyle.
type cellOverride struct {
	border        pdf.Border
	bg            *pdf.Color
	textColor     *pdf.Color
	fontName      string
	fontSize      string
	halign        string
	valign        string
	topPadding    string
	rightPadding  string
	bottomPadding string
	leftPadding   string
}

// applyToRange applies fn to every cell in the row/col range, clamped to matrix bounds.
func applyToRange(matrix [][]cellOverride, r0, r1, c0, c1, numRows, numCols int, fn func(*cellOverride)) {
	for r := r0; r <= r1 && r < numRows; r++ {
		for c := c0; c <= c1 && c < numCols; c++ {
			fn(&matrix[r][c])
		}
	}
}
