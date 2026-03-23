// Command celltag demonstrates automatic table-cell generation from struct tags.
//
// Instead of building []pdf.Cell by hand, you annotate a struct with `cell`
// tags that carry both the column style (alignment, background, border, font)
// and — for header rows — the column label.  Two helper functions then turn
// any struct value into a ready-to-use []pdf.Cell slice:
//
//	pdf.HeaderCellsFromStruct(row)  → header cells (text from header= tag)
//	pdf.CellsFromStruct(row)        → data cells   (text from field value)
//
// The example generates a two-table PDF:
//   - An invoice table driven by struct tags
//   - A side-by-side comparison showing the equivalent hand-written code
//
// # Usage
//
//	go run ./examples/celltag \
//	    -font /Library/Fonts/Lato-Medium.ttf \
//	    -bold /Library/Fonts/Lato-Black.ttf \
//	    -out celltag.pdf
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gvanbeck/nautilus/pdf"
)

// ─── Data model ──────────────────────────────────────────────────────────────

// invoiceItem describes one line on an invoice.
// The `cell` tag on each field drives both the header label and the per-column
// style that is applied automatically when building table rows.
type invoiceItem struct {
	// header= sets the column title; halign/valign control text placement.
	Ref      string  `cell:"header=Ref;halign=left"`
	Desc     string  `cell:"header=Description;halign=left"`
	Qty      int     `cell:"header=Qty;halign=center"`
	Unit     float64 `cell:"header=Unit price;halign=right;format=€%.2f"`
	Subtotal float64 `cell:"header=Subtotal;halign=right;format=€%.2f;bold;bg=240,248,255"`
}

// summaryLine is a two-cell summary row (label + amount).
// colspan=3 collapses the first three columns into a single label cell.
type summaryLine struct {
	Label  string  `cell:"colspan=3;halign=right;bold"`
	Amount float64 `cell:"halign=right;bold;format=€%.2f;bg=240,248,255"`
}

// ─── Sample data ─────────────────────────────────────────────────────────────

var items = []invoiceItem{
	{"A-001", "Mechanical keyboard, TKL, Cherry MX Brown", 2, 89.95, 179.90},
	{"A-002", "27\" 4K IPS monitor, 144 Hz, USB-C", 1, 549.00, 549.00},
	{"B-010", "USB-C to DisplayPort cable, 2 m", 3, 14.99, 44.97},
	{"B-011", "Wireless mouse, ergonomic, rechargeable", 5, 34.50, 172.50},
	{"C-003", "Laptop stand, aluminium, adjustable height", 2, 49.00, 98.00},
	{"C-007", "Desk cable management tray, black", 4, 18.75, 75.00},
}

var summary = []summaryLine{
	{"Subtotal (excl. VAT)", 1119.37},
	{"VAT 21%", 235.07},
	{"Total (incl. VAT)", 1354.44},
}

// ─── Main ─────────────────────────────────────────────────────────────────────

