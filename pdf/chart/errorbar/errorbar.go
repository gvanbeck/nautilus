// Package errorbar provides an error bar chart renderer for the Nautilus PDF library.
//
// Error bars show uncertainty around data points.  Each Point has Low and High
// values; an I-beam is drawn between them.  Use XAxis.Categories to label each
// position on the x-axis.
//
// Error bars are typically overlaid on a column or line series. This package
// can also render them standalone.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Measurement Error"},
//	    XAxis:    &chart.Axis{Categories: []string{"Jan", "Feb", "Mar"}},
//	    YAxis:    &chart.Axis{},
//	    Series: []chart.Series{{
//	        Name: "Error",
//	        Points: []chart.Point{
//	            {Low: 48, High: 51},
//	            {Low: 68, High: 73},
//	            {Low: 92, High: 110},
//	        },
//	    }},
//	}
//	ec := &errorbar.ErrorbarChart{Options: opts}
//	ec.Draw(doc, 50, 50, 400, 200)
package errorbar

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// ErrorbarChart renders error bars onto a pdf.Document.
type ErrorbarChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *ErrorbarChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	eo := errorbarOptions(opts)
	layout := render.ComputeLayout(opts, x, y, width, height)

	render.DrawBackground(doc, opts, x, y, width, height)
	if err := render.DrawTitle(doc, opts, x, y, width); err != nil {
		return err
	}

	// Collect all low/high for y-range.
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

	lw := eo.LineWidth
	if lw <= 0 {
		lw = 1.5
	}
	whisker := eo.WhiskerLength
	if whisker <= 0 {
		whisker = 0.25
	}

	bucketW := 0.0
	if n > 0 {
		bucketW = layout.Plot.W / float64(n)
	}
	capW := bucketW * whisker

	for si, s := range opts.Series {
		color := chart.SeriesColor(opts, si)
		if eo.Color != nil {
			color = *eo.Color
		}
		if s.Color != nil {
			color = *s.Color
		}
		nPts := len(s.Points)
		for i, p := range s.Points {
			cx := render.CategoryCenterX(i, nPts, layout.Plot)
			pyLow := render.ValueToY(p.Low, yMin, yMax, layout.Plot)
			pyHigh := render.ValueToY(p.High, yMin, yMax, layout.Plot)

			// Vertical stem.
			doc.DrawLine(cx, pyHigh, cx, pyLow, lw, color)
			// Top cap.
			doc.DrawLine(cx-capW/2, pyHigh, cx+capW/2, pyHigh, lw, color)
			// Bottom cap.
			doc.DrawLine(cx-capW/2, pyLow, cx+capW/2, pyLow, lw, color)
		}
	}

	return render.DrawLegend(doc, opts, layout.Legend)
}

func errorbarOptions(opts chart.Options) *chart.ErrorbarOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Errorbar != nil {
		return opts.PlotOptions.Errorbar
	}
	return &chart.ErrorbarOptions{}
}

var _ chart.Drawable = (*ErrorbarChart)(nil)
