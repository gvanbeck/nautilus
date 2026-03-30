package emoji

import (
	"strings"
	"unicode/utf8"

	"github.com/forPelevin/gomoji"
	"github.com/rivo/uniseg"
)

// SegmentKind identifies the content type of a Segment.
type SegmentKind int

const (
	// KindText represents a run of plain Unicode text (no emoji).
	KindText SegmentKind = iota

	// KindEmoji represents a single emoji grapheme cluster.
	// The cluster may span multiple Unicode codepoints (e.g. skin-tone
	// modifiers or zero-width-joiner sequences such as 👨‍👩‍👧).
	KindEmoji
)

// Segment is a contiguous run of either plain text or a single emoji.
type Segment struct {
	// Kind indicates whether this segment is plain text or an emoji.
	Kind SegmentKind

	// Value is the raw UTF-8 string for this segment.
	// For KindEmoji it contains the full grapheme cluster including any
	// combining codepoints (modifiers, ZWJ, variation selectors).
	Value string
}

// Split breaks s into a slice of Segment values by iterating over Unicode
// grapheme clusters.  Adjacent clusters of the same kind are merged into a
// single Segment.
//
// The function correctly handles:
//   - Multi-codepoint emoji sequences joined by U+200D (ZWJ): 👨‍👩‍👧
//   - Skin-tone modifier sequences: 👍🏼
//   - Variation-selector sequences: ❤️
//   - Keycap sequences: 1️⃣
//   - Standard ASCII and Unicode text
func Split(s string) []Segment {
	var segments []Segment
	var buf strings.Builder
	currentKind := KindText

	g := uniseg.NewGraphemes(s)
	for g.Next() {
		cluster := g.Str()

		kind := KindText
		if couldBeEmoji(cluster) && gomoji.ContainsEmoji(cluster) {
			kind = KindEmoji
		}

		// When the kind changes, flush the current buffer as a new segment.
		if kind != currentKind && buf.Len() > 0 {
			segments = append(segments, Segment{Kind: currentKind, Value: buf.String()})
			buf.Reset()
		}

		currentKind = kind
		buf.WriteString(cluster)
	}

	// Flush any remaining content.
	if buf.Len() > 0 {
		segments = append(segments, Segment{Kind: currentKind, Value: buf.String()})
	}

	return segments
}

// couldBeEmoji is a fast pre-check that returns false when the first rune of
// cluster is definitely not an emoji.  It checks common emoji Unicode ranges
// to short-circuit the more expensive gomoji.ContainsEmoji call for plain
// ASCII and most ordinary text.
func couldBeEmoji(cluster string) bool {
	r, _ := utf8.DecodeRuneInString(cluster)
	if r == utf8.RuneError {
		return false
	}
	// Fast reject: plain ASCII printable text (and control chars) below
	// the handful of ASCII-range emoji codepoints.
	// Known ASCII-range emoji: #(0x23), *(0x2A), 0-9 (0x30-0x39) — these
	// only become emoji when followed by U+FE0F U+20E3 (keycap sequence),
	// but the grapheme cluster will contain those extra runes, so len > 1.
	if r < 0x80 {
		if len(cluster) > 1 && (r == '#' || r == '*' || (r >= '0' && r <= '9')) {
			return true
		}
		return false
	}
	// Common emoji ranges (non-exhaustive but covers the vast majority):
	//   U+00A9, U+00AE                 — © ®
	//   U+200D                         — ZWJ (shouldn't appear alone but be safe)
	//   U+203C - U+3299                — misc symbols
	//   U+FE0F                         — variation selector
	//   U+1F000 - U+1FAFF              — main emoji blocks
	//   U+E0020 - U+E007F              — tag sequences
	if r == 0xA9 || r == 0xAE {
		return true
	}
	if r >= 0x200D && r <= 0x3299 {
		return true
	}
	if r >= 0xFE00 && r <= 0xFE0F {
		return true
	}
	if r >= 0x1F000 && r <= 0x1FAFF {
		return true
	}
	if r >= 0xE0020 && r <= 0xE007F {
		return true
	}
	return false
}
