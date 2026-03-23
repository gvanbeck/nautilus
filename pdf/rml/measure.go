package rml

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gvanbeck/nautilus/pdf"
)

// pt converts a measurement string to points.
// Supported formats: bare number (points), "21cm", "210mm", "8.27in", "595pt".
func pt(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	for suf, factor := range map[string]float64{
		"cm": 28.3465,
		"mm": 2.83465,
		"in": 72.0,
		"pt": 1.0,
	} {
		if strings.HasSuffix(s, suf) {
			v, err := strconv.ParseFloat(strings.TrimSuffix(s, suf), 64)
			if err != nil {
				return 0, fmt.Errorf("invalid measurement %q: %w", s, err)
			}
			return v * factor, nil
		}
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid measurement %q: %w", s, err)
	}
	return v, nil
}

func ptDef(s string, def float64) float64 {
	if s == "" {
		return def
	}
	v, err := pt(s)
	if err != nil {
		return def
	}
	return v
}

// parsePageSize parses "A4", "A3", "A5", "letter", "legal", or "(w,h)".
func parsePageSize(s string) (pdf.PageSize, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "a4":
		return pdf.PageSizeA4, nil
	case "a3":
		return pdf.PageSizeA3, nil
	case "a5":
		return pdf.PageSizeA5, nil
	case "letter":
		return pdf.PageSizeLetter, nil
	case "legal":
		return pdf.PageSizeLegal, nil
	}
	s = strings.Trim(s, "() ")
	parts := strings.SplitN(s, ",", 2)
	if len(parts) != 2 {
		return pdf.PageSizeA4, fmt.Errorf("unknown page size %q", s)
	}
	w, err := pt(strings.TrimSpace(parts[0]))
	if err != nil {
		return pdf.PageSizeA4, err
	}
	h, err := pt(strings.TrimSpace(parts[1]))
	if err != nil {
		return pdf.PageSizeA4, err
	}
	return pdf.PageSize{Width: w, Height: h}, nil
}

// parseColor parses named colors, "#rrggbb", and "r,g,b".
// Returns (color, true) on success or (zero, false) if s is empty or unknown.
func parseColor(s string) (pdf.Color, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return pdf.Color{}, false
	}
	named := map[string]pdf.Color{
		"black":     pdf.ColorBlack,
		"white":     pdf.ColorWhite,
		"lightgray": pdf.ColorLightGray,
		"lightgrey": pdf.ColorLightGray,
		"gray":      pdf.ColorGray,
		"grey":      pdf.ColorGray,
		"darkgray":  pdf.ColorDarkGray,
		"darkgrey":  pdf.ColorDarkGray,
		"red":       pdf.ColorRed,
		"green":     pdf.ColorGreen,
		"blue":      pdf.ColorBlue,
		"navy":      pdf.ColorNavy,
		"orange":    pdf.ColorOrange,
	}
	if c, ok := named[strings.ToLower(s)]; ok {
		return c, true
	}
	if strings.HasPrefix(s, "#") {
		hex := strings.TrimPrefix(s, "#")
		if len(hex) == 6 {
			r, e1 := strconv.ParseUint(hex[0:2], 16, 8)
			g, e2 := strconv.ParseUint(hex[2:4], 16, 8)
			b, e3 := strconv.ParseUint(hex[4:6], 16, 8)
			if e1 == nil && e2 == nil && e3 == nil {
				return pdf.Color{R: uint8(r), G: uint8(g), B: uint8(b)}, true
			}
		}
		return pdf.Color{}, false
	}
	parts := strings.Split(s, ",")
	if len(parts) == 3 {
		r, e1 := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 8)
		g, e2 := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 8)
		b, e3 := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 8)
		if e1 == nil && e2 == nil && e3 == nil {
			return pdf.Color{R: uint8(r), G: uint8(g), B: uint8(b)}, true
		}
	}
	return pdf.Color{}, false
}

// colorPtr returns a heap-allocated Color or nil when s is empty/unknown.
func colorPtr(s string) *pdf.Color {
	c, ok := parseColor(s)
	if !ok {
		return nil
	}
	return &c
}

// parseHAlignPDF maps an RML alignment string to pdf.HAlign (for table cells).
func parseHAlignPDF(s string) pdf.HAlign {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "center", "centre":
		return pdf.HAlignCenter
	case "right":
		return pdf.HAlignRight
	default:
		return pdf.HAlignLeft
	}
}

// parseVAlignPDF maps an RML valign string to pdf.VAlign.
func parseVAlignPDF(s string) pdf.VAlign {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "middle":
		return pdf.VAlignMiddle
	case "bottom":
		return pdf.VAlignBottom
	default:
		return pdf.VAlignTop
	}
}

// splitWidths parses "150,200,100" into []float64 of points.
func splitWidths(s string) ([]float64, error) {
	var result []float64
	for _, p := range strings.Split(s, ",") {
		v, err := pt(strings.TrimSpace(p))
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

// parseCoord parses "col,row" into (col, row int); negative values allowed.
func parseCoord(s string) (col, row int, err error) {
	parts := strings.SplitN(strings.TrimSpace(s), ",", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid coord %q (expected col,row)", s)
	}
	c, e1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	r, e2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if e1 != nil || e2 != nil {
		return 0, 0, fmt.Errorf("invalid coord %q", s)
	}
	return c, r, nil
}

// resolveIdx converts a possibly-negative index to a concrete index.
func resolveIdx(idx, length int) int {
	if idx < 0 {
		v := length + idx
		if v < 0 {
			return 0
		}
		return v
	}
	if idx >= length {
		return length - 1
	}
	return idx
}
