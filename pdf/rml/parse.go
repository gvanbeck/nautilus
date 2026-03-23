package rml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// parseRML decodes an RML document from r into the internal rmlDoc AST.
func parseRML(r io.Reader) (*rmlDoc, error) {
	d := xml.NewDecoder(r)
	doc := &rmlDoc{}

	// Advance to the root <document> element.
	if err := advanceTo(d, "document"); err != nil {
		return nil, fmt.Errorf("rml: %w", err)
	}

	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			if end, ok := tok.(xml.EndElement); ok && end.Name.Local == "document" {
				break
			}
			continue
		}

		switch start.Name.Local {
		case "docinit":
			if err := parseDocinit(d, &doc.docinit); err != nil {
				return nil, fmt.Errorf("rml <docinit>: %w", err)
			}
		case "template":
			if err := parseTemplate(d, start, &doc.template); err != nil {
				return nil, fmt.Errorf("rml <template>: %w", err)
			}
		case "stylesheet":
			if err := parseStylesheet(d, &doc.stylesheet); err != nil {
				return nil, fmt.Errorf("rml <stylesheet>: %w", err)
			}
		case "story":
			nodes, err := parseStory(d)
			if err != nil {
				return nil, fmt.Errorf("rml <story>: %w", err)
			}
			doc.story = nodes
		default:
			if err := d.Skip(); err != nil {
				return nil, err
			}
		}
	}
	return doc, nil
}

// ─── Docinit ────────────────────────────────────────────────────────────────

func parseDocinit(d *xml.Decoder, di *docinit) error {
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "registerTTFont", "registerFont":
				fr := fontReg{
					name: attr(t, "fontName"),
					file: attr(t, "fontFile"),
				}
				di.fonts = append(di.fonts, fr)
				if err := d.Skip(); err != nil {
					return err
				}
			case "registerFontFamily":
				ff := fontFamilyReg{
					name:       attr(t, "name"),
					fontName:   attr(t, "fontName"),
					bold:       attr(t, "bold"),
					italic:     attr(t, "italic"),
					boldItalic: attr(t, "boldItalic"),
				}
				di.families = append(di.families, ff)
				if err := d.Skip(); err != nil {
					return err
				}
			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "docinit" {
				return nil
			}
		}
	}
}

// ─── Template ───────────────────────────────────────────────────────────────

func parseTemplate(d *xml.Decoder, start xml.StartElement, t *tmpl) error {
	t.pageSize = attr(start, "pageSize")
	t.leftMargin = attr(start, "leftMargin")
	t.rightMargin = attr(start, "rightMargin")
	t.topMargin = attr(start, "topMargin")
	t.bottomMargin = attr(start, "bottomMargin")
	t.firstPageTemplate = attr(start, "firstPageTemplate")
	t.title = attr(start, "title")
	t.author = attr(start, "author")
	t.subject = attr(start, "subject")
	t.creator = attr(start, "creator")

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			if tok.Name.Local == "pageTemplate" {
				pt, err := parsePageTemplate(d, tok)
				if err != nil {
					return err
				}
				t.templates = append(t.templates, pt)
			} else {
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if tok.Name.Local == "template" {
				return nil
			}
		}
	}
}

func parsePageTemplate(d *xml.Decoder, start xml.StartElement) (pageTmpl, error) {
	pt := pageTmpl{id: attr(start, "id")}
	for {
		tok, err := d.Token()
		if err != nil {
			return pt, err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case "frame":
				fd := frameDef{
					id:     attr(tok, "id"),
					x1:     attr(tok, "x1"),
					y1:     attr(tok, "y1"),
					width:  attr(tok, "width"),
					height: attr(tok, "height"),
				}
				pt.frames = append(pt.frames, fd)
				if err := d.Skip(); err != nil {
					return pt, err
				}
			case "pageGraphics":
				gb, err := parsePageGraphics(d)
				if err != nil {
					return pt, err
				}
				pt.pageGraphics = &gb
			default:
				if err := d.Skip(); err != nil {
					return pt, err
				}
			}
		case xml.EndElement:
			if tok.Name.Local == "pageTemplate" {
				return pt, nil
			}
		}
	}
}

