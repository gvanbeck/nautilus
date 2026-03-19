// Package funnel provides funnel and pyramid chart renderers for the Nautilus
// PDF library.
//
// A funnel chart displays a series of stages narrowing downward.  Set
// PlotOptions.Funnel.Reversed = true to render a pyramid (wide at top).
// Stage data is stored in Series.Points with Name and Y fields.
//
// Example — funnel:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Sales Funnel"},
//	    Series: []chart.Series{{
//	        Points: []chart.Point{
//	            {Name: "Website visits",        Y: 15654},
//	            {Name: "Downloads",             Y: 4064},
//	            {Name: "Price list requested",  Y: 1987},
//	            {Name: "Invoice sent",          Y: 976},
//	            {Name: "Finalized",             Y: 846},
//	        },
//	    }},
//	    Legend: &chart.Legend{},
//	}
//	fc := &funnel.FunnelChart{Options: opts}
//	fc.Draw(doc, 50, 50, 300, 280)
package funnel

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// FunnelChart renders a funnel or pyramid chart onto a pdf.Document.
type FunnelChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *FunnelChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	fo := funnelOptions(opts)
	fs := render.EffectiveFontSize(opts)

	render.DrawBackground(doc, opts, x, y, width, height)
	if err := render.DrawTitle(doc, opts, x, y, width); err != nil {
		return err
	}

	if len(opts.Series) == 0 {
		return nil
	}
	points := opts.Series[0].Points

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

	areaH := (y + height - legendH) - top
	areaW := width

	// Parse geometry parameters.
	funnelW := render.ParsePercent(fo.Width)
	if funnelW == 0 {
		funnelW = 0.8
	}
	neckW := render.ParsePercent(fo.NeckWidth)
	if neckW == 0 {
		neckW = 0.3
	}
	neckH := render.ParsePercent(fo.NeckHeight)
	if neckH == 0 {
		neckH = 0.25
	}

	maxHalfW := areaW * funnelW / 2
	neckHalfW := areaW * neckW / 2
	neckStartY := top + areaH*(1-neckH)
	cx := x + areaW/2

	n := len(points)
	if n == 0 {
		return render.DrawLegend(doc, opts, legendArea)
	}

	// Compute total for proportional heights.
	total := 0.0
	for _, p := range points {
		total += p.Y
	}
	if total == 0 {
		return render.DrawLegend(doc, opts, legendArea)
	}

	// Draw trapezoids from top to bottom.
	curY := top
	for i, p := range points {
		sliceFrac := p.Y / total
		sliceH := areaH * sliceFrac

		// Half-width at the top and bottom of this slice.
		tFracTop := (curY - top) / areaH
		tFracBot := (curY + sliceH - top) / areaH

		halfWTop := halfWidthAt(tFracTop, neckStartY, top, areaH, maxHalfW, neckHalfW)
		halfWBot := halfWidthAt(tFracBot, neckStartY, top, areaH, maxHalfW, neckHalfW)

		if fo.Reversed {
			// For pyramid, invert so widest at top.
			halfWTop, halfWBot = maxHalfW-halfWTop+neckHalfW, maxHalfW-halfWBot+neckHalfW
		}

		color := chart.SeriesColor(opts, i)
		if p.Color != nil {
			color = *p.Color
		}

		pts := []pdf.Point{
			{X: cx - halfWTop, Y: curY},
			{X: cx + halfWTop, Y: curY},
			{X: cx + halfWBot, Y: curY + sliceH},
			{X: cx - halfWBot, Y: curY + sliceH},
		}
		doc.FillAndStrokePolygon(pts, color, 0.5, pdf.ColorWhite)

		// Data label.
		midY := curY + sliceH/2
		if fo.DataLabels != nil && render.BoolVal(fo.DataLabels.Enabled, false) && opts.FontName != "" {
			lfs := fs
			fn := opts.FontName
			if fo.DataLabels.FontSize > 0 {
				lfs = fo.DataLabels.FontSize
			}
			if fo.DataLabels.FontName != "" {
				fn = fo.DataLabels.FontName
			}
			if err := doc.SetFont(fn, lfs); err != nil {
				return err
			}
			c := pdf.ColorBlack
			if fo.DataLabels.Color != nil {
				c = *fo.DataLabels.Color
			}
			doc.SetTextColor(c.R, c.G, c.B)
			label := p.Name + ": " + render.FormatFloat(p.Y)
			lw, _ := doc.MeasureText(label)
			doc.WriteLine(label, cx-lw/2, midY-lfs/2) //nolint:errcheck
		}

		curY += sliceH
	}

	return render.DrawLegend(doc, opts, legendArea)
}

// halfWidthAt returns the half-width of the funnel at a vertical fraction t
// of the total funnel height.
func halfWidthAt(tFrac float64, neckStartY, top, areaH, maxHalfW, neckHalfW float64) float64 {
	neckFrac := (neckStartY - top) / areaH
	if tFrac >= neckFrac {
		return neckHalfW
	}
	// Linear interpolation from maxHalfW at top to neckHalfW at neckFrac.
	if neckFrac == 0 {
		return neckHalfW
	}
	t := tFrac / neckFrac
	return maxHalfW*(1-t) + neckHalfW*t
}

func funnelOptions(opts chart.Options) *chart.FunnelOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Funnel != nil {
		return opts.PlotOptions.Funnel
	}
	return &chart.FunnelOptions{}
}

var _ chart.Drawable = (*FunnelChart)(nil)
