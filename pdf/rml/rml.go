// Package rml implements a subset of ReportLab's Report Markup Language (RML)
// for the nautilus PDF library.
//
// RML is an XML dialect that separates document structure into three sections:
//
//   - <template>  — page layout: page size, margins, frames
//   - <stylesheet> — named styles: paraStyle, blockTableStyle
//   - <story>      — content: para, spacer, pageBreak, blockTable
//
// # Supported elements
//
//	<docinit>
//	  <registerTTFont fontName="…" fontFile="…"/>
//
//	<template pageSize="A4" leftMargin="55" rightMargin="55"
//	          topMargin="55" bottomMargin="55">
//	  <pageTemplate id="…">
//	    <frame id="…" x1="…" y1="…" width="…" height="…"/>
//	  </pageTemplate>
//
//	<stylesheet>
//	  <paraStyle name="…" fontName="…" fontSize="…" leading="…"
//	             alignment="left|center|right" spaceBefore="…" spaceAfter="…"
//	             textColor="…" leftIndent="…" rightIndent="…"/>
//	  <blockTableStyle id="…">
//	    <lineStyle    kind="GRID|OUTLINE|INNERGRID|BOX"
//	                 colorName="…" thickness="…"
//	                 start="col,row" stop="col,row"/>
//	    <blockBackground  colorName="…" start="col,row" stop="col,row"/>
//	    <blockFont        name="…" size="…" start="col,row" stop="col,row"/>
//	    <blockTextColor   colorName="…" start="col,row" stop="col,row"/>
//	    <blockAlignment   value="left|center|right" start="col,row" stop="col,row"/>
//	    <blockValign      value="top|middle|bottom" start="col,row" stop="col,row"/>
//
//	<story>
//	  <para style="…">text</para>
//	  <spacer length="…"/>
//	  <pageBreak/>
//	  <frameBreak/>
//	  <hr thickness="…" colorName="…"/>
//	  <blockTable colWidths="…" rowHeights="…" style="…">
//	    <tr height="…" bg="…">
//	      <td colspan="…" rowspan="…" bold="1" halign="…" valign="…"
//	          fontName="…" fontSize="…" bg="…" textColor="…">text</td>
//
// # Coordinate system
//
// RML uses PDF coordinates (origin at bottom-left). <frame> x1/y1 are
// automatically converted to nautilus top-left coordinates.
//
// # Units
//
// All measurements accept bare numbers (points), "21cm", "210mm", "8.27in".
//
// # Colors
//
// Colors can be named (black, white, navy, gray, …), hex (#rrggbb), or
// "r,g,b" triples.
//
// # Example
//
//	doc, err := rml.ParseFile("invoice.rml", rml.Options{FontDir: "/fonts"})
//	if err != nil { log.Fatal(err) }
//	if err := doc.Save("invoice.pdf"); err != nil { log.Fatal(err) }
package rml

import (
	"io"
	"os"

	"github.com/gvanbeck/nautilus/pdf"
)

// Options configures the RML renderer.
type Options struct {
	// FontDir is prepended to relative font file paths in <registerTTFont>.
	// Defaults to the current working directory.
	FontDir string
}

// Parse reads an RML document from r and returns a rendered *pdf.Document
// ready to be saved with doc.Save(path).
func Parse(r io.Reader, opts Options) (*pdf.Document, error) {
	doc, err := parseRML(r)
	if err != nil {
		return nil, err
	}
	return render(doc, opts)
}

// ParseFile reads an RML document from the file at path.
func ParseFile(path string, opts Options) (*pdf.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f, opts)
}
