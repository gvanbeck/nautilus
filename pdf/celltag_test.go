package pdf

import (
	"testing"
)

type invoiceRow struct {
	Item     string  `cell:"halign=left;header=Omschrijving"`
	Qty      int     `cell:"halign=center;header=Aantal"`
	Price    float64 `cell:"halign=right;header=Prijs;format=%.2f"`
	Total    float64 `cell:"halign=right;bold;bg=240,240,240;border=bottom;header=Totaal;format=%.2f"`
	internal string  // unexported: skipped automatically
	Secret   string  `cell:"-"` // explicitly skipped
}

func TestCellsFromStruct_data(t *testing.T) {
	row := invoiceRow{Item: "Widget A", Qty: 3, Price: 9.99, Total: 29.97}
	cells, err := CellsFromStruct(row)
	if err != nil {
		t.Fatal(err)
	}
	if len(cells) != 4 {
		t.Fatalf("want 4 cells, got %d", len(cells))
	}

	if cells[0].Text != "Widget A" {
		t.Errorf("cell[0] text = %q", cells[0].Text)
	}
	if cells[0].Style.HAlign != HAlignLeft {
		t.Errorf("cell[0] halign = %v", cells[0].Style.HAlign)
	}
	if cells[1].Text != "3" {
		t.Errorf("cell[1] text = %q", cells[1].Text)
	}
	if cells[2].Text != "9.99" {
		t.Errorf("cell[2] text = %q, want 9.99", cells[2].Text)
	}
	if cells[3].Text != "29.97" {
		t.Errorf("cell[3] text = %q, want 29.97", cells[3].Text)
	}
	if cells[3].Style.Background == nil {
		t.Error("cell[3] background should be set")
	} else if *cells[3].Style.Background != (Color{240, 240, 240}) {
		t.Errorf("cell[3] bg = %v", *cells[3].Style.Background)
	}
	if cells[3].Style.Border.Bottom == nil {
		t.Error("cell[3] bottom border should be set")
	}
}

func TestHeaderCellsFromStruct(t *testing.T) {
	cells, err := HeaderCellsFromStruct(invoiceRow{})
	if err != nil {
		t.Fatal(err)
	}
	if len(cells) != 4 {
		t.Fatalf("want 4 header cells, got %d", len(cells))
	}

	want := []string{"Omschrijving", "Aantal", "Prijs", "Totaal"}
	for i, w := range want {
		if cells[i].Text != w {
			t.Errorf("header[%d] = %q, want %q", i, cells[i].Text, w)
		}
	}
}

func TestCellsFromStruct_pointer(t *testing.T) {
	row := &invoiceRow{Item: "Test"}
	cells, err := CellsFromStruct(row)
	if err != nil {
		t.Fatal(err)
	}
	if len(cells) != 4 {
		t.Fatalf("want 4 cells, got %d", len(cells))
	}
}

func TestCellsFromStruct_colspan(t *testing.T) {
	type spanRow struct {
		Label string `cell:"colspan=2;text=Merged"`
		Value string `cell:""`
	}
	cells, err := CellsFromStruct(spanRow{Label: "ignored", Value: "v"})
	if err != nil {
		t.Fatal(err)
	}
	if cells[0].ColSpan != 2 {
		t.Errorf("colspan = %d, want 2", cells[0].ColSpan)
	}
	if cells[0].Text != "Merged" {
		t.Errorf("text = %q, want Merged", cells[0].Text)
	}
}
