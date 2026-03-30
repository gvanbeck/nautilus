// Package bubble provides a bubble chart renderer for the Nautilus PDF library.
//
// A bubble chart is a scatter chart where a third value (Z) controls each
// circle's radius.  Data is stored in Series.Points with X, Y, and Z fields.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Countries"},
//	    XAxis:    &chart.Axis{},
//	    YAxis:    &chart.Axis{},
//	    Series: []chart.Series{
//	        {Name: "Europe", Points: []chart.Point{
//	            {X: 95, Y: 95, Z: 13.8, Name: "Belgium"},
//	            {X: 86, Y: 102, Z: 14.7, Name: "Germany"},
//	        }},
//	    },
//	    Legend: &chart.Legend{},
//	}
//	bc := &bubble.BubbleChart{Options: opts}
//	bc.Draw(doc, 50, 50, 400, 250)
package bubble

import (
	"math"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// BubbleChart renders a bubble chart onto a pdf.Document.
type BubbleChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *BubbleChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	bo := bubbleOptions(opts)
	layout := render.ComputeLayout(opts, x, y, width, height)

	render.DrawBackground(doc, opts, x, y, width, height)
	if err := render.DrawTitle(doc, opts, x, y, width); err != nil {
		return err
	}

	xDataMin, xDataMax := render.XDataRange(opts.Series)
	yDataMin, yDataMax := render.PointYRange(opts.Series)
	zDataMin, zDataMax := render.ZDataRange(opts.Series)

	xMin, xMax, xStep := render.NiceRange(xDataMin, xDataMax, opts.XAxis)
	yMin, yMax, yStep := render.NiceRange(yDataMin, yDataMax, opts.YAxis)

	// Z range for radius scaling.
	zMin, zMax := zDataMin, zDataMax
	if bo.ZMin != nil {
		zMin = *bo.ZMin
	}
	if bo.ZMax != nil {
		zMax = *bo.ZMax
	}
	if zMax == zMin {
		zMax = zMin + 1
	}

	minR := bo.MinSize
	if minR <= 0 {
		minR = 4
	}
	maxR := bo.MaxSize
	if maxR <= 0 {
		maxR = 30
	}

	if err := render.DrawYAxis(doc, opts, layout.Plot, layout.YAxis, yMin, yMax, yStep); err != nil {
		return err
	}
	if err := render.DrawXAxisNumeric(doc, opts, layout.Plot, layout.XAxis, xMin, xMax, xStep); err != nil {
		return err
	}

	for si, s := range opts.Series {
		color := chart.SeriesColor(opts, si)
		if s.Color != nil {
			color = *s.Color
		}
		fillColor := render.LightenColor(color, 0.6)

		for _, p := range s.Points {
			px := render.ValueToX(p.X, xMin, xMax, layout.Plot)
			py := render.ValueToY(p.Y, yMin, yMax, layout.Plot)

			t := (p.Z - zMin) / (zMax - zMin)
			r := minR + math.Sqrt(t)*(maxR-minR)

			doc.FillCircle(px, py, r, fillColor)
			doc.StrokeCircle(px, py, r, 1, color)

			if err := render.DrawDataLabel(doc, opts, bo.DataLabels, p.Y, px, py); err != nil {
				return err
			}
		}
	}

	return render.DrawLegend(doc, opts, layout.Legend)
}

func bubbleOptions(opts chart.Options) *chart.BubbleOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Bubble != nil {
		return opts.PlotOptions.Bubble
	}
	return &chart.BubbleOptions{}
}

var _ chart.Drawable = (*BubbleChart)(nil)
