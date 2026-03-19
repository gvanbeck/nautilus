package rtl

import (
	"strings"

	"golang.org/x/text/unicode/bidi"
)

// reorder applies the Unicode Bidirectional Algorithm to text and returns
// the string reordered for visual (left-to-right) display.
//
// RTL runs have their runes reversed so that a left-to-right renderer (such
// as gopdf) produces the correct visual result.  LTR runs are left as-is.
// Mixed paragraphs (e.g. Arabic embedded in Latin) are handled correctly.
func reorder(text string) string {
	if text == "" {
		return text
	}

	p := &bidi.Paragraph{}
	if _, err := p.SetString(text); err != nil {
		// Fallback: treat the whole string as a single RTL run.
		return reverseRunes(text)
	}

	order, err := p.Order()
	if err != nil {
		return reverseRunes(text)
	}

	var buf strings.Builder
	buf.Grow(len(text))
	for i := 0; i < order.NumRuns(); i++ {
		run := order.Run(i)
		s := run.String()
		if run.Direction() == bidi.RightToLeft {
			buf.WriteString(reverseRunes(s))
		} else {
			buf.WriteString(s)
		}
	}
	return buf.String()
}

// reverseRunes returns s with its Unicode runes in reversed order.
func reverseRunes(s string) string {
	rs := []rune(s)
	for l, r := 0, len(rs)-1; l < r; l, r = l+1, r-1 {
		rs[l], rs[r] = rs[r], rs[l]
	}
	return string(rs)
}
