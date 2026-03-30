package rml

import (
	"strconv"
	"strings"
	"unicode"

	pdfhtml "github.com/gvanbeck/nautilus/pdf/html"
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/layout"
)

// ─── htmlParagraph ───────────────────────────────────────────────────────────
//
// htmlParagraph is a layout.Flowable that renders a paragraph whose text may
// contain inline HTML markup (<b>, <i>, <u>, <s>, <code>, …).  It honours the
// same ParagraphStyle as layout.Paragraph; per-span font variants are resolved
// via the fontFamily map.
//
// monoFont, when non-empty, is the base font name used for monospace spans
// (<code>, <tt>, <kbd>, <samp>).  Register a font family named "mono" in the
// RML <docinit> to activate automatic monospace font substitution.

type htmlParagraph struct {
	text      string // raw inner-XML text (may contain <b>…</b> etc.)
	style     layout.ParagraphStyle
	families  map[string]fontFamily // font family overrides
	monoFont  string                // font name for monospace spans; "" = no substitution
	computedW float64
	computedH float64
	lines     []htmlLine
}

type htmlLine []htmlSpan

type htmlSpan struct {
	text          string
	bold          bool
	italic        bool
	underline     bool
	strikethrough bool
	monospace     bool
}

// SpaceBefore / SpaceAfter / KeepWithNext / MinWidth satisfy Flowable.
func (p *htmlParagraph) SpaceBefore() float64 { return p.style.SpaceBefore }
func (p *htmlParagraph) SpaceAfter() float64  { return p.style.SpaceAfter }
func (p *htmlParagraph) KeepWithNext() bool   { return p.style.KeepWithNextPara }
func (p *htmlParagraph) MinWidth() float64    { return 0 }

func (p *htmlParagraph) Wrap(doc *pdf.Document, availW, _ float64) (float64, float64) {
	innerW := availW - p.style.LeftIndent - p.style.RightIndent
	if innerW <= 0 {
		p.computedW, p.computedH = availW, 0
		return availW, 0
	}
	p.activateFont(doc)
	lh := p.leading(doc)
	p.lines = p.buildLines(doc, innerW)
	p.computedW = availW
	p.computedH = float64(len(p.lines)) * lh
	return availW, p.computedH
}

func (p *htmlParagraph) Draw(doc *pdf.Document, x, y float64) error {
	if len(p.lines) == 0 {
		return nil
	}
	x += p.style.LeftIndent
	innerW := p.computedW - p.style.LeftIndent - p.style.RightIndent
	lh := p.leading(doc)

	// Font size used for decoration line positioning and thickness.
	fs := p.style.FontSize
	if fs == 0 {
		fs, _ = doc.CurrentFontSize()
	}

	// Decoration lines follow the text color.
	lineColor := pdf.Color{R: 0, G: 0, B: 0}
	if p.style.TextColor != nil {
		doc.SetTextColor(p.style.TextColor.R, p.style.TextColor.G, p.style.TextColor.B)
		lineColor = *p.style.TextColor
	} else {
		doc.SetTextColor(0, 0, 0)
	}

	for _, line := range p.lines {
		lineW := p.measureLine(doc, line)
		lineX := x
		switch p.style.Alignment {
		case layout.AlignCenter:
			lineX = x + (innerW-lineW)/2
		case layout.AlignRight:
			lineX = x + innerW - lineW
		}
		cx := lineX
		for _, span := range line {
			p.activateFontForSpan(doc, span)
			doc.WriteLine(span.text, cx, y) //nolint:errcheck
			w, _ := doc.MeasureText(span.text)
			thickness := fs * 0.07
			if span.underline {
				doc.DrawLine(cx, y+fs*0.9, cx+w, y+fs*0.9, thickness, lineColor)
			}
			if span.strikethrough {
				doc.DrawLine(cx, y+fs*0.45, cx+w, y+fs*0.45, thickness, lineColor)
			}
			cx += w
		}
		y += lh
	}
	return nil
}