// ─── pageGraphics ────────────────────────────────────────────────────────────

func parsePageGraphics(d *xml.Decoder) (graphicsBlock, error) {
	var gb graphicsBlock
	for {
		tok, err := d.Token()
		if err != nil {
			return gb, err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			cmd, err := parseGfxElement(d, tok)
			if err != nil {
				return gb, err
			}
			if cmd != nil {
				gb.commands = append(gb.commands, cmd)
			}
		case xml.EndElement:
			if tok.Name.Local == "pageGraphics" {
				return gb, nil
			}
		}
	}
}

func parseGfxElement(d *xml.Decoder, el xml.StartElement) (gfxCmd, error) {
	defer func() {}() // skip errors from d.Skip
	switch el.Name.Local {
	case "saveState":
		d.Skip() //nolint:errcheck
		return &gfxSaveState{}, nil
	case "restoreState":
		d.Skip() //nolint:errcheck
		return &gfxRestoreState{}, nil
	case "setFont":
		d.Skip() //nolint:errcheck
		return &gfxSetFont{
			name: attrDef(el, "name", attr(el, "fontName")),
			size: attrDef(el, "size", attr(el, "fontSize")),
		}, nil
	case "fill":
		d.Skip() //nolint:errcheck
		return &gfxSetFill{color: attrDef(el, "color", attr(el, "colorName"))}, nil
	case "stroke":
		d.Skip() //nolint:errcheck
		return &gfxSetStroke{
			color: attrDef(el, "color", attr(el, "colorName")),
			width: attr(el, "width"),
		}, nil
	case "drawString":
		text, _ := innerText(d, "drawString")
		return &gfxDrawString{x: attr(el, "x"), y: attr(el, "y"), text: text}, nil
	case "drawRightString":
		text, _ := innerText(d, "drawRightString")
		return &gfxDrawRString{x: attr(el, "x"), y: attr(el, "y"), text: text}, nil
	case "drawCentredString", "drawCenteredString":
		text, _ := innerText(d, el.Name.Local)
		return &gfxDrawCString{x: attr(el, "x"), y: attr(el, "y"), text: text}, nil
	case "rect":
		d.Skip() //nolint:errcheck
		return &gfxRect{
			x:      attr(el, "x"),
			y:      attr(el, "y"),
			w:      attr(el, "width"),
			h:      attr(el, "height"),
			fill:   attrDef(el, "fill", attrDef(el, "fillColor", attr(el, "colorName"))),
			stroke: attrDef(el, "stroke", attr(el, "strokeColor")),
			round:  attr(el, "round"),
		}, nil
	case "circle":
		d.Skip() //nolint:errcheck
		return &gfxCircle{
			x:      attr(el, "x"),
			y:      attr(el, "y"),
			r:      attr(el, "radius"),
			fill:   attrDef(el, "fill", attr(el, "fillColor")),
			stroke: attrDef(el, "stroke", attr(el, "strokeColor")),
		}, nil
	case "line":
		d.Skip() //nolint:errcheck
		return &gfxLine{
			x1:    attr(el, "x1"), y1: attr(el, "y1"),
			x2:    attr(el, "x2"), y2: attr(el, "y2"),
			width: attrDef(el, "width", attr(el, "lineWidth")),
			color: attrDef(el, "color", attr(el, "colorName")),
		}, nil
	case "lines":
		coords, _ := innerText(d, "lines")
		return &gfxLines{
			coords: coords,
			width:  attrDef(el, "width", attr(el, "lineWidth")),
			color:  attrDef(el, "color", attr(el, "colorName")),
		}, nil
	default:
		d.Skip() //nolint:errcheck
		return nil, nil
	}
}

// ─── Stylesheet ─────────────────────────────────────────────────────────────

