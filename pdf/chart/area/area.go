// Package area provides an area (filled line) chart renderer for the Nautilus
// PDF library.  An area chart is a line chart with the region between the line
// and the zero baseline filled with a semi-transparent color.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Visitors per Month"},
//	    XAxis:    &chart.Axis{Categories: []string{"Jan", "Feb", "Mar"}},
//	    YAxis:    &chart.Axis{},
//	    Series: []chart.Series{
//	        {Name: "Mobile", Data: []float64{800, 1200, 950}},
//	        {Name: "Desktop", Data: []float64{500, 600, 700}},
//	    },
//	    PlotOptions: &chart.PlotOptions{
//	        Area: &chart.AreaOptions{FillAlpha: 0.25},
//	    },
//	}
//	ac := &area.AreaChart{Options: opts}
//	ac.Draw(doc, 50, 100, 400, 200)
package area

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// AreaChart renders a multi-series area (filled line) chart onto a pdf.Document.
type AreaChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *AreaChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
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

	ao := areaOptions(opts)
	zeroY := render.ValueToY(0, yMin, yMax, layout.Plot)

	for si, s := range opts.Series {
		color := chart.SeriesColor(opts, si)
		if s.Color != nil {
			color = *s.Color
		}
		if len(s.Data) == 0 {
			continue
		}

		pts := make([]pdf.Point, len(s.Data))
		for i, v := range s.Data {
			pts[i] = pdf.Point{
				X: render.CategoryCenterX(i, n, layout.Plot),
				Y: render.ValueToY(v, yMin, yMax, layout.Plot),
			}
		}

		// Build the fill polygon: start at bottom-left, trace the line,
		// then return along the baseline to bottom-left.
		alpha := ao.FillAlpha
		if alpha <= 0 {
			alpha = 0.3
		}
		fillColor := render.LightenColor(color, alpha)
		fill := make([]pdf.Point, 0, len(pts)+2)
		fill = append(fill, pdf.Point{X: pts[0].X, Y: zeroY})
		fill = append(fill, pts...)
		fill = append(fill, pdf.Point{X: pts[len(pts)-1].X, Y: zeroY})
		doc.FillPolygon(fill, fillColor)

		// Draw the line on top of the fill.
		lw := ao.LineWidth
		if lw <= 0 {
			lw = 2
		}
		for i := 1; i < len(pts); i++ {
			doc.DrawLine(pts[i-1].X, pts[i-1].Y, pts[i].X, pts[i].Y, lw, color)
		}

		// Draw markers.
		m := markerOrDefault(ao.Marker)
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
			if err := render.DrawDataLabel(doc, opts, ao.DataLabels, s.Data[i], p.X, p.Y); err != nil {
				return err
			}
		}
	}

	return render.DrawLegend(doc, opts, layout.Legend)
}

func areaOptions(opts chart.Options) *chart.AreaOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Area != nil {
		return opts.PlotOptions.Area
	}
	return &chart.AreaOptions{}
}

func markerOrDefault(m *chart.Marker) *chart.Marker {
	if m != nil {
		return m
	}
	return &chart.Marker{}
}
