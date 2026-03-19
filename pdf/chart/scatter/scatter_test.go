package scatter_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/scatter"
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

func TestScatterChart_Draw(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular", FontSize: 9,
		Title: &chart.Title{Text: "Height vs Weight"},
		XAxis: &chart.Axis{}, YAxis: &chart.Axis{}, Legend: &chart.Legend{},
		Series: []chart.Series{
			{Name: "Group A", Points: []chart.Point{{X: 161, Y: 51}, {X: 175, Y: 72}, {X: 155, Y: 46}}},
			{Name: "Group B", Points: []chart.Point{{X: 183, Y: 88}, {X: 170, Y: 77}, {X: 185, Y: 90}}},
		},
	}
	if err := (&scatter.ScatterChart{Options: opts}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	assertValidPDF(t, doc)
}

func TestScatterChart_EmptyPoints(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{FontName: "regular", XAxis: &chart.Axis{}, YAxis: &chart.Axis{},
		Series: []chart.Series{{Name: "S", Points: []chart.Point{}}},
	}
	if err := (&scatter.ScatterChart{Options: opts}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw: %v", err)
	}
}
