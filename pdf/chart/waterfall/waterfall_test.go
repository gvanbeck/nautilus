package waterfall_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/waterfall"
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

func TestWaterfallChart_Draw(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular", FontSize: 9,
		Title: &chart.Title{Text: "Financials"},
		YAxis: &chart.Axis{},
		Series: []chart.Series{{Points: []chart.Point{
			{Name: "Start",    Y: 120000},
			{Name: "Revenue",  Y: 569000},
			{Name: "Costs",    Y: -342000},
			{Name: "Subtotal", IsIntermediateSum: true},
			{Name: "Balance",  IsSum: true},
		}}},
	}
	if err := (&waterfall.WaterfallChart{Options: opts}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	assertValidPDF(t, doc)
}

func TestWaterfallChart_EmptySeries(t *testing.T) {
	doc := newDoc(t)
	if err := (&waterfall.WaterfallChart{Options: chart.Options{FontName: "regular"}}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw empty: %v", err)
	}
}

func TestWaterfallChart_AllNegative(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular", YAxis: &chart.Axis{},
		Series: []chart.Series{{Points: []chart.Point{
			{Name: "A", Y: -100}, {Name: "B", Y: -50}, {Name: "Total", IsSum: true},
		}}},
	}
	if err := (&waterfall.WaterfallChart{Options: opts}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw: %v", err)
	}
}
