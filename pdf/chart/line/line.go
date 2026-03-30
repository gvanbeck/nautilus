// Package line provides a line chart renderer for the Nautilus PDF library.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Monthly Revenue"},
//	    XAxis:    &chart.Axis{Categories: []string{"Jan", "Feb", "Mar", "Apr"}},
//	    YAxis:    &chart.Axis{},
//	    Series: []chart.Series{
//	        {Name: "2023", Data: []float64{120, 150, 130, 180}},
//	        {Name: "2024", Data: []float64{140, 160, 175, 210}},
//	    },
//	}
//	lc := &line.LineChart{Options: opts}
//	lc.Draw(doc, 50, 100, 400, 200)
package line

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// LineChart renders a multi-series line chart onto a pdf.Document.
type LineChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *LineChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
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

	lo := lineOptions(opts)

	for si, s := range opts.Series {
		color := chart.SeriesColor(opts, si)
		if s.Color != nil {
			color = *s.Color
		}
		if len(s.Data) == 0 {
			continue
		}

		// Compute point coordinates.
		pts := make([]pdf.Point, len(s.Data))
		for i, v := range s.Data {
			pts[i] = pdf.Point{
				X: render.CategoryCenterX(i, n, layout.Plot),
				Y: render.ValueToY(v, yMin, yMax, layout.Plot),
			}
		}

		// Draw connecting lines.
		lw := lo.LineWidth
		if lw <= 0 {
			lw = 2
		}
		for i := 1; i < len(pts); i++ {
			doc.DrawLine(pts[i-1].X, pts[i-1].Y, pts[i].X, pts[i].Y, lw, color)
		}

		// Draw markers.
		m := render.MarkerOrDefault(lo.Marker)
		if render.BoolVal(m.Enabled, true) {
			r := m.Radius
			if r <= 0 {
				r = 3
			}
			sym := m.Symbol
			if sym == "" {
				sym = "circle"
			}
			for _, p := range pts {
				render.DrawMarker(doc, sym, p.X, p.Y, r, color)
			}
		}

		// Draw data labels.
		for i, p := range pts {
			if err := render.DrawDataLabel(doc, opts, lo.DataLabels, s.Data[i], p.X, p.Y); err != nil {
				return err
			}
		}
	}

	return render.DrawLegend(doc, opts, layout.Legend)
}

func lineOptions(opts chart.Options) *chart.LineOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Line != nil {
		return opts.PlotOptions.Line
	}
	return &chart.LineOptions{}
}

