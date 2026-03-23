package rml

// graphics.go defines the types for <pageGraphics> drawing commands and
// the functions that execute them on a pdf.Document.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gvanbeck/nautilus/pdf"
)

// ─── AST types ────────────────────────────────────────────────────────────────

// graphicsBlock holds all commands from a <pageGraphics> element.
type graphicsBlock struct {
	commands []gfxCmd
}

// gfxCmd is one drawing command inside <pageGraphics>.
type gfxCmd interface{ isGfxCmd() }

type gfxSetFont      struct{ name, size string }
type gfxSetFill      struct{ color string }
type gfxSetStroke    struct{ color, width string }
type gfxSaveState    struct{}
type gfxRestoreState struct{}
type gfxDrawString   struct{ x, y, text string }
type gfxDrawRString  struct{ x, y, text string } // right-aligned
type gfxDrawCString  struct{ x, y, text string } // centred
type gfxRect         struct{ x, y, w, h, fill, stroke, round string }
type gfxCircle       struct{ x, y, r, fill, stroke string }
type gfxLine         struct{ x1, y1, x2, y2, width, color string }
type gfxLines        struct{ coords string; width, color string }
type gfxPlace        struct{ x, y, w, h string; inner []gfxCmd }

func (*gfxSetFont) isGfxCmd()      {}
func (*gfxSetFill) isGfxCmd()      {}
func (*gfxSetStroke) isGfxCmd()    {}
func (*gfxSaveState) isGfxCmd()    {}
func (*gfxRestoreState) isGfxCmd() {}
func (*gfxDrawString) isGfxCmd()   {}
func (*gfxDrawRString) isGfxCmd()  {}
func (*gfxDrawCString) isGfxCmd()  {}
func (*gfxRect) isGfxCmd()         {}
func (*gfxCircle) isGfxCmd()       {}
func (*gfxLine) isGfxCmd()         {}
func (*gfxLines) isGfxCmd()        {}
func (*gfxPlace) isGfxCmd()        {}

// ─── Parser ──────────────────────────────────────────────────────────────────

// parseGraphicsCommands reads child elements from a decoder until endTag is
// reached and returns them as []gfxCmd.  Used for <pageGraphics> and <place>.
func parseGraphicsCommands(d interface {
	Token() (interface{}, error)
}, endTag string) ([]gfxCmd, error) {
	// We use a real *xml.Decoder — import it via the parent parse.go helpers.
	return nil, nil // placeholder: actual implementation is inlined in parse.go
}

// ─── Executor ────────────────────────────────────────────────────────────────

// execGraphics runs all commands in a graphicsBlock on doc.
// pageH is used to convert RML bottom-left y coordinates to top-left.
func execGraphics(doc *pdf.Document, block graphicsBlock, pageH float64) {
	execGraphicsWithPage(doc, block, pageH, 0)
}

// execGraphicsWithPage runs all commands with page number available for %p/%P substitution.
func execGraphicsWithPage(doc *pdf.Document, block graphicsBlock, pageH float64, pageNum int) {
	for _, cmd := range block.commands {
		execCmdWithPage(doc, cmd, pageH, pageNum)
	}
}

