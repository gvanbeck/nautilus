package boxplot_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/boxplot"
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

func TestBoxplotChart_Draw(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular", FontSize: 9,
		Title: &chart.Title{Text: "Observations"},
		XAxis: &chart.Axis{Categories: []string{"Loc A", "Loc B", "Loc C"}},
		YAxis: &chart.Axis{}, Legend: &chart.Legend{},
		Series: []chart.Series{{Name: "Obs", Points: []chart.Point{
			{Low: 760, Q1: 801, Median: 848, Q3: 895, High: 965},
			{Low: 733, Q1: 853, Median: 939, Q3: 980, High: 1080},
			{Low: 714, Q1: 762, Median: 817, Q3: 870, High: 918},
		}}},
	}
	if err := (&boxplot.BoxplotChart{Options: opts}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	assertValidPDF(t, doc)
}

func TestBoxplotChart_EmptySeries(t *testing.T) {
	doc := newDoc(t)
	if err := (&boxplot.BoxplotChart{Options: chart.Options{FontName: "regular"}}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw empty: %v", err)
	}
}

func TestBoxplotChart_MultipleSeries(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular", FontSize: 9,
		XAxis: &chart.Axis{Categories: []string{"A", "B"}}, YAxis: &chart.Axis{},
		Series: []chart.Series{
			{Name: "S1", Points: []chart.Point{{Low: 1, Q1: 2, Median: 3, Q3: 4, High: 5}, {Low: 2, Q1: 3, Median: 4, Q3: 5, High: 6}}},
			{Name: "S2", Points: []chart.Point{{Low: 0, Q1: 1, Median: 2, Q3: 3, High: 4}, {Low: 1, Q1: 2, Median: 3, Q3: 4, High: 5}}},
		},
	}
	if err := (&boxplot.BoxplotChart{Options: opts}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw: %v", err)
	}
}
