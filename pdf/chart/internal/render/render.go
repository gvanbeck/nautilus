// Package render provides shared drawing utilities for the chart sub-packages.
// It is an internal package — only code within pdf/chart/... may import it.
package render

import (
	"math"
	"strconv"
	"strings"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
)

// Area is a rectangular region described by its top-left corner and size.
type Area struct {
	X, Y, W, H float64
}

// Right returns the X coordinate of the right edge.
func (a Area) Right() float64 { return a.X + a.W }

// Bottom returns the Y coordinate of the bottom edge.
func (a Area) Bottom() float64 { return a.Y + a.H }

// Layout describes how a chart bounding box is divided into functional zones.
type Layout struct {
	Plot   Area // inner rectangle where data is drawn
	YAxis  Area // left column reserved for y-axis labels
	XAxis  Area // bottom row reserved for x-axis labels
	Legend Area // legend strip
}

// DefaultGridColor is the default color for axis grid lines.
var DefaultGridColor = pdf.Color{R: 224, G: 224, B: 224}

// DefaultAxisColor is the color for axis baselines and tick marks.
var DefaultAxisColor = pdf.Color{R: 140, G: 140, B: 140}

// Spacing constants (points).
const (
	titlePad      = 6.0  // gap below title / subtitle
	legendPad     = 8.0  // gap above legend strip
	legendSwatch  = 8.0  // swatch box size in legend
	legendGap     = 4.0  // gap between swatch and text
	legendSpacing = 12.0 // gap between legend items
	yLabelPad     = 6.0  // gap between y-label right edge and plot left edge
	xLabelPad     = 4.0  // gap between plot bottom and x-label top
)

// EffectiveFontSize returns the base font size, defaulting to 9 when 0.
func EffectiveFontSize(opts chart.Options) float64 {
	if opts.FontSize > 0 {
		return opts.FontSize
	}
	return 9
}

