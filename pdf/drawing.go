package pdf

import (
	"math"

	"github.com/signintech/gopdf"
)

// Point represents a 2D coordinate in points (1 pt = 1/72 inch).
// The origin is at the top-left corner of the page; X increases rightward
// and Y increases downward.
type Point struct {
	X, Y float64
}

// SaveGraphicsState saves the current PDF graphics state onto the stack.
// Use RestoreGraphicsState to return to the saved state.
// Calls must be balanced.  No-op during the counting pass of Build.
func (d *Document) SaveGraphicsState() {
	if !d.countingMode {
		d.pdf.SaveGraphicsState()
	}
}

// RestoreGraphicsState pops the most recently saved graphics state from
// the stack.  No-op during the counting pass of Build.
func (d *Document) RestoreGraphicsState() {
	if !d.countingMode {
		d.pdf.RestoreGraphicsState()
	}
}

// DrawLine draws a straight line from (x1,y1) to (x2,y2) with the given
// line width in points and stroke color.
// No-op during the counting pass of Build.
func (d *Document) DrawLine(x1, y1, x2, y2, lineWidth float64, color Color) {
	if d.countingMode {
		return
	}
	d.pdf.SetLineWidth(lineWidth)
	d.pdf.SetStrokeColor(color.R, color.G, color.B)
	d.pdf.SetLineType("")
	d.pdf.Line(x1, y1, x2, y2)
}

// FillPolygon draws a filled polygon with the given vertices.
// The polygon is automatically closed (last point connects to first).
// No-op when fewer than 3 points are provided or during the counting pass.
func (d *Document) FillPolygon(points []Point, color Color) {
	if d.countingMode || len(points) < 3 {
		return
	}
	d.pdf.SetFillColor(color.R, color.G, color.B)
	d.pdf.Polygon(toGopdfPoints(points), "F")
}

// FillAndStrokePolygon draws a filled polygon with a stroked outline.
// No-op when fewer than 3 points are provided or during the counting pass.
func (d *Document) FillAndStrokePolygon(points []Point, fillColor Color, lineWidth float64, strokeColor Color) {
	if d.countingMode || len(points) < 3 {
		return
	}
	d.pdf.SetFillColor(fillColor.R, fillColor.G, fillColor.B)
	d.pdf.SetLineWidth(lineWidth)
	d.pdf.SetStrokeColor(strokeColor.R, strokeColor.G, strokeColor.B)
	d.pdf.SetLineType("")
	d.pdf.Polygon(toGopdfPoints(points), "DF")
}

// FillCircle draws a filled circle centered at (cx,cy) with the given radius.
// No-op during the counting pass of Build.
func (d *Document) FillCircle(cx, cy, r float64, color Color) {
	if d.countingMode || r <= 0 {
		return
	}
	d.FillPolygon(circlePoints(cx, cy, r, 32), color)
}

// StrokeCircle draws the outline of a circle centered at (cx,cy) with the
// given radius, line width, and stroke color.
// No-op during the counting pass of Build.
func (d *Document) StrokeCircle(cx, cy, r, lineWidth float64, color Color) {
	if d.countingMode || r <= 0 {
		return
	}
	d.pdf.SetLineWidth(lineWidth)
	d.pdf.SetStrokeColor(color.R, color.G, color.B)
	d.pdf.SetLineType("")
	d.pdf.Polygon(toGopdfPoints(circlePoints(cx, cy, r, 32)), "D")
}

// circlePoints returns n points approximating a circle centered at (cx,cy).
func circlePoints(cx, cy, r float64, n int) []Point {
	pts := make([]Point, n)
	for i := range pts {
		a := 2 * math.Pi * float64(i) / float64(n)
		pts[i] = Point{X: cx + r*math.Cos(a), Y: cy + r*math.Sin(a)}
	}
	return pts
}

// DrawImage renders a PNG or JPEG image file into a rectangle at (x,y)
// with the given width and height in points.
// No-op during the counting pass of Build.
func (d *Document) DrawImage(path string, x, y, width, height float64) error {
	if d.countingMode {
		return nil
	}
	return d.pdf.Image(path, x, y, &gopdf.Rect{W: width, H: height})
}

// toGopdfPoints converts []pdf.Point to []gopdf.Point.
func toGopdfPoints(pts []Point) []gopdf.Point {
	out := make([]gopdf.Point, len(pts))
	for i, p := range pts {
		out[i] = gopdf.Point{X: p.X, Y: p.Y}
	}
	return out
}
