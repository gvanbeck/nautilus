// Package waterfall provides a waterfall chart renderer for the Nautilus PDF library.
//
// A waterfall chart shows how an initial value is affected by a series of
// positive or negative increments.  Each step is a Point with a Name and Y
// value.  Set IsSum=true for a total bar, or IsIntermediateSum=true for a
// running subtotal.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Company Financials"},
//	    YAxis:    &chart.Axis{},
//	    Series: []chart.Series{{
//	        Name: "Balance",
//	        Points: []chart.Point{
//	            {Name: "Start",     Y: 120000},
//	            {Name: "Revenue",   Y: 569000},
//	            {Name: "Costs",     Y: -342000},
//	            {Name: "Subtotal",  IsIntermediateSum: true},
//	            {Name: "More costs",Y: -233000},
//	            {Name: "Balance",   IsSum: true},
//	        },
//	    }},
//	}
//	wc := &waterfall.WaterfallChart{Options: opts}
//	wc.Draw(doc, 50, 50, 400, 250)
package waterfall

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// WaterfallChart renders a waterfall (running total) chart onto a pdf.Document.
type WaterfallChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *WaterfallChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	wo := waterfallOptions(opts)

	render.DrawBackground(doc, opts, x, y, width, height)
	if err := render.DrawTitle(doc, opts, x, y, width); err != nil {
		return err
	}

	if len(opts.Series) == 0 || len(opts.Series[0].Points) == 0 {
		return nil
	}

	points := opts.Series[0].Points
	n := len(points)

	// Compute running totals and value range.
	type bar struct {
		name   string
		bottom float64 // lower edge (may be negative)
		top    float64 // upper edge
		color  pdf.Color
		isSum  bool
	}

	seriesColor := chart.SeriesColor(opts, 0)
	if opts.Series[0].Color != nil {
		seriesColor = *opts.Series[0].Color
	}
	upColor := seriesColor
	if wo.UpColor != nil {
		upColor = *wo.UpColor
	}
	negColor := pdf.Color{R: 223, G: 83, B: 83}
	if wo.NegativeColor != nil {
		negColor = *wo.NegativeColor
	}
	sumColor := pdf.Color{R: 124, G: 181, B: 236}

	running := 0.0
	bars := make([]bar, n)
	yMin, yMax := 0.0, 0.0

	for i, p := range points {
		var b bar
		b.name = p.Name
		if p.IsSum {
			b.isSum = true
			b.bottom = 0
			b.top = running
			b.color = sumColor
			if p.Color != nil {
				b.color = *p.Color
			}
		} else if p.IsIntermediateSum {
			b.isSum = true
			b.bottom = 0
			b.top = running
			b.color = sumColor
			if p.Color != nil {
				b.color = *p.Color
			}
		} else {
			if p.Y >= 0 {
				b.bottom = running
				b.top = running + p.Y
				b.color = upColor
			} else {
				b.bottom = running + p.Y
				b.top = running
				b.color = negColor
			}
			if p.Color != nil {
				b.color = *p.Color
			}
			running += p.Y
		}
		bars[i] = b
		if b.bottom < yMin {
			yMin = b.bottom
		}
		if b.top > yMax {
			yMax = b.top
		}
	}

	yMin, yMax, yStep := render.NiceRange(yMin, yMax, opts.YAxis)

	// Compute layout.
	fs := render.EffectiveFontSize(opts)
	top := y
	if opts.Title != nil && opts.Title.Text != "" {
		tfs := fs * 1.5
		if opts.Title.FontSize > 0 {
			tfs = opts.Title.FontSize
		}
		top += tfs + 6
	}
	if opts.Subtitle != nil && opts.Subtitle.Text != "" {
		sfs := fs * 1.1
		if opts.Subtitle.FontSize > 0 {
			sfs = opts.Subtitle.FontSize
		}
		top += sfs + 6
	}

	legendH := 0.0
	if (opts.Legend == nil || render.BoolVal(opts.Legend.Enabled, true)) && len(opts.Series) > 0 {
		legendH = fs + 12
	}
	legendArea := render.Area{X: x, Y: y + height - legendH, W: width, H: legendH}

	yAxisW := fs*0.6*7 + 6
	xAxisH := fs*1.2 + 4

	plot := render.Area{
		X: x + yAxisW,
		Y: top,
		W: width - yAxisW,
		H: (y + height - legendH - xAxisH) - top,
	}
	yAxisArea := render.Area{X: x, Y: top, W: yAxisW, H: plot.H}
	xLabelY := plot.Bottom() + 4
	_ = xAxisH

	if err := render.DrawYAxis(doc, opts, plot, yAxisArea, yMin, yMax, yStep); err != nil {
		return err
	}
	// X baseline.
	doc.DrawLine(plot.X, plot.Bottom(), plot.Right(), plot.Bottom(), 0.5, render.DefaultAxisColor)

	// Draw bars and connectors.
	groupPad := 0.15
	slotW := plot.W / float64(n)
	barW := slotW * (1 - 2*groupPad)

	for i, b := range bars {
		bx := plot.X + (float64(i)+groupPad)*slotW
		by := render.ValueToY(b.top, yMin, yMax, plot)
		bh := render.ValueToY(b.bottom, yMin, yMax, plot) - by

		doc.FillRect(bx, by, barW, bh, b.color)
		if wo.LineWidth > 0 {
			doc.DrawLine(bx, by, bx+barW, by, wo.LineWidth, render.DefaultAxisColor)
			doc.DrawLine(bx, by+bh, bx+barW, by+bh, wo.LineWidth, render.DefaultAxisColor)
			doc.DrawLine(bx, by, bx, by+bh, wo.LineWidth, render.DefaultAxisColor)
			doc.DrawLine(bx+barW, by, bx+barW, by+bh, wo.LineWidth, render.DefaultAxisColor)
		}

		// Connector line to next bar (non-sum bars).
		if i < n-1 && !b.isSum {
			nextTop := render.ValueToY(bars[i+1].bottom, yMin, yMax, plot)
			if bars[i+1].isSum {
				nextTop = render.ValueToY(b.top, yMin, yMax, plot)
			}
			doc.DrawLine(bx+barW, render.ValueToY(b.top, yMin, yMax, plot),
				plot.X+float64(i+1)*slotW, nextTop,
				0.5, render.DefaultGridColor)
		}

		// Data label above/below bar.
		if wo.DataLabels != nil && render.BoolVal(wo.DataLabels.Enabled, false) {
			v := b.top - b.bottom
			if err := render.DrawDataLabel(doc, opts, wo.DataLabels, v, bx+barW/2, by); err != nil {
				return err
			}
		}
	}

	// X-axis category labels.
	if opts.FontName != "" {
		if err := doc.SetFont(opts.FontName, fs); err != nil {
			return err
		}
		doc.SetTextColor(100, 100, 100)
		for i, b := range bars {
			cx := plot.X + (float64(i)+0.5)*slotW
			lw, _ := doc.MeasureText(b.name)
			doc.WriteLine(b.name, cx-lw/2, xLabelY) //nolint:errcheck
		}
	}

	return render.DrawLegend(doc, opts, legendArea)
}

func waterfallOptions(opts chart.Options) *chart.WaterfallOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Waterfall != nil {
		return opts.PlotOptions.Waterfall
	}
	return &chart.WaterfallOptions{}
}

var _ chart.Drawable = (*WaterfallChart)(nil)
