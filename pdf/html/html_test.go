package html

import (
	"testing"
)

func TestParse_PlainText(t *testing.T) {
	spans, err := Parse("hello world", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("want 1 span, got %d", len(spans))
	}
	if spans[0].Text != "hello world" {
		t.Errorf("want %q, got %q", "hello world", spans[0].Text)
	}
	if spans[0].Style != (Style{}) {
		t.Errorf("want empty style, got %+v", spans[0].Style)
	}
}

func TestParse_Bold(t *testing.T) {
	spans, err := Parse("normal <b>bold</b> normal", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 3 {
		t.Fatalf("want 3 spans, got %d", len(spans))
	}
	if spans[1].Style.Bold != true {
		t.Error("middle span should be bold")
	}
	if spans[0].Style.Bold || spans[2].Style.Bold {
		t.Error("surrounding spans should not be bold")
	}
}

func TestParse_Italic(t *testing.T) {
	spans, err := Parse("<i>italic</i>", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 || !spans[0].Style.Italic {
		t.Error("want italic span")
	}
}

func TestParse_Underline(t *testing.T) {
	spans, err := Parse("<u>under</u>", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 || !spans[0].Style.Underline {
		t.Error("want underline span")
	}
}

func TestParse_Strong_Em(t *testing.T) {
	spans, err := Parse("<strong>s</strong><em>e</em>", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 2 {
		t.Fatalf("want 2 spans, got %d", len(spans))
	}
	if !spans[0].Style.Bold {
		t.Error("strong should be bold")
	}
	if !spans[1].Style.Italic {
		t.Error("em should be italic")
	}
}

func TestParse_Nested(t *testing.T) {
	spans, err := Parse("<b><i>bold italic</i></b>", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("want 1 span, got %d", len(spans))
	}
	s := spans[0].Style
	if !s.Bold || !s.Italic {
		t.Errorf("want bold+italic, got %+v", s)
	}
}

func TestParse_ClassPreserved(t *testing.T) {
	spans, err := Parse(`<span class="highlight">text</span>`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("want 1 span, got %d", len(spans))
	}
	if spans[0].Class != "highlight" {
		t.Errorf("want class %q, got %q", "highlight", spans[0].Class)
	}
	if spans[0].Style != (Style{}) {
		t.Errorf("want empty style without ClassStyle map, got %+v", spans[0].Style)
	}
}

func TestParse_ClassStyle(t *testing.T) {
	cs := ClassStyle{
		"highlight": {Bold: true},
	}
	spans, err := Parse(`<span class="highlight">text</span>`, cs)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("want 1 span, got %d", len(spans))
	}
	if spans[0].Class != "highlight" {
		t.Errorf("want class %q, got %q", "highlight", spans[0].Class)
	}
	if !spans[0].Style.Bold {
		t.Error("want bold from class style")
	}
}

func TestParse_TagWithClass(t *testing.T) {
	// class on a b tag should combine tag style with class style
	cs := ClassStyle{
		"extra": {Underline: true},
	}
	spans, err := Parse(`<b class="extra">text</b>`, cs)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("want 1 span, got %d", len(spans))
	}
	s := spans[0].Style
	if !s.Bold || !s.Underline {
		t.Errorf("want bold+underline, got %+v", s)
	}
}

func TestParse_Strikethrough(t *testing.T) {
	for _, tag := range []string{"s", "strike", "del"} {
		spans, err := Parse("<"+tag+">struck</"+tag+">", nil)
		if err != nil {
			t.Fatalf("%s: %v", tag, err)
		}
		if len(spans) != 1 || !spans[0].Style.Strikethrough {
			t.Errorf("<%s> should produce strikethrough span", tag)
		}
	}
}

func TestParse_Monospace(t *testing.T) {
	for _, tag := range []string{"code", "tt", "kbd", "samp"} {
		spans, err := Parse("<"+tag+">mono</"+tag+">", nil)
		if err != nil {
			t.Fatalf("%s: %v", tag, err)
		}
		if len(spans) != 1 || !spans[0].Style.Monospace {
			t.Errorf("<%s> should produce monospace span", tag)
		}
	}
}

func TestParse_SemanticItalic(t *testing.T) {
	for _, tag := range []string{"cite", "var", "dfn"} {
		spans, err := Parse("<"+tag+">text</"+tag+">", nil)
		if err != nil {
			t.Fatalf("%s: %v", tag, err)
		}
		if len(spans) != 1 || !spans[0].Style.Italic {
			t.Errorf("<%s> should produce italic span", tag)
		}
	}
}

func TestParse_Ins(t *testing.T) {
	spans, err := Parse("<ins>inserted</ins>", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 || !spans[0].Style.Underline {
		t.Error("<ins> should produce underline span")
	}
}

func TestParse_UnclosedTag(t *testing.T) {
	_, err := Parse("text <incomplete", nil)
	if err == nil {
		t.Error("want error for tag without closing >")
	}
}

func TestParse_InnermostClass(t *testing.T) {
	spans, err := Parse(`<span class="outer"><span class="inner">text</span></span>`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("want 1 span, got %d", len(spans))
	}
	if spans[0].Class != "inner" {
		t.Errorf("want innermost class %q, got %q", "inner", spans[0].Class)
	}
}

func TestParse_EmptyInput(t *testing.T) {
	spans, err := Parse("", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 0 {
		t.Errorf("want 0 spans, got %d", len(spans))
	}
}

func TestParse_SelfClosingTagIgnored(t *testing.T) {
	// Self-closing tags must not be pushed onto the style stack.
	spans, err := Parse("before<br/>after", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 2 {
		t.Fatalf("want 2 spans, got %d", len(spans))
	}
	if spans[0].Text != "before" || spans[1].Text != "after" {
		t.Errorf("unexpected texts: %q %q", spans[0].Text, spans[1].Text)
	}
}

func TestParse_CaseInsensitiveTags(t *testing.T) {
	spans, err := Parse("<B>bold</B> <STRONG>strong</STRONG>", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 3 {
		t.Fatalf("want 3 spans, got %d", len(spans))
	}
	if !spans[0].Style.Bold {
		t.Error("<B> should produce bold")
	}
	if !spans[2].Style.Bold {
		t.Error("<STRONG> should produce bold")
	}
}

func TestParse_UnknownClass(t *testing.T) {
	// An unknown class must be preserved in Span.Class but must not change Style.
	cs := ClassStyle{
		"known": {Bold: true},
	}
	spans, err := Parse(`<span class="unknown">text</span>`, cs)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("want 1 span, got %d", len(spans))
	}
	if spans[0].Class != "unknown" {
		t.Errorf("want class %q, got %q", "unknown", spans[0].Class)
	}
	if spans[0].Style != (Style{}) {
		t.Errorf("want empty style for unknown class, got %+v", spans[0].Style)
	}
}

func TestParse_ClassSingleQuote(t *testing.T) {
	spans, err := Parse(`<span class='sq'>text</span>`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 || spans[0].Class != "sq" {
		t.Errorf("want class %q, got %q", "sq", spans[0].Class)
	}
}

func TestParse_ClassUnquoted(t *testing.T) {
	spans, err := Parse(`<span class=uq>text</span>`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 || spans[0].Class != "uq" {
		t.Errorf("want class %q, got %q", "uq", spans[0].Class)
	}
}

func TestParse_ParentClassInheritedByChild(t *testing.T) {
	// Text inside a classed parent but below an unclassed child should still
	// report the parent's class via effectiveClass.
	spans, err := Parse(`<span class="outer"><b>text</b></span>`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("want 1 span, got %d", len(spans))
	}
	if spans[0].Class != "outer" {
		t.Errorf("want class %q, got %q", "outer", spans[0].Class)
	}
	if !spans[0].Style.Bold {
		t.Error("want bold from <b> inside classed span")
	}
}

func TestParse_StrikethroughAndItalic(t *testing.T) {
	spans, err := Parse("<s><i>struck italic</i></s>", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("want 1 span, got %d", len(spans))
	}
	s := spans[0].Style
	if !s.Strikethrough || !s.Italic {
		t.Errorf("want strikethrough+italic, got %+v", s)
	}
}

func TestParse_MonospaceAndBold(t *testing.T) {
	spans, err := Parse("<code><b>bold mono</b></code>", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("want 1 span, got %d", len(spans))
	}
	s := spans[0].Style
	if !s.Monospace || !s.Bold {
		t.Errorf("want monospace+bold, got %+v", s)
	}
}