func main() {
	fontPath := flag.String("font", "/Library/Fonts/Lato-Medium.ttf", "path to regular TTF font")
	boldPath := flag.String("bold", "/Library/Fonts/Lato-Black.ttf", "path to bold TTF font (optional)")
	outPath := flag.String("out", "celltag.pdf", "output PDF path")
	flag.Parse()

	// ── Document ──────────────────────────────────────────────────────────
	doc, err := pdf.New(pdf.Config{
		PageSize:         pdf.PageSizeA4,
		DefaultFontSize:  11,
		LineHeightFactor: 1.4,
	})
	if err != nil {
		log.Fatalf("create document: %v", err)
	}

	// ── Fonts ─────────────────────────────────────────────────────────────
	if _, err := os.Stat(*fontPath); err != nil {
		log.Fatalf("font not found at %q — use -font to specify a TTF file", *fontPath)
	}
	if err := doc.RegisterFont("regular", *fontPath); err != nil {
		log.Fatalf("register regular font: %v", err)
	}

	boldFont := "regular" // fallback when no bold font is provided
	if _, err := os.Stat(*boldPath); err == nil {
		if err := doc.RegisterFont("regularBold", *boldPath); err != nil {
			log.Printf("warning: bold font: %v (continuing without bold)", err)
		} else {
			boldFont = "regularBold"
		}
	}
	_ = boldFont // used via tag bold= → "regularBold" suffix

	// ── Layout constants ──────────────────────────────────────────────────
	const (
		marginLeft  = 55.0
		marginRight = 55.0
	)
	pageW    := doc.PageWidth()
	contentW := pageW - marginLeft - marginRight

	// ── Page ──────────────────────────────────────────────────────────────
	doc.AddPage()

	y := 50.0

	// Title
	if err := doc.SetFont("regular", 18); err != nil {
		log.Fatal(err)
	}
	doc.SetTextColor(20, 20, 80)
	doc.WriteLine("Invoice — struct-tag cell generation", marginLeft, y) //nolint
	y += 28

	// Subtitle
	if err := doc.SetFont("regular", 10); err != nil {
		log.Fatal(err)
	}
	doc.SetTextColor(100, 100, 100)
	doc.WriteLine(
		"Header row and data rows are built with pdf.HeaderCellsFromStruct / pdf.CellsFromStruct.",
		marginLeft, y,
	) //nolint
	y += 22

	// ── Table ─────────────────────────────────────────────────────────────
	// Column widths must match the number of exported fields in invoiceItem.
	colWidths := []float64{55, 220, 40, 80, 80}

	cellBorder := pdf.NewUniformBorder(pdf.BorderSpec{
		Thickness: 0.4,
		Color:     pdf.ColorLightGray,
	})
	navy  := pdf.ColorNavy
	white := pdf.ColorWhite

	tbl := doc.NewTable(pdf.TableConfig{
		X:         marginLeft,
		Y:         y,
		ColWidths: colWidths,
		Border: pdf.NewUniformBorder(pdf.BorderSpec{
			Thickness: 1.2,
			Color:     pdf.ColorNavy,
		}),
		DefaultCellStyle: pdf.CellStyle{
			Padding:  pdf.Padding{Top: 5, Right: 7, Bottom: 5, Left: 7},
			Border:   cellBorder,
			FontName: "regular",
			FontSize: 10,
		},
		PageBottom:    doc.PageHeight() - 55,
		ContinuationY: 55,
	})

	// ── Header row (one call — no manual Cell literals needed) ────────────
	headerCells, err := pdf.HeaderCellsFromStruct(invoiceItem{})
	if err != nil {
		log.Fatalf("header cells: %v", err)
	}
	// Apply a shared header style on top of each cell's tag-driven style.
	for i := range headerCells {
		headerCells[i].Style.Background = &navy
		headerCells[i].Style.TextColor = &white
		headerCells[i].Style.FontName = "regularBold"
	}
	tbl.AddRow(pdf.Row{Height: 22, Cells: headerCells})

	// ── Data rows ─────────────────────────────────────────────────────────
	lightBlue := pdf.Color{R: 245, G: 248, B: 255}
	for i, item := range items {
		cells, err := pdf.CellsFromStruct(item)
		if err != nil {
			log.Fatalf("row %d cells: %v", i, err)
		}
		var bg *pdf.Color
		if i%2 == 0 {
			bg = &lightBlue
		}
		tbl.AddRow(pdf.Row{Background: bg, Cells: cells})
	}

	// ── Summary rows (different struct, same helper) ───────────────────────
	sepSpec := &pdf.BorderSpec{Thickness: 1, Color: pdf.ColorNavy}
	for j, s := range summary {
		cells, err := pdf.CellsFromStruct(s)
		if err != nil {
			log.Fatalf("summary row %d: %v", j, err)
		}
		// Draw a separator above the first summary row.
		if j == 0 {
			for i := range cells {
				cells[i].Style.Border.Top = sepSpec
			}
		}
		tbl.AddRow(pdf.Row{Cells: cells})
	}

	if err := tbl.Draw(); err != nil {
		log.Fatalf("draw table: %v", err)
	}

	// ── Annotation block below the table ──────────────────────────────────
	y = tbl.CurrentY() + 14

	if err := doc.SetFont("regular", 9); err != nil {
		log.Fatal(err)
	}
	doc.SetTextColor(80, 80, 80)

	annotation := "The table above was built with:\n\n" +
		"  headerCells, _ := pdf.HeaderCellsFromStruct(invoiceItem{})\n" +
		"  for _, item := range items {\n" +
		"      cells, _ := pdf.CellsFromStruct(item)\n" +
		"      tbl.AddRow(pdf.Row{Cells: cells})\n" +
		"  }\n\n" +
		"All alignment, background, format, and border options are read from\n" +
		"the `cell` struct tag on each field.  No manual pdf.Cell{} literals needed."

	bgColor := pdf.Color{R: 245, G: 245, B: 235}
	frame := doc.NewFrame(pdf.FrameConfig{
		X:          marginLeft,
		Y:          y,
		Width:      contentW,
		Background: &bgColor,
		Border: pdf.Border{
			Left: &pdf.BorderSpec{Thickness: 3, Color: pdf.ColorNavy},
		},
		Padding: pdf.Padding{Top: 10, Right: 14, Bottom: 10, Left: 14},
	})
	frame.SetFont("regular", 9) //nolint
	frame.SetTextColor(50, 50, 50)
	frame.WriteText(annotation) //nolint
	frame.Close()               //nolint

	// ── Save ──────────────────────────────────────────────────────────────
	if err := doc.Save(*outPath); err != nil {
		log.Fatalf("save: %v", err)
	}
	fmt.Printf("written %s (%d page(s))\n", *outPath, doc.PageCount())
}
