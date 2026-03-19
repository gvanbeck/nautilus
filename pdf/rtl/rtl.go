// Package rtl provides Arabic character shaping and Unicode Bidirectional
// Algorithm (BiDi) reordering for right-to-left scripts.
//
// # Supported scripts
//
//   - Arabic: contextual letter forms (isolated/initial/medial/final) and
//     mandatory lam-alef ligatures are applied before BiDi reordering.
//   - Hebrew: no contextual shaping is required; only BiDi reordering is applied.
//   - Mixed RTL/LTR: the full BiDi paragraph algorithm handles correct run ordering.
//
// # Typical usage
//
// Single line:
//
//	shaped := rtl.Shape("مرحبا بالعالم")
//	w, _ := doc.MeasureText(shaped)
//	doc.WriteLineRTL(shaped, rightEdge, y)
//
// Multi-line (word-wrapped) paragraph – use WriteTextRTL on the Document or
// Frame directly; it calls ShapeOnly + Reorder internally per line:
//
//	doc.WriteTextRTL("مرحبا\nكيف حالك", rightEdge, y, maxWidth)
//
// # Font requirements
//
// Arabic: the registered font must include the Unicode Arabic Presentation
// Forms-B block (U+FE70–U+FEFF).  Fonts such as Noto Naskh Arabic or
// Amiri carry these code points alongside their OpenType GSUB tables.
//
// Hebrew: any font covering the Hebrew block (U+0590–U+05FF) is sufficient.
package rtl

// Shape prepares a string for single-line RTL rendering in a PDF.
//
// It applies Arabic contextual shaping (presentation forms + lam-alef
// ligatures) and then reorders all runs into visual (left-to-right glyph)
// order via the Unicode Bidirectional Algorithm.
//
// The returned string can be passed directly to Document.WriteLineRTL or
// measured with Document.MeasureText.
func Shape(text string) string {
	return reorder(shapeArabic(text))
}

// ShapeOnly applies Arabic contextual shaping to text without BiDi
// reordering.  The result stays in logical (storage) order with shaped
// code points.
//
// This is intended for use with Document.WriteTextRTL and Frame.WriteTextRTL,
// which wrap the shaped text into lines and call Reorder on each line
// individually.  Using ShapeOnly + Reorder per line preserves the correct
// word order across line breaks.
func ShapeOnly(text string) string {
	return shapeArabic(text)
}

// Reorder applies the Unicode Bidirectional Algorithm to text and returns
// the string in visual (left-to-right display) order.
//
// It is safe to call on already-shaped Arabic text (presentation forms are
// treated as RTL by the BiDi algorithm).  For Hebrew text no shaping is
// needed, so Reorder alone is sufficient.
func Reorder(text string) string {
	return reorder(text)
}
