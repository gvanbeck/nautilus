package layout

import (
	"strings"

	"github.com/gvanbeck/nautilus/pdf"
)

// HAlign defines horizontal text alignment within the available width.
type HAlign int

const (
	// AlignLeft aligns text to the left edge of the available width.
	AlignLeft HAlign = iota

	// AlignCenter centres text within the available width.
	AlignCenter

	// AlignRight aligns text to the right edge of the available width.
	AlignRight
)

// ParagraphStyle defines the visual appearance of a Paragraph.
type ParagraphStyle struct {
	// FontName is the registered font name to activate before rendering.
	// Leave empty to use whichever font is currently active on the Document.
	FontName string

	// FontSize is the font size in points.  0 uses the Document default.
	FontSize float64

	// Leading is the line height in points.  0 defaults to FontSize × 1.2.
	Leading float64

	// Alignment controls horizontal placement of each rendered line.
	Alignment HAlign

	// SpaceBefore and SpaceAfter are extra whitespace above/below the paragraph.
	SpaceBefore float64
	SpaceAfter  float64

	// KeepWithNext prevents a frame/page break between this paragraph and
	// the immediately following flowable.
	KeepWithNextPara bool

	// LeftIndent and RightIndent reduce the usable text width from each side.
	LeftIndent  float64
	RightIndent float64

	// TextColor, when non-nil, sets the text colour for this paragraph.
	TextColor *pdf.Color
}

// Paragraph is a Flowable that renders word-wrapped text using a ParagraphStyle.
//
// It supports explicit line breaks (\n), per-style fonts and colours, and
// horizontal alignment.  Long paragraphs can be split across frames.
type Paragraph struct {
	baseFlowable

	// Text is the content to render.  Use \n for explicit paragraph breaks.
	Text string

	// Style controls the visual appearance.
	Style ParagraphStyle

	// lines is populated by Wrap and reused by Draw and Split.
	lines []string
	// lineBreaks tracks the index of lines that start a new \n-paragraph
	// (i.e. the first wrapped line of each explicit-newline block).
	lineBreaks []int
	// wrapWidth is the availWidth used during the last Wrap call.
	wrapWidth float64
}

// Wrap measures the paragraph and stores the wrapped lines for later use.
func (p *Paragraph) Wrap(doc *pdf.Document, availWidth, _ float64) (float64, float64) {
	p.applyFont(doc)

	innerW := availWidth - p.Style.LeftIndent - p.Style.RightIndent
	p.wrapWidth = availWidth

	var allLines []string
	var lineBreaks []int
	for _, para := range strings.Split(p.Text, "\n") {
		lineBreaks = append(lineBreaks, len(allLines))
		wrapped, err := doc.WrapLine(para, innerW)
		if err != nil || len(wrapped) == 0 {
			allLines = append(allLines, "")
		} else {
			allLines = append(allLines, wrapped...)
		}
	}
	p.lines = allLines
	p.lineBreaks = lineBreaks

	return availWidth, float64(len(p.lines)) * p.leading(doc)
}

// Draw renders the paragraph at (x, y) — the top-left corner.
func (p *Paragraph) Draw(doc *pdf.Document, x, y float64) error {
	p.applyFont(doc)
	if p.Style.TextColor != nil {
		doc.SetTextColor(p.Style.TextColor.R, p.Style.TextColor.G, p.Style.TextColor.B)
	}

	leading := p.leading(doc)
	contentX := x + p.Style.LeftIndent
	contentW := p.wrapWidth - p.Style.LeftIndent - p.Style.RightIndent

	for i, line := range p.lines {
		lineY := y + float64(i)*leading
		lineX := contentX

		switch p.Style.Alignment {
		case AlignCenter:
			lw, _ := doc.MeasureText(line)
			lineX = contentX + (contentW-lw)/2
		case AlignRight:
			lw, _ := doc.MeasureText(line)
			lineX = contentX + contentW - lw
		}

		if _, err := doc.WriteLine(line, lineX, lineY); err != nil {
			return err
		}
	}
	return nil
}

// Split divides the paragraph so that the first part fits within availHeight.
// Returns nil when not even a single line fits.
func (p *Paragraph) Split(doc *pdf.Document, availWidth, availHeight float64) []Flowable {
	p.applyFont(doc)

	leading := p.leading(doc)
	fitLines := int(availHeight / leading)

	if fitLines <= 0 {
		return nil // nothing fits; engine will move us to the next frame
	}
	if fitLines >= len(p.lines) {
		return []Flowable{p} // everything fits (defensive)
	}

	// Reconstruct text from the pre-wrapped lines, preserving explicit \n
	// breaks.  Lines within the same \n-paragraph are joined with a space;
	// \n-paragraph boundaries are joined with \n.
	p1 := &Paragraph{
		Text:  rejoinLines(p.lines[:fitLines], 0, p.lineBreaks),
		Style: p.Style,
	}
	p1.spaceBefore = p.baseFlowable.spaceBefore
	// No spaceAfter on the first part to avoid extra gap at the frame boundary.

	p2 := &Paragraph{
		Text:  rejoinLines(p.lines[fitLines:], fitLines, p.lineBreaks),
		Style: p.Style,
	}
	p2.spaceAfter = p.baseFlowable.spaceAfter

	return []Flowable{p1, p2}
}

// SpaceBefore returns the space before this paragraph (style takes precedence).
func (p *Paragraph) SpaceBefore() float64 {
	if p.Style.SpaceBefore != 0 {
		return p.Style.SpaceBefore
	}
	return p.baseFlowable.spaceBefore
}

// SpaceAfter returns the space after this paragraph (style takes precedence).
func (p *Paragraph) SpaceAfter() float64 {
	if p.Style.SpaceAfter != 0 {
		return p.Style.SpaceAfter
	}
	return p.baseFlowable.spaceAfter
}

// KeepWithNext returns true when this paragraph must not be separated from
// the following flowable.
func (p *Paragraph) KeepWithNext() bool {
	return p.Style.KeepWithNextPara || p.baseFlowable.keepWithNext
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (p *Paragraph) applyFont(doc *pdf.Document) {
	if p.Style.FontName != "" && p.Style.FontSize > 0 {
		_ = doc.SetFont(p.Style.FontName, p.Style.FontSize)
	} else if p.Style.FontName != "" {
		_ = doc.SetFont(p.Style.FontName, 12)
	}
}

func (p *Paragraph) leading(doc *pdf.Document) float64 {
	if p.Style.Leading > 0 {
		return p.Style.Leading
	}
	if p.Style.FontSize > 0 {
		return p.Style.FontSize * 1.2
	}
	// Fall back to the document's current line height.
	return 12 * 1.2
}

// rejoinLines reconstructs text from wrapped lines at global offset startIdx,
// reinserting \n at the original explicit newline boundaries recorded in
// lineBreaks (which stores global line indices where \n-paragraphs begin).
func rejoinLines(lines []string, startIdx int, lineBreaks []int) string {
	if len(lines) == 0 {
		return ""
	}
	breakSet := make(map[int]bool, len(lineBreaks))
	for _, idx := range lineBreaks {
		breakSet[idx] = true
	}
	var b strings.Builder
	b.WriteString(lines[0])
	for i := 1; i < len(lines); i++ {
		globalIdx := startIdx + i
		if breakSet[globalIdx] {
			b.WriteByte('\n')
		} else {
			b.WriteByte(' ')
		}
		b.WriteString(lines[i])
	}
	return b.String()
}
