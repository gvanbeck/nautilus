package treemap_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/treemap"
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

func TestTreemapChart_Draw(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular", FontSize: 9,
		Title: &chart.Title{Text: "Revenue by Region"},
		Series: []chart.Series{{Points: []chart.Point{
			{Name: "North America", Y: 42},
			{Name: "Europe",        Y: 35},
			{Name: "Asia Pacific",  Y: 18},
			{Name: "Latin America", Y: 5},
		}}},
	}
	if err := (&treemap.TreemapChart{Options: opts}).Draw(doc, 10, 10, 400, 280); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	assertValidPDF(t, doc)
}

func TestTreemapChart_EmptySeries(t *testing.T) {
	doc := newDoc(t)
	if err := (&treemap.TreemapChart{Options: chart.Options{FontName: "regular"}}).Draw(doc, 10, 10, 400, 280); err != nil {
		t.Fatalf("Draw empty: %v", err)
	}
}

func TestTreemapChart_SingleCell(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular",
		Series: []chart.Series{{Points: []chart.Point{{Name: "Only", Y: 100}}}},
	}
	if err := (&treemap.TreemapChart{Options: opts}).Draw(doc, 10, 10, 400, 280); err != nil {
		t.Fatalf("Draw single: %v", err)
	}
}

func TestTreemapChart_ManyItems(t *testing.T) {
	doc := newDoc(t)
	pts := make([]chart.Point, 20)
	for i := range pts {
		pts[i] = chart.Point{Name: "Item", Y: float64(i + 1)}
	}
	opts := chart.Options{FontName: "regular", Series: []chart.Series{{Points: pts}}}
	if err := (&treemap.TreemapChart{Options: opts}).Draw(doc, 10, 10, 400, 280); err != nil {
		t.Fatalf("Draw many: %v", err)
	}
}
