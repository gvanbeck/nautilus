// Package bar provides a horizontal bar chart renderer for the Nautilus PDF
// library.  Categories appear on the Y axis; values extend horizontally.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Top Products"},
//	    // XAxis holds value-axis config (numeric, rendered at the bottom).
//	    XAxis: &chart.Axis{},
//	    // YAxis holds category-axis config (rendered on the left).
//	    YAxis: &chart.Axis{
//	        Categories: []string{"Product A", "Product B", "Product C"},
//	    },
//	    Series: []chart.Series{
//	        {Name: "2024", Data: []float64{43, 71, 55}},
//	    },
//	}
//	bc := &bar.BarChart{Options: opts}
//	bc.Draw(doc, 50, 100, 400, 200)
package bar

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// BarChart renders a horizontal bar chart onto a pdf.Document.
// For bar charts, YAxis.Categories lists the row labels and the XAxis
// represents the numeric value range.
type BarChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *BarChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	bo := barOptions(opts)

	render.DrawBackground(doc, opts, x, y, width, height)
	if err := render.DrawTitle(doc, opts, x, y, width); err != nil {
		return err
	}

	fs := render.EffectiveFontSize(opts)

	// Reserve title space.
	top := y
	if opts.Title != nil && opts.Title.Text != "" {
		top += fs*1.5 + 6
	}
	if opts.Subtitle != nil && opts.Subtitle.Text != "" {
		top += fs*1.1 + 6
	}

	// Reserve legend space at bottom.
	legendH := 0.0
	if (opts.Legend == nil || render.BoolVal(opts.Legend.Enabled, true)) && len(opts.Series) > 0 {
		legendH = fs + 12
	}

	// Reserve value-axis labels at bottom.
	xAxisH := fs*1.2 + 4

	// Reserve category labels on the left.
	categories := barCategories(opts)
	n := len(categories)
	yAxisW := estimateLabelWidth(categories, fs) + 6

	plotX := x + yAxisW
	plotY := top
	plotW := width - yAxisW
	plotH := (y + height - legendH - xAxisH) - plotY

	plot := render.Area{X: plotX, Y: plotY, W: plotW, H: plotH}
	xAxisArea := render.Area{X: plotX, Y: plotY + plotH, W: plotW, H: xAxisH}
	yAxisArea := render.Area{X: x, Y: plotY, W: yAxisW, H: plotH}
	legendArea := render.Area{X: x, Y: y + height - legendH, W: width, H: legendH}

	// Compute x (value) range.
	dataMin, dataMax := render.DataRange(opts.Series)
	xMin, xMax, xStep := render.NiceRange(dataMin, dataMax, opts.XAxis)

	// Draw vertical grid lines and value labels.
	if err := drawValueAxis(doc, opts, plot, xAxisArea, xMin, xMax, xStep, fs); err != nil {
		return err
	}

	// Draw category labels on left.
	if err := drawCategoryAxis(doc, opts, plot, yAxisArea, categories, fs); err != nil {
		return err
	}

	// Draw bars.
	ns := len(opts.Series)
	if ns == 0 || n == 0 {
		return render.DrawLegend(doc, opts, legendArea)
	}

	groupPad := bo.GroupPadding
	if groupPad <= 0 {
		groupPad = 0.2
	}
	pointPad := bo.PointPadding
	if pointPad <= 0 {
		pointPad = 0.1
	}

	slotH := plot.H / float64(n)
	innerH := slotH * (1 - groupPad*2)
	barH := innerH / float64(ns) * (1 - pointPad*2)
	if barH < 1 {
		barH = 1
	}

	zeroX := valueToX(0, xMin, xMax, plot)

	for si, s := range opts.Series {
		color := chart.SeriesColor(opts, si)
		if s.Color != nil {
			color = *s.Color
		}
		for i, v := range s.Data {
			if i >= n {
				break
			}
			slotTop := plot.Y + float64(i)*slotH
			groupTop := slotTop + slotH*groupPad
			barTop := groupTop + float64(si)*innerH/float64(ns) + innerH/float64(ns)*pointPad

			rightX := valueToX(v, xMin, xMax, plot)
			barW := rightX - zeroX
			barStartX := zeroX
			if barW < 0 {
				barW = -barW
				barStartX = rightX
			}

			doc.FillRect(barStartX, barTop, barW, barH, color)

			if bo.BorderWidth > 0 {
				bc := pdf.ColorWhite
				if bo.BorderColor != nil {
					bc = *bo.BorderColor
				}
				spec := pdf.BorderSpec{Thickness: bo.BorderWidth, Color: bc}
				doc.DrawBorder(zeroX, barTop, barW, barH, pdf.NewUniformBorder(spec)) //nolint:errcheck
			}

			if bo.DataLabels != nil && render.BoolVal(bo.DataLabels.Enabled, false) {
				render.DrawDataLabel(doc, opts, bo.DataLabels, v, rightX+4, barTop+barH/2) //nolint:errcheck
			}
		}
	}

	return render.DrawLegend(doc, opts, legendArea)
}

