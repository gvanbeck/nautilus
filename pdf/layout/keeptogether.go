package layout

import (
	"math"

	"github.com/gvanbeck/nautilus/pdf"
)

// KeepTogether prevents a group of flowables from being split across frames.
//
// When the group does not fit in the remaining space of the current frame, the
// engine inserts a FrameBreak and retries on the next frame.  If the group
// still does not fit (i.e. it is taller than a full frame), the individual
// flowables are returned so each can be split independently.
//
// Example — keep a heading together with the first paragraph of its body:
//
//	story = append(story, &layout.KeepTogether{
//	    Flowables: []layout.Flowable{heading, firstParagraph},
//	})
type KeepTogether struct {
	baseFlowable

	// Flowables is the group of elements to keep on the same frame/page.
	Flowables []Flowable

	// moved is set to true after the engine has already attempted a
	// FrameBreak for this group.  On a second failed attempt we fall back
	// to splitting the individual flowables rather than looping forever.
	moved bool

	// cached measurement from the last Wrap call
	cachedWidth float64
	cachedTotal float64
}

// Wrap measures the total height of all contained flowables.
//
// Returns the actual total height so the engine can decide whether the group
// fits.  If it doesn't fit, the engine will call Split.
func (kt *KeepTogether) Wrap(doc *pdf.Document, availWidth, availHeight float64) (float64, float64) {
	total := kt.measureTotal(doc, availWidth)
	kt.cachedWidth = availWidth
	kt.cachedTotal = total
	return availWidth, total
}

// Draw renders all contained flowables sequentially from (x, y) downward.
func (kt *KeepTogether) Draw(doc *pdf.Document, x, y float64) error {
	curY := y
	for i, f := range kt.Flowables {
		spaceBefore := f.SpaceBefore()
		if i == 0 {
			spaceBefore = 0 // suppress leading space at top of the group
		}
		_, h := f.Wrap(doc, kt.cachedWidth, math.MaxFloat64)
		if err := f.Draw(doc, x, curY+spaceBefore); err != nil {
			return err
		}
		curY += spaceBefore + h + f.SpaceAfter()
	}
	return nil
}

// Split either inserts a FrameBreak (first attempt) or falls back to
// returning the individual flowables (subsequent attempts) to avoid loops.
func (kt *KeepTogether) Split(doc *pdf.Document, availWidth, availHeight float64) []Flowable {
	total := kt.measureTotal(doc, availWidth)

	if total <= availHeight {
		// Fits in remaining space after all — return as-is.
		return []Flowable{kt}
	}

	if kt.moved {
		// We already tried moving to a new frame and it still doesn't fit.
		// Fall back to letting the individual flowables be split normally.
		return kt.Flowables
	}

	// First time: request a frame break so the full frame is available.
	kt.moved = true
	return []Flowable{&FrameBreak{}, kt}
}

// measureTotal sums the heights (including inter-element spacing) of all
// contained flowables using Wrap for measurement.
func (kt *KeepTogether) measureTotal(doc *pdf.Document, availWidth float64) float64 {
	total := 0.0
	for i, f := range kt.Flowables {
		_, h := f.Wrap(doc, availWidth, math.MaxFloat64)
		if i > 0 {
			total += f.SpaceBefore()
		}
		total += h + f.SpaceAfter()
	}
	return total
}
