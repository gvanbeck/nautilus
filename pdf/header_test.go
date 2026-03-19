package pdf_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
)

// TestSetHeader_calledOnEveryPage verifies that the header function is invoked
// once for each AddPage call.
func TestSetHeader_calledOnEveryPage(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}

	calls := 0
	doc.SetHeader(func(d *pdf.Document, info pdf.PageInfo) {
		calls++
	})

	for range 3 {
		doc.AddPage()
	}

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}

	if calls != 3 {
		t.Errorf("header called %d times, want 3", calls)
	}
}

// TestSetFooter_calledOnEveryPage verifies that the footer function is invoked
// once per page.
func TestSetFooter_calledOnEveryPage(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}

	calls := 0
	doc.SetFooter(func(d *pdf.Document, info pdf.PageInfo) {
		calls++
	})

	for range 4 {
		doc.AddPage()
	}

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}

	if calls != 4 {
		t.Errorf("footer called %d times, want 4", calls)
	}
}

// TestSetHeader_pageNumbers verifies that PageInfo.Number increments correctly
// and that PageInfo.Total matches SetTotalPages.
func TestSetHeader_pageNumbers(t *testing.T) {
	fontPath := systemFont(t)

	const total = 5
	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}

	doc.SetTotalPages(total)

	var numbers []int
	var totals []int
	doc.SetHeader(func(d *pdf.Document, info pdf.PageInfo) {
		numbers = append(numbers, info.Number)
		totals = append(totals, info.Total)
	})

	for range total {
		doc.AddPage()
	}

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}

	for i, n := range numbers {
		want := i + 1
		if n != want {
			t.Errorf("header[%d].Number = %d, want %d", i, n, want)
		}
		if totals[i] != total {
			t.Errorf("header[%d].Total = %d, want %d", i, totals[i], total)
		}
	}
}

// TestBuild_totalPagesKnown verifies that Build correctly determines the total
// page count and passes it to header/footer functions.
func TestBuild_totalPagesKnown(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}

	const pageCount = 3
	var headerInfos []pdf.PageInfo

	doc.SetHeader(func(d *pdf.Document, info pdf.PageInfo) {
		headerInfos = append(headerInfos, info)
	})

	doc.Build(func() {
		for range pageCount {
			doc.AddPage()
			// Simulate writing content (no-op during counting pass).
			doc.SetFont("regular", 12)  //nolint
			doc.WriteText("body", 50, 60, 495) //nolint
		}
	})

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(headerInfos) != pageCount {
		t.Fatalf("header called %d times, want %d", len(headerInfos), pageCount)
	}
	for i, info := range headerInfos {
		if info.Number != i+1 {
			t.Errorf("headerInfos[%d].Number = %d, want %d", i, info.Number, i+1)
		}
		if info.Total != pageCount {
			t.Errorf("headerInfos[%d].Total = %d, want %d", i, info.Total, pageCount)
		}
	}
}

// TestBuild_footerReceivesTotal verifies that footer functions also receive the
// correct Total via Build.
func TestBuild_footerReceivesTotal(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}

	const pageCount = 2
	var footerTotals []int

	doc.SetFooter(func(d *pdf.Document, info pdf.PageInfo) {
		footerTotals = append(footerTotals, info.Total)
	})

	doc.Build(func() {
		for range pageCount {
			doc.AddPage()
		}
	})

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(footerTotals) != pageCount {
		t.Fatalf("footer called %d times, want %d", len(footerTotals), pageCount)
	}
	for i, tot := range footerTotals {
		if tot != pageCount {
			t.Errorf("footerTotals[%d] = %d, want %d", i, tot, pageCount)
		}
	}
}

// TestSetHeader_rendersText is an integration test that writes page number text
// inside a header and verifies the resulting PDF is valid.
func TestSetHeader_rendersText(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}

	doc.Build(func() {
		doc.SetHeader(func(d *pdf.Document, info pdf.PageInfo) {
			d.SetFont("regular", 9)   //nolint
			d.SetTextColor(120, 120, 120)
			label := fmt.Sprintf("Page %d of %d", info.Number, info.Total)
			d.WriteLine(label, 50, 15) //nolint
		})

		for range 3 {
			doc.AddPage()
			doc.SetFont("regular", 12)   //nolint
			doc.WriteText("Content", 50, 60, 495) //nolint
		}
	})

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Fatal("output is not a valid PDF")
	}
}

// TestPageCount verifies that PageCount returns the number of AddPage calls.
func TestPageCount(t *testing.T) {
	doc, _ := pdf.New(pdf.Config{})
	for i := range 5 {
		doc.AddPage()
		if got := doc.PageCount(); got != i+1 {
			t.Errorf("after %d AddPage calls: PageCount = %d, want %d", i+1, got, i+1)
		}
	}
}

// TestSetHeader_nil verifies that setting header to nil removes it without panic.
func TestSetHeader_nil(t *testing.T) {
	doc, _ := pdf.New(pdf.Config{})
	doc.SetHeader(func(d *pdf.Document, info pdf.PageInfo) {})
	doc.SetHeader(nil)
	doc.AddPage() // must not panic
}

// TestSetFooter_nil verifies that setting footer to nil removes it without panic.
func TestSetFooter_nil(t *testing.T) {
	doc, _ := pdf.New(pdf.Config{})
	doc.SetFooter(func(d *pdf.Document, info pdf.PageInfo) {})
	doc.SetFooter(nil)
	doc.AddPage()

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output after nil footer: %v", err)
	}
}
