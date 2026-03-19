package column_test

import (
	"bytes"; "os"; "testing"
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/column"
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
func TestColumnChart_Draw(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{FontName: "regular", FontSize: 9,
		Title: &chart.Title{Text: "Sales"},
		XAxis: &chart.Axis{Categories: []string{"Q1","Q2","Q3","Q4"}},
		YAxis: &chart.Axis{}, Legend: &chart.Legend{},
		Series: []chart.Series{{Name: "North", Data: []float64{43,55,57,60}},{Name: "South", Data: []float64{23,35,41,47}}},
	}
	if err := (&column.ColumnChart{Options: opts}).Draw(doc, 10, 10, 400, 250); err != nil { t.Fatalf("Draw: %v", err) }
	assertValidPDF(t, doc)
}
func TestColumnChart_Stacked(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{FontName: "regular", FontSize: 9,
		XAxis: &chart.Axis{Categories: []string{"A","B"}}, YAxis: &chart.Axis{},
		Series: []chart.Series{{Name: "X", Data: []float64{10,20}},{Name: "Y", Data: []float64{5,15}}},
		PlotOptions: &chart.PlotOptions{Column: &chart.ColumnOptions{Stacking: "normal"}},
	}
	if err := (&column.ColumnChart{Options: opts}).Draw(doc, 10, 10, 400, 250); err != nil { t.Fatalf("Draw stacked: %v", err) }
}
func TestColumnChart_EmptySeries(t *testing.T) {
	doc := newDoc(t)
	if err := (&column.ColumnChart{Options: chart.Options{FontName: "regular"}}).Draw(doc, 10, 10, 400, 250); err != nil { t.Fatalf("Draw empty: %v", err) }
}
