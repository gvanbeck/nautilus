package pdf

import (
	"fmt"
	"strings"

	"github.com/gvanbeck/nautilus/pdf/emoji"
	"github.com/gvanbeck/nautilus/pdf/rtl"
	"github.com/signintech/gopdf"
)

// WriteLine renders text on a single line starting at (x, y).
//
// The text may contain any Unicode characters including emoji.  Emoji
// grapheme clusters are replaced with PNG images when an EmojiResolver is
// configured; otherwise they are silently skipped.
//
// Returns the X coordinate immediately after the last rendered element,
// which can be used to continue writing on the same line.
//
// Returns x unchanged during the counting pass of Build.
func (d *Document) WriteLine(text string, x, y float64) (float64, error) {
	if d.countingMode {
		return x, nil
	}
	segments := emoji.Split(text)
	currentX := x

	for _, seg := range segments {
		var err error
		switch seg.Kind {
		case emoji.KindText:
			currentX, err = d.renderText(seg.Value, currentX, y)
		case emoji.KindEmoji:
			currentX, err = d.renderEmoji(seg.Value, currentX, y)
		}
		if err != nil {
			return currentX, err
		}
	}

	return currentX, nil
}

// WriteText renders text with automatic word wrapping within maxWidth points.
//
// Explicit newline characters (\n) are always honoured.  Long lines are broken
// at word boundaries (spaces).  The Y position is advanced by one line height
// after each rendered line.
//
// Returns the Y coordinate immediately below the last rendered line.  This
// value can be used to continue writing content below the text block.
//
// When maxWidth is 0 or negative, no wrapping is applied and the text is
// rendered as-is (newlines still apply).
//
// Returns y unchanged during the counting pass of Build.
func (d *Document) WriteText(text string, x, y, maxWidth float64) (float64, error) {
	if d.countingMode {
		return y, nil
	}
	// Honour explicit newlines first.
	paragraphs := strings.Split(text, "\n")
	currentY := y

	for _, para := range paragraphs {
		lines, err := d.wrapLine(para, maxWidth)
		if err != nil {
			return currentY, err
		}
		for _, line := range lines {
			if _, err := d.WriteLine(line, x, currentY); err != nil {
				return currentY, err
			}
			currentY += d.lineHeight()
		}
	}

	return currentY, nil
}

// measureWord returns the width in points of word, accounting for any emoji
// segments within the word.  Each emoji is assumed to be a square with side
// equal to the current font size.
func (d *Document) measureWord(word string) (float64, error) {
	segments := emoji.Split(word)
	total := 0.0

	for _, seg := range segments {
		switch seg.Kind {
		case emoji.KindText:
			w, err := d.pdf.MeasureTextWidth(seg.Value)
			if err != nil {
				return 0, fmt.Errorf("pdf: measure word %q: %w", word, err)
			}
			total += w
		case emoji.KindEmoji:
			// Emoji are rendered as squares sized to the current font size.
			total += d.fontSize
		}
	}

	return total, nil
}

// wrapLine splits a single line of text (no newlines) into a slice of lines
// that each fit within maxWidth.  When maxWidth is 0 or negative, the
// original text is returned unchanged.
func (d *Document) wrapLine(text string, maxWidth float64) ([]string, error) {
	if maxWidth <= 0 || text == "" {
		return []string{text}, nil
	}

	// Split on spaces to obtain individual tokens.
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}, nil
	}

	spaceWidth, err := d.pdf.MeasureTextWidth(" ")
	if err != nil {
		return nil, fmt.Errorf("pdf: measure space: %w", err)
	}

	var lines []string
	var currentWords []string
	currentWidth := 0.0

	for _, word := range words {
		wordWidth, err := d.measureWord(word)
		if err != nil {
			return nil, err
		}

		if len(currentWords) == 0 {
			// First word on a fresh line — always placed even if it overflows.
			currentWords = append(currentWords, word)
			currentWidth = wordWidth
		} else if currentWidth+spaceWidth+wordWidth <= maxWidth {
			// Word fits on the current line.
			currentWords = append(currentWords, word)
			currentWidth += spaceWidth + wordWidth
		} else {
			// Word does not fit — start a new line.
			lines = append(lines, strings.Join(currentWords, " "))
			currentWords = currentWords[:0]
			currentWords = append(currentWords, word)
			currentWidth = wordWidth
		}
	}

	if len(currentWords) > 0 {
		lines = append(lines, strings.Join(currentWords, " "))
	}

	return lines, nil
}

