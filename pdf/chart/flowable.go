package chart

import (
	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/layout"
)

// NewFlowable wraps a Drawable chart as a layout.Flowable so that it can be
// placed in a DocTemplate story.
//
// width is the desired chart width in points; pass 0 to fill the available
// frame width.  height is the fixed chart height in points.
//
//	story := []layout.Flowable{
//	    chart.NewFlowable(myPieChart, 0, 220),
//	    &layout.Spacer{Height: 12},
//	}
func NewFlowable(c Drawable, width, height float64) layout.Flowable {
	return &chartFlowable{chart: c, width: width, height: height}
}

type chartFlowable struct {
	chart         Drawable
	width, height float64
	ww            float64 // resolved width, set by Wrap
}

func (f *chartFlowable) Wrap(_ *pdf.Document, availWidth, _ float64) (float64, float64) {
	w := f.width
	if w <= 0 || w > availWidth {
		w = availWidth
	}
	f.ww = w
	return w, f.height
}

func (f *chartFlowable) Draw(doc *pdf.Document, x, y float64) error {
	return f.chart.Draw(doc, x, y, f.ww, f.height)
}

func (f *chartFlowable) Split(_ *pdf.Document, _, _ float64) []layout.Flowable { return nil }
func (f *chartFlowable) SpaceBefore() float64                                  { return 0 }
func (f *chartFlowable) SpaceAfter() float64                                   { return 0 }
func (f *chartFlowable) KeepWithNext() bool                                    { return false }
func (f *chartFlowable) MinWidth() float64                                     { return 0 }