// BoolVal returns *b if non-nil, else def.
func BoolVal(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

// LightenColor blends color c towards white by (1-alpha).
// alpha=1 → original color; alpha=0 → white.
func LightenColor(c pdf.Color, alpha float64) pdf.Color {
	blend := func(v uint8) uint8 {
		return uint8(math.Round(255 + (float64(v)-255)*alpha))
	}
	return pdf.Color{R: blend(c.R), G: blend(c.G), B: blend(c.B)}
}

// FormatFloat formats a float64 removing trailing zeros.
func FormatFloat(v float64) string {
	if v == math.Trunc(v) {
		return strconv.FormatInt(int64(v), 10)
	}
	s := strconv.FormatFloat(v, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// FormatAxisValue formats v according to any axis label format string.
func FormatAxisValue(v float64, axis *chart.Axis) string {
	s := FormatFloat(v)
	if axis != nil && axis.Labels != nil && axis.Labels.Format != "" {
		s = strings.ReplaceAll(axis.Labels.Format, "{value}", s)
	}
	return s
}

// AutoCategories returns category labels for n items when no explicit
// categories are configured ("1", "2", …, n).
func AutoCategories(n int) []string {
	cats := make([]string, n)
	for i := range cats {
		cats[i] = strconv.Itoa(i + 1)
	}
	return cats
}

// CategoriesFor returns the configured categories or auto-generated ones.
func CategoriesFor(opts chart.Options) []string {
	if opts.XAxis != nil && len(opts.XAxis.Categories) > 0 {
		return opts.XAxis.Categories
	}
	if len(opts.Series) > 0 {
		return AutoCategories(len(opts.Series[0].Data))
	}
	return nil
}

// DataRange returns the min and max values across all series.
func DataRange(series []chart.Series) (min, max float64) {
	first := true
	for _, s := range series {
		for _, v := range s.Data {
			if first {
				min = v
				max = v
				first = false
				continue
			}
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
	}
	return
}

// NiceRange computes a clean min/max/step that covers [dataMin, dataMax].
// The result always starts at 0 when all data is non-negative.
func NiceRange(dataMin, dataMax float64, axis *chart.Axis) (niceMin, niceMax, step float64) {
	if axis != nil {
		if axis.Min != nil {
			dataMin = *axis.Min
		}
		if axis.Max != nil {
			dataMax = *axis.Max
		}
		if axis.TickInterval != nil {
			step = *axis.TickInterval
			niceMin = math.Floor(dataMin/step) * step
			niceMax = math.Ceil(dataMax/step) * step
			return
		}
	}
	if dataMin == dataMax {
		if dataMin == 0 {
			dataMax = 1
		} else {
			dataMin = 0
		}
	}
	if dataMin > 0 {
		dataMin = 0
	}
	r := dataMax - dataMin
	magnitude := math.Pow(10, math.Floor(math.Log10(r/5)))
	for _, f := range []float64{1, 2, 2.5, 5, 10} {
		s := magnitude * f
		if r/s <= 6 {
			step = s
			break
		}
	}
	if step == 0 {
		step = r / 5
	}
	niceMin = math.Floor(dataMin/step) * step
	niceMax = math.Ceil(dataMax/step) * step
	return
}

// ComputeLayout divides (x,y,w,h) into title, axes, plot, and legend areas.
func ComputeLayout(opts chart.Options, x, y, w, h float64) Layout {
	fs := EffectiveFontSize(opts)

	// Reserve space for title.
	top := y
	if opts.Title != nil && opts.Title.Text != "" {
		tfs := fs * 1.5
		if opts.Title.FontSize > 0 {
			tfs = opts.Title.FontSize
		}
		top += tfs + titlePad
	}
	if opts.Subtitle != nil && opts.Subtitle.Text != "" {
		sfs := fs * 1.1
		if opts.Subtitle.FontSize > 0 {
			sfs = opts.Subtitle.FontSize
		}
		top += sfs + titlePad
	}

	// Reserve space for legend.
	legendEnabled := opts.Legend == nil || BoolVal(opts.Legend.Enabled, true)
	legendH := 0.0
	if legendEnabled && len(opts.Series) > 0 {
		lfs := fs
		if opts.Legend != nil && opts.Legend.FontSize > 0 {
			lfs = opts.Legend.FontSize
		}
		legendH = lfs + legendPad + 4
	}

	legendOnTop := opts.Legend != nil && opts.Legend.VerticalAlign == "top"
	plotTop := top
	if legendOnTop {
		plotTop += legendH
	}
	plotBottom := y + h
	if !legendOnTop {
		plotBottom -= legendH
	}

	// Reserve space for y-axis labels (left side).
	yAxisVisible := opts.YAxis == nil || BoolVal(opts.YAxis.Visible, true)
	yAxisW := 0.0
	if yAxisVisible {
		yAxisW = fs*0.6*7 + yLabelPad // ~7 chars wide
	}

	// Reserve space for x-axis labels + optional title (bottom).
	xAxisVisible := opts.XAxis == nil || BoolVal(opts.XAxis.Visible, true)
	xAxisH := 0.0
	if xAxisVisible {
		xAxisH = fs*1.2 + xLabelPad
		if opts.XAxis != nil && opts.XAxis.Title != nil && opts.XAxis.Title.Text != "" {
			xAxisH += fs*1.2 + 4
		}
	}

	plotX := x + yAxisW
	plotY := plotTop
	plotW := w - yAxisW
	plotH := (plotBottom - xAxisH) - plotY

	var legendArea Area
	if legendOnTop {
		legendArea = Area{X: x, Y: top - legendH, W: w, H: legendH}
	} else {
		legendArea = Area{X: x, Y: y + h - legendH, W: w, H: legendH}
	}

	return Layout{
		Plot:   Area{X: plotX, Y: plotY, W: plotW, H: plotH},
		YAxis:  Area{X: x, Y: plotY, W: yAxisW, H: plotH},
		XAxis:  Area{X: plotX, Y: plotY + plotH, W: plotW, H: xAxisH},
		Legend: legendArea,
	}
}

// ValueToY converts a data value to a Y pixel coordinate within plot.
// yMin maps to plot.Bottom(); yMax maps to plot.Y.
func ValueToY(v, yMin, yMax float64, plot Area) float64 {
	if yMax == yMin {
		return plot.Y + plot.H/2
	}
	frac := (v - yMin) / (yMax - yMin)
	return plot.Bottom() - frac*plot.H
}

// CategoryCenterX returns the centre X of the i-th bucket (0-based) when
// there are n equal-width buckets spanning plot.X to plot.Right().
func CategoryCenterX(i, n int, plot Area) float64 {
	if n <= 0 {
		return plot.X
	}
	step := plot.W / float64(n)
	return plot.X + (float64(i)+0.5)*step
}

// CategoryLeftX returns the left edge X of the i-th bucket.
func CategoryLeftX(i, n int, plot Area) float64 {
	if n <= 0 {
		return plot.X
	}
	return plot.X + float64(i)*(plot.W/float64(n))
}

// DrawBackground fills the chart bounding box if a background color is set.
func DrawBackground(doc *pdf.Document, opts chart.Options, x, y, w, h float64) {
	if opts.Background != nil {
		doc.FillRect(x, y, w, h, *opts.Background)
	}
}

// DrawTitle renders the chart title and subtitle.
func DrawTitle(doc *pdf.Document, opts chart.Options, x, y, w float64) error {
	fs := EffectiveFontSize(opts)
	curY := y

	if opts.Title != nil && opts.Title.Text != "" {
		tfs := fs * 1.5
		if opts.Title.FontSize > 0 {
			tfs = opts.Title.FontSize
		}
		fn := opts.FontName
		if opts.Title.FontName != "" {
			fn = opts.Title.FontName
		}
		if fn != "" {
			if err := doc.SetFont(fn, tfs); err != nil {
				return err
			}
			c := pdf.ColorBlack
			if opts.Title.Color != nil {
				c = *opts.Title.Color
			}
			doc.SetTextColor(c.R, c.G, c.B)
			tw, _ := doc.MeasureText(opts.Title.Text)
			if _, err := doc.WriteLine(opts.Title.Text, x+(w-tw)/2, curY); err != nil {
				return err
			}
		}
		curY += tfs + titlePad
	}

	if opts.Subtitle != nil && opts.Subtitle.Text != "" {
		sfs := fs * 1.1
		if opts.Subtitle.FontSize > 0 {
			sfs = opts.Subtitle.FontSize
		}
		fn := opts.FontName
		if opts.Subtitle.FontName != "" {
			fn = opts.Subtitle.FontName
		}
		if fn != "" {
			if err := doc.SetFont(fn, sfs); err != nil {
				return err
			}
			c := pdf.ColorGray
			if opts.Subtitle.Color != nil {
				c = *opts.Subtitle.Color
			}
			doc.SetTextColor(c.R, c.G, c.B)
			tw, _ := doc.MeasureText(opts.Subtitle.Text)
			if _, err := doc.WriteLine(opts.Subtitle.Text, x+(w-tw)/2, curY); err != nil {
				return err
			}
		}
	}

	return nil
}

// DrawYAxis draws y-axis gridlines, baseline, and value labels.
func DrawYAxis(doc *pdf.Document, opts chart.Options, plot, yAxisArea Area, yMin, yMax, step float64) error {
	if opts.YAxis != nil && !BoolVal(opts.YAxis.Visible, true) {
		return nil
	}

	fs := EffectiveFontSize(opts)

	// Grid settings.
	gridW := 0.5
	gridColor := DefaultGridColor
	if opts.YAxis != nil {
		if opts.YAxis.GridLineWidth > 0 {
			gridW = opts.YAxis.GridLineWidth
		} else if opts.YAxis.GridLineWidth < 0 {
			gridW = 0
		}
		if opts.YAxis.GridLineColor != nil {
			gridColor = *opts.YAxis.GridLineColor
		}
	}

	// Label settings.
	labelFont := opts.FontName
	labelSize := fs
	labelsEnabled := true
	if opts.YAxis != nil && opts.YAxis.Labels != nil {
		if !BoolVal(opts.YAxis.Labels.Enabled, true) {
			labelsEnabled = false
		}
		if opts.YAxis.Labels.FontName != "" {
			labelFont = opts.YAxis.Labels.FontName
		}
		if opts.YAxis.Labels.FontSize > 0 {
			labelSize = opts.YAxis.Labels.FontSize
		}
	}

	if labelsEnabled && labelFont != "" {
		if err := doc.SetFont(labelFont, labelSize); err != nil {
			return err
		}
	}

	// Draw tick grid lines and labels.
	for v := yMin; v <= yMax+step*0.001; v += step {
		py := ValueToY(v, yMin, yMax, plot)

		if gridW > 0 {
			doc.DrawLine(plot.X, py, plot.Right(), py, gridW, gridColor)
		}

		if labelsEnabled && labelFont != "" {
			label := FormatAxisValue(v, opts.YAxis)
			doc.SetTextColor(100, 100, 100)
			lw, _ := doc.MeasureText(label)
			lx := yAxisArea.X + yAxisArea.W - yLabelPad - lw
			ly := py - labelSize*0.4
			if _, err := doc.WriteLine(label, lx, ly); err != nil {
				return err
			}
		}
	}

	// Y-axis title (horizontal, above the axis area).
	if opts.YAxis != nil && opts.YAxis.Title != nil && opts.YAxis.Title.Text != "" {
		t := opts.YAxis.Title
		fn := opts.FontName
		if t.FontName != "" {
			fn = t.FontName
		}
		tfs := fs
		if t.FontSize > 0 {
			tfs = t.FontSize
		}
		if fn != "" {
			if err := doc.SetFont(fn, tfs); err != nil {
				return err
			}
			c := pdf.ColorGray
			if t.Color != nil {
				c = *t.Color
			}
			doc.SetTextColor(c.R, c.G, c.B)
			tw, _ := doc.MeasureText(t.Text)
			if _, err := doc.WriteLine(t.Text, yAxisArea.X+(yAxisArea.W-tw)/2, yAxisArea.Y-tfs-2); err != nil {
				return err
			}
		}
	}

	// Left baseline.
	doc.DrawLine(plot.X, plot.Y, plot.X, plot.Bottom(), 0.5, DefaultAxisColor)
	return nil
}

// DrawXAxis draws the x-axis baseline and category labels.
func DrawXAxis(doc *pdf.Document, opts chart.Options, plot, xAxisArea Area, categories []string) error {
	if opts.XAxis != nil && !BoolVal(opts.XAxis.Visible, true) {
		return nil
	}

	// Bottom baseline.
	doc.DrawLine(plot.X, plot.Bottom(), plot.Right(), plot.Bottom(), 0.5, DefaultAxisColor)

	n := len(categories)
	if n == 0 {
		return nil
	}

	fs := EffectiveFontSize(opts)
	labelFont := opts.FontName
	labelSize := fs
	labelsEnabled := true
	if opts.XAxis != nil && opts.XAxis.Labels != nil {
		if !BoolVal(opts.XAxis.Labels.Enabled, true) {
			labelsEnabled = false
		}
		if opts.XAxis.Labels.FontName != "" {
			labelFont = opts.XAxis.Labels.FontName
		}
		if opts.XAxis.Labels.FontSize > 0 {
			labelSize = opts.XAxis.Labels.FontSize
		}
	}

	if labelsEnabled && labelFont != "" {
		if err := doc.SetFont(labelFont, labelSize); err != nil {
			return err
		}
		doc.SetTextColor(100, 100, 100)
		for i, cat := range categories {
			cx := CategoryCenterX(i, n, plot)
			lw, _ := doc.MeasureText(cat)
			if _, err := doc.WriteLine(cat, cx-lw/2, xAxisArea.Y+xLabelPad); err != nil {
				return err
			}
		}
	}

	// Optional x-axis title.
	if opts.XAxis != nil && opts.XAxis.Title != nil && opts.XAxis.Title.Text != "" {
		t := opts.XAxis.Title
		fn := opts.FontName
		if t.FontName != "" {
			fn = t.FontName
		}
		tfs := fs
		if t.FontSize > 0 {
			tfs = t.FontSize
		}
		if fn != "" {
			if err := doc.SetFont(fn, tfs); err != nil {
				return err
			}
			c := pdf.ColorGray
			if t.Color != nil {
				c = *t.Color
			}
			doc.SetTextColor(c.R, c.G, c.B)
			tw, _ := doc.MeasureText(t.Text)
			ty := xAxisArea.Bottom() - tfs - 2
			if _, err := doc.WriteLine(t.Text, plot.X+(plot.W-tw)/2, ty); err != nil {
				return err
			}
		}
	}

	return nil
}

// DrawLegend renders the chart legend into legendArea.
func DrawLegend(doc *pdf.Document, opts chart.Options, legendArea Area) error {
	if opts.Legend != nil && !BoolVal(opts.Legend.Enabled, true) {
		return nil
	}
	if len(opts.Series) == 0 {
		return nil
	}

	fs := EffectiveFontSize(opts)
	lfn := opts.FontName
	lfs := fs
	if opts.Legend != nil {
		if opts.Legend.FontName != "" {
			lfn = opts.Legend.FontName
		}
		if opts.Legend.FontSize > 0 {
			lfs = opts.Legend.FontSize
		}
	}
	if lfn == "" {
		return nil
	}
	if err := doc.SetFont(lfn, lfs); err != nil {
		return err
	}

	type item struct {
		color pdf.Color
		name  string
		textW float64
	}
	items := make([]item, len(opts.Series))
	totalW := 0.0
	for i, s := range opts.Series {
		c := chart.SeriesColor(opts, i)
		if s.Color != nil {
			c = *s.Color
		}
		tw, _ := doc.MeasureText(s.Name)
		items[i] = item{color: c, name: s.Name, textW: tw}
		totalW += legendSwatch + legendGap + tw
		if i < len(opts.Series)-1 {
			totalW += legendSpacing
		}
	}

	align := "center"
	if opts.Legend != nil && opts.Legend.Align != "" {
		align = opts.Legend.Align
	}

	startX := legendArea.X
	switch align {
	case "center":
		startX = legendArea.X + (legendArea.W-totalW)/2
	case "right":
		startX = legendArea.Right() - totalW
	}

	cy := legendArea.Y + legendPad
	cx := startX
	doc.SetTextColor(60, 60, 60)

	for _, it := range items {
		doc.FillRect(cx, cy, legendSwatch, legendSwatch, it.color)
		ty := cy + legendSwatch - lfs - 1
		if ty < cy {
			ty = cy
		}
		if _, err := doc.WriteLine(it.name, cx+legendSwatch+legendGap, ty); err != nil {
			return err
		}
		cx += legendSwatch + legendGap + it.textW + legendSpacing
	}

	return nil
}

// DrawDataLabel renders a formatted value label above (cx, cy).
func DrawDataLabel(doc *pdf.Document, opts chart.Options, dl *chart.DataLabels, value, cx, cy float64) error {
	if dl == nil || !BoolVal(dl.Enabled, false) {
		return nil
	}
	fn := opts.FontName
	if dl.FontName != "" {
		fn = dl.FontName
	}
	if fn == "" {
		return nil
	}
	fs := EffectiveFontSize(opts)
	lfs := fs
	if dl.FontSize > 0 {
		lfs = dl.FontSize
	}
	if err := doc.SetFont(fn, lfs); err != nil {
		return err
	}
	c := pdf.ColorBlack
	if dl.Color != nil {
		c = *dl.Color
	}
	doc.SetTextColor(c.R, c.G, c.B)

	label := FormatFloat(value)
	if dl.Format != "" {
		label = strings.ReplaceAll(dl.Format, "{y}", label)
	}
	lw, _ := doc.MeasureText(label)
	if _, err := doc.WriteLine(label, cx-lw/2, cy-lfs-2); err != nil {
		return err
	}
	return nil
}

// DrawMarker draws a marker symbol centered at (cx, cy) with radius r.
func DrawMarker(doc *pdf.Document, symbol string, cx, cy, r float64, color pdf.Color) {
	switch symbol {
	case "square":
		doc.FillRect(cx-r, cy-r, r*2, r*2, color)
	case "diamond":
		pts := []pdf.Point{
			{X: cx, Y: cy - r},
			{X: cx + r, Y: cy},
			{X: cx, Y: cy + r},
			{X: cx - r, Y: cy},
		}
		doc.FillPolygon(pts, color)
	default: // "circle"
		doc.FillCircle(cx, cy, r, color)
	}
}

// PieSlicePolygon returns the polygon for a pie slice.
// startAngle and endAngle are in radians (0 = east, increases clockwise).
func PieSlicePolygon(cx, cy, r, startAngle, endAngle float64) []pdf.Point {
	arcAngle := endAngle - startAngle
	n := int(math.Ceil(arcAngle / (2 * math.Pi) * 64))
	if n < 4 {
		n = 4
	}
	pts := make([]pdf.Point, 0, n+2)
	pts = append(pts, pdf.Point{X: cx, Y: cy})
	for i := 0; i <= n; i++ {
		a := startAngle + arcAngle*float64(i)/float64(n)
		pts = append(pts, pdf.Point{X: cx + r*math.Cos(a), Y: cy + r*math.Sin(a)})
	}
	return pts
}

// DonutSlicePolygon returns the polygon for a donut slice (ring segment).
func DonutSlicePolygon(cx, cy, outerR, innerR, startAngle, endAngle float64) []pdf.Point {
	arcAngle := endAngle - startAngle
	n := int(math.Ceil(arcAngle / (2 * math.Pi) * 64))
	if n < 4 {
		n = 4
	}
	pts := make([]pdf.Point, 0, (n+1)*2)
	for i := 0; i <= n; i++ {
		a := startAngle + arcAngle*float64(i)/float64(n)
		pts = append(pts, pdf.Point{X: cx + outerR*math.Cos(a), Y: cy + outerR*math.Sin(a)})
	}
	for i := n; i >= 0; i-- {
		a := startAngle + arcAngle*float64(i)/float64(n)
		pts = append(pts, pdf.Point{X: cx + innerR*math.Cos(a), Y: cy + innerR*math.Sin(a)})
	}
	return pts
}

// ValueToX converts a data value to an X pixel coordinate within plot.
// xMin maps to plot.X; xMax maps to plot.Right().
func ValueToX(v, xMin, xMax float64, plot Area) float64 {
	if xMax == xMin {
		return plot.X + plot.W/2
	}
	frac := (v - xMin) / (xMax - xMin)
	return plot.X + frac*plot.W
}

// XDataRange returns the min and max X values across all series points.
func XDataRange(series []chart.Series) (min, max float64) {
	first := true
	for _, s := range series {
		for _, p := range s.Points {
			if first {
				min = p.X
				max = p.X
				first = false
				continue
			}
			if p.X < min {
				min = p.X
			}
			if p.X > max {
				max = p.X
			}
		}
	}
	return
}

// ZDataRange returns the min and max Z values across all series points.
func ZDataRange(series []chart.Series) (min, max float64) {
	first := true
	for _, s := range series {
		for _, p := range s.Points {
			if first {
				min = p.Z
				max = p.Z
				first = false
				continue
			}
			if p.Z < min {
				min = p.Z
			}
			if p.Z > max {
				max = p.Z
			}
		}
	}
	return
}

// DrawXAxisNumeric draws a numeric x-axis (for scatter/bubble charts).
func DrawXAxisNumeric(doc *pdf.Document, opts chart.Options, plot, xAxisArea Area, xMin, xMax, xStep float64) error {
	if opts.XAxis != nil && !BoolVal(opts.XAxis.Visible, true) {
		return nil
	}

	// Bottom baseline.
	doc.DrawLine(plot.X, plot.Bottom(), plot.Right(), plot.Bottom(), 0.5, DefaultAxisColor)

	fs := EffectiveFontSize(opts)
	labelFont := opts.FontName
	labelSize := fs
	labelsEnabled := true
	gridW := 0.5
	gridColor := DefaultGridColor

	if opts.XAxis != nil {
		if opts.XAxis.GridLineWidth > 0 {
			gridW = opts.XAxis.GridLineWidth
		} else if opts.XAxis.GridLineWidth < 0 {
			gridW = 0
		}
		if opts.XAxis.GridLineColor != nil {
			gridColor = *opts.XAxis.GridLineColor
		}
		if opts.XAxis.Labels != nil {
			if !BoolVal(opts.XAxis.Labels.Enabled, true) {
				labelsEnabled = false
			}
			if opts.XAxis.Labels.FontName != "" {
				labelFont = opts.XAxis.Labels.FontName
			}
			if opts.XAxis.Labels.FontSize > 0 {
				labelSize = opts.XAxis.Labels.FontSize
			}
		}
	}

	if labelsEnabled && labelFont != "" {
		if err := doc.SetFont(labelFont, labelSize); err != nil {
			return err
		}
		doc.SetTextColor(100, 100, 100)
	}

	for v := xMin; v <= xMax+xStep*0.001; v += xStep {
		px := ValueToX(v, xMin, xMax, plot)
		if gridW > 0 {
			doc.DrawLine(px, plot.Y, px, plot.Bottom(), gridW, gridColor)
		}
		if labelsEnabled && labelFont != "" {
			label := FormatAxisValue(v, opts.XAxis)
			lw, _ := doc.MeasureText(label)
			if _, err := doc.WriteLine(label, px-lw/2, xAxisArea.Y+xLabelPad); err != nil {
				return err
			}
		}
	}
	return nil
}

// MarkerOrDefault returns marker if non-nil, otherwise a zero-value Marker.
func MarkerOrDefault(m *chart.Marker) *chart.Marker {
	if m != nil {
		return m
	}
	return &chart.Marker{}
}

// PointYRange returns the min and max Y values across all series points.
func PointYRange(series []chart.Series) (min, max float64) {
	first := true
	for _, s := range series {
		for _, p := range s.Points {
			if first {
				min = p.Y
				max = p.Y
				first = false
				continue
			}
			if p.Y < min {
				min = p.Y
			}
			if p.Y > max {
				max = p.Y
			}
		}
	}
	return
}

// LowHighRange returns the min and max across Low and High fields of all series points.
func LowHighRange(series []chart.Series) (min, max float64) {
	first := true
	for _, s := range series {
		for _, p := range s.Points {
			for _, v := range []float64{p.Low, p.High} {
				if first {
					min = v
					max = v
					first = false
					continue
				}
				if v < min {
					min = v
				}
				if v > max {
					max = v
				}
			}
		}
	}
	return
}

// BlendColor interpolates between a and b by t (0=a, 1=b).
func BlendColor(a, b pdf.Color, t float64) pdf.Color {
	lerp := func(x, y uint8) uint8 {
		return uint8(math.Round(float64(x) + (float64(y)-float64(x))*t))
	}
	return pdf.Color{R: lerp(a.R, b.R), G: lerp(a.G, b.G), B: lerp(a.B, b.B)}
}

// ParsePercent parses a percentage string like "50%" and returns the value
// divided by 100.  Returns 0 for empty or invalid strings.
func ParsePercent(s string) float64 {
	s = strings.TrimSpace(strings.TrimSuffix(s, "%"))
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v / 100
}
