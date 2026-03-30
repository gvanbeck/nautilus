// Package pie provides a pie and donut chart renderer for the Nautilus PDF
// library.
//
// A pie chart uses one series; the values define the slice sizes.
// Set PlotOptions.Pie.InnerSize to a percentage (e.g. "50%") for a donut chart.
//
// Example — pie chart:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Market Share"},
//	    Series: []chart.Series{
//	        {Name: "Chrome",  Data: []float64{65}},
//	        {Name: "Firefox", Data: []float64{15}},
//	        {Name: "Safari",  Data: []float64{12}},
//	        {Name: "Other",   Data: []float64{8}},
//	    },
//	}
//	pc := &pie.PieChart{Options: opts}
//	pc.Draw(doc, 50, 100, 300, 220)
//
// Example — donut chart:
//
//	opts.PlotOptions = &chart.PlotOptions{
//	    Pie: &chart.PieOptions{InnerSize: "50%"},
//	}
package pie

import (
	"math"
	"strings"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// PieChart renders a pie or donut chart onto a pdf.Document.
type PieChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *PieChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	po := pieOptions(opts)

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

	// Reserve legend strip at bottom.
	legendH := 0.0
	if (opts.Legend == nil || render.BoolVal(opts.Legend.Enabled, true)) && len(opts.Series) > 0 {
		legendH = fs + 12
	}
	legendArea := render.Area{X: x, Y: y + height - legendH, W: width, H: legendH}

	// The pie fits in the remaining space.
	pieAreaH := (y + height - legendH) - top
	pieAreaW := width

	// Collect slice values from first element of each series (Highcharts pie
	// convention: each series entry is one slice).
	type slice struct {
		name  string
		value float64
		color pdf.Color
	}
	var slices []slice
	total := 0.0
	for i, s := range opts.Series {
		v := 0.0
		if len(s.Data) > 0 {
			v = s.Data[0]
		}
		if v < 0 {
			v = -v
		}
		c := chart.SeriesColor(opts, i)
		if s.Color != nil {
			c = *s.Color
		}
		slices = append(slices, slice{name: s.Name, value: v, color: c})
		total += v
	}
	if total == 0 || len(slices) == 0 {
		return render.DrawLegend(doc, opts, legendArea)
	}

	// Determine inner radius (donut).
	outerR := math.Min(pieAreaW, pieAreaH) / 2 * 0.85
	innerR := 0.0
	if po.InnerSize != "" && po.InnerSize != "0%" {
		innerFrac := render.ParsePercent(po.InnerSize)
		if innerFrac > 0 {
			innerR = outerR * innerFrac
		}
	}

	cx := x + pieAreaW/2
	cy := top + pieAreaH/2

	// Start angle: default −90° (top/12 o'clock) in radians.
	startDeg := -90.0
	if po.StartAngle != nil {
		startDeg = *po.StartAngle
	}
	angle := startDeg * math.Pi / 180

	// Draw slices.
	for _, sl := range slices {
		sweep := sl.value / total * 2 * math.Pi
		endAngle := angle + sweep

		if innerR > 0 {
			pts := render.DonutSlicePolygon(cx, cy, outerR, innerR, angle, endAngle)
			doc.FillAndStrokePolygon(pts, sl.color, 0.5, pdf.ColorWhite)
		} else {
			pts := render.PieSlicePolygon(cx, cy, outerR, angle, endAngle)
			doc.FillAndStrokePolygon(pts, sl.color, 0.5, pdf.ColorWhite)
		}

		// Data label (name + percentage) inside or outside slice.
		if po.DataLabels != nil && render.BoolVal(po.DataLabels.Enabled, false) {
			midAngle := angle + sweep/2
			labelR := outerR * 0.65
			if innerR > 0 {
				labelR = (outerR + innerR) / 2
			}
			lx := cx + labelR*math.Cos(midAngle)
			ly := cy + labelR*math.Sin(midAngle)
			drawPieLabel(doc, opts, po.DataLabels, sl.name, sl.value, total, lx, ly)
		}

		angle = endAngle
	}

	return render.DrawLegend(doc, opts, legendArea)
}

// drawPieLabel renders a slice label centered at (lx, ly).
func drawPieLabel(doc *pdf.Document, opts chart.Options, dl *chart.DataLabels, name string, value, total, lx, ly float64) {
	fn := opts.FontName
	if dl.FontName != "" {
		fn = dl.FontName
	}
	if fn == "" {
		return
	}
	fs := render.EffectiveFontSize(opts)
	lfs := fs
	if dl.FontSize > 0 {
		lfs = dl.FontSize
	}
	if err := doc.SetFont(fn, lfs); err != nil {
		return
	}
	c := pdf.ColorBlack
	if dl.Color != nil {
		c = *dl.Color
	}
	doc.SetTextColor(c.R, c.G, c.B)

	pct := value / total * 100
	label := render.FormatFloat(pct) + "%"
	if dl.Format != "" {
		label = strings.ReplaceAll(dl.Format, "{point.name}", name)
		label = strings.ReplaceAll(label, "{y}", render.FormatFloat(value))
		label = strings.ReplaceAll(label, "{percentage:.0f}", render.FormatFloat(pct))
	}

	lw, _ := doc.MeasureText(label)
	doc.WriteLine(label, lx-lw/2, ly-lfs/2) //nolint:errcheck
}

func pieOptions(opts chart.Options) *chart.PieOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Pie != nil {
		return opts.PlotOptions.Pie
	}
	return &chart.PieOptions{}
}
