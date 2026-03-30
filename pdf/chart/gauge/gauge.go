// Package gauge provides gauge and solid-gauge chart renderers for the Nautilus
// PDF library.
//
// A gauge chart displays a single value on a circular arc scale with a needle.
// Set PlotOptions.Gauge.Solid = true to draw a solid-gauge (filled arc) instead.
//
// Example — gauge with color bands:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Speed"},
//	    YAxis:    &chart.Axis{Min: chart.Float(0), Max: chart.Float(200)},
//	    Series: []chart.Series{{Name: "Speed", Data: []float64{120}}},
//	    PlotOptions: &chart.PlotOptions{Gauge: &chart.GaugeOptions{
//	        PaneStartAngle: -150, PaneEndAngle: 150,
//	        PlotBands: []chart.GaugePlotBand{
//	            {From: 0,   To: 120, Color: pdf.Color{R: 85, G: 191, B: 59}},
//	            {From: 120, To: 160, Color: pdf.Color{R: 221, G: 223, B: 13}},
//	            {From: 160, To: 200, Color: pdf.Color{R: 223, G: 83, B: 83}},
//	        },
//	    }},
//	}
//	gc := &gauge.GaugeChart{Options: opts}
//	gc.Draw(doc, 50, 50, 250, 200)
package gauge

import (
	"math"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// GaugeChart renders a gauge or solid-gauge chart onto a pdf.Document.
type GaugeChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *GaugeChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	go_ := gaugeOptions(opts)
	fs := render.EffectiveFontSize(opts)

	render.DrawBackground(doc, opts, x, y, width, height)
	if err := render.DrawTitle(doc, opts, x, y, width); err != nil {
		return err
	}

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

	// Gauge geometry.
	startDeg := -150.0
	if go_.PaneStartAngle != nil {
		startDeg = *go_.PaneStartAngle
	}
	endDeg := 150.0
	if go_.PaneEndAngle != nil {
		endDeg = *go_.PaneEndAngle
	}

	// Angles in our system: 0=east, clockwise positive.
	// Highcharts: 0=top, clockwise. Convert: ourAngle = hcAngle - 90.
	startRad := (startDeg - 90) * math.Pi / 180
	endRad := (endDeg - 90) * math.Pi / 180

	outerR := math.Min(areaW, areaH) / 2 * 0.85
	trackThick := outerR * 0.15
	innerR := outerR - trackThick

	cx := x + areaW/2
	cy := top + areaH/2

	// Data range.
	_, dataMax := render.DataRange(opts.Series)
	yMin, yMax, _ := render.NiceRange(0, dataMax, opts.YAxis)
	if opts.YAxis != nil && opts.YAxis.Min != nil {
		yMin = *opts.YAxis.Min
	}
	if opts.YAxis != nil && opts.YAxis.Max != nil {
		yMax = *opts.YAxis.Max
	}
	if yMax == yMin {
		yMax = yMin + 1
	}

	valueToAngle := func(v float64) float64 {
		t := (v - yMin) / (yMax - yMin)
		return startRad + t*(endRad-startRad)
	}

	// Track background (light gray).
	drawArcBand(doc, cx, cy, outerR, innerR, startRad, endRad, render.DefaultGridColor)

	// Plot bands.
	for _, pb := range go_.PlotBands {
		thick := pb.Thickness
		if thick == 0 {
			thick = trackThick
		}
		pOuter := outerR
		pInner := outerR - thick
		pStart := valueToAngle(pb.From)
		pEnd := valueToAngle(pb.To)
		drawArcBand(doc, cx, cy, pOuter, pInner, pStart, pEnd, pb.Color)
	}

	// Get value from first series.
	value := yMin
	if len(opts.Series) > 0 && len(opts.Series[0].Data) > 0 {
		value = opts.Series[0].Data[0]
	}

	if go_.Solid {
		// Solid gauge: fill from start to value angle.
		seriesColor := chart.SeriesColor(opts, 0)
		if len(opts.Series) > 0 && opts.Series[0].Color != nil {
			seriesColor = *opts.Series[0].Color
		}
		vAngle := valueToAngle(value)
		drawArcBand(doc, cx, cy, outerR, innerR, startRad, vAngle, seriesColor)
	} else {
		// Needle gauge.
		vAngle := valueToAngle(value)
		needleLen := innerR * 0.85
		needleW := outerR * 0.04

		// Needle triangle.
		npts := []pdf.Point{
			{X: cx + needleLen*math.Cos(vAngle), Y: cy + needleLen*math.Sin(vAngle)},
			{X: cx + needleW*math.Cos(vAngle+math.Pi/2), Y: cy + needleW*math.Sin(vAngle+math.Pi/2)},
			{X: cx + needleW*math.Cos(vAngle-math.Pi/2), Y: cy + needleW*math.Sin(vAngle-math.Pi/2)},
		}
		needleColor := pdf.Color{R: 60, G: 60, B: 60}
		doc.FillPolygon(npts, needleColor)

		// Pivot circle.
		doc.FillCircle(cx, cy, outerR*0.06, needleColor)
		doc.FillCircle(cx, cy, outerR*0.03, pdf.ColorWhite)

		// Tick labels on the arc.
		if opts.FontName != "" {
			if err := doc.SetFont(opts.FontName, fs*0.85); err != nil {
				return err
			}
			doc.SetTextColor(80, 80, 80)
			nTicks := 5
			for i := 0; i <= nTicks; i++ {
				t := float64(i) / float64(nTicks)
				v := yMin + t*(yMax-yMin)
				a := startRad + t*(endRad-startRad)
				labelR := outerR + fs*1.2
				lx := cx + labelR*math.Cos(a)
				ly := cy + labelR*math.Sin(a)
				label := render.FormatAxisValue(v, opts.YAxis)
				lw, _ := doc.MeasureText(label)
				doc.WriteLine(label, lx-lw/2, ly-fs*0.4) //nolint:errcheck
			}
		}
	}

	// Center value label.
	if go_.DataLabels != nil && render.BoolVal(go_.DataLabels.Enabled, false) && opts.FontName != "" {
		lfs := fs * 1.5
		if go_.DataLabels.FontSize > 0 {
			lfs = go_.DataLabels.FontSize
		}
		fn := opts.FontName
		if go_.DataLabels.FontName != "" {
			fn = go_.DataLabels.FontName
		}
		if err := doc.SetFont(fn, lfs); err != nil {
			return err
		}
		c := pdf.ColorBlack
		if go_.DataLabels.Color != nil {
			c = *go_.DataLabels.Color
		}
		doc.SetTextColor(c.R, c.G, c.B)
		label := render.FormatFloat(value)
		lw, _ := doc.MeasureText(label)
		doc.WriteLine(label, cx-lw/2, cy+outerR*0.35) //nolint:errcheck
	}

	return render.DrawLegend(doc, opts, legendArea)
}

// drawArcBand draws a filled arc ring segment.
func drawArcBand(doc *pdf.Document, cx, cy, outerR, innerR, startAngle, endAngle float64, color pdf.Color) {
	if endAngle <= startAngle {
		return
	}
	pts := render.DonutSlicePolygon(cx, cy, outerR, innerR, startAngle, endAngle)
	doc.FillPolygon(pts, color)
}

func gaugeOptions(opts chart.Options) *chart.GaugeOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Gauge != nil {
		return opts.PlotOptions.Gauge
	}
	return &chart.GaugeOptions{}
}

var _ chart.Drawable = (*GaugeChart)(nil)
