// Package treemap provides a treemap chart renderer for the Nautilus PDF library.
//
// A treemap displays hierarchical data as nested rectangles.  Each rectangle's
// area is proportional to its value.  The squarified layout algorithm is used
// for the most visually balanced result.
//
// Data is stored in Series.Points where each Point has a Name and Y (value)
// field.  Colors are assigned automatically from the palette (one per cell) or
// from Point.Color.
//
// Example:
//
//	opts := chart.Options{
//	    FontName: "regular",
//	    Title:    &chart.Title{Text: "Global Sales"},
//	    Series: []chart.Series{{
//	        Points: []chart.Point{
//	            {Name: "North America", Y: 42},
//	            {Name: "Europe",        Y: 35},
//	            {Name: "Asia Pacific",  Y: 18},
//	            {Name: "Latin America", Y: 5},
//	        },
//	    }},
//	}
//	tc := &treemap.TreemapChart{Options: opts}
//	tc.Draw(doc, 50, 50, 400, 280)
package treemap

import (
	"math"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// TreemapChart renders a treemap chart onto a pdf.Document.
type TreemapChart struct {
	Options chart.Options
}

// Draw renders the chart into the rectangle at (x, y) with the given size.
func (c *TreemapChart) Draw(doc *pdf.Document, x, y, width, height float64) error {
	doc.SaveGraphicsState()
	defer doc.RestoreGraphicsState()

	opts := c.Options
	to := treemapOptions(opts)
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

	plotX := x
	plotY := top
	plotW := width
	plotH := (y + height - legendH) - top

	if plotW <= 0 || plotH <= 0 {
		return render.DrawLegend(doc, opts, legendArea)
	}

	// Collect items.
	var items []tmCell
	total := 0.0
	colorIdx := 0
	for _, s := range opts.Series {
		for _, p := range s.Points {
			if p.Y <= 0 {
				continue
			}
			c := chart.SeriesColor(opts, colorIdx)
			if p.Color != nil {
				c = *p.Color
			}
			items = append(items, tmCell{name: p.Name, value: p.Y, color: c})
			total += p.Y
			colorIdx++
		}
	}
	if len(items) == 0 || total == 0 {
		return render.DrawLegend(doc, opts, legendArea)
	}

	// Normalize values to total plot area.
	totalArea := plotW * plotH
	for i := range items {
		items[i].value = items[i].value / total * totalArea
	}

	border := to.BorderWidth
	if border == 0 {
		border = 1
	}
	borderColor := pdf.ColorWhite
	if to.BorderColor != nil {
		borderColor = *to.BorderColor
	}

	// Run squarified treemap layout.
	type rect struct {
		x, y, w, h float64
		idx        int
	}
	rects := squarify(items, plotX, plotY, plotW, plotH)

	// Draw cells.
	for _, r := range rects {
		it := items[r.idx]
		cx := r.x + border/2
		cy := r.y + border/2
		cw := r.w - border
		ch := r.h - border
		if cw <= 0 || ch <= 0 {
			continue
		}
		doc.FillRect(cx, cy, cw, ch, it.color)
		if border > 0 {
			doc.FillRect(r.x, r.y, r.w, border/2, borderColor)
			doc.FillRect(r.x, r.y+r.h-border/2, r.w, border/2, borderColor)
			doc.FillRect(r.x, r.y, border/2, r.h, borderColor)
			doc.FillRect(r.x+r.w-border/2, r.y, border/2, r.h, borderColor)
		}

		// Label (if it fits).
		if it.name != "" && opts.FontName != "" {
			lfs := fs
			if to.DataLabels != nil && to.DataLabels.FontSize > 0 {
				lfs = to.DataLabels.FontSize
			}
			if cw >= lfs*2 && ch >= lfs*1.5 {
				if err := doc.SetFont(opts.FontName, lfs); err != nil {
					return err
				}
				lc := pdf.ColorWhite
				if to.DataLabels != nil && to.DataLabels.Color != nil {
					lc = *to.DataLabels.Color
				}
				doc.SetTextColor(lc.R, lc.G, lc.B)
				lw, _ := doc.MeasureText(it.name)
				if lw > cw-4 {
					// Try smaller font.
					smallFS := lfs * (cw - 4) / lw
					if smallFS >= 6 {
						doc.SetFont(opts.FontName, smallFS) //nolint:errcheck
						lw, _ = doc.MeasureText(it.name)
						lfs = smallFS
					} else {
						continue
					}
				}
				doc.WriteLine(it.name, cx+(cw-lw)/2, cy+ch/2-lfs*0.4) //nolint:errcheck
			}
		}
	}

	return render.DrawLegend(doc, opts, legendArea)
}

type tmItem struct {
	value float64
	idx   int
}

type tmRect struct {
	x, y, w, h float64
	idx        int
}

type tmCell struct {
	name  string
	value float64
	color pdf.Color
}

// squarify implements the squarified treemap algorithm.
func squarify(items []tmCell, x, y, w, h float64) []tmRect {
	if len(items) == 0 || w <= 0 || h <= 0 {
		return nil
	}

	tms := make([]tmItem, len(items))
	for i, it := range items {
		tms[i] = tmItem{value: it.value, idx: i}
	}

	var result []tmRect
	squarifySlice(tms, x, y, w, h, &result)
	return result
}

func squarifySlice(items []tmItem, x, y, w, h float64, result *[]tmRect) {
	if len(items) == 0 {
		return
	}
	if len(items) == 1 {
		*result = append(*result, tmRect{x: x, y: y, w: w, h: h, idx: items[0].idx})
		return
	}

	// Choose shorter side for the current row.
	shortSide := math.Min(w, h)
	row := []tmItem{}
	remaining := items

	for len(remaining) > 0 {
		candidate := append(row, remaining[0])
		if len(row) > 0 && worstAspect(row, shortSide) < worstAspect(candidate, shortSide) {
			break
		}
		row = candidate
		remaining = remaining[1:]
	}

	// Lay out the row.
	rowSum := 0.0
	for _, it := range row {
		rowSum += it.value
	}

	totalArea := w * h
	rowTotal := 0.0
	for _, it := range items {
		rowTotal += it.value
	}

	rowH := rowSum / rowTotal * h
	rowW := rowSum / rowTotal * w

	var rx, ry, rw, rh float64
	var nx, ny, nw, nh float64

	if w >= h {
		// Lay row as a column on the left.
		rw = rowSum / rowTotal * w
		rh = h
		rx = x
		ry = y
		nx = x + rw
		ny = y
		nw = w - rw
		nh = h
		_ = rowH
	} else {
		// Lay row as a strip at the top.
		rw = w
		rh = rowSum / rowTotal * h
		rx = x
		ry = y
		nx = x
		ny = y + rh
		nw = w
		nh = h - rh
		_ = rowW
	}

	// Place items within the row.
	cursor := 0.0
	for _, it := range row {
		frac := it.value / rowSum
		var ir tmRect
		if w >= h {
			ir = tmRect{x: rx, y: ry + cursor*rh, w: rw, h: frac * rh, idx: it.idx}
			cursor += frac
		} else {
			ir = tmRect{x: rx + cursor*rw, y: ry, w: frac * rw, h: rh, idx: it.idx}
			cursor += frac
		}
		*result = append(*result, ir)
	}

	_ = totalArea
	squarifySlice(remaining, nx, ny, nw, nh, result)
}

func worstAspect(row []tmItem, sideLen float64) float64 {
	if len(row) == 0 || sideLen == 0 {
		return math.MaxFloat64
	}
	sum := 0.0
	for _, it := range row {
		sum += it.value
	}
	worst := 0.0
	for _, it := range row {
		w := it.value * sideLen / sum
		h := sum / sideLen
		if w == 0 || h == 0 {
			continue
		}
		ratio := math.Max(w/h, h/w)
		if ratio > worst {
			worst = ratio
		}
	}
	return worst
}

func treemapOptions(opts chart.Options) *chart.TreemapOptions {
	if opts.PlotOptions != nil && opts.PlotOptions.Treemap != nil {
		return opts.PlotOptions.Treemap
	}
	return &chart.TreemapOptions{}
}

var _ chart.Drawable = (*TreemapChart)(nil)
