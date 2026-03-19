// Package columnrange provides a column-range chart renderer for the Nautilus
// PDF library.
//
// Each bar spans a low–high range instead of starting from zero.
// Data is stored in Series.Points with Low and High fields.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Temperature Range"},
//	    XAxis:    &chart.Axis{Categories: []string{"Jan", "Feb", "Mar"}},
//	    YAxis:    &chart.Axis{},
//	    Series: []chart.Series{{
//	        Name: "Temp °C",
//	        Points: []chart.Point{
//	            {Low: -9.5, High: 8.0},
//	            {Low: -7.8, High: 8.3},
//	            {Low: -13.1, High: 9.2},
//	        },
//	    }},
//	    Legend: &chart.Legend{},
//	}
//	cr := &columnrange.ColumnRangeChart{Options: opts}
//	cr.Draw(doc, 50, 50, 400, 250)
package columnrange

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// ColumnRangeChart renders a column range chart onto a pdf.Document.
type ColumnRangeChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *ColumnRangeChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	co := columnRangeOptions(opts)
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

	groupPad := co.GroupPadding
	if groupPad == 0 {
		groupPad = 0.2
	}
	pointPad := co.PointPadding
	if pointPad == 0 {
		pointPad = 0.1
	}
	bw := co.BorderWidth
	borderColor := pdf.ColorWhite
	if co.BorderColor != nil {
		borderColor = *co.BorderColor
	}

	nSeries := len(opts.Series)
	actualN := n
	if actualN == 0 {
		for _, s := range opts.Series {
			if len(s.Points) > actualN {
				actualN = len(s.Points)
			}
		}
	}

	slotW := layout.Plot.W / float64(actualN)
	barSlot := slotW * (1 - 2*groupPad) / float64(nSeries)
	barW := barSlot * (1 - 2*pointPad)

	for si, s := range opts.Series {
		color := chart.SeriesColor(opts, si)
		if s.Color != nil {
			color = *s.Color
		}
		for i, p := range s.Points {
			bx := layout.Plot.X + (float64(i)+groupPad)*slotW + float64(si)*barSlot + pointPad*barSlot
			by := render.ValueToY(p.High, yMin, yMax, layout.Plot)
			bh := render.ValueToY(p.Low, yMin, yMax, layout.Plot) - by

			doc.FillRect(bx, by, barW, bh, color)
			if bw > 0 {
				doc.DrawLine(bx, by, bx+barW, by, bw, borderColor)
				doc.DrawLine(bx, by+bh, bx+barW, by+bh, bw, borderColor)
				doc.DrawLine(bx, by, bx, by+bh, bw, borderColor)
				doc.DrawLine(bx+barW, by, bx+barW, by+bh, bw, borderColor)
			}

			// Data labels.
			if co.DataLabels != nil && render.BoolVal(co.DataLabels.Enabled, false) {
				if err := render.DrawDataLabel(doc, opts, co.DataLabels, p.High, bx+barW/2, by); err != nil {
					return err
				}
			}
		}
	}

	return render.DrawLegend(doc, opts, layout.Legend)
}

func columnRangeOptions(opts chart.Options) *chart.ColumnRangeOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.ColumnRange != nil {
		return opts.PlotOptions.ColumnRange
	}
	return &chart.ColumnRangeOptions{}
}

var _ chart.Drawable = (*ColumnRangeChart)(nil)
