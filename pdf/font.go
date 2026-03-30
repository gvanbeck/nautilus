package pdf

import (
	"fmt"
	"path/filepath"
	"strings"
)

// fontFormat enumerates the font formats understood by the library.
type fontFormat int

const (
	fontFormatTTF fontFormat = iota
)

// detectFontFormat infers the font format from the file extension.
// Returns an error when the extension is not recognised.
func detectFontFormat(path string) (fontFormat, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ttf":
		return fontFormatTTF, nil
	default:
		return 0, fmt.Errorf("pdf: unsupported font format %q (only .ttf is supported)", filepath.Ext(path))
	}
}

// RegisterFont loads a TrueType (.ttf) font from path and registers it under
// name.  The name is used later when calling SetFont.
//
// Multiple fonts can be registered under different names, for example:
//
//	doc.RegisterFont("regular", "NotoSans-Regular.ttf")
//	doc.RegisterFont("bold",    "NotoSans-Bold.ttf")
//	doc.RegisterFont("italic",  "NotoSans-Italic.ttf")
func (d *Document) RegisterFont(name, path string) error {
	if name == "" {
		return fmt.Errorf("pdf: font name must not be empty")
	}

	format, err := detectFontFormat(path)
	if err != nil {
		return err
	}

	switch format {
	case fontFormatTTF:
		if err := d.pdf.AddTTFFont(name, path); err != nil {
			return fmt.Errorf("pdf: register font %q from %q: %w", name, path, err)
		}
	}

	return nil
}

// SetFont activates the previously registered font identified by name at the
// given size in points.  The font must have been registered with RegisterFont
// before calling SetFont.
//
// Changing the font size also affects the line height computed by WriteText.
//
// SetFont always calls gopdf even during the counting pass of Build so that
// text measurement (MeasureText, WrapLine) remains accurate for layout
// calculations such as paragraph height estimation.
func (d *Document) SetFont(name string, size float64) error {
	if name == "" {
		return fmt.Errorf("pdf: font name must not be empty")
	}
	if size <= 0 {
		return fmt.Errorf("pdf: font size must be positive, got %f", size)
	}

	// Always track font state so lineHeight() and table measurements work.
	d.fontSize = size
	d.currentFont = name

	if err := d.pdf.SetFont(name, "", size); err != nil {
		return fmt.Errorf("pdf: set font %q size %.1f: %w", name, size, err)
	}

	return nil
}

// CurrentFontSize returns the currently active font size in points.
// Returns 0 when no font has been set yet.
func (d *Document) CurrentFontSize() (float64, bool) {
	return d.fontSize, d.fontSize > 0
}

// MeasureText returns the width in points that text would occupy when rendered
// with the currently active font and size.  A font must be set before calling
// this method.
//
// Returns 0 during the counting pass of Build.
func (d *Document) MeasureText(text string) (float64, error) {
	if d.countingMode {
		return 0, nil
	}
	w, err := d.pdf.MeasureTextWidth(text)
	if err != nil {
		return 0, fmt.Errorf("pdf: measure text: %w", err)
	}
	return w, nil
}
