package pie_test

import (
	"bytes"; "os"; "testing"
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/pie"
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
func TestPieChart_Draw(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{FontName: "regular", FontSize: 9,
		Title: &chart.Title{Text: "Market share"}, Legend: &chart.Legend{},
		Series: []chart.Series{
			{Name: "Chrome", Data: []float64{65}}, {Name: "Firefox", Data: []float64{15}},
			{Name: "Safari", Data: []float64{12}}, {Name: "Other",  Data: []float64{8}},
		},
	}
	if err := (&pie.PieChart{Options: opts}).Draw(doc, 10, 10, 300, 250); err != nil { t.Fatalf("Draw: %v", err) }
	assertValidPDF(t, doc)
}
func TestPieChart_Donut(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{FontName: "regular", FontSize: 9,
		Series: []chart.Series{{Name: "A", Data: []float64{60}},{Name: "B", Data: []float64{40}}},
		PlotOptions: &chart.PlotOptions{Pie: &chart.PieOptions{InnerSize: "50%"}},
	}
	if err := (&pie.PieChart{Options: opts}).Draw(doc, 10, 10, 300, 250); err != nil { t.Fatalf("Draw donut: %v", err) }
}
func TestPieChart_EmptySeries(t *testing.T) {
	doc := newDoc(t)
	if err := (&pie.PieChart{Options: chart.Options{FontName: "regular"}}).Draw(doc, 10, 10, 300, 250); err != nil { t.Fatalf("Draw empty: %v", err) }
}
