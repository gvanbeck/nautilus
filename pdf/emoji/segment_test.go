package emoji_test

import (
	"testing"

	"github.com/gvanbeck/nautilus/pdf/emoji"
)

func TestSplit_plainText(t *testing.T) {
	segments := emoji.Split("Hello, World!")
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
	assertSegment(t, segments[0], emoji.KindText, "Hello, World!")
}

func TestSplit_emojiOnly(t *testing.T) {
	segments := emoji.Split("😀")
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
	assertSegment(t, segments[0], emoji.KindEmoji, "😀")
}

func TestSplit_textAndEmoji(t *testing.T) {
	// "Hello " + emoji + " World"
	segments := emoji.Split("Hello 👋 World")
	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d: %+v", len(segments), segments)
	}
	assertSegment(t, segments[0], emoji.KindText, "Hello ")
	assertSegment(t, segments[1], emoji.KindEmoji, "👋")
	assertSegment(t, segments[2], emoji.KindText, " World")
}

func TestSplit_multipleEmojis(t *testing.T) {
	segments := emoji.Split("😀😂🎉")
	// All three are consecutive emojis → merged into one emoji segment.
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment (consecutive emojis merged), got %d: %+v", len(segments), segments)
	}
	assertSegment(t, segments[0], emoji.KindEmoji, "😀😂🎉")
}

func TestSplit_skinToneModifier(t *testing.T) {
	// 👍🏼 is a base emoji + skin-tone modifier: should be one emoji segment.
	thumbsUp := "👍🏼"
	segments := emoji.Split(thumbsUp)
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment for skin-tone emoji, got %d: %+v", len(segments), segments)
	}
	assertSegment(t, segments[0], emoji.KindEmoji, thumbsUp)
}

func TestSplit_zwjSequence(t *testing.T) {
	// 👨‍👩‍👧 is a ZWJ sequence spanning multiple codepoints.
	family := "👨‍👩‍👧"
	segments := emoji.Split(family)
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment for ZWJ sequence, got %d: %+v", len(segments), segments)
	}
	assertSegment(t, segments[0], emoji.KindEmoji, family)
}

func TestSplit_variationSelector(t *testing.T) {
	// ❤️ = U+2764 + U+FE0F (variation selector-16).
	heart := "❤️"
	segments := emoji.Split(heart)
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment for variation-selector emoji, got %d: %+v", len(segments), segments)
	}
	assertSegment(t, segments[0], emoji.KindEmoji, heart)
}

func TestSplit_empty(t *testing.T) {
	segments := emoji.Split("")
	if len(segments) != 0 {
		t.Fatalf("expected 0 segments for empty string, got %d", len(segments))
	}
}

func TestSplit_unicodeText(t *testing.T) {
	// CJK and accented characters are plain text, not emoji.
	text := "こんにちは café"
	segments := emoji.Split(text)
	if len(segments) != 1 {
		t.Fatalf("expected 1 text segment for unicode text, got %d: %+v", len(segments), segments)
	}
	assertSegment(t, segments[0], emoji.KindText, text)
}

func TestSplit_mixedUnicodeAndEmoji(t *testing.T) {
	segments := emoji.Split("Bonjour 🌍 café")
	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d: %+v", len(segments), segments)
	}
	assertSegment(t, segments[0], emoji.KindText, "Bonjour ")
	assertSegment(t, segments[1], emoji.KindEmoji, "🌍")
	assertSegment(t, segments[2], emoji.KindText, " café")
}

// assertSegment is a test helper that verifies both the Kind and Value of a Segment.
func assertSegment(t *testing.T, seg emoji.Segment, wantKind emoji.SegmentKind, wantValue string) {
	t.Helper()
	if seg.Kind != wantKind {
		kindName := map[emoji.SegmentKind]string{
			emoji.KindText:  "KindText",
			emoji.KindEmoji: "KindEmoji",
		}
		t.Errorf("segment kind: got %s, want %s", kindName[seg.Kind], kindName[wantKind])
	}
	if seg.Value != wantValue {
		t.Errorf("segment value: got %q, want %q", seg.Value, wantValue)
	}
}
