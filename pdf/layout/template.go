package layout

import (
	"fmt"

	"github.com/gvanbeck/nautilus/pdf"
)

// PageDecorator is a callback invoked by the engine at the start or end of a
// page.  Use it to draw headers, footers, watermarks, or page numbers.
//
// pageNum is 1-based.
type PageDecorator func(doc *pdf.Document, pageNum int)

// PageTemplate defines the layout of one or more pages as an ordered list of
// LayoutFrames.  The engine fills frames sequentially; when all frames on a
// page are full it starts a new page using the same (or the next) template.
type PageTemplate struct {
	// ID is the name of this template, used by NextPageTemplate.
	ID string

	// Frames is the ordered list of layout regions on the page.
	// At least one frame is required.
	Frames []*LayoutFrame

	// OnPage, when set, is called immediately after AddPage so that the caller
	// can render headers, watermarks, or static decorations.
	OnPage PageDecorator

	// OnPageEnd, when set, is called just before the page is finalised.
	OnPageEnd PageDecorator

	// AutoNextTemplate is the ID of the template to switch to automatically
	// after this page ends.  Leave empty to keep the same template.
	AutoNextTemplate string
}

// DocTemplate is the document engine that processes a story ([]Flowable) and
// flows its contents across frames and pages, adding pages automatically.
//
// Create with NewDocTemplate, register at least one PageTemplate via
// AddPageTemplate, then call Build.
type DocTemplate struct {
	doc           *pdf.Document
	pageTemplates []*PageTemplate

	// engine state during Build
	pageNum         int
	templateIdx     int
	frameIdx        int
	pendingTemplate string // deferred switch from NextPageTemplate action
}

// NewDocTemplate creates a DocTemplate that renders into doc.
func NewDocTemplate(doc *pdf.Document) *DocTemplate {
	return &DocTemplate{doc: doc}
}

// AddPageTemplate registers a page template with the engine.
// Templates are identified by their ID field.
// The first registered template is used for the first page.
func (dt *DocTemplate) AddPageTemplate(pt *PageTemplate) {
	dt.pageTemplates = append(dt.pageTemplates, pt)
}

// Build processes story and generates the PDF content.
//
// It flows flowables sequentially across frames and pages.  When a frame is
// full the engine advances to the next frame; when all frames on a page are
// exhausted a new page is started using the current (or next) template.
//
// ActionFlowables embedded in story (PageBreak, FrameBreak, …) are executed
// immediately as they are encountered.
//
// Build returns an error if no page templates are registered or if a flowable
// cannot be placed in any frame after repeated attempts.
func (dt *DocTemplate) Build(story []Flowable) error {
	if len(dt.pageTemplates) == 0 {
		return fmt.Errorf("layout: no page templates registered")
	}

	// Copy story into a mutable queue.
	queue := make([]Flowable, len(story))
	copy(queue, story)

	// Start the first page.
	if err := dt.startPage(); err != nil {
		return err
	}

	noProgress := 0 // consecutive attempts without placing a flowable

	for len(queue) > 0 {
		f := queue[0]
		queue = queue[1:]

		// ActionFlowables are executed immediately without taking space.
		if af, ok := f.(ActionFlowable); ok {
			af.apply(dt)
			noProgress = 0
			continue
		}

		// Collect keepWithNext chains and wrap them in KeepTogether.
		if f.KeepWithNext() {
			group := []Flowable{f}
			for len(queue) > 0 && group[len(group)-1].KeepWithNext() {
				group = append(group, queue[0])
				queue = queue[1:]
			}
			if len(group) > 1 {
				f = &KeepTogether{Flowables: group}
			}
		}

		frame := dt.currentFrame()

		// Attempt to place the flowable without splitting.
		if frame.add(f, dt.doc) {
			noProgress = 0
			continue
		}

		// Try to split the flowable.
		parts := frame.split(f, dt.doc)
		if len(parts) > 0 {
			// Re-queue all parts (including first) so ActionFlowables
			// among them are processed correctly by the main loop.
			queue = append(parts, queue...)
			noProgress = 0
			continue
		}

		// Nothing fits and nothing can be split — advance frame/page.
		noProgress++
		if noProgress > 10 {
			return fmt.Errorf("layout: flowable %T cannot fit in any frame", f)
		}
		// Re-queue the flowable for the next frame.
		queue = append([]Flowable{f}, queue...)
		if err := dt.advanceFrame(); err != nil {
			return err
		}
	}

	dt.endPage()
	return nil
}

// ── internal page / frame management ────────────────────────────────────────

func (dt *DocTemplate) currentTemplate() *PageTemplate {
	return dt.pageTemplates[dt.templateIdx]
}

func (dt *DocTemplate) currentFrame() *LayoutFrame {
	return dt.currentTemplate().Frames[dt.frameIdx]
}

func (dt *DocTemplate) startPage() error {
	dt.pageNum++
	dt.doc.AddPage()

	tmpl := dt.currentTemplate()
	for _, fr := range tmpl.Frames {
		fr.reset()
		if fr.ShowBoundary {
			if err := fr.drawBoundary(dt.doc); err != nil {
				return err
			}
		}
	}
	dt.frameIdx = 0

	if tmpl.OnPage != nil {
		tmpl.OnPage(dt.doc, dt.pageNum)
	}
	return nil
}

func (dt *DocTemplate) endPage() {
	tmpl := dt.currentTemplate()
	if tmpl.OnPageEnd != nil {
		tmpl.OnPageEnd(dt.doc, dt.pageNum)
	}

	// Apply any pending template switch.
	next := dt.pendingTemplate
	if next == "" {
		next = tmpl.AutoNextTemplate
	}
	if next != "" {
		dt.applyTemplateSwitch(next)
	}
	dt.pendingTemplate = ""
}

func (dt *DocTemplate) advanceFrame() error {
	tmpl := dt.currentTemplate()
	dt.frameIdx++
	if dt.frameIdx < len(tmpl.Frames) {
		return nil // more frames on this page
	}
	// All frames exhausted — start a new page.
	dt.endPage()
	return dt.startPage()
}

func (dt *DocTemplate) applyTemplateSwitch(id string) {
	for i, t := range dt.pageTemplates {
		if t.ID == id {
			dt.templateIdx = i
			return
		}
	}
}

// ── methods called by ActionFlowables ───────────────────────────────────────

// forcePageBreak immediately starts a new page, optionally switching template.
func (dt *DocTemplate) forcePageBreak(nextTemplateID string) error {
	dt.endPage()
	if nextTemplateID != "" {
		dt.applyTemplateSwitch(nextTemplateID)
		dt.pendingTemplate = ""
	}
	return dt.startPage()
}

// forceFrameBreak advances to the next frame (or next page).
func (dt *DocTemplate) forceFrameBreak() error {
	return dt.advanceFrame()
}

// scheduleTemplateSwitch defers a template switch until the next page break.
func (dt *DocTemplate) scheduleTemplateSwitch(id string) {
	dt.pendingTemplate = id
}

// remainingFrameHeight returns the available height left in the current frame.
func (dt *DocTemplate) remainingFrameHeight() float64 {
	return dt.currentFrame().availHeight()
}
