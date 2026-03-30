// Package polar provides a polar (spider/radar) chart renderer for the
// Nautilus PDF library.
//
// A polar chart plots multi-dimensional data on a web of equally-spaced spokes.
// Each spoke represents one category; the distance from the centre represents
// the value.  Multiple series are drawn as overlapping filled polygons,
// mirroring the Highcharts polar-spider demo.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Budget vs Spending"},
//	    XAxis: &chart.Axis{
//	        Categories: []string{"Sales", "Marketing", "Development",
//	            "Customer Support", "IT", "Administration"},
//	    },
//	    YAxis: &chart.Axis{Min: chart.Float(0)},
//	    Series: []chart.Series{
//	        {Name: "Allocated Budget", Data: []float64{43000, 19000, 60000, 35000, 17000, 10000}},
//	        {Name: "Actual Spending",  Data: []float64{50000, 39000, 42000, 31000, 26000, 14000}},
//	    },
//	    Legend: &chart.Legend{},
//	    PlotOptions: &chart.PlotOptions{
//	        Polar: &chart.PolarOptions{GridLineInterpolation: "polygon"},
//	    },
//	}
//	pc := &polar.PolarChart{Options: opts}
//	pc.Draw(doc, 50, 50, 400, 300)
package polar

import (
	"math"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// PolarChart renders a polar (spider/radar) chart onto a pdf.Document.
type PolarChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *PolarChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	po := polarOptions(opts)
	fs := render.EffectiveFontSize(opts)

	render.DrawBackground(doc, opts, x, y, width, height)
	if err := render.DrawTitle(doc, opts, x, y, width); err != nil {
		return err
	}

	// Reserve space consumed by title.
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

	// Reserve legend strip at bottom.
	legendH := 0.0
	if (opts.Legend == nil || render.BoolVal(opts.Legend.Enabled, true)) && len(opts.Series) > 0 {
		legendH = fs + 12
	}
	legendArea := render.Area{X: x, Y: y + height - legendH, W: width, H: legendH}

	// Available area for the radar circle.
	areaH := (y + height - legendH) - top
	areaW := width

	// Categories become the spokes.
	categories := render.CategoriesFor(opts)
	n := len(categories)
	if n < 3 {
		return render.DrawLegend(doc, opts, legendArea)
	}

	// Compute data range.
	dataMin, dataMax := render.DataRange(opts.Series)
	yMin, yMax, yStep := render.NiceRange(dataMin, dataMax, opts.YAxis)

	// Centre and outer radius.
	// Leave ~25 % of half-dimension for spoke labels.
	labelPad := math.Max(fs*3, math.Min(areaW, areaH)*0.15)
	r := math.Min(areaW, areaH)/2 - labelPad
	if r <= 0 {
		return render.DrawLegend(doc, opts, legendArea)
	}
	cx := x + areaW/2
	cy := top + areaH/2

	// spokeAngle returns the angle (radians) for spoke i.
	// Angle 0 is east; we start at −π/2 (top/12 o'clock) and go clockwise.
	spokeAngle := func(i int) float64 {
		return -math.Pi/2 + float64(i)*2*math.Pi/float64(n)
	}

	// valueToR maps a data value to a radius.
	valueToR := func(v float64) float64 {
		if yMax == yMin {
			return 0
		}
		return (v - yMin) / (yMax - yMin) * r
	}

	// ── Grid settings ─────────────────────────────────────────────────────
	gridColor := render.DefaultGridColor
	if opts.YAxis != nil && opts.YAxis.GridLineColor != nil {
		gridColor = *opts.YAxis.GridLineColor
	}
	gridW := 0.5
	if opts.YAxis != nil {
		if opts.YAxis.GridLineWidth > 0 {
			gridW = opts.YAxis.GridLineWidth
		} else if opts.YAxis.GridLineWidth < 0 {
			gridW = 0
		}
	}

	useCircle := po.GridLineInterpolation == "circle"

	// ── Draw concentric grid rings ─────────────────────────────────────────
	labelFont := opts.FontName
	labelSize := fs
	yLabelsEnabled := opts.YAxis == nil || render.BoolVal(opts.YAxis.Visible, true)
	if opts.YAxis != nil && opts.YAxis.Labels != nil {
		if !render.BoolVal(opts.YAxis.Labels.Enabled, true) {
			yLabelsEnabled = false
		}
		if opts.YAxis.Labels.FontName != "" {
			labelFont = opts.YAxis.Labels.FontName
		}
		if opts.YAxis.Labels.FontSize > 0 {
			labelSize = opts.YAxis.Labels.FontSize
		}
	}
	if yLabelsEnabled && labelFont != "" {
		if err := doc.SetFont(labelFont, labelSize); err != nil {
			return err
		}
	}

	for v := yMin + yStep; v <= yMax+yStep*0.001; v += yStep {
		gr := valueToR(v)
		if gr <= 0 {
			continue
		}
		if gridW > 0 {
			if useCircle {
				doc.StrokeCircle(cx, cy, gr, gridW, gridColor)
			} else {
				pts := spokePolygon(cx, cy, gr, n, spokeAngle)
				drawPolyline(doc, pts, true, gridW, gridColor)
			}
		}
		// Y-axis value label along the top spoke (slightly right of center).
		if yLabelsEnabled && labelFont != "" {
			label := render.FormatAxisValue(v, opts.YAxis)
			doc.SetTextColor(120, 120, 120)
			doc.WriteLine(label, cx+2, cy-gr-labelSize*0.4) //nolint:errcheck
		}
	}

	// ── Draw spokes ────────────────────────────────────────────────────────
	for i := 0; i < n; i++ {
		a := spokeAngle(i)
		doc.DrawLine(cx, cy, cx+r*math.Cos(a), cy+r*math.Sin(a), 0.5, render.DefaultGridColor)
	}

	// ── Draw series polygons ───────────────────────────────────────────────
	fillAlpha := po.FillAlpha
	if fillAlpha == 0 {
		fillAlpha = 0.3
	}
	lineW := po.LineWidth
	if lineW == 0 {
		lineW = 2
	}

	m := render.MarkerOrDefault(po.Marker)
	markerEnabled := render.BoolVal(m.Enabled, true)
	markerR := m.Radius
	if markerR <= 0 {
		markerR = 3
	}
	markerSym := m.Symbol
	if markerSym == "" {
		markerSym = "circle"
	}

	for si, s := range opts.Series {
		if len(s.Data) == 0 {
			continue
		}
		seriesColor := chart.SeriesColor(opts, si)
		if s.Color != nil {
			seriesColor = *s.Color
		}

		// Build the polygon: one point per spoke.
		pts := make([]pdf.Point, n)
		for i := 0; i < n; i++ {
			v := 0.0
			if i < len(s.Data) {
				v = s.Data[i]
			}
			rv := valueToR(v)
			a := spokeAngle(i)
			pts[i] = pdf.Point{X: cx + rv*math.Cos(a), Y: cy + rv*math.Sin(a)}
		}

		// Filled area (lightened series color).
		doc.FillPolygon(pts, render.LightenColor(seriesColor, fillAlpha))
		// Outlined polygon.
		drawPolyline(doc, pts, true, lineW, seriesColor)

		// Markers.
		if markerEnabled {
			for _, pt := range pts {
				render.DrawMarker(doc, markerSym, pt.X, pt.Y, markerR, seriesColor)
			}
		}

		// Data labels.
		for i, pt := range pts {
			v := 0.0
			if i < len(s.Data) {
				v = s.Data[i]
			}
			if err := render.DrawDataLabel(doc, opts, po.DataLabels, v, pt.X, pt.Y); err != nil {
				return err
			}
		}
	}

	// ── Spoke labels ───────────────────────────────────────────────────────
	spokeLabelFont := opts.FontName
	spokeLabelSize := fs
	spokesEnabled := opts.XAxis == nil || render.BoolVal(opts.XAxis.Visible, true)
	if opts.XAxis != nil && opts.XAxis.Labels != nil {
		if !render.BoolVal(opts.XAxis.Labels.Enabled, true) {
			spokesEnabled = false
		}
		if opts.XAxis.Labels.FontName != "" {
			spokeLabelFont = opts.XAxis.Labels.FontName
		}
		if opts.XAxis.Labels.FontSize > 0 {
			spokeLabelSize = opts.XAxis.Labels.FontSize
		}
	}
	if spokesEnabled && spokeLabelFont != "" {
		if err := doc.SetFont(spokeLabelFont, spokeLabelSize); err != nil {
			return err
		}
		doc.SetTextColor(60, 60, 60)
		for i, cat := range categories {
			a := spokeAngle(i)
			// Place label just outside the outer radius.
			lx := cx + (r+labelPad*0.35)*math.Cos(a)
			ly := cy + (r+labelPad*0.35)*math.Sin(a)
			lw, _ := doc.MeasureText(cat)
			// Horizontal alignment: centre by default; shift right/left for
			// labels that sit clearly on one side.
			ax := lx - lw/2
			cosA := math.Cos(a)
			if cosA < -0.3 {
				ax = lx - lw - 2
			} else if cosA > 0.3 {
				ax = lx + 2
			}
			ay := ly - spokeLabelSize*0.5
			doc.WriteLine(cat, ax, ay) //nolint:errcheck
		}
	}

	return render.DrawLegend(doc, opts, legendArea)
}

// spokePolygon returns the n vertices of a regular polygon with the given
// radius, each vertex placed at the corresponding spoke angle.
func spokePolygon(cx, cy, r float64, n int, angle func(int) float64) []pdf.Point {
	pts := make([]pdf.Point, n)
	for i := range pts {
		a := angle(i)
		pts[i] = pdf.Point{X: cx + r*math.Cos(a), Y: cy + r*math.Sin(a)}
	}
	return pts
}

// drawPolyline draws lines connecting pts in order.
// When closed is true it also draws the segment from the last point back to
// the first.
func drawPolyline(doc *pdf.Document, pts []pdf.Point, closed bool, lineW float64, color pdf.Color) {
	if len(pts) < 2 {
		return
	}
	last := len(pts)
	if !closed {
		last--
	}
	for i := 0; i < last; i++ {
		j := (i + 1) % len(pts)
		doc.DrawLine(pts[i].X, pts[i].Y, pts[j].X, pts[j].Y, lineW, color)
	}
}

func polarOptions(opts chart.Options) *chart.PolarOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Polar != nil {
		return opts.PlotOptions.Polar
	}
	return &chart.PolarOptions{}
}
