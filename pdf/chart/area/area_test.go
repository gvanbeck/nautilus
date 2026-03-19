package area_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/area"
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
	t.Skip("no system font found"); return ""
}

func newDoc(t *testing.T) *pdf.Document {
	t.Helper()
	doc, err := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	if err != nil { t.Fatalf("pdf.New: %v", err) }
	doc.AddPage()
	if err := doc.RegisterFont("regular", systemFont(t)); err != nil { t.Fatalf("RegisterFont: %v", err) }
	return doc
}

func assertValidPDF(t *testing.T, doc *pdf.Document) {
	t.Helper()
	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil { t.Fatalf("Output: %v", err) }
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) { t.Fatal("not a valid PDF") }
}

func TestAreaChart_Draw(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular", FontSize: 9,
		Title: &chart.Title{Text: "Visitors"},
		XAxis: &chart.Axis{Categories: []string{"Jan", "Feb", "Mar"}},
		YAxis: &chart.Axis{}, Legend: &chart.Legend{},
		Series: []chart.Series{
			{Name: "Mobile", Data: []float64{800, 950, 1100}},
			{Name: "Desktop", Data: []float64{500, 580, 620}},
		},
	}
	if err := (&area.AreaChart{Options: opts}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	assertValidPDF(t, doc)
}

func TestAreaChart_EmptySeries(t *testing.T) {
	doc := newDoc(t)
	if err := (&area.AreaChart{Options: chart.Options{FontName: "regular"}}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw empty: %v", err)
	}
}

func TestAreaChart_FillAlpha(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular", FontSize: 9,
		XAxis: &chart.Axis{Categories: []string{"A", "B"}}, YAxis: &chart.Axis{},
		Series: []chart.Series{{Name: "S", Data: []float64{10, 20}}},
		PlotOptions: &chart.PlotOptions{Area: &chart.AreaOptions{FillAlpha: 0.5}},
	}
	if err := (&area.AreaChart{Options: opts}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw: %v", err)
	}
}
