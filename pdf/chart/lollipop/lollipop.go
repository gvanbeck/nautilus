// Package lollipop provides a lollipop chart renderer for the Nautilus PDF library.
//
// A lollipop chart draws a thin line from zero (or baseline) to each value,
// with a circle at the data point.  It is a cleaner alternative to a bar chart
// for comparing point values.  Data comes from Series.Data (y-values).
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Top Products"},
//	    XAxis:    &chart.Axis{Categories: []string{"A", "B", "C", "D", "E"}},
//	    YAxis:    &chart.Axis{},
//	    Series: []chart.Series{
//	        {Name: "Units", Data: []float64{143, 112, 98, 87, 76}},
//	    },
//	    Legend: &chart.Legend{},
//	}
//	lc := &lollipop.LollipopChart{Options: opts}
//	lc.Draw(doc, 50, 50, 400, 250)
package lollipop

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// LollipopChart renders a lollipop chart onto a pdf.Document.
type LollipopChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *LollipopChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	lo := lollipopOptions(opts)
	layout := render.ComputeLayout(opts, x, y, width, height)

	render.DrawBackground(doc, opts, x, y, width, height)
	if err := render.DrawTitle(doc, opts, x, y, width); err != nil {
		return err
	}

	dataMin, dataMax := render.DataRange(opts.Series)
	yMin, yMax, yStep := render.NiceRange(dataMin, dataMax, opts.YAxis)
	categories := render.CategoriesFor(opts)
	n := len(categories)

	if err := render.DrawYAxis(doc, opts, layout.Plot, layout.YAxis, yMin, yMax, yStep); err != nil {
		return err
	}
	if err := render.DrawXAxis(doc, opts, layout.Plot, layout.XAxis, categories); err != nil {
		return err
	}

	lw := lo.LineWidth
	if lw <= 0 {
		lw = 1.5
	}
	m := render.MarkerOrDefault(lo.Marker)
	markerR := m.Radius
	if markerR <= 0 {
		markerR = 5
	}
	sym := m.Symbol
	if sym == "" {
		sym = "circle"
	}

	baseline := render.ValueToY(0, yMin, yMax, layout.Plot)

	for si, s := range opts.Series {
		color := chart.SeriesColor(opts, si)
		if s.Color != nil {
			color = *s.Color
		}
		for i, v := range s.Data {
			cx := render.CategoryCenterX(i, n, layout.Plot)
			py := render.ValueToY(v, yMin, yMax, layout.Plot)

			doc.DrawLine(cx, baseline, cx, py, lw, color)
			render.DrawMarker(doc, sym, cx, py, markerR, color)

			if err := render.DrawDataLabel(doc, opts, lo.DataLabels, v, cx, py); err != nil {
				return err
			}
		}
	}

	return render.DrawLegend(doc, opts, layout.Legend)
}

func lollipopOptions(opts chart.Options) *chart.LollipopOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Lollipop != nil {
		return opts.PlotOptions.Lollipop
	}
	return &chart.LollipopOptions{}
}

var _ chart.Drawable = (*LollipopChart)(nil)
