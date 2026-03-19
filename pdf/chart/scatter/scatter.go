// Package scatter provides a scatter chart renderer for the Nautilus PDF library.
//
// Each data point is a [x, y] pair stored in Series.Points.
// No line connects the points; each is shown as a marker symbol.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Height vs Weight"},
//	    XAxis:    &chart.Axis{},
//	    YAxis:    &chart.Axis{},
//	    Series: []chart.Series{
//	        {Name: "Group A", Points: []chart.Point{
//	            {X: 1.74, Y: 67}, {X: 1.81, Y: 88}, {X: 1.69, Y: 55},
//	        }},
//	    },
//	    Legend: &chart.Legend{},
//	}
//	sc := &scatter.ScatterChart{Options: opts}
//	sc.Draw(doc, 50, 50, 400, 250)
package scatter

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// ScatterChart renders a scatter (X-Y point cloud) chart onto a pdf.Document.
type ScatterChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *ScatterChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	so := scatterOptions(opts)
	layout := render.ComputeLayout(opts, x, y, width, height)

	render.DrawBackground(doc, opts, x, y, width, height)
	if err := render.DrawTitle(doc, opts, x, y, width); err != nil {
		return err
	}

	// Compute axis ranges from point data.
	xMin, xMax := render.XDataRange(opts.Series)
	yDataMin, yDataMax := pointYRange(opts.Series)
	xMin, xMax, xStep := render.NiceRange(xMin, xMax, opts.XAxis)
	yMin, yMax, yStep := render.NiceRange(yDataMin, yDataMax, opts.YAxis)

	if err := render.DrawYAxis(doc, opts, layout.Plot, layout.YAxis, yMin, yMax, yStep); err != nil {
		return err
	}
	if err := render.DrawXAxisNumeric(doc, opts, layout.Plot, layout.XAxis, xMin, xMax, xStep); err != nil {
		return err
	}

	m := markerOrDefault(so.Marker)
	markerEnabled := render.BoolVal(m.Enabled, true)
	markerR := m.Radius
	if markerR <= 0 {
		markerR = 3
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
		for _, p := range s.Points {
			px := render.ValueToX(p.X, xMin, xMax, layout.Plot)
			py := render.ValueToY(p.Y, yMin, yMax, layout.Plot)
			if markerEnabled {
				render.DrawMarker(doc, sym, px, py, markerR, color)
			}
			if err := render.DrawDataLabel(doc, opts, so.DataLabels, p.Y, px, py); err != nil {
				return err
			}
		}
	}

	return render.DrawLegend(doc, opts, layout.Legend)
}

func pointYRange(series []chart.Series) (min, max float64) {
	first := true
	for _, s := range series {
		for _, p := range s.Points {
			if first || p.Y < min {
				min = p.Y
				first = false
			}
			if p.Y > max {
				max = p.Y
			}
		}
	}
	return
}

func scatterOptions(opts chart.Options) *chart.ScatterOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Scatter != nil {
		return opts.PlotOptions.Scatter
	}
	return &chart.ScatterOptions{}
}

func markerOrDefault(m *chart.Marker) *chart.Marker {
	if m != nil {
		return m
	}
	return &chart.Marker{}
}

// Ensure ScatterChart implements chart.Drawable.
var _ chart.Drawable = (*ScatterChart)(nil)
