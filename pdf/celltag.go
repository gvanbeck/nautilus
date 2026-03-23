package pdf

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// CellsFromStruct generates a slice of Cells from an exported struct value.
// Each exported field becomes one cell; the field's formatted value is the
// cell text.  Style is driven by the `cell` struct tag.
//
// Tag syntax (semicolon-separated key=value pairs or bare flags):
//
//	cell:"halign=center;valign=middle;bg=240,240,240;color=0,0,128"
//	cell:"border=all;colspan=2;rowspan=1;font=Bold;size=10"
//	cell:"text=Override;format=%.2f;bold"
//	cell:"-"   ← skip this field entirely
//
// Supported keys:
//
//	text=…       override cell text (ignores field value)
//	format=…     fmt.Sprintf format for numeric fields (e.g. "%.2f")
//	header=…     alternate label used by HeaderCellsFromStruct
//	halign=      left | center | right
//	valign=      top | middle | bottom
//	bold         (bare flag) sets FontName to the "Bold" variant registered in the doc
//	font=…       explicit FontName
//	size=…       FontSize in points
//	bg=r,g,b     background color
//	color=r,g,b  text color
//	border=…     all | none | top | bottom | left | right (comma-separated sides)
//	colspan=n    ColSpan value (default 1)
//	rowspan=n    RowSpan value (default 1)
//	-            skip field
func CellsFromStruct(v any) ([]Cell, error) {
	return cellsFromValue(reflect.ValueOf(v), false)
}

// HeaderCellsFromStruct generates a header row from the struct's field names
// and `cell` tags.  The cell text is taken from the `header=` tag key when
// present; otherwise the exported field name is used.  All other style options
// from the `cell` tag are applied normally.
func HeaderCellsFromStruct(v any) ([]Cell, error) {
	return cellsFromValue(reflect.ValueOf(v), true)
}

// ─── internal ───────────────────────────────────────────────────────────────

func cellsFromValue(rv reflect.Value, headerMode bool) ([]Cell, error) {
	// dereference pointer
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, fmt.Errorf("celltag: nil pointer passed")
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("celltag: expected struct, got %s", rv.Kind())
	}

	rt := rv.Type()
	var cells []Cell

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("cell")
		if tag == "-" {
			continue
		}

		opts := parseTag(tag)

		// ── text ────────────────────────────────────────────────────────
		var text string
		if headerMode {
			if h, ok := opts["header"]; ok {
				text = h
			} else {
				text = field.Name
			}
		} else if override, ok := opts["text"]; ok {
			text = override
		} else {
			fv := rv.Field(i)
			if fmt_, ok := opts["format"]; ok {
				text = fmt.Sprintf(fmt_, fv.Interface())
			} else {
				text = fmt.Sprintf("%v", fv.Interface())
			}
		}

		// ── colspan / rowspan ────────────────────────────────────────────
		colspan := optInt(opts, "colspan", 1)
		rowspan := optInt(opts, "rowspan", 1)

		// ── style ────────────────────────────────────────────────────────
		style, err := styleFromOpts(opts)
		if err != nil {
			return nil, fmt.Errorf("celltag: field %s: %w", field.Name, err)
		}

		cells = append(cells, Cell{
			Text:    text,
			ColSpan: colspan,
			RowSpan: rowspan,
			Style:   style,
		})
	}
	return cells, nil
}

// parseTag splits "key=value;key2=value2;flag" into a map.
// Bare flags (no '=') are stored as key → "".
func parseTag(tag string) map[string]string {
	opts := make(map[string]string)
	for part := range strings.SplitSeq(tag, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if k, v, found := strings.Cut(part, "="); found {
			opts[k] = v
		} else {
			opts[part] = ""
		}
	}
	return opts
}

// styleFromOpts builds a CellStyle from parsed tag options.
func styleFromOpts(opts map[string]string) (CellStyle, error) {
	var s CellStyle

	// halign
	if v, ok := opts["halign"]; ok {
		switch strings.ToLower(v) {
		case "left":
			s.HAlign = HAlignLeft
		case "center":
			s.HAlign = HAlignCenter
		case "right":
			s.HAlign = HAlignRight
		default:
			return s, fmt.Errorf("unknown halign %q", v)
		}
	}

	// valign
	if v, ok := opts["valign"]; ok {
		switch strings.ToLower(v) {
		case "top":
			s.VAlign = VAlignTop
		case "middle":
			s.VAlign = VAlignMiddle
		case "bottom":
			s.VAlign = VAlignBottom
		default:
			return s, fmt.Errorf("unknown valign %q", v)
		}
	}

	// font / bold
	if v, ok := opts["font"]; ok {
		s.FontName = v
	}
	if _, ok := opts["bold"]; ok {
		// Append "Bold" suffix convention used by most registered font families.
		// e.g. "Helvetica" → "HelveticaBold"
		s.FontName = s.FontName + "Bold"
	}

	// size
	if v, ok := opts["size"]; ok {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return s, fmt.Errorf("invalid size %q: %w", v, err)
		}
		s.FontSize = f
	}

	// bg
	if v, ok := opts["bg"]; ok {
		c, err := parseColor(v)
		if err != nil {
			return s, fmt.Errorf("bg: %w", err)
		}
		s.Background = &c
	}

	// color
	if v, ok := opts["color"]; ok {
		c, err := parseColor(v)
		if err != nil {
			return s, fmt.Errorf("color: %w", err)
		}
		s.TextColor = &c
	}

	// border
	if v, ok := opts["border"]; ok {
		b, err := parseBorder(v)
		if err != nil {
			return s, fmt.Errorf("border: %w", err)
		}
		s.Border = b
	}

	return s, nil
}

// parseColor parses "r,g,b" into a Color.
func parseColor(s string) (Color, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return Color{}, fmt.Errorf("expected r,g,b got %q", s)
	}
	var rgb [3]uint8
	for i, p := range parts {
		n, err := strconv.ParseUint(strings.TrimSpace(p), 10, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid component %q: %w", p, err)
		}
		rgb[i] = uint8(n)
	}
	return Color{R: rgb[0], G: rgb[1], B: rgb[2]}, nil
}

// parseBorder parses "all", "none", or a comma-separated subset of
// "top,right,bottom,left" and returns a Border with a default 1pt solid spec
// on each requested side.
func parseBorder(s string) (Border, error) {
	defaultSpec := func() *BorderSpec { b := BorderSpec{Thickness: 1}; return &b }

	var b Border
	for side := range strings.SplitSeq(strings.ToLower(s), ",") {
		side = strings.TrimSpace(side)
		switch side {
		case "all":
			b.Top, b.Right, b.Bottom, b.Left = defaultSpec(), defaultSpec(), defaultSpec(), defaultSpec()
			return b, nil
		case "none":
			return Border{}, nil
		case "top":
			b.Top = defaultSpec()
		case "right":
			b.Right = defaultSpec()
		case "bottom":
			b.Bottom = defaultSpec()
		case "left":
			b.Left = defaultSpec()
		default:
			return b, fmt.Errorf("unknown border side %q", side)
		}
	}
	return b, nil
}

// optInt reads an integer option, returning def if absent or invalid.
func optInt(opts map[string]string, key string, def int) int {
	if v, ok := opts[key]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
