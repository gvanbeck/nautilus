// Package column provides a vertical bar (column) chart renderer for the
// Nautilus PDF library.
//
// Multiple series are rendered as grouped bars by default.  Set
// PlotOptions.Column.Stacking to "normal" for stacked bars or "percent" for
// 100 % stacked bars.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Quarterly Sales"},
//	    XAxis:    &chart.Axis{Categories: []string{"Q1", "Q2", "Q3", "Q4"}},
//	    YAxis:    &chart.Axis{},
//	    Series: []chart.Series{
//	        {Name: "Product A", Data: []float64{43, 55, 57, 60}},
//	        {Name: "Product B", Data: []float64{23, 35, 41, 47}},
//	    },
//	}
//	cc := &column.ColumnChart{Options: opts}
//	cc.Draw(doc, 50, 100, 400, 200)
package column

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// ColumnChart renders a grouped or stacked column chart onto a pdf.Document.
type ColumnChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *ColumnChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	layout := render.ComputeLayout(opts, x, y, width, height)

	render.DrawBackground(doc, opts, x, y, width, height)
	if err := render.DrawTitle(doc, opts, x, y, width); err != nil {
		return err
	}

	co := colOptions(opts)
	stacking := co.Stacking
	categories := render.CategoriesFor(opts)
	n := len(categories)

	// Compute y range (stacked charts use summed values).
	var yMin, yMax, yStep float64
	if stacking == "percent" {
		yMin, yMax, yStep = 0, 100, 20
	} else if stacking == "normal" {
		dataMin, dataMax := stackedRange(opts.Series, n)
		yMin, yMax, yStep = render.NiceRange(dataMin, dataMax, opts.YAxis)
	} else {
		dataMin, dataMax := render.DataRange(opts.Series)
		yMin, yMax, yStep = render.NiceRange(dataMin, dataMax, opts.YAxis)
	}

	if err := render.DrawYAxis(doc, opts, layout.Plot, layout.YAxis, yMin, yMax, yStep); err != nil {
		return err
	}
	if err := render.DrawXAxis(doc, opts, layout.Plot, layout.XAxis, categories); err != nil {
		return err
	}

	zeroY := render.ValueToY(0, yMin, yMax, layout.Plot)
	ns := len(opts.Series)
	if ns == 0 {
		return nil
	}

	groupPad := co.GroupPadding
	if groupPad <= 0 {
		groupPad = 0.2
	}
	pointPad := co.PointPadding
	if pointPad <= 0 {
		pointPad = 0.1
	}

	slotW := layout.Plot.W / float64(n)

	switch stacking {
	case "normal", "percent":
		drawStackedColumns(doc, opts, layout.Plot, categories, yMin, yMax, zeroY, slotW, groupPad, co, stacking)
	default:
		drawGroupedColumns(doc, opts, layout.Plot, categories, yMin, yMax, zeroY, slotW, groupPad, pointPad, co)
	}

	return render.DrawLegend(doc, opts, layout.Legend)
}

func drawGroupedColumns(doc *pdf.Document, opts chart.Options, plot render.Area,
	categories []string, yMin, yMax, zeroY, slotW, groupPad, pointPad float64, co *chart.ColumnOptions) {

	n := len(categories)
	ns := len(opts.Series)

	innerW := slotW * (1 - groupPad*2)
	barW := innerW / float64(ns) * (1 - pointPad*2)
	if barW < 1 {
		barW = 1
	}

	for si, s := range opts.Series {
		color := chart.SeriesColor(opts, si)
		if s.Color != nil {
			color = *s.Color
		}
		for i, v := range s.Data {
			if i >= n {
				break
			}
			slotLeft := plot.X + float64(i)*slotW
			groupLeft := slotLeft + slotW*groupPad
			barLeft := groupLeft + float64(si)*innerW/float64(ns) + innerW/float64(ns)*pointPad

			topY := render.ValueToY(v, yMin, yMax, plot)
			barH := zeroY - topY
			if barH < 0 {
				barH = -barH
				topY = zeroY
			}

			drawBar(doc, opts, co, barLeft, topY, barW, barH, v, color)
		}
	}
}

func drawStackedColumns(doc *pdf.Document, opts chart.Options, plot render.Area,
	categories []string, yMin, yMax, zeroY, slotW, groupPad float64, co *chart.ColumnOptions, stacking string) {

	n := len(categories)
	ns := len(opts.Series)

	innerW := slotW * (1 - groupPad*2)
	if innerW < 1 {
		innerW = 1
	}

	// Pre-compute column totals for percent stacking.
	totals := make([]float64, n)
	if stacking == "percent" {
		for _, s := range opts.Series {
			for i, v := range s.Data {
				if i < n {
					totals[i] += v
				}
			}
		}
	}

	for i := 0; i < n; i++ {
		cumSum := 0.0
		for si := 0; si < ns; si++ {
			s := opts.Series[si]
			if i >= len(s.Data) {
				continue
			}
			v := s.Data[i]
			plotV := v
			if stacking == "percent" && totals[i] != 0 {
				plotV = v / totals[i] * 100
			}

			color := chart.SeriesColor(opts, si)
			if s.Color != nil {
				color = *s.Color
			}

			barLeft := plot.X + float64(i)*slotW + slotW*groupPad
			topY := render.ValueToY(cumSum+plotV, yMin, yMax, plot)
			bottomY := render.ValueToY(cumSum, yMin, yMax, plot)
			barH := bottomY - topY
			if barH < 0 {
				barH = 0
			}

			drawBar(doc, opts, co, barLeft, topY, innerW, barH, v, color)
			cumSum += plotV
		}
	}
}

func drawBar(doc *pdf.Document, opts chart.Options, co *chart.ColumnOptions,
	bx, by, bw, bh, value float64, color pdf.Color) {

	doc.FillRect(bx, by, bw, bh, color)

	if co.BorderWidth > 0 {
		bc := pdf.ColorWhite
		if co.BorderColor != nil {
			bc = *co.BorderColor
		}
		spec := pdf.BorderSpec{Thickness: co.BorderWidth, Color: bc}
		doc.DrawBorder(bx, by, bw, bh, pdf.NewUniformBorder(spec)) //nolint:errcheck
	}

	if co.DataLabels != nil && render.BoolVal(co.DataLabels.Enabled, false) {
		render.DrawDataLabel(doc, opts, co.DataLabels, value, bx+bw/2, by) //nolint:errcheck
	}
}

func colOptions(opts chart.Options) *chart.ColumnOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Column != nil {
		return opts.PlotOptions.Column
	}
	return &chart.ColumnOptions{}
}

// stackedRange returns the min/max summed value across all categories.
func stackedRange(series []chart.Series, n int) (min, max float64) {
	for i := 0; i < n; i++ {
		sum := 0.0
		for _, s := range series {
			if i < len(s.Data) {
				sum += s.Data[i]
			}
		}
		if i == 0 || sum > max {
			max = sum
		}
		if i == 0 || sum < min {
			min = sum
		}
	}
	return
}