func (p *htmlParagraph) Split(doc *pdf.Document, availW, availH float64) []layout.Flowable {
	p.activateFont(doc)
	lh := p.leading(doc)
	innerW := availW - p.style.LeftIndent - p.style.RightIndent
	lines := p.buildLines(doc, innerW)
	maxLines := int(availH / lh)
	if maxLines <= 0 || maxLines >= len(lines) {
		return nil
	}
	p1 := &htmlParagraph{text: "", style: p.style, families: p.families, monoFont: p.monoFont, lines: lines[:maxLines]}
	p1.computedW = availW
	p1.computedH = float64(maxLines) * lh
	p1.style.SpaceAfter = 0

	p2 := &htmlParagraph{text: p.text, style: p.style, families: p.families, monoFont: p.monoFont}
	p2.style.SpaceBefore = 0
	// Reconstruct p2 text from remaining lines (approx; re-parse on Wrap)
	remaining := make([]string, 0, len(lines)-maxLines)
	for _, l := range lines[maxLines:] {
		parts := make([]string, len(l))
		for i, s := range l {
			parts[i] = s.text
		}
		remaining = append(remaining, strings.Join(parts, ""))
	}
	p2.text = strings.Join(remaining, " ")
	return []layout.Flowable{p1, p2}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (p *htmlParagraph) leading(doc *pdf.Document) float64 {
	if p.style.Leading > 0 {
		return p.style.Leading
	}
	fs := p.style.FontSize
	if fs == 0 {
		fs, _ = doc.CurrentFontSize()
	}
	return fs * 1.3
}

func (p *htmlParagraph) activateFont(doc *pdf.Document) {
	if p.style.FontName != "" {
		fs := p.style.FontSize
		if fs == 0 {
			fs, _ = doc.CurrentFontSize()
		}
		doc.SetFont(p.style.FontName, fs) //nolint:errcheck
	}
}

func (p *htmlParagraph) activateFontForSpan(doc *pdf.Document, span htmlSpan) {
	base := p.style.FontName
	if span.monospace && p.monoFont != "" {
		base = p.monoFont
	}
	fs := p.style.FontSize
	if fs == 0 {
		fs, _ = doc.CurrentFontSize()
	}
	name := p.resolveVariant(base, span.bold, span.italic)
	doc.SetFont(name, fs) //nolint:errcheck
}

// resolveVariant resolves bold/italic font name from the families map or by
// applying the conventional "Bold"/"Italic"/"BoldItalic" suffix.
func (p *htmlParagraph) resolveVariant(base string, bold, italic bool) string {
	if fam, ok := p.families[base]; ok {
		switch {
		case bold && italic && fam.boldItalic != "":
			return fam.boldItalic
		case bold && fam.bold != "":
			return fam.bold
		case italic && fam.italic != "":
			return fam.italic
		}
	}
	switch {
	case bold && italic:
		return base + "BoldItalic"
	case bold:
		return base + "Bold"
	case italic:
		return base + "Italic"
	default:
		return base
	}
}

// buildLines word-wraps the inner HTML into lines that fit within innerW.
func (p *htmlParagraph) buildLines(doc *pdf.Document, innerW float64) []htmlLine {
	// Parse inline HTML into spans.
	spans, err := pdfhtml.Parse(p.text, nil)
	if err != nil {
		// Fall back to plain text.
		spans = []pdfhtml.Span{{Text: p.text}}
	}

	// Tokenise: split spans into words preserving span style.
	type word struct {
		text   string
		bold   bool
		italic bool
		under  bool
		strike bool
		mono   bool
	}
	var words []word
	for _, sp := range spans {
		// Split on whitespace while preserving span style.
		parts := strings.FieldsFunc(sp.Text, unicode.IsSpace)
		for _, w := range parts {
			words = append(words, word{
				text:   w,
				bold:   sp.Style.Bold,
				italic: sp.Style.Italic,
				under:  sp.Style.Underline,
				strike: sp.Style.Strikethrough,
				mono:   sp.Style.Monospace,
			})
		}
		// Honour explicit newlines as paragraph breaks within the text.
		if strings.Contains(sp.Text, "\n") {
			words = append(words, word{text: "\n"})
		}
	}

	var lines []htmlLine
	var curLine htmlLine
	var curW float64

	flush := func() {
		if len(curLine) > 0 {
			lines = append(lines, curLine)
			curLine = nil
			curW = 0
		}
	}

	spaceW := p.measureWordStyled(doc, " ", false, false, false)

	for _, w := range words {
		if w.text == "\n" {
			flush()
			continue
		}
		ww := p.measureWordStyled(doc, w.text, w.bold, w.italic, w.mono)
		if curW+ww > innerW && len(curLine) > 0 {
			flush()
		}
		// Append to current span or start new span.
		if len(curLine) > 0 {
			last := &curLine[len(curLine)-1]
			if last.bold == w.bold && last.italic == w.italic && last.underline == w.under &&
				last.strikethrough == w.strike && last.monospace == w.mono {
				last.text += " " + w.text
				curW += spaceW + ww
				continue
			}
			// Measure space with the previous span's font for accurate width.
			prevSpaceW := p.measureWordStyled(doc, " ", last.bold, last.italic, last.monospace)
			last.text += " "
			curW += prevSpaceW
		}
		curLine = append(curLine, htmlSpan{
			text:          w.text,
			bold:          w.bold,
			italic:        w.italic,
			underline:     w.under,
			strikethrough: w.strike,
			monospace:     w.mono,
		})
		curW += ww
	}
	flush()
	return lines
}

func (p *htmlParagraph) measureWordStyled(doc *pdf.Document, w string, bold, italic, mono bool) float64 {
	base := p.style.FontName
	if mono && p.monoFont != "" {
		base = p.monoFont
	}
	name := p.resolveVariant(base, bold, italic)
	fs := p.style.FontSize
	if fs == 0 {
		fs, _ = doc.CurrentFontSize()
	}
	doc.SetFont(name, fs) //nolint:errcheck
	width, _ := doc.MeasureText(w)
	return width
}

func (p *htmlParagraph) measureLine(doc *pdf.Document, line htmlLine) float64 {
	w := 0.0
	for _, span := range line {
		p.activateFontForSpan(doc, span)
		sw, _ := doc.MeasureText(span.text)
		w += sw
	}
	return w
}

// ─── imageFlowable ───────────────────────────────────────────────────────────

type imageFlowable struct {
	path        string
	width       float64 // desired width  (0 = use natural or scale to availW)
	height      float64 // desired height (0 = proportional from width)
	align       layout.HAlign
	spaceBefore float64
	spaceAfter  float64
	computedW   float64
	computedH   float64
}

func (im *imageFlowable) SpaceBefore() float64 { return im.spaceBefore }
func (im *imageFlowable) SpaceAfter() float64  { return im.spaceAfter }
func (im *imageFlowable) KeepWithNext() bool   { return false }
func (im *imageFlowable) MinWidth() float64    { return 0 }

func (im *imageFlowable) Wrap(_ *pdf.Document, availW, _ float64) (float64, float64) {
	w := im.width
	h := im.height
	if w == 0 {
		w = availW
	}
	if w > availW {
		h = h * (availW / w)
		w = availW
	}
	if h == 0 {
		h = w // square fallback when no height given
	}
	im.computedW = w
	im.computedH = h
	return availW, h
}

func (im *imageFlowable) Draw(doc *pdf.Document, x, y float64) error {
	drawX := x
	switch im.align {
	case layout.AlignCenter:
		drawX = x + (im.computedW-im.width)/2
	case layout.AlignRight:
		drawX = x + im.computedW - im.width
	}
	return doc.DrawImage(im.path, drawX, y, im.computedW, im.computedH)
}

func (im *imageFlowable) Split(_ *pdf.Document, _, _ float64) []layout.Flowable {
	return nil // images don't split
}

// ─── listFlowable ────────────────────────────────────────────────────────────
//
// listFlowable renders an ordered or unordered list by converting each item
// into a Paragraph prefixed with a bullet or number.

type listFlowable struct {
	items       []listItem
	ordered     bool
	start       int // first number for ordered lists
	style       layout.ParagraphStyle
	bulletIndent float64
	spaceBefore float64
	spaceAfter  float64
	inner       []layout.Flowable // built during Wrap
	wrapWidth   float64           // stored from last Wrap call
}

type listItem struct {
	text  string
	style layout.ParagraphStyle // per-item style override (inherits from list)
}

func (l *listFlowable) SpaceBefore() float64 { return l.spaceBefore }
func (l *listFlowable) SpaceAfter() float64  { return l.spaceAfter }
func (l *listFlowable) KeepWithNext() bool   { return false }
func (l *listFlowable) MinWidth() float64    { return 0 }

const defaultBulletIndent = 18.0

func (l *listFlowable) buildInner() {
	indent := l.bulletIndent
	if indent == 0 {
		indent = defaultBulletIndent
	}
	l.inner = make([]layout.Flowable, len(l.items))
	for i, item := range l.items {
		var bullet string
		if l.ordered {
			bullet = orderedBullet(l.start+i) + "  "
		} else {
			bullet = "•  "
		}
		s := item.style
		if s.FontName == "" {
			s = l.style
		}
		s.LeftIndent += indent
		s.SpaceBefore = 0
		s.SpaceAfter = 2
		l.inner[i] = &layout.Paragraph{Text: bullet + item.text, Style: s}
	}
}

func orderedBullet(n int) string {
	return strconv.Itoa(n) + "."
}

func (l *listFlowable) Wrap(doc *pdf.Document, availW, availH float64) (float64, float64) {
	if l.inner == nil {
		l.buildInner()
	}
	l.wrapWidth = availW
	h := 0.0
	for _, f := range l.inner {
		_, fh := f.Wrap(doc, availW, availH-h)
		h += fh + f.SpaceBefore() + f.SpaceAfter()
	}
	return availW, h
}

func (l *listFlowable) Draw(doc *pdf.Document, x, y float64) error {
	for _, f := range l.inner {
		y += f.SpaceBefore()
		_, h := f.Wrap(doc, l.wrapWidth, 9999)
		if err := f.Draw(doc, x, y); err != nil {
			return err
		}
		y += h + f.SpaceAfter()
	}
	return nil
}

func (l *listFlowable) Split(doc *pdf.Document, availW, availH float64) []layout.Flowable {
	if l.inner == nil {
		l.buildInner()
	}
	h := 0.0
	splitAt := 0
	for i, f := range l.inner {
		_, fh := f.Wrap(doc, availW, availH)
		item := fh + f.SpaceBefore() + f.SpaceAfter()
		if h+item > availH {
			splitAt = i
			break
		}
		h += item
		splitAt = i + 1
	}
	if splitAt <= 0 || splitAt >= len(l.items) {
		return nil
	}
	l1 := *l
	l1.items = l.items[:splitAt]
	l1.inner = l.inner[:splitAt]
	l1.spaceAfter = 0
	l2 := *l
	l2.items = l.items[splitAt:]
	l2.inner = nil
	l2.spaceBefore = 0
	if l.ordered {
		l2.start = l.start + splitAt
	}
	return []layout.Flowable{&l1, &l2}
}

// ─── indentFlowable ──────────────────────────────────────────────────────────
//
// indentFlowable wraps a group of flowables, adding left/right indent.

type indentFlowable struct {
	left, right float64
	inner       []layout.Flowable
	wrapWidth   float64 // stored from last Wrap call
}

func (ind *indentFlowable) SpaceBefore() float64 { return 0 }
func (ind *indentFlowable) SpaceAfter() float64  { return 0 }
func (ind *indentFlowable) KeepWithNext() bool   { return false }
func (ind *indentFlowable) MinWidth() float64    { return 0 }

func (ind *indentFlowable) innerW(availW float64) float64 {
	return availW - ind.left - ind.right
}

func (ind *indentFlowable) Wrap(doc *pdf.Document, availW, availH float64) (float64, float64) {
	ind.wrapWidth = availW
	iw := ind.innerW(availW)
	h := 0.0
	for _, f := range ind.inner {
		_, fh := f.Wrap(doc, iw, availH-h)
		h += fh + f.SpaceBefore() + f.SpaceAfter()
	}
	return availW, h
}

func (ind *indentFlowable) Draw(doc *pdf.Document, x, y float64) error {
	iw := ind.innerW(ind.wrapWidth)
	ix := x + ind.left
	for _, f := range ind.inner {
		y += f.SpaceBefore()
		_, h := f.Wrap(doc, iw, 9999)
		if err := f.Draw(doc, ix, y); err != nil {
			return err
		}
		y += h + f.SpaceAfter()
	}
	return nil
}

func (ind *indentFlowable) Split(_ *pdf.Document, _, _ float64) []layout.Flowable {
	return nil
}
