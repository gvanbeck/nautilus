package polar_test

import (
	"bytes"; "os"; "testing"
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/polar"
)
func systemFont(t *testing.T) string {
	t.Helper()
	for _, p := range []string{"/Library/Fonts/Lato-Regular.ttf","/Library/Fonts/Arial.ttf","/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"} {
		if _, err := os.Stat(p); err == nil { return p }
	}
	t.Skip("no system font found"); return ""
}
func newDoc(t *testing.T) *pdf.Document {
	t.Helper()
	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4}); doc.AddPage()
	if err := doc.RegisterFont("regular", systemFont(t)); err != nil { t.Fatalf("RegisterFont: %v", err) }
	return doc
}
func assertValidPDF(t *testing.T, doc *pdf.Document) {
	t.Helper(); var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil { t.Fatalf("Output: %v", err) }
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) { t.Fatal("not a valid PDF") }
}
func TestPolarChart_Draw(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{FontName: "regular", FontSize: 9,
		Title: &chart.Title{Text: "Budget vs Spending"},
		XAxis: &chart.Axis{Categories: []string{"Sales","Marketing","Development","Support","IT","Admin"}},
		YAxis: &chart.Axis{Min: chart.Float(0)}, Legend: &chart.Legend{},
		Series: []chart.Series{
			{Name: "Budget",  Data: []float64{43000,19000,60000,35000,17000,10000}},
			{Name: "Spending",Data: []float64{50000,39000,42000,31000,26000,14000}},
		},
		PlotOptions: &chart.PlotOptions{Polar: &chart.PolarOptions{GridLineInterpolation: "polygon"}},
	}
	if err := (&polar.PolarChart{Options: opts}).Draw(doc, 10, 10, 400, 300); err != nil { t.Fatalf("Draw: %v", err) }
	assertValidPDF(t, doc)
}
func TestPolarChart_CircleGrid(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{FontName: "regular", FontSize: 9,
		XAxis: &chart.Axis{Categories: []string{"A","B","C","D"}},
		YAxis: &chart.Axis{Min: chart.Float(0)},
		Series: []chart.Series{{Name: "S", Data: []float64{10,20,15,25}}},
		PlotOptions: &chart.PlotOptions{Polar: &chart.PolarOptions{GridLineInterpolation: "circle"}},
	}
	if err := (&polar.PolarChart{Options: opts}).Draw(doc, 10, 10, 400, 300); err != nil { t.Fatalf("Draw circle grid: %v", err) }
}
func TestPolarChart_TooFewSpokes(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{FontName: "regular",
		XAxis:  &chart.Axis{Categories: []string{"A","B"}}, // < 3 spokes
		Series: []chart.Series{{Data: []float64{10,20}}},
	}
	if err := (&polar.PolarChart{Options: opts}).Draw(doc, 10, 10, 400, 300); err != nil { t.Fatalf("Draw: %v", err) }
}
