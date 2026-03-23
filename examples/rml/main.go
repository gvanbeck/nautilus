// Command rml demonstrates the nautilus RML parser.
//
// It reads an RML file and renders it to a PDF without writing any Go PDF
// code — all layout and styling is described in the XML document itself.
//
// # Usage
//
//	go run ./examples/rml \
//	    -rml    examples/rml/invoice.rml \
//	    -fontdir /Library/Fonts \
//	    -out    invoice.pdf
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gvanbeck/nautilus/pdf/rml"
)

func main() {
	rmlPath := flag.String("rml", "examples/rml/invoice.rml", "path to .rml file")
	fontDir := flag.String("fontdir", "/Library/Fonts", "directory containing font files")
	outPath := flag.String("out", "invoice.pdf", "output PDF path")
	flag.Parse()

	doc, err := rml.ParseFile(*rmlPath, rml.Options{FontDir: *fontDir})
	if err != nil {
		log.Fatalf("parse rml: %v", err)
	}

	if err := doc.Save(*outPath); err != nil {
		log.Fatalf("save pdf: %v", err)
	}

	fmt.Printf("written %s (%d page(s))\n", *outPath, doc.PageCount())
}
