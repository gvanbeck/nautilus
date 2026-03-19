// Package boxplot provides a box-and-whisker chart renderer for the Nautilus
// PDF library.
//
// Each data point is a Point with Low, Q1, Median, Q3, and High fields.
// The box spans Q1–Q3 with a line at the median.  Whiskers extend to Low/High.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Temperature Distribution"},
//	    XAxis:    &chart.Axis{Categories: []string{"Jan", "Feb", "Mar"}},
//	    YAxis:    &chart.Axis{},
//	    Series: []chart.Series{{
//	        Name: "Temp",
//	        Points: []chart.Point{
//	            {Low: -9, Q1: -3, Median: 2,  Q3: 8,  High: 14},
//	            {Low: -6, Q1: 0,  Median: 6,  Q3: 11, High: 18},
//	            {Low: -2, Q1: 5,  Median: 10, Q3: 15, High: 22},
//	        },
//	    }},
//	    Legend: &chart.Legend{},
//	}
//	bp := &boxplot.BoxplotChart{Options: opts}
//	bp.Draw(doc, 50, 50, 400, 250)
package boxplot

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// BoxplotChart renders a box-and-whisker chart onto a pdf.Document.
type BoxplotChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *BoxplotChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	bo := boxplotOptions(opts)
	layout := render.ComputeLayout(opts, x, y, width, height)

	render.DrawBackground(doc, opts, x, y, width, height)
	if err := render.DrawTitle(doc, opts, x, y, width); err != nil {
		return err
	}

	// Compute value range from point data.
	dataMin, dataMax := 0.0, 0.0
	first := true
	for _, s := range opts.Series {
		for _, p := range s.Points {
			vals := []float64{p.Low, p.Q1, p.Median, p.Q3, p.High}
			for _, v := range vals {
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

	lw := bo.LineWidth
	if lw <= 0 {
		lw = 1.5
	}
	whiskerFrac := bo.WhiskerLength
	if whiskerFrac <= 0 {
		whiskerFrac = 0.25
	}
	fillColor := pdf.ColorWhite
	if bo.FillColor != nil {
		fillColor = *bo.FillColor
	}

	nSeries := len(opts.Series)
	groupPad := 0.15
	pointPad := 0.05

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

		slotW := layout.Plot.W / float64(actualN)
		barSlot := slotW * (1 - 2*groupPad) / float64(nSeries)
		barW := barSlot * (1 - 2*pointPad)
		capW := barW * whiskerFrac

		for i, p := range s.Points {
			bx := layout.Plot.X + (float64(i)+groupPad)*slotW + float64(si)*barSlot + pointPad*barSlot
			cx := bx + barW/2

			pyLow := render.ValueToY(p.Low, yMin, yMax, layout.Plot)
			pyQ1 := render.ValueToY(p.Q1, yMin, yMax, layout.Plot)
			pyMed := render.ValueToY(p.Median, yMin, yMax, layout.Plot)
			pyQ3 := render.ValueToY(p.Q3, yMin, yMax, layout.Plot)
			pyHigh := render.ValueToY(p.High, yMin, yMax, layout.Plot)

			// Box (Q1 to Q3).
			boxH := pyQ1 - pyQ3
			doc.FillRect(bx, pyQ3, barW, boxH, fillColor)
			doc.DrawLine(bx, pyQ3, bx+barW, pyQ3, lw, color)
			doc.DrawLine(bx, pyQ1, bx+barW, pyQ1, lw, color)
			doc.DrawLine(bx, pyQ3, bx, pyQ1, lw, color)
			doc.DrawLine(bx+barW, pyQ3, bx+barW, pyQ1, lw, color)

			// Median line.
			doc.DrawLine(bx, pyMed, bx+barW, pyMed, lw*1.5, color)

			// Upper whisker (Q3 → High).
			doc.DrawLine(cx, pyQ3, cx, pyHigh, lw, color)
			doc.DrawLine(cx-capW, pyHigh, cx+capW, pyHigh, lw, color)

			// Lower whisker (Q1 → Low).
			doc.DrawLine(cx, pyQ1, cx, pyLow, lw, color)
			doc.DrawLine(cx-capW, pyLow, cx+capW, pyLow, lw, color)
		}
	}

	return render.DrawLegend(doc, opts, layout.Legend)
}

func boxplotOptions(opts chart.Options) *chart.BoxplotOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Boxplot != nil {
		return opts.PlotOptions.Boxplot
	}
	return &chart.BoxplotOptions{}
}

var _ chart.Drawable = (*BoxplotChart)(nil)
