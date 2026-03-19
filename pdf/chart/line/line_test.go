package line_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/line"
)

func systemFont(t *testing.T) string {
	t.Helper()
	for _, p := range []string{
		"/Library/Fonts/Lato-Regular.ttf",
		"/Library/Fonts/Arial.ttf",
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
		t.Fatal("output is not a valid PDF")
	}
}

func TestLineChart_Draw(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular",
		FontSize: 9,
		Title:    &chart.Title{Text: "Monthly Revenue"},
		XAxis:    &chart.Axis{Categories: []string{"Jan", "Feb", "Mar", "Apr"}},
		YAxis:    &chart.Axis{},
		Legend:   &chart.Legend{},
		Series: []chart.Series{
			{Name: "2023", Data: []float64{120, 150, 130, 180}},
			{Name: "2024", Data: []float64{140, 165, 175, 210}},
		},
	}
	lc := &line.LineChart{Options: opts}
	if err := lc.Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	assertValidPDF(t, doc)
}

func TestLineChart_EmptySeries(t *testing.T) {
	doc := newDoc(t)
	lc := &line.LineChart{Options: chart.Options{FontName: "regular", FontSize: 9}}
	if err := lc.Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw with empty series: %v", err)
	}
}

func TestLineChart_DataLabels(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular",
		FontSize: 9,
		XAxis:    &chart.Axis{Categories: []string{"A", "B", "C"}},
		YAxis:    &chart.Axis{},
		Series:   []chart.Series{{Name: "S", Data: []float64{1, 2, 3}}},
		PlotOptions: &chart.PlotOptions{
			Line: &chart.LineOptions{
				DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
			},
		},
	}
	if err := (&line.LineChart{Options: opts}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw: %v", err)
	}
}
