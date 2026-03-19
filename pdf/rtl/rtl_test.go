package rtl

import (
	"testing"
)

// ── Arabic shaping ─────────────────────────────────────────────────────────

func TestShapeArabic_IsolatedLetter(t *testing.T) {
	// A lone ب (ba) must become its isolated form FE8F.
	got := shapeArabic("ب")
	want := string([]rune{0xFE8F})
	if got != want {
		t.Errorf("isolated ب: got %U want %U", []rune(got), []rune(want))
	}
}

func TestShapeArabic_InitialFinalPair(t *testing.T) {
	// بر = ba (initial) + ra (final, right-joining only).
	got := shapeArabic("بر")
	want := string([]rune{0xFE91, 0xFEAD}) // ba-initial, ra-isolated (ra is right-joining → isolated when nothing further connects)
	// Note: ra has no initial/medial form; ba is initial (connects left to ra which is R).
	// ra is isolated here because nothing further to its left connects.
	_ = want
	rs := []rune(got)
	if len(rs) != 2 {
		t.Fatalf("بر: expected 2 runes, got %d: %U", len(rs), rs)
	}
	if rs[0] != 0xFE91 {
		t.Errorf("ب in بر: want initial (FE91), got %U", rs[0])
	}
	// ra: isolated because it is right-joining (no left-side connection) and
	// nothing precedes it connects in "right-connect" direction yet.
	// Actually ra IS right-joining, so: prev=ba(D)→connectsRight=true, no initial/medial → final form (FEAE).
	if rs[1] != 0xFEAE {
		t.Errorf("ر in بر: want final (FEAE), got %U", rs[1])
	}
}

func TestShapeArabic_Word_Kitab(t *testing.T) {
	// كتاب (book): ك ت ا ب in logical order.
	// Expected forms: ك=initial(FEDB), ت=medial(FE98), ا=final(FE8E), ب=isolated(FE8F)
	got := shapeArabic("كتاب")
	rs := []rune(got)
	if len(rs) != 4 {
		t.Fatalf("كتاب: expected 4 runes, got %d: %U", len(rs), rs)
	}
	wantForms := []rune{0xFEDB, 0xFE98, 0xFE8E, 0xFE8F}
	for i, w := range wantForms {
		if rs[i] != w {
			t.Errorf("كتاب[%d]: want %U, got %U", i, w, rs[i])
		}
	}
}

func TestShapeArabic_LamAlefLigature(t *testing.T) {
	// لا (lam + alef) must collapse into a single lam-alef ligature.
	// Isolated (no preceding dual): FEFB.
	got := shapeArabic("لا")
	rs := []rune(got)
	if len(rs) != 1 {
		t.Fatalf("لا: expected 1 rune (ligature), got %d: %U", len(rs), rs)
	}
	if rs[0] != 0xFEFB {
		t.Errorf("لا isolated ligature: want FEFB, got %U", rs[0])
	}
}

func TestShapeArabic_LamAlefLigature_Final(t *testing.T) {
	// كلا: ك (dual) + ل + ا → lam-alef in final form because ك precedes it.
	got := shapeArabic("كلا")
	rs := []rune(got)
	// Expected: ك=initial, لا-ligature-final(FEFC)
	if len(rs) != 2 {
		t.Fatalf("كلا: expected 2 runes, got %d: %U", len(rs), rs)
	}
	if rs[1] != 0xFEFC {
		t.Errorf("لا in كلا: want FEFC (final ligature), got %U", rs[1])
	}
}

func TestShapeArabic_NonArabicPassThrough(t *testing.T) {
	// Latin and digits must pass through unchanged.
	got := shapeArabic("Hello 123")
	if got != "Hello 123" {
		t.Errorf("non-Arabic text changed: %q", got)
	}
}

func TestShapeArabic_MixedArabicLatin(t *testing.T) {
	// Latin surrounded by Arabic should not affect Arabic shaping.
	// ب X ب – the two ba letters should be isolated (broken by X).
	got := shapeArabic("ب X ب")
	rs := []rune(got)
	// ب (isolated=FE8F), space, X, space, ب (isolated=FE8F)
	if rs[0] != 0xFE8F {
		t.Errorf("first ب: want isolated (FE8F), got %U", rs[0])
	}
	if rs[4] != 0xFE8F {
		t.Errorf("second ب: want isolated (FE8F), got %U", rs[4])
	}
}

// ── BiDi reordering ────────────────────────────────────────────────────────

func TestReorder_PureHebrew(t *testing.T) {
	// שלום = shin lamed vav mem.  After reorder the runes must be reversed.
	input := "שלום"
	got := reorder(input)
	want := reverseRunes(input)
	if got != want {
		t.Errorf("Hebrew reorder: got %q want %q", got, want)
	}
}

func TestReorder_LTRUnchanged(t *testing.T) {
	input := "Hello World"
	got := reorder(input)
	if got != input {
		t.Errorf("LTR text should be unchanged: got %q want %q", got, input)
	}
}

func TestReorder_Empty(t *testing.T) {
	if got := reorder(""); got != "" {
		t.Errorf("empty string: got %q", got)
	}
}

// ── Public API ─────────────────────────────────────────────────────────────

func TestShape_HebrewRoundtrip(t *testing.T) {
	// Shape of pure Hebrew should simply reverse (no Arabic shaping needed).
	input := "שלום"
	got := Shape(input)
	want := reverseRunes(input)
	if got != want {
		t.Errorf("Shape(Hebrew): got %q want %q", got, want)
	}
}

func TestShape_ArabicProducesVisualOrder(t *testing.T) {
	// For a single Arabic word the shaped+reordered result must differ from
	// the original (characters reversed + presentation forms).
	input := "كتاب"
	shaped := Shape(input)
	original := []rune(input)
	result := []rune(shaped)
	if len(original) != len(result) {
		t.Fatalf("length changed: %d → %d", len(original), len(result))
	}
	// The result must be in visual order: the last shaped form comes first.
	// كتاب shaped logical: FEDB FE98 FE8E FE8F → visual reversed: FE8F FE8E FE98 FEDB
	wantVisual := []rune{0xFE8F, 0xFE8E, 0xFE98, 0xFEDB}
	for i, w := range wantVisual {
		if result[i] != w {
			t.Errorf("visual[%d]: want %U, got %U", i, w, result[i])
		}
	}
}

func TestShapeOnly_PreservesLogicalOrder(t *testing.T) {
	// ShapeOnly must produce shaped code points in logical (storage) order.
	got := ShapeOnly("كتاب")
	rs := []rune(got)
	// Logical order: ك=initial, ت=medial, ا=final, ب=isolated
	wantLogical := []rune{0xFEDB, 0xFE98, 0xFE8E, 0xFE8F}
	for i, w := range wantLogical {
		if rs[i] != w {
			t.Errorf("logical[%d]: want %U, got %U", i, w, rs[i])
		}
	}
}

func TestReorder_ExposedAPI(t *testing.T) {
	// Reorder on a pure Hebrew string should reverse it.
	input := "שלום"
	if got := Reorder(input); got != reverseRunes(input) {
		t.Errorf("Reorder(Hebrew): got %q want %q", got, reverseRunes(input))
	}
}
