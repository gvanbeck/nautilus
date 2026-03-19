package columnrange_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/columnrange"
)

func systemFont(t *testing.T) string {
	t.Helper()
	for _, p := range []string{
		"/Library/Fonts/Lato-Regular.ttf", "/Library/Fonts/Arial.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	t.Skip("no system font found")
	return ""
}

func newDoc(t *testing.T) *pdf.Document {
	t.Helper()
	doc, err := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	if err != nil {
		t.Fatalf("pdf.New: %v", err)
	}
	doc.AddPage()
	if err := doc.RegisterFont("regular", systemFont(t)); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}
	return doc
}

func assertValidPDF(t *testing.T, doc *pdf.Document) {
	t.Helper()
	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Fatal("not a valid PDF")
	}
}

func TestColumnRangeChart_Draw(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular", FontSize: 9,
		Title: &chart.Title{Text: "Temp Range"},
		XAxis: &chart.Axis{Categories: []string{"Jan", "Feb", "Mar", "Apr"}},
		YAxis: &chart.Axis{}, Legend: &chart.Legend{},
		Series: []chart.Series{{Name: "Temp", Points: []chart.Point{
			{Low: -9.5, High: 8.0}, {Low: -7.8, High: 8.3},
			{Low: -4.1, High: 13.0}, {Low: 0.4, High: 18.2},
		}}},
	}
	if err := (&columnrange.ColumnRangeChart{Options: opts}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	assertValidPDF(t, doc)
}

func TestColumnRangeChart_EmptySeries(t *testing.T) {
	doc := newDoc(t)
	if err := (&columnrange.ColumnRangeChart{Options: chart.Options{FontName: "regular"}}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw empty: %v", err)
	}
}