func parseStylesheet(d *xml.Decoder, ss *stylesheet) error {
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case "paraStyle":
				ps := paraStyle{
					name:            attr(tok, "name"),
					parent:          attr(tok, "parent"),
					fontName:        attr(tok, "fontName"),
					fontSize:        attr(tok, "fontSize"),
					leading:         attr(tok, "leading"),
					alignment:       attr(tok, "alignment"),
					spaceBefore:     attr(tok, "spaceBefore"),
					spaceAfter:      attr(tok, "spaceAfter"),
					textColor:       attr(tok, "textColor"),
					backColor:       attrDef(tok, "backColor", attr(tok, "backgroundColor")),
					leftIndent:      attr(tok, "leftIndent"),
					rightIndent:     attr(tok, "rightIndent"),
					firstLineIndent: attr(tok, "firstLineIndent"),
					underline:       attr(tok, "underline"),
					strike:          attr(tok, "strike"),
					keepWithNext:    attr(tok, "keepWithNext"),
				}
				ss.paraStyles = append(ss.paraStyles, ps)
				if err := d.Skip(); err != nil {
					return err
				}
			case "blockTableStyle":
				bts, err := parseBlockTableStyle(d, tok)
				if err != nil {
					return err
				}
				ss.tableStyles = append(ss.tableStyles, bts)
			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if tok.Name.Local == "stylesheet" {
				return nil
			}
		}
	}
}

func parseBlockTableStyle(d *xml.Decoder, start xml.StartElement) (blockTableStyle, error) {
	bts := blockTableStyle{id: attr(start, "id")}
	for {
		tok, err := d.Token()
		if err != nil {
			return bts, err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			cmd, err := parseTableCmd(tok)
			if err != nil {
				return bts, err
			}
			if cmd != nil {
				bts.commands = append(bts.commands, *cmd)
			}
			if err := d.Skip(); err != nil {
				return bts, err
			}
		case xml.EndElement:
			if tok.Name.Local == "blockTableStyle" {
				return bts, nil
			}
		}
	}
}

func parseTableCmd(el xml.StartElement) (*tableCmd, error) {
	cmd := &tableCmd{kind: el.Name.Local}

	startC, startR, err1 := parseCoord(attrDef(el, "start", "0,0"))
	stopC, stopR, err2 := parseCoord(attrDef(el, "stop", "-1,-1"))
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("blockTableStyle command %s: invalid range", el.Name.Local)
	}
	cmd.startCol, cmd.startRow = startC, startR
	cmd.stopCol, cmd.stopRow = stopC, stopR

	switch el.Name.Local {
	case "lineStyle":
		cmd.lineKind = attrDef(el, "kind", "GRID")
		cmd.colorName = attrDef(el, "colorName", attrDef(el, "color", "black"))
		th, _ := strconv.ParseFloat(attr(el, "thickness"), 64)
		if th == 0 {
			th = 0.5
		}
		cmd.thickness = th
	case "blockBackground":
		cmd.colorName = attrDef(el, "colorName", attr(el, "color"))
	case "blockFont":
		cmd.fontName = attrDef(el, "name", attr(el, "fontName"))
		cmd.fontSize = attr(el, "size")
	case "blockTextColor":
		cmd.colorName = attrDef(el, "colorName", attr(el, "color"))
	case "blockAlignment":
		cmd.alignment = attrDef(el, "value", attr(el, "alignment"))
	case "blockValign":
		cmd.valign = attrDef(el, "value", attr(el, "valign"))
	case "blockLeftPadding", "blockRightPadding", "blockTopPadding", "blockBottomPadding",
		"blockPadding":
		cmd.padding = attrDef(el, "length", attr(el, "value"))
	case "blockLeading":
		cmd.padding = attrDef(el, "length", attr(el, "value")) // reuse padding field for leading
	default:
		return nil, nil // unknown command: silently ignore
	}
	return cmd, nil
}

// ─── Story ──────────────────────────────────────────────────────────────────

func parseStory(d *xml.Decoder) ([]storyNode, error) {
	var nodes []storyNode
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			node, err := parseStoryElement(d, tok)
			if err != nil {
				return nil, fmt.Errorf("<%s>: %w", tok.Name.Local, err)
			}
			if node != nil {
				nodes = append(nodes, node)
			}
		case xml.EndElement:
			if tok.Name.Local == "story" {
				return nodes, nil
			}
		}
	}
}

