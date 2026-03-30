package layout

import "github.com/gvanbeck/nautilus/pdf"

// LayoutFrame is a rectangular region on a page that receives Flowables.
//
// The frame maintains an internal Y cursor that advances downward as content
// is added.  When the cursor reaches the bottom of the usable area the frame
// is considered full; the DocTemplate then activates the next frame or starts
// a new page.
//
// Coordinate origin is the top-left corner of the page (Y increases downward),
// consistent with the rest of the Nautilus library.
type LayoutFrame struct {
	// X and Y are the top-left corner of the frame in page coordinates (points).
	X, Y float64

	// Width and Height are the outer dimensions of the frame in points.
	Width, Height float64

	// Padding reduces the usable interior of the frame.
	Padding pdf.Padding

	// ID is an optional name used for debugging and NextFrameFlowable.
	ID string

	// ShowBoundary, when true, draws a thin rectangle around the frame
	// for debugging.  Drawn by the DocTemplate at page-start time.
	ShowBoundary bool

	// internal cursor state — reset at the start of each page
	curY  float64 // absolute page Y of the next content line
	atTop bool    // true until the first flowable is successfully placed
}

// reset prepares the frame to receive content from the top of its area.
// Called by the DocTemplate at the beginning of each page.
func (f *LayoutFrame) reset() {
	f.curY = f.Y + f.Padding.Top
	f.atTop = true
}

// innerWidth returns the usable width after subtracting horizontal padding.
func (f *LayoutFrame) innerWidth() float64 {
	return f.Width - f.Padding.Left - f.Padding.Right
}

// contentX returns the absolute page X of the left edge of the content area.
func (f *LayoutFrame) contentX() float64 {
	return f.X + f.Padding.Left
}

// bottomY returns the absolute page Y of the bottom of the usable content area.
func (f *LayoutFrame) bottomY() float64 {
	return f.Y + f.Height - f.Padding.Bottom
}

// availHeight returns the remaining vertical space in the frame.
func (f *LayoutFrame) availHeight() float64 {
	return f.bottomY() - f.curY
}

// add attempts to place flowable in this frame without splitting it.
//
// It calls Wrap to measure the flowable.  If the flowable fits within the
// remaining space (accounting for spaceBefore), it calls Draw and advances
// the cursor.  Returns true on success.
func (f *LayoutFrame) add(flowable Flowable, doc *pdf.Document) (bool, error) {
	avW := f.innerWidth()

	spaceBefore := flowable.SpaceBefore()
	if f.atTop {
		spaceBefore = 0 // suppress leading space at the top of a frame
	}

	avH := f.availHeight() - spaceBefore
	if avH < 0 {
		return false, nil
	}

	_, h := flowable.Wrap(doc, avW, avH)
	if h > avH {
		return false, nil
	}

	drawY := f.curY + spaceBefore
	if err := flowable.Draw(doc, f.contentX(), drawY); err != nil {
		return false, err
	}

	f.curY = drawY + h + flowable.SpaceAfter()
	f.atTop = false
	return true, nil
}

// split asks flowable to split itself to fit within the frame's remaining space.
// The spaceBefore of the flowable is subtracted from available height before
// delegating to flowable.Split, mirroring the logic in add.
func (f *LayoutFrame) split(flowable Flowable, doc *pdf.Document) []Flowable {
	avW := f.innerWidth()
	spaceBefore := flowable.SpaceBefore()
	if f.atTop {
		spaceBefore = 0
	}
	avH := f.availHeight() - spaceBefore
	if avH < 0 {
		return nil
	}
	return flowable.Split(doc, avW, avH)
}

// drawBoundary draws a thin rectangle around the frame outline.
// Called when ShowBoundary is true.
func (f *LayoutFrame) drawBoundary(doc *pdf.Document) error {
	spec := pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}
	return doc.DrawBorder(f.X, f.Y, f.Width, f.Height, pdf.NewUniformBorder(spec))
}
