// Package html converts inline HTML into text spans for PDF rendering.
//
// Supported inline tags:
//   - <b>, <strong>              → bold
//   - <i>, <em>, <cite>, <var>, <dfn> → italic
//   - <u>, <ins>                 → underline
//   - <s>, <strike>, <del>       → strikethrough
//   - <code>, <tt>, <kbd>, <samp> → monospace
//   - any tag with a class attribute
//
// Tags may be freely nested.
package html

import (
	"errors"
	"strings"
)

// Style describes text formatting for a span of text.
type Style struct {
	Bold          bool
	Italic        bool
	Underline     bool
	Strikethrough bool
	Monospace     bool
}

// Span is a piece of text with associated formatting.
// Class holds the CSS class name of the innermost classed tag, if any.
// It is always preserved regardless of whether a ClassStyle map is provided.
type Span struct {
	Text  string
	Style Style
	Class string
}

// ClassStyle maps CSS class names to Style overrides.
// When passed to Parse, matching class styles are merged into Span.Style.
// The class name is always preserved in Span.Class regardless.
type ClassStyle map[string]Style

// Parse converts a string of inline HTML into a slice of Spans.
// If classes is nil, class attributes are preserved in Span.Class but
// do not affect Span.Style.
func Parse(input string, classes ClassStyle) ([]Span, error) {
	p := &parser{input: input, classes: classes}
	return p.parse()
}

type styleFrame struct {
	tagName       string
	bold          bool
	italic        bool
	underline     bool
	strikethrough bool
	monospace     bool
	class         string
}

type parser struct {
	input   string
	classes ClassStyle
	stack   []styleFrame
}

func (p *parser) parse() ([]Span, error) {
	var spans []Span
	i := 0
	for i < len(p.input) {
		if p.input[i] != '<' {
			j := strings.IndexByte(p.input[i:], '<')
			var text string
			if j == -1 {
				text = p.input[i:]
				i = len(p.input)
			} else {
				text = p.input[i : i+j]
				i += j
			}
			if text != "" {
				spans = append(spans, Span{
					Text:  text,
					Style: p.effectiveStyle(),
					Class: p.effectiveClass(),
				})
			}
			continue
		}

		end := strings.IndexByte(p.input[i:], '>')
		if end == -1 {
			return nil, errors.New("html: unclosed tag")
		}
		raw := p.input[i+1 : i+end]
		i += end + 1

		if strings.HasPrefix(raw, "/") {
			tagName := strings.ToLower(strings.TrimSpace(raw[1:]))
			p.pop(tagName)
		} else if !strings.HasSuffix(strings.TrimSpace(raw), "/") {
			p.push(raw)
		}
	}
	return spans, nil
}

func (p *parser) push(raw string) {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return
	}
	tagName := strings.ToLower(fields[0])
	frame := styleFrame{tagName: tagName}

	switch tagName {
	case "b", "strong":
		frame.bold = true
	case "i", "em", "cite", "var", "dfn":
		frame.italic = true
	case "u", "ins":
		frame.underline = true
	case "s", "strike", "del":
		frame.strikethrough = true
	case "code", "tt", "kbd", "samp":
		frame.monospace = true
	}

	frame.class = extractClass(raw)
	if frame.class != "" && p.classes != nil {
		if cs, ok := p.classes[frame.class]; ok {
			frame.bold = frame.bold || cs.Bold
			frame.italic = frame.italic || cs.Italic
			frame.underline = frame.underline || cs.Underline
			frame.strikethrough = frame.strikethrough || cs.Strikethrough
			frame.monospace = frame.monospace || cs.Monospace
		}
	}

	p.stack = append(p.stack, frame)
}

func (p *parser) pop(tagName string) {
	for i := len(p.stack) - 1; i >= 0; i-- {
		if p.stack[i].tagName == tagName {
			p.stack = append(p.stack[:i], p.stack[i+1:]...)
			return
		}
	}
}

func (p *parser) effectiveStyle() Style {
	var s Style
	for _, f := range p.stack {
		if f.bold {
			s.Bold = true
		}
		if f.italic {
			s.Italic = true
		}
		if f.underline {
			s.Underline = true
		}
		if f.strikethrough {
			s.Strikethrough = true
		}
		if f.monospace {
			s.Monospace = true
		}
	}
	return s
}

// effectiveClass returns the innermost class name on the stack.
func (p *parser) effectiveClass() string {
	for i := len(p.stack) - 1; i >= 0; i-- {
		if p.stack[i].class != "" {
			return p.stack[i].class
		}
	}
	return ""
}

// extractClass parses the value of a class attribute from a raw tag string,
// e.g. `span class="foo"` → "foo".
func extractClass(raw string) string {
	lower := strings.ToLower(raw)
	idx := strings.Index(lower, "class=")
	if idx == -1 {
		return ""
	}
	rest := raw[idx+6:]
	if len(rest) == 0 {
		return ""
	}
	if quote := rest[0]; quote == '"' || quote == '\'' {
		end := strings.IndexByte(rest[1:], quote)
		if end == -1 {
			return rest[1:]
		}
		return rest[1 : end+1]
	}
	// unquoted value
	end := strings.IndexAny(rest, " \t>")
	if end == -1 {
		return rest
	}
	return rest[:end]
}