// valueToX maps a numeric value to an X coordinate within the plot area.
func valueToX(v, xMin, xMax float64, plot render.Area) float64 {
	if xMax == xMin {
		return plot.X + plot.W/2
	}
	frac := (v - xMin) / (xMax - xMin)
	return plot.X + frac*plot.W
}

// drawValueAxis draws vertical gridlines and value labels along the bottom.
func drawValueAxis(doc *pdf.Document, opts chart.Options, plot, xAxisArea render.Area, xMin, xMax, step, fs float64) error {
	gridW := 0.5
	gridColor := render.DefaultGridColor
	if opts.XAxis != nil {
		if opts.XAxis.GridLineWidth > 0 {
			gridW = opts.XAxis.GridLineWidth
		} else if opts.XAxis.GridLineWidth < 0 {
			gridW = 0
		}
		if opts.XAxis.GridLineColor != nil {
			gridColor = *opts.XAxis.GridLineColor
		}
	}

	labelFont := opts.FontName
	labelSize := fs
	if opts.XAxis != nil && opts.XAxis.Labels != nil {
		if opts.XAxis.Labels.FontName != "" {
			labelFont = opts.XAxis.Labels.FontName
		}
		if opts.XAxis.Labels.FontSize > 0 {
			labelSize = opts.XAxis.Labels.FontSize
		}
	}

	if labelFont != "" {
		if err := doc.SetFont(labelFont, labelSize); err != nil {
			return err
		}
	}

	// Bottom baseline.
	doc.DrawLine(plot.X, plot.Bottom(), plot.Right(), plot.Bottom(), 0.5, render.DefaultAxisColor)

	for v := xMin; v <= xMax+step*0.001; v += step {
		px := valueToX(v, xMin, xMax, plot)
		if gridW > 0 {
			doc.DrawLine(px, plot.Y, px, plot.Bottom(), gridW, gridColor)
		}
		if labelFont != "" {
			label := render.FormatAxisValue(v, opts.XAxis)
			doc.SetTextColor(100, 100, 100)
			lw, _ := doc.MeasureText(label)
			if _, err := doc.WriteLine(label, px-lw/2, xAxisArea.Y+4); err != nil {
				return err
			}
		}
	}

	// Left baseline.
	doc.DrawLine(plot.X, plot.Y, plot.X, plot.Bottom(), 0.5, render.DefaultAxisColor)
	return nil
}

// drawCategoryAxis draws category labels on the left of the plot.
func drawCategoryAxis(doc *pdf.Document, opts chart.Options, plot, yAxisArea render.Area, categories []string, fs float64) error {
	labelFont := opts.FontName
	labelSize := fs
	if opts.YAxis != nil && opts.YAxis.Labels != nil {
		if opts.YAxis.Labels.FontName != "" {
			labelFont = opts.YAxis.Labels.FontName
		}
		if opts.YAxis.Labels.FontSize > 0 {
			labelSize = opts.YAxis.Labels.FontSize
		}
	}
	if labelFont == "" {
		return nil
	}
	if err := doc.SetFont(labelFont, labelSize); err != nil {
		return err
	}
	doc.SetTextColor(100, 100, 100)

	n := len(categories)
	slotH := plot.H / float64(n)
	for i, cat := range categories {
		cy := plot.Y + (float64(i)+0.5)*slotH
		lw, _ := doc.MeasureText(cat)
		lx := yAxisArea.X + yAxisArea.W - 6 - lw
		if _, err := doc.WriteLine(cat, lx, cy-labelSize/2); err != nil {
			return err
		}
	}
	return nil
}

// barCategories returns the categories from YAxis (bar charts list them there).
func barCategories(opts chart.Options) []string {
	if opts.YAxis != nil && len(opts.YAxis.Categories) > 0 {
		return opts.YAxis.Categories
	}
	if len(opts.Series) > 0 {
		return render.AutoCategories(len(opts.Series[0].Data))
	}
	return nil
}

// estimateLabelWidth estimates the pixel width of the longest label.
func estimateLabelWidth(labels []string, fs float64) float64 {
	maxLen := 0
	for _, l := range labels {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}
	return float64(maxLen) * fs * 0.55
}

func barOptions(opts chart.Options) *chart.BarOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Bar != nil {
		return opts.PlotOptions.Bar
	}
	return &chart.BarOptions{}
}
