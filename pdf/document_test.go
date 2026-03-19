package pdf_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
)

// systemFont returns the path of a TTF font present on the system, or skips
// the test when no candidate font can be found.  This keeps the test suite
// self-contained without bundling a font in the repository.
func systemFont(t *testing.T) string {
	t.Helper()

	candidates := []string{
		// macOS (common bundles)
		"/Library/Fonts/Arial.ttf",
		"/System/Library/Fonts/Supplemental/Arial.ttf",
		"/Library/Fonts/Lato-Medium.ttf",
		"/Library/Fonts/Helvetica.ttf",
		// Linux
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	t.Skip("no system TTF font found; skipping font-dependent test")
	return ""
}

// TestNew_defaults verifies that New accepts a zero-valued Config and applies
// sensible defaults.
func TestNew_defaults(t *testing.T) {
	doc, err := pdf.New(pdf.Config{})
	if err != nil {
		t.Fatalf("New(Config{}) error: %v", err)
	}
	if doc == nil {
		t.Fatal("New returned nil document")
	}
	// Default page size is A4.
	if doc.PageWidth() != pdf.PageSizeA4.Width {
		t.Errorf("PageWidth = %f, want %f", doc.PageWidth(), pdf.PageSizeA4.Width)
	}
	if doc.PageHeight() != pdf.PageSizeA4.Height {
		t.Errorf("PageHeight = %f, want %f", doc.PageHeight(), pdf.PageSizeA4.Height)
	}
}

// TestNew_customPageSize verifies that a non-default page size is applied.
func TestNew_customPageSize(t *testing.T) {
	doc, err := pdf.New(pdf.Config{PageSize: pdf.PageSizeA3})
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	if doc.PageWidth() != pdf.PageSizeA3.Width {
		t.Errorf("PageWidth = %f, want %f (A3)", doc.PageWidth(), pdf.PageSizeA3.Width)
	}
}

// TestRegisterFont_ttf verifies successful registration of a TTF font.
func TestRegisterFont_ttf(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{})
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont TTF: %v", err)
	}
}

// TestRegisterFont_invalidExtension verifies that an unrecognised extension is
// rejected with a descriptive error.
func TestRegisterFont_invalidExtension(t *testing.T) {
	doc, _ := pdf.New(pdf.Config{})
	err := doc.RegisterFont("bad", "/some/font.woff")
	if err == nil {
		t.Fatal("expected error for .woff font, got nil")
	}
}

// TestSetFont_noFont verifies that SetFont returns an error for an empty name.
func TestSetFont_emptyName(t *testing.T) {
	doc, _ := pdf.New(pdf.Config{})
	if err := doc.SetFont("", 12); err == nil {
		t.Fatal("expected error for empty font name, got nil")
	}
}

// TestSetFont_zeroSize verifies that SetFont rejects a zero font size.
func TestSetFont_zeroSize(t *testing.T) {
	doc, _ := pdf.New(pdf.Config{})
	if err := doc.SetFont("regular", 0); err == nil {
		t.Fatal("expected error for zero font size, got nil")
	}
}

// TestWriteLine_noEmoji is an integration test that generates a single-page PDF
// with a line of plain Unicode text and verifies the output is non-empty.
func TestWriteLine_noEmoji(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	doc.AddPage()

	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}
	if err := doc.SetFont("regular", 14); err != nil {
		t.Fatalf("SetFont: %v", err)
	}

	if _, err := doc.WriteLine("Hello, Unicode! こんにちは café", 50, 100); err != nil {
		t.Fatalf("WriteLine: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("Output produced empty PDF")
	}
	// Verify PDF header.
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Fatalf("output does not start with %%PDF- header")
	}
}

// TestWriteText_wordWrap is an integration test that exercises automatic word
// wrapping and verifies the returned Y position advances correctly.
func TestWriteText_wordWrap(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	doc.AddPage()
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}
	if err := doc.SetFont("regular", 12); err != nil {
		t.Fatalf("SetFont: %v", err)
	}

	startY := 50.0
	maxWidth := 200.0
	longText := "This is a long sentence that should be wrapped across multiple lines in the PDF document."

	endY, err := doc.WriteText(longText, 50, startY, maxWidth)
	if err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	if endY <= startY {
		t.Errorf("WriteText endY = %f, want > startY (%f): expected line wrapping", endY, startY)
	}
}

// TestWriteText_explicitNewlines verifies that \n characters produce line breaks.
func TestWriteText_explicitNewlines(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	doc.AddPage()
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}
	if err := doc.SetFont("regular", 12); err != nil {
		t.Fatalf("SetFont: %v", err)
	}

	startY := 50.0
	endY, err := doc.WriteText("Line one\nLine two\nLine three", 50, startY, 0)
	if err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	if endY <= startY {
		t.Errorf("WriteText endY = %f, want > startY (%f)", endY, startY)
	}
}

// TestSave writes a complete PDF to a temporary file and checks the file exists
// and contains a valid PDF header.
func TestSave(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	doc.AddPage()
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}
	if err := doc.SetFont("regular", 12); err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	if _, err := doc.WriteLine("Test", 50, 50); err != nil {
		t.Fatalf("WriteLine: %v", err)
	}

	out := filepath.Join(t.TempDir(), "test.pdf")
	if err := doc.Save(out); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read saved PDF: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatalf("saved file does not start with %%PDF- header")
	}
}

// TestSetLineHeightFactor_invalid verifies that non-positive factors are rejected.
func TestSetLineHeightFactor_invalid(t *testing.T) {
	doc, _ := pdf.New(pdf.Config{})
	if err := doc.SetLineHeightFactor(0); err == nil {
		t.Fatal("expected error for factor=0, got nil")
	}
	if err := doc.SetLineHeightFactor(-1); err == nil {
		t.Fatal("expected error for factor=-1, got nil")
	}
}

// TestMultiplePageSizes verifies that all named page sizes can be used without error.
func TestMultiplePageSizes(t *testing.T) {
	sizes := map[string]pdf.PageSize{
		"A3":     pdf.PageSizeA3,
		"A4":     pdf.PageSizeA4,
		"A5":     pdf.PageSizeA5,
		"Letter": pdf.PageSizeLetter,
		"Legal":  pdf.PageSizeLegal,
	}
	for name, size := range sizes {
		t.Run(name, func(t *testing.T) {
			doc, err := pdf.New(pdf.Config{PageSize: size})
			if err != nil {
				t.Fatalf("New(%s): %v", name, err)
			}
			doc.AddPage()
			if doc.PageWidth() != size.Width {
				t.Errorf("PageWidth = %f, want %f", doc.PageWidth(), size.Width)
			}
		})
	}
}