func parseStoryElement(d *xml.Decoder, start xml.StartElement) (storyNode, error) {
	switch start.Name.Local {
	case "para":
		text, err := innerText(d, "para")
		if err != nil {
			return nil, err
		}
		return &paraNode{style: attr(start, "style"), text: text}, nil

	case "spacer":
		if err := d.Skip(); err != nil {
			return nil, err
		}
		return &spacerNode{
			length: attrDef(start, "length", attr(start, "height")),
			width:  attr(start, "width"),
		}, nil

	case "pageBreak":
		if err := d.Skip(); err != nil {
			return nil, err
		}
		return &pageBreakNode{}, nil

	case "frameBreak":
		if err := d.Skip(); err != nil {
			return nil, err
		}
		return &frameBreakNode{}, nil

	case "hr", "hRule":
		if err := d.Skip(); err != nil {
			return nil, err
		}
		return &hrNode{
			width:     attrDef(start, "width", "100%"),
			thickness: attrDef(start, "thickness", "1"),
			colorName: attrDef(start, "color", attrDef(start, "colorName", "black")),
		}, nil

	case "blockTable":
		return parseBlockTable(d, start)

	case "image", "img":
		if err := d.Skip(); err != nil {
			return nil, err
		}
		return &imageNode{
			file:        attrDef(start, "file", attrDef(start, "filename", attr(start, "src"))),
			width:       attr(start, "width"),
			height:      attr(start, "height"),
			align:       attrDef(start, "align", attr(start, "hAlign")),
			spaceBefore: attr(start, "spaceBefore"),
			spaceAfter:  attr(start, "spaceAfter"),
		}, nil

	case "keepTogether":
		return parseKeepTogether(d, start)

	case "condPageBreak":
		if err := d.Skip(); err != nil {
			return nil, err
		}
		return &condPageBreakNode{height: attrDef(start, "height", attr(start, "minHeight"))}, nil

	case "nextPageTemplate":
		if err := d.Skip(); err != nil {
			return nil, err
		}
		return &nextPageTemplateNode{id: attr(start, "id")}, nil

	case "indent":
		return parseIndent(d, start)

	case "ul":
		return parseList(d, start, false)

	case "ol":
		return parseList(d, start, true)

	// heading shortcuts: map to para with style "h1"…"h6"
	case "h1", "h2", "h3", "h4", "h5", "h6":
		text, err := innerText(d, start.Name.Local)
		if err != nil {
			return nil, err
		}
		styleName := attrDef(start, "style", start.Name.Local)
		return &paraNode{style: styleName, text: text}, nil

	default:
		// unknown story element: skip
		if err := d.Skip(); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func parseKeepTogether(d *xml.Decoder, start xml.StartElement) (*keepTogetherNode, error) {
	kt := &keepTogetherNode{maxHeight: attr(start, "maxHeight")}
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			node, err := parseStoryElement(d, tok)
			if err != nil {
				return nil, err
			}
			if node != nil {
				kt.children = append(kt.children, node)
			}
		case xml.EndElement:
			if tok.Name.Local == "keepTogether" {
				return kt, nil
			}
		}
	}
}

func parseIndent(d *xml.Decoder, start xml.StartElement) (*indentNode, error) {
	ind := &indentNode{
		left:  attrDef(start, "left", attr(start, "leftIndent")),
		right: attrDef(start, "right", attr(start, "rightIndent")),
	}
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			node, err := parseStoryElement(d, tok)
			if err != nil {
				return nil, err
			}
			if node != nil {
				ind.children = append(ind.children, node)
			}
		case xml.EndElement:
			if tok.Name.Local == "indent" {
				return ind, nil
			}
		}
	}
}

func parseList(d *xml.Decoder, start xml.StartElement, ordered bool) (*listNode, error) {
	ln := &listNode{
		ordered:      ordered,
		start:        attrDef(start, "start", "1"),
		bulletIndent: attr(start, "bulletIndent"),
		style:        attr(start, "style"),
	}
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			if tok.Name.Local == "li" {
				text, err := innerText(d, "li")
				if err != nil {
					return nil, err
				}
				ln.items = append(ln.items, listItemNode{
					style: attr(tok, "style"),
					text:  strings.TrimSpace(text),
				})
			} else {
				if err := d.Skip(); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if tok.Name.Local == "ul" || tok.Name.Local == "ol" {
				return ln, nil
			}
		}
	}
}

