package gauge_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/gauge"
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

func TestGaugeChart_Draw(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular", FontSize: 9,
		Title:  &chart.Title{Text: "Speed"},
		YAxis:  &chart.Axis{Min: chart.Float(0), Max: chart.Float(200)},
		Series: []chart.Series{{Name: "Speed", Data: []float64{120}}},
		PlotOptions: &chart.PlotOptions{Gauge: &chart.GaugeOptions{
			PaneStartAngle: -150, PaneEndAngle: 150,
			PlotBands: []chart.GaugePlotBand{
				{From: 0, To: 80, Color: pdf.Color{R: 85, G: 191, B: 59}},
				{From: 80, To: 140, Color: pdf.Color{R: 221, G: 223, B: 13}},
				{From: 140, To: 200, Color: pdf.Color{R: 223, G: 83, B: 83}},
			},
			DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
		}},
	}
	if err := (&gauge.GaugeChart{Options: opts}).Draw(doc, 10, 10, 250, 200); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	assertValidPDF(t, doc)
}

func TestGaugeChart_Solid(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular", FontSize: 9,
		YAxis:  &chart.Axis{Min: chart.Float(0), Max: chart.Float(100)},
		Series: []chart.Series{{Data: []float64{75}}},
		PlotOptions: &chart.PlotOptions{Gauge: &chart.GaugeOptions{
			PaneStartAngle: -90, PaneEndAngle: 90, Solid: true,
		}},
	}
	if err := (&gauge.GaugeChart{Options: opts}).Draw(doc, 10, 10, 250, 200); err != nil {
		t.Fatalf("Draw solid: %v", err)
	}
}

func TestGaugeChart_DefaultAngles(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular",
		YAxis:  &chart.Axis{Min: chart.Float(0), Max: chart.Float(100)},
		Series: []chart.Series{{Data: []float64{50}}},
	}
	if err := (&gauge.GaugeChart{Options: opts}).Draw(doc, 10, 10, 250, 200); err != nil {
		t.Fatalf("Draw defaults: %v", err)
	}
}
