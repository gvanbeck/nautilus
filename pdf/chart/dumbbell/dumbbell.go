// Package dumbbell provides a dumbbell chart renderer for the Nautilus PDF library.
//
// A dumbbell chart connects two values (Low and High) per category with a line,
// drawing a circle marker at each end.  It shows the range or change between
// two states.  Data is stored in Series.Points with Low and High fields.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Life Expectancy"},
//	    XAxis:    &chart.Axis{Categories: []string{"Austria", "Belgium", "Germany"}},
//	    YAxis:    &chart.Axis{},
//	    Series: []chart.Series{
//	        {Name: "1990", Points: []chart.Point{{Low: 70.1}, {Low: 71.0}, {Low: 70.8}}},
//	        {Name: "2020", Points: []chart.Point{{High: 81.3}, {High: 81.9}, {High: 81.2}}},
//	    },
//	}
//
// More commonly used with a single series per point:
//
//	Series: []chart.Series{{
//	    Points: []chart.Point{
//	        {Name: "Austria", Low: 70.1, High: 81.3},
//	        {Name: "Belgium", Low: 71.0, High: 81.9},
//	    },
//	}},
package dumbbell

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// DumbbellChart renders a dumbbell chart onto a pdf.Document.
type DumbbellChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *DumbbellChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	do := dumbbellOptions(opts)
	layout := render.ComputeLayout(opts, x, y, width, height)

	render.DrawBackground(doc, opts, x, y, width, height)
	if err := render.DrawTitle(doc, opts, x, y, width); err != nil {
		return err
	}

	dataMin, dataMax := 0.0, 0.0
	first := true
	for _, s := range opts.Series {
		for _, p := range s.Points {
			for _, v := range []float64{p.Low, p.High} {
				if first || v < dataMin {
					dataMin = v
					first = false
				}
				if v > dataMax {
					dataMax = v
				}
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

	lw := do.LineWidth
	if lw <= 0 {
		lw = 2
	}
	m := render.MarkerOrDefault(do.Marker)
	markerR := m.Radius
	if markerR <= 0 {
		markerR = 4
	}
	sym := m.Symbol
	if sym == "" {
		sym = "circle"
	}

	for si, s := range opts.Series {
		color := chart.SeriesColor(opts, si)
		if s.Color != nil {
			color = *s.Color
		}
		nPts := len(s.Points)
		actualN := n
		if actualN == 0 {
			actualN = nPts
		}
		for i, p := range s.Points {
			cx := render.CategoryCenterX(i, actualN, layout.Plot)
			pyLow := render.ValueToY(p.Low, yMin, yMax, layout.Plot)
			pyHigh := render.ValueToY(p.High, yMin, yMax, layout.Plot)

			doc.DrawLine(cx, pyLow, cx, pyHigh, lw, render.LightenColor(color, 0.5))
			render.DrawMarker(doc, sym, cx, pyLow, markerR, color)
			render.DrawMarker(doc, sym, cx, pyHigh, markerR, color)

			if err := render.DrawDataLabel(doc, opts, do.DataLabels, p.High, cx, pyHigh); err != nil {
				return err
			}
		}
	}

	return render.DrawLegend(doc, opts, layout.Legend)
}

func dumbbellOptions(opts chart.Options) *chart.DumbbellOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Dumbbell != nil {
		return opts.PlotOptions.Dumbbell
	}
	return &chart.DumbbellOptions{}
}

var _ chart.Drawable = (*DumbbellChart)(nil)
