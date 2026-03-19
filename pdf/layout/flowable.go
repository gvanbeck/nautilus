// Package layout implements a high-level document layout engine inspired by
// the Platypus system from the Python ReportLab library.
//
// The central concept is the Flowable: any piece of content (paragraph, image,
// table, spacer, …) that can be measured and drawn.  A Story is simply a
// slice of Flowables.  The DocTemplate engine consumes a story and flows its
// contents across one or more LayoutFrames on each page, adding new pages
// automatically when a frame is full.
//
// # Minimal example
//
//	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4, Margins: pdf.UniformMargins(50)})
//	doc.RegisterFont("regular", "NotoSans-Regular.ttf")
//	doc.SetFont("regular", 12)
//
//	style := layout.ParagraphStyle{FontName: "regular", FontSize: 12}
//	story := []layout.Flowable{
//	    &layout.Paragraph{Text: "Hello, Nautilus!", Style: style},
//	    &layout.Spacer{Height: 12},
//	    &layout.Paragraph{Text: "Second paragraph.", Style: style},
//	}
//
//	frame := &layout.LayoutFrame{
//	    X: doc.ContentX(), Y: doc.ContentY(),
//	    Width: doc.ContentWidth(), Height: doc.ContentHeight(),
//	}
//	tmpl := &layout.PageTemplate{ID: "main", Frames: []*layout.LayoutFrame{frame}}
//	dt := layout.NewDocTemplate(doc)
//	dt.AddPageTemplate(tmpl)
//	dt.Build(story)
//	doc.Save("output.pdf")
package layout

import "github.com/gvanbeck/nautilus/pdf"

// Flowable is the core interface for any element that can be placed on a page.
//
// The engine always calls Wrap before Draw or Split on the same flowable.
// Implementations must be idempotent with respect to repeated Wrap calls.
type Flowable interface {
	// Wrap measures the flowable within the given available space.
	// It returns the actual (width, height) the flowable will occupy.
	// The height returned must not exceed availHeight for the flowable
	// to be accepted by a LayoutFrame; return a larger value to signal
	// that the flowable does not fit and should be split or moved.
	Wrap(doc *pdf.Document, availWidth, availHeight float64) (float64, float64)

	// Draw renders the flowable with its top-left corner at (x, y).
	// It is only called after a successful Wrap.
	Draw(doc *pdf.Document, x, y float64) error

	// Split attempts to divide the flowable so that the first returned
	// part fits within availHeight.  Returns nil when splitting is not
	// possible; the engine will then move the flowable to the next frame.
	// The returned parts must together reproduce all original content.
	Split(doc *pdf.Document, availWidth, availHeight float64) []Flowable

	// SpaceBefore returns the extra whitespace to add above this flowable.
	SpaceBefore() float64

	// SpaceAfter returns the extra whitespace to add below this flowable.
	SpaceAfter() float64

	// KeepWithNext signals that a frame/page break must not be inserted
	// between this flowable and the one that follows it in the story.
	KeepWithNext() bool

	// MinWidth returns the minimum width required by this flowable.
	MinWidth() float64
}

// ActionFlowable is a zero-height flowable that controls the document engine
// rather than rendering visible content (e.g. PageBreak, FrameBreak).
type ActionFlowable interface {
	Flowable
	apply(doc *DocTemplate)
}

// baseFlowable provides default no-op implementations for the optional parts
// of the Flowable interface.  Embed it in concrete types and override only
// the methods you need.
type baseFlowable struct {
	spaceBefore  float64
	spaceAfter   float64
	keepWithNext bool
}

func (b *baseFlowable) SpaceBefore() float64                                   { return b.spaceBefore }
func (b *baseFlowable) SpaceAfter() float64                                    { return b.spaceAfter }
func (b *baseFlowable) KeepWithNext() bool                                     { return b.keepWithNext }
func (b *baseFlowable) MinWidth() float64                                      { return 0 }
func (b *baseFlowable) Split(_ *pdf.Document, _, _ float64) []Flowable        { return nil }
