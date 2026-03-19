// Package layout implements a Platypus-inspired high-level document layout
// engine on top of the Nautilus pdf.Document drawing layer.
//
// # Architecture
//
// The system is organised in four layers:
//
//	Story ([]Flowable)
//	    ↓ consumed by
//	DocTemplate  (page/frame scheduler)
//	    ↓ manages
//	PageTemplate  (page geometry: ordered LayoutFrames + decorators)
//	    ↓ contains
//	LayoutFrame  (rectangular region with a downward-moving Y cursor)
//	    ↓ draws into
//	pdf.Document  (the underlying PDF canvas)
//
// # Flowable interface
//
// Every piece of content implements Flowable:
//
//	type Flowable interface {
//	    Wrap(doc, availWidth, availHeight) (width, height)
//	    Draw(doc, x, y) error
//	    Split(doc, availWidth, availHeight) []Flowable
//	    SpaceBefore() float64
//	    SpaceAfter()  float64
//	    KeepWithNext() bool
//	    MinWidth() float64
//	}
//
// The engine always calls Wrap before Draw or Split on the same flowable.
//
// # Built-in flowables
//
//   - [Paragraph] — word-wrapped text with style (font, size, colour, alignment)
//   - [Spacer]    — reserves vertical space without drawing
//   - [HRFlowable] — solid horizontal rule bar
//   - [KeepTogether] — prevents splitting a group across frames
//
// # Action flowables (zero-height control signals)
//
//   - [PageBreak]       — force a page break, optionally switch template
//   - [FrameBreak]      — advance to the next frame
//   - [CondPageBreak]   — page break only when < MinHeight points remain
//   - [NextPageTemplate] — deferred template switch on the next page break
//
// # Minimal usage
//
//	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
//	doc.RegisterFont("regular", "NotoSans-Regular.ttf")
//	doc.SetFont("regular", 11)
//
//	style := layout.ParagraphStyle{FontName: "regular", FontSize: 11}
//	story := []layout.Flowable{
//	    &layout.Paragraph{Text: "Hello, layout engine!", Style: style},
//	    &layout.Spacer{Height: 12},
//	    &layout.Paragraph{Text: "Second paragraph.", Style: style},
//	}
//
//	frame := &layout.LayoutFrame{
//	    X: 50, Y: 50,
//	    Width:  doc.ContentWidth(),
//	    Height: doc.ContentHeight(),
//	}
//	dt := layout.NewDocTemplate(doc)
//	dt.AddPageTemplate(&layout.PageTemplate{ID: "main", Frames: []*layout.LayoutFrame{frame}})
//	dt.Build(story)
//	doc.Save("output.pdf")
package layout
