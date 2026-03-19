package funnel_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/funnel"
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

func TestFunnelChart_Draw(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular", FontSize: 9,
		Title: &chart.Title{Text: "Sales Funnel"}, Legend: &chart.Legend{},
		Series: []chart.Series{{Points: []chart.Point{
			{Name: "Visits",    Y: 15654},
			{Name: "Downloads", Y: 4064},
			{Name: "Invoices",  Y: 976},
			{Name: "Final",     Y: 846},
		}}},
		PlotOptions: &chart.PlotOptions{Funnel: &chart.FunnelOptions{
			DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
		}},
	}
	if err := (&funnel.FunnelChart{Options: opts}).Draw(doc, 10, 10, 300, 280); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	assertValidPDF(t, doc)
}

func TestFunnelChart_Pyramid(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular",
		Series: []chart.Series{{Points: []chart.Point{
			{Name: "Top", Y: 100}, {Name: "Mid", Y: 60}, {Name: "Base", Y: 20},
		}}},
		PlotOptions: &chart.PlotOptions{Funnel: &chart.FunnelOptions{Reversed: true}},
	}
	if err := (&funnel.FunnelChart{Options: opts}).Draw(doc, 10, 10, 300, 280); err != nil {
		t.Fatalf("Draw pyramid: %v", err)
	}
}

func TestFunnelChart_EmptySeries(t *testing.T) {
	doc := newDoc(t)
	if err := (&funnel.FunnelChart{Options: chart.Options{FontName: "regular"}}).Draw(doc, 10, 10, 300, 280); err != nil {
		t.Fatalf("Draw empty: %v", err)
	}
}
