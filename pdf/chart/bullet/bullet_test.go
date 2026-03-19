package bullet_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/bullet"
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

func TestBulletChart_Draw(t *testing.T) {
	doc := newDoc(t)
	opts := chart.Options{
		FontName: "regular", FontSize: 9,
		Title: &chart.Title{Text: "Performance"},
		YAxis: &chart.Axis{Min: chart.Float(0), Max: chart.Float(300)},
		Series: []chart.Series{{Points: []chart.Point{
			{Name: "Q1", Y: 180, Target: 220},
			{Name: "Q2", Y: 210, Target: 200},
			{Name: "Q3", Y: 150, Target: 240},
		}}},
		PlotOptions: &chart.PlotOptions{Bullet: &chart.BulletOptions{
			PlotBands: []chart.GaugePlotBand{
				{From: 0, To: 150, Color: pdf.Color{R: 200, G: 200, B: 200}},
				{From: 150, To: 300, Color: pdf.Color{R: 170, G: 170, B: 170}},
			},
		}},
	}
	if err := (&bullet.BulletChart{Options: opts}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	assertValidPDF(t, doc)
}

func TestBulletChart_EmptySeries(t *testing.T) {
	doc := newDoc(t)
	if err := (&bullet.BulletChart{Options: chart.Options{FontName: "regular"}}).Draw(doc, 10, 10, 400, 250); err != nil {
		t.Fatalf("Draw empty: %v", err)
	}
}