// WrapLine splits a single line of text (no newlines) into a slice of lines
// that each fit within maxWidth using the current font.
// When maxWidth is 0 or negative the original text is returned unchanged.
// This is the exported version of wrapLine for use by layout packages.
func (d *Document) WrapLine(text string, maxWidth float64) ([]string, error) {
	return d.wrapLine(text, maxWidth)
}

// measureLines counts how many lines text would occupy when wrapped to
// maxWidth using the current font.  Explicit newlines are counted as paragraph
// breaks.  Returns at least 1 even for empty text.
//
// A font must be set before calling this method.  During the counting pass of
// Build, wrapLine's MeasureTextWidth call is not guarded, so measureLines
// should only be called during the rendering pass.
func (d *Document) measureLines(text string, maxWidth float64) (int, error) {
	if text == "" {
		return 1, nil
	}
	total := 0
	for _, para := range strings.Split(text, "\n") {
		lines, err := d.wrapLine(para, maxWidth)
		if err != nil {
			return 0, fmt.Errorf("pdf: measure lines: %w", err)
		}
		if len(lines) == 0 {
			total++ // empty paragraph counts as one blank line
		} else {
			total += len(lines)
		}
	}
	return total, nil
}

// WriteLineRTL renders a single line of right-to-left text with its right
// edge at rightX.
//
// The text must already be in visual (display) order.  Use rtl.Shape before
// calling this method to apply Arabic contextual shaping and BiDi reordering.
//
//	shaped := rtl.Shape("مرحبا بالعالم")
//	doc.WriteLineRTL(shaped, rightEdge, y)
//
// Returns the left-edge X of the rendered text.
// Returns rightX unchanged during the counting pass of Build.
func (d *Document) WriteLineRTL(text string, rightX, y float64) (float64, error) {
	if d.countingMode {
		return rightX, nil
	}
	w, err := d.MeasureText(text)
	if err != nil {
		return rightX, err
	}
	if _, err := d.WriteLine(text, rightX-w, y); err != nil {
		return rightX - w, err
	}
	return rightX - w, nil
}

// WriteTextRTL renders text with automatic word wrapping for right-to-left
// scripts, right-aligned within maxWidth points to the left of rightX.
//
// The text must be in its original (logical) form — WriteTextRTL calls
// rtl.ShapeOnly and rtl.Reorder internally on each wrapped line so that
// word order across line breaks is preserved correctly.
//
// Explicit newline characters (\n) are treated as paragraph breaks.
//
// Returns the Y coordinate below the last rendered line.
// Returns y unchanged during the counting pass of Build.
func (d *Document) WriteTextRTL(text string, rightX, y, maxWidth float64) (float64, error) {
	if d.countingMode {
		return y, nil
	}
	currentY := y
	for _, para := range strings.Split(text, "\n") {
		// Shape Arabic in logical order first, then wrap by width.
		shaped := rtl.ShapeOnly(para)
		lines, err := d.wrapLine(shaped, maxWidth)
		if err != nil {
			return currentY, err
		}
		for _, line := range lines {
			visual := rtl.Reorder(line)
			w, err := d.MeasureText(visual)
			if err != nil {
				return currentY, err
			}
			if _, err := d.WriteLine(visual, rightX-w, currentY); err != nil {
				return currentY, err
			}
			currentY += d.lineHeight()
		}
	}
	return currentY, nil
}

// renderText writes a plain-text string at (x, y) using the current font and
// returns the X position after the text.
func (d *Document) renderText(text string, x, y float64) (float64, error) {
	d.pdf.SetX(x)
	d.pdf.SetY(y)

	if err := d.pdf.Cell(nil, text); err != nil {
		return x, fmt.Errorf("pdf: render text %q: %w", text, err)
	}

	return d.pdf.GetX(), nil
}

// renderEmoji places the PNG image for the emoji at (x, y) and returns the X
// position after the image.  The image is sized as a square with side equal to
// the current font size so that it visually matches surrounding text.
//
// If no resolver is set or the emoji cannot be resolved, the function returns
// x unchanged (the emoji is silently skipped).
func (d *Document) renderEmoji(cluster string, x, y float64) (float64, error) {
	if d.resolver == nil {
		return x, nil
	}

	path, found := d.resolver.Resolve(cluster)
	if !found {
		return x, nil
	}

	size := d.fontSize
	err := d.pdf.Image(path, x, y, &gopdf.Rect{W: size, H: size})
	if err != nil {
		return x, fmt.Errorf("pdf: render emoji %q from %q: %w", cluster, path, err)
	}

	return x + size, nil
}
