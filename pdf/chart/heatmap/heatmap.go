// Package heatmap provides a heatmap chart renderer for the Nautilus PDF library.
//
// Each cell is a Point{X: colIndex, Y: rowIndex, Z: value}.
// The cell color is interpolated between MinColor (lowest value) and MaxColor
// (highest value).  Row and column labels come from XAxis.Categories and
// YAxis.Categories respectively.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Sales per weekday"},
//	    XAxis:    &chart.Axis{Categories: []string{"Alice", "Bob", "Carol"}},
//	    YAxis:    &chart.Axis{Categories: []string{"Mon", "Tue", "Wed"}},
//	    Series: []chart.Series{{
//	        Name: "Sales",
//	        Points: []chart.Point{
//	            {X: 0, Y: 0, Z: 10}, {X: 1, Y: 0, Z: 19}, {X: 2, Y: 0, Z: 8},
//	            {X: 0, Y: 1, Z: 92}, {X: 1, Y: 1, Z: 58}, {X: 2, Y: 1, Z: 78},
//	            {X: 0, Y: 2, Z: 35}, {X: 1, Y: 2, Z: 15}, {X: 2, Y: 2, Z: 50},
//	        },
//	    }},
//	}
//	hc := &heatmap.HeatmapChart{Options: opts}
//	hc.Draw(doc, 50, 50, 400, 250)
package heatmap

import (
	"math"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// HeatmapChart renders a heatmap onto a pdf.Document.
type HeatmapChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *HeatmapChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	ho := heatmapOptions(opts)

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

	// Determine grid dimensions from categories.
	xCats := []string{}
	if opts.XAxis != nil {
		xCats = opts.XAxis.Categories
	}
	yCats := []string{}
	if opts.YAxis != nil {
		yCats = opts.YAxis.Categories
	}

	// Collect all points from all series.
	type cell struct {
		xi, yi int
		z      float64
		color  *pdf.Color
	}
	var cells []cell
	zMin, zMax := 0.0, 0.0
	first := true
	for _, s := range opts.Series {
		for _, p := range s.Points {
			z := p.Z
			if first || z < zMin {
				zMin = z
				first = false
			}
			if z > zMax {
				zMax = z
			}
			cells = append(cells, cell{xi: int(math.Round(p.X)), yi: int(math.Round(p.Y)), z: z, color: p.Color})
		}
	}
	if len(cells) == 0 {
		return render.DrawLegend(doc, opts, legendArea)
	}
	if zMax == zMin {
		zMax = zMin + 1
	}

	// Infer grid size from data if categories are missing.
	nCols := len(xCats)
	nRows := len(yCats)
	if nCols == 0 || nRows == 0 {
		for _, cl := range cells {
			if cl.xi+1 > nCols {
				nCols = cl.xi + 1
			}
			if cl.yi+1 > nRows {
				nRows = cl.yi + 1
			}
		}
	}
	if nCols == 0 || nRows == 0 {
		return nil
	}

	// Available area.
	xLabelW := 0.0
	if len(yCats) > 0 {
		xLabelW = fs*0.6*8 + 4 // approximate width for y-category labels
	}
	xLabelH := 0.0
	if len(xCats) > 0 {
		xLabelH = fs*1.2 + 4
	}

	plotX := x + xLabelW
	plotY := top
	plotW := width - xLabelW
	plotH := (y + height - legendH) - top - xLabelH

	cellW := plotW / float64(nCols)
	cellH := plotH / float64(nRows)

	border := ho.BorderWidth
	if border == 0 {
		border = 1
	}
	borderColor := pdf.ColorWhite
	if ho.BorderColor != nil {
		borderColor = *ho.BorderColor
	}

	// Default color scale: white → series[0] color.
	seriesColor := chart.SeriesColor(opts, 0)
	minColor := pdf.ColorWhite
	if ho.MinColor != nil {
		minColor = *ho.MinColor
	}
	maxColor := seriesColor
	if ho.MaxColor != nil {
		maxColor = *ho.MaxColor
	}

	// Draw cells.
	for _, cl := range cells {
		if cl.xi < 0 || cl.xi >= nCols || cl.yi < 0 || cl.yi >= nRows {
			continue
		}
		cx := plotX + float64(cl.xi)*cellW
		cy := plotY + float64(cl.yi)*cellH

		t := (cl.z - zMin) / (zMax - zMin)
		fillColor := render.BlendColor(minColor, maxColor, t)
		if cl.color != nil {
			fillColor = *cl.color
		}

		doc.FillRect(cx+border/2, cy+border/2, cellW-border, cellH-border, fillColor)
		if border > 0 {
			doc.FillRect(cx, cy, cellW, border/2, borderColor)
			doc.FillRect(cx, cy+cellH-border/2, cellW, border/2, borderColor)
			doc.FillRect(cx, cy, border/2, cellH, borderColor)
			doc.FillRect(cx+cellW-border/2, cy, border/2, cellH, borderColor)
		}

		// Data label.
		if ho.DataLabels != nil && render.BoolVal(ho.DataLabels.Enabled, false) && opts.FontName != "" {
			lfs := fs
			if ho.DataLabels.FontSize > 0 {
				lfs = ho.DataLabels.FontSize
			}
			fn := opts.FontName
			if ho.DataLabels.FontName != "" {
				fn = ho.DataLabels.FontName
			}
			if err := doc.SetFont(fn, lfs); err != nil {
				return err
			}
			c := pdf.ColorBlack
			if ho.DataLabels.Color != nil {
				c = *ho.DataLabels.Color
			}
			doc.SetTextColor(c.R, c.G, c.B)
			label := render.FormatFloat(cl.z)
			lw, _ := doc.MeasureText(label)
			doc.WriteLine(label, cx+cellW/2-lw/2, cy+cellH/2-lfs/2) //nolint:errcheck
		}
	}

	// Y-axis category labels (rows).
	if len(yCats) > 0 && opts.FontName != "" {
		if err := doc.SetFont(opts.FontName, fs); err != nil {
			return err
		}
		doc.SetTextColor(60, 60, 60)
		for i, cat := range yCats {
			if i >= nRows {
				break
			}
			cy := plotY + float64(i)*cellH + cellH/2 - fs*0.4
			lw, _ := doc.MeasureText(cat)
			doc.WriteLine(cat, plotX-lw-4, cy) //nolint:errcheck
		}
	}

	// X-axis category labels (columns).
	if len(xCats) > 0 && opts.FontName != "" {
		if err := doc.SetFont(opts.FontName, fs); err != nil {
			return err
		}
		doc.SetTextColor(60, 60, 60)
		for i, cat := range xCats {
			if i >= nCols {
				break
			}
			cx := plotX + float64(i)*cellW + cellW/2
			lw, _ := doc.MeasureText(cat)
			doc.WriteLine(cat, cx-lw/2, plotY+plotH+4) //nolint:errcheck
		}
	}

	return render.DrawLegend(doc, opts, legendArea)
}

func heatmapOptions(opts chart.Options) *chart.HeatmapOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Heatmap != nil {
		return opts.PlotOptions.Heatmap
	}
	return &chart.HeatmapOptions{}
}

var _ chart.Drawable = (*HeatmapChart)(nil)