func execCmdWithPage(doc *pdf.Document, cmd gfxCmd, pageH float64, pageNum int) {
	switch c := cmd.(type) {
	case *gfxSaveState:
		doc.SaveGraphicsState()

	case *gfxRestoreState:
		doc.RestoreGraphicsState()

	case *gfxSetFont:
		size := ptDef(c.size, 10)
		if c.name != "" {
			doc.SetFont(c.name, size) //nolint:errcheck
		}

	case *gfxSetFill:
		if col, ok := parseColor(c.color); ok {
			doc.SetTextColor(col.R, col.G, col.B)
		}

	case *gfxSetStroke:
		// Stroke color is implicitly set via DrawLine / DrawBorder.
		// Nothing extra needed here for now.

	case *gfxDrawString:
		x := ptDef(c.x, 0)
		y := rmlY(ptDef(c.y, 0), pageH)
		text := substitutePageVars(c.text, pageNum, 0)
		doc.WriteLine(text, x, y) //nolint:errcheck

	case *gfxDrawRString:
		x := ptDef(c.x, 0)
		y := rmlY(ptDef(c.y, 0), pageH)
		text := substitutePageVars(c.text, pageNum, 0)
		w, _ := doc.MeasureText(text)
		doc.WriteLine(text, x-w, y) //nolint:errcheck

	case *gfxDrawCString:
		x := ptDef(c.x, 0)
		y := rmlY(ptDef(c.y, 0), pageH)
		text := substitutePageVars(c.text, pageNum, 0)
		w, _ := doc.MeasureText(text)
		doc.WriteLine(text, x-w/2, y) //nolint:errcheck

	case *gfxRect:
		x := ptDef(c.x, 0)
		y := rmlY(ptDef(c.y, 0)+ptDef(c.h, 0), pageH) // convert bottom-left to top-left
		w := ptDef(c.w, 0)
		h := ptDef(c.h, 0)
		if c.fill != "" {
			if col, ok := parseColor(c.fill); ok {
				doc.FillRect(x, y, w, h, col)
			}
		}
		if c.stroke != "" {
			if col, ok := parseColor(c.stroke); ok {
				thick := 1.0
				doc.DrawBorder(x, y, w, h, pdf.NewUniformBorder(pdf.BorderSpec{ //nolint:errcheck
					Thickness: thick,
					Color:     col,
				}))
			}
		}

	case *gfxCircle:
		cx := ptDef(c.x, 0)
		cy := rmlY(ptDef(c.y, 0), pageH)
		r := ptDef(c.r, 0)
		if c.fill != "" {
			if col, ok := parseColor(c.fill); ok {
				doc.FillCircle(cx, cy, r, col)
			}
		}
		if c.stroke != "" {
			if col, ok := parseColor(c.stroke); ok {
				doc.StrokeCircle(cx, cy, r, 1, col)
			}
		}

	case *gfxLine:
		x1 := ptDef(c.x1, 0)
		y1 := rmlY(ptDef(c.y1, 0), pageH)
		x2 := ptDef(c.x2, 0)
		y2 := rmlY(ptDef(c.y2, 0), pageH)
		lw := ptDef(c.width, 1)
		col := pdf.ColorBlack
		if cc, ok := parseColor(c.color); ok {
			col = cc
		}
		doc.DrawLine(x1, y1, x2, y2, lw, col)

	case *gfxLines:
		// coords: "x1 y1 x2 y2 x3 y3 ..." pairs of coordinates
		lw := ptDef(c.width, 1)
		col := pdf.ColorBlack
		if cc, ok := parseColor(c.color); ok {
			col = cc
		}
		pts := parseCoordList(c.coords, pageH)
		for i := 0; i+3 < len(pts); i += 4 {
			doc.DrawLine(pts[i], pts[i+1], pts[i+2], pts[i+3], lw, col)
		}
	}
}

// rmlY converts an RML bottom-left y coordinate to a nautilus top-left y.
func rmlY(y, pageH float64) float64 {
	return pageH - y
}

// parseCoordList parses "x1 y1 x2 y2 ..." converting y values.
func parseCoordList(s string, pageH float64) []float64 {
	fields := strings.Fields(s)
	result := make([]float64, 0, len(fields))
	for i, f := range fields {
		v, err := strconv.ParseFloat(f, 64)
		if err != nil {
			continue
		}
		// Odd indices are y coordinates (0-based: 1, 3, 5, …)
		if i%2 == 1 {
			v = pageH - v
		}
		result = append(result, v)
	}
	return result
}

// execGraphicsPageInfo returns a string for use in drawString substitutions.
// %p = page number, %P = total pages.
func substitutePageVars(text string, pageNum, totalPages int) string {
	text = strings.ReplaceAll(text, "%p", fmt.Sprintf("%d", pageNum))
	text = strings.ReplaceAll(text, "%P", fmt.Sprintf("%d", totalPages))
	return text
}
