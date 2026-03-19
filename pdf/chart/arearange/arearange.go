// Package arearange provides an area range chart renderer for the Nautilus PDF
// library.
//
// An area range chart fills the band between a low and high value over
// discrete categories.  Data is stored in Series.Points with Low and High
// fields.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Temperature Range"},
//	    XAxis:    &chart.Axis{Categories: []string{"Jan", "Feb", "Mar", "Apr"}},
//	    YAxis:    &chart.Axis{},
//	    Series: []chart.Series{{
//	        Name: "Temp range",
//	        Points: []chart.Point{
//	            {Low: -9.5, High: 8.0},
//	            {Low: -7.8, High: 8.3},
//	            {Low: 0.4,  High: 13.1},
//	            {Low: 6.6,  High: 18.2},
//	        },
//	    }},
//	    Legend: &chart.Legend{},
//	}
//	ar := &arearange.AreaRangeChart{Options: opts}
//	ar.Draw(doc, 50, 50, 400, 250)
package arearange

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// AreaRangeChart renders an area range chart onto a pdf.Document.
type AreaRangeChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *AreaRangeChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	ao := areaRangeOptions(opts)
	layout := render.ComputeLayout(opts, x, y, width, height)

	render.DrawBackground(doc, opts, x, y, width, height)
	if err := render.DrawTitle(doc, opts, x, y, width); err != nil {
		return err
	}

	dataMin, dataMax := 0.0, 0.0
	first := true
	for _, s := range opts.Series {
		for _, p := range s.Points {
			if first || p.Low < dataMin {
				dataMin = p.Low
				first = false
			}
			if p.High > dataMax {
				dataMax = p.High
			}
		}
	}
	yMin, yMax, yStep := render.NiceRange(dataMin, dataMax, opts.YAxis)
	categories := render.CategoriesFor(opts)
	n := len(categories)

	if err := render.DrawYAxis(doc, opts, layout.Plot, layout.YAxis, yMin, yMax, yStep); err != nil {
		return err
	}
	if err := render.DrawXAxis(doc, opts, layout.Plot, layout.XAxis, categories); err != nil {
		return err
	}

	fillAlpha := ao.FillAlpha
	if fillAlpha == 0 {
		fillAlpha = 0.3
	}
	lw := ao.LineWidth
	if lw <= 0 {
		lw = 1.5
	}

	for si, s := range opts.Series {
		color := chart.SeriesColor(opts, si)
		if s.Color != nil {
			color = *s.Color
		}
		nPts := len(s.Points)
		if nPts == 0 {
			continue
		}
		actualN := n
		if actualN == 0 {
			actualN = nPts
		}

		// Build upper and lower edge coordinates.
		upper := make([]pdf.Point, nPts)
		lower := make([]pdf.Point, nPts)
		for i, p := range s.Points {
			cx := render.CategoryCenterX(i, actualN, layout.Plot)
			upper[i] = pdf.Point{X: cx, Y: render.ValueToY(p.High, yMin, yMax, layout.Plot)}
			lower[i] = pdf.Point{X: cx, Y: render.ValueToY(p.Low, yMin, yMax, layout.Plot)}
		}

		// Fill polygon: upper points forward, lower points backward.
		polygon := make([]pdf.Point, nPts*2)
		for i, pt := range upper {
			polygon[i] = pt
		}
		for i, pt := range lower {
			polygon[nPts*2-1-i] = pt
		}
		doc.FillPolygon(polygon, render.LightenColor(color, fillAlpha))

		// Upper edge line.
		for i := 1; i < len(upper); i++ {
			doc.DrawLine(upper[i-1].X, upper[i-1].Y, upper[i].X, upper[i].Y, lw, color)
		}
		// Lower edge line.
		for i := 1; i < len(lower); i++ {
			doc.DrawLine(lower[i-1].X, lower[i-1].Y, lower[i].X, lower[i].Y, lw, color)
		}

		// Data labels.
		for i, p := range s.Points {
			if err := render.DrawDataLabel(doc, opts, ao.DataLabels, p.High, upper[i].X, upper[i].Y); err != nil {
				return err
			}
		}
	}

	return render.DrawLegend(doc, opts, layout.Legend)
}

func areaRangeOptions(opts chart.Options) *chart.AreaRangeOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.AreaRange != nil {
		return opts.PlotOptions.AreaRange
	}
	return &chart.AreaRangeOptions{}
}

var _ chart.Drawable = (*AreaRangeChart)(nil)