func parseBlockTable(d *xml.Decoder, start xml.StartElement) (*blockTableNode, error) {
	bt := &blockTableNode{
		colWidths:   attr(start, "colWidths"),
		rowHeights:  attr(start, "rowHeights"),
		style:       attr(start, "style"),
		repeatRows:  attr(start, "repeatRows"),
		align:       attr(start, "align"),
		spaceBefore: attr(start, "spaceBefore"),
		spaceAfter:  attr(start, "spaceAfter"),
	}
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			if tok.Name.Local == "tr" {
				tr, err := parseTR(d, tok)
				if err != nil {
					return nil, err
				}
				bt.rows = append(bt.rows, tr)
			} else {
				if err := d.Skip(); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if tok.Name.Local == "blockTable" {
				return bt, nil
			}
		}
	}
}

func parseTR(d *xml.Decoder, start xml.StartElement) (trNode, error) {
	tr := trNode{
		height: attr(start, "height"),
		bg:     attrDef(start, "background", attrDef(start, "bg", attr(start, "colorName"))),
	}
	for {
		tok, err := d.Token()
		if err != nil {
			return tr, err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			if tok.Name.Local == "td" {
				td, err := parseTD(d, tok)
				if err != nil {
					return tr, err
				}
				tr.cells = append(tr.cells, td)
			} else {
				if err := d.Skip(); err != nil {
					return tr, err
				}
			}
		case xml.EndElement:
			if tok.Name.Local == "tr" {
				return tr, nil
			}
		}
	}
}

func parseTD(d *xml.Decoder, start xml.StartElement) (tdNode, error) {
	text, err := innerText(d, "td")
	if err != nil {
		return tdNode{}, err
	}
	return tdNode{
		colSpan:       attrDef(start, "colspan", attrDef(start, "colSpan", "1")),
		rowSpan:       attrDef(start, "rowspan", attrDef(start, "rowSpan", "1")),
		style:         attr(start, "style"),
		fontName:      attr(start, "fontName"),
		fontSize:      attr(start, "fontSize"),
		bold:          attr(start, "bold"),
		bg:            attrDef(start, "background", attrDef(start, "bg", attr(start, "colorName"))),
		textColor:     attrDef(start, "textColor", attr(start, "color")),
		halign:        attrDef(start, "alignment", attr(start, "align")),
		valign:        attr(start, "valign"),
		topPadding:    attrDef(start, "topPadding", attr(start, "tp")),
		rightPadding:  attrDef(start, "rightPadding", attr(start, "rp")),
		bottomPadding: attrDef(start, "bottomPadding", attr(start, "bp")),
		leftPadding:   attrDef(start, "leftPadding", attr(start, "lp")),
		text:          strings.TrimSpace(text),
	}, nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

// innerText collects all character data inside an element, stripping child tags.
func innerText(d *xml.Decoder, endTag string) (string, error) {
	var buf strings.Builder
	depth := 0
	for {
		tok, err := d.Token()
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.CharData:
			buf.Write(t)
		case xml.StartElement:
			depth++
		case xml.EndElement:
			if depth == 0 {
				if t.Name.Local != endTag {
					return "", fmt.Errorf("expected </%s>, got </%s>", endTag, t.Name.Local)
				}
				return buf.String(), nil
			}
			depth--
		}
	}
}

// advanceTo skips tokens until a StartElement with the given local name.
func advanceTo(d *xml.Decoder, name string) error {
	for {
		tok, err := d.Token()
		if err != nil {
			return fmt.Errorf("looking for <%s>: %w", name, err)
		}
		if s, ok := tok.(xml.StartElement); ok && s.Name.Local == name {
			return nil
		}
	}
}

// attr returns the value of a named attribute, or "" if absent.
func attr(el xml.StartElement, name string) string {
	for _, a := range el.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

// attrDef returns the value of a named attribute, or def if absent/empty.
func attrDef(el xml.StartElement, name, def string) string {
	v := attr(el, name)
	if v == "" {
		return def
	}
	return v
}
