// Package bullet provides a bullet chart renderer for the Nautilus PDF library.
//
// A bullet chart shows a primary bar (actual value), a target marker, and
// optional qualitative background bands.  Data is stored in Series.Points
// with Y (actual) and Target fields.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Revenue"},
//	    YAxis:    &chart.Axis{Min: chart.Float(0), Max: chart.Float(300)},
//	    Series: []chart.Series{{
//	        Name: "Actual",
//	        Points: []chart.Point{
//	            {Name: "Q1", Y: 180, Target: 220},
//	            {Name: "Q2", Y: 210, Target: 200},
//	            {Name: "Q3", Y: 150, Target: 240},
//	        },
//	    }},
//	    PlotOptions: &chart.PlotOptions{Bullet: &chart.BulletOptions{
//	        PlotBands: []chart.GaugePlotBand{
//	            {From: 0,   To: 150, Color: pdf.Color{R: 200, G: 200, B: 200}},
//	            {From: 150, To: 225, Color: pdf.Color{R: 180, G: 180, B: 180}},
//	            {From: 225, To: 300, Color: pdf.Color{R: 160, G: 160, B: 160}},
//	        },
//	    }},
//	}
//	bc := &bullet.BulletChart{Options: opts}
//	bc.Draw(doc, 50, 50, 400, 250)
package bullet

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// BulletChart renders a bullet chart onto a pdf.Document.
type BulletChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *BulletChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	bo := bulletOptions(opts)
	layout := render.ComputeLayout(opts, x, y, width, height)

	render.DrawBackground(doc, opts, x, y, width, height)
	if err := render.DrawTitle(doc, opts, x, y, width); err != nil {
		return err
	}

	// Compute value range (include targets).
	dataMax := 0.0
	for _, s := range opts.Series {
		for _, p := range s.Points {
			if p.Y > dataMax {
				dataMax = p.Y
			}
			if p.Target > dataMax {
				dataMax = p.Target
			}
		}
	}
	yMin, yMax, yStep := render.NiceRange(0, dataMax, opts.YAxis)

	if err := render.DrawYAxis(doc, opts, layout.Plot, layout.YAxis, yMin, yMax, yStep); err != nil {
		return err
	}

	// Collect all points across series into bullet rows.
	type row struct {
		name   string
		actual float64
		target float64
		color  pdf.Color
	}
	var rows []row
	for si, s := range opts.Series {
		color := chart.SeriesColor(opts, si)
		if s.Color != nil {
			color = *s.Color
		}
		for _, p := range s.Points {
			pc := color
			if p.Color != nil {
				pc = *p.Color
			}
			rows = append(rows, row{name: p.Name, actual: p.Y, target: p.Target, color: pc})
		}
	}
	if len(rows) == 0 {
		return render.DrawLegend(doc, opts, layout.Legend)
	}

	n := len(rows)
	slotH := layout.Plot.H / float64(n)
	barPad := slotH * 0.25
	barH := slotH - 2*barPad

	targetW := bo.TargetWidth
	if targetW == 0 {
		targetW = 0.15
	}
	targetColor := pdf.Color{R: 40, G: 40, B: 40}
	if bo.TargetColor != nil {
		targetColor = *bo.TargetColor
	}

	// Draw plot bands (vertical bands, full-height background).
	for _, pb := range bo.PlotBands {
		bx := render.ValueToX(pb.From, yMin, yMax, layout.Plot)
		ex := render.ValueToX(pb.To, yMin, yMax, layout.Plot)
		doc.FillRect(bx, layout.Plot.Y, ex-bx, layout.Plot.H, pb.Color)
	}

	fs := render.EffectiveFontSize(opts)

	for i, r := range rows {
		ry := layout.Plot.Y + float64(i)*slotH + barPad
		// Primary bar.
		bw := render.ValueToX(r.actual, yMin, yMax, layout.Plot) - layout.Plot.X
		doc.FillRect(layout.Plot.X, ry, bw, barH, r.color)

		// Target marker.
		tx := render.ValueToX(r.target, yMin, yMax, layout.Plot)
		tHalf := slotH * targetW
		doc.FillRect(tx-1, ry-tHalf/2+barH/2, 2, barH+tHalf, targetColor)

		// Row label.
		if r.name != "" && opts.FontName != "" {
			if err := doc.SetFont(opts.FontName, fs); err != nil {
				return err
			}
			doc.SetTextColor(60, 60, 60)
			lw, _ := doc.MeasureText(r.name)
			doc.WriteLine(r.name, layout.YAxis.X+layout.YAxis.W-lw-2, ry+barH/2-fs*0.4) //nolint:errcheck
		}
	}

	// Bottom baseline.
	doc.DrawLine(layout.Plot.X, layout.Plot.Bottom(), layout.Plot.Right(), layout.Plot.Bottom(), 0.5, render.DefaultAxisColor)

	return render.DrawLegend(doc, opts, layout.Legend)
}

func bulletOptions(opts chart.Options) *chart.BulletOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Bullet != nil {
		return opts.PlotOptions.Bullet
	}
	return &chart.BulletOptions{}
}

var _ chart.Drawable = (*BulletChart)(nil)
