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
