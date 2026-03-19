package emoji

import (
	"strings"

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
		if gomoji.ContainsEmoji(cluster) {
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
