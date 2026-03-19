package pdf

// PageInfo holds the page numbering context passed to header and footer
// functions.
type PageInfo struct {
	// Number is the 1-based index of the current page.
	Number int

	// Total is the total number of pages in the document.
	// It is 0 when the total is not yet known (e.g. during the counting pass
	// of Build, or when SetTotalPages has not been called).
	Total int
}

// HeaderFunc is a callback invoked at the start of every page, before any
// user content is written.  Use it to draw logos, document titles, or
// decorative rules in the top margin area.
//
// The function receives the Document (so it can call SetFont, WriteLine, etc.)
// and a PageInfo with the current and total page counts.
type HeaderFunc func(doc *Document, info PageInfo)

// FooterFunc is a callback invoked at the end of every page, after all user
// content has been written for that page.  Use it to render page numbers,
// dates, or horizontal rules in the bottom margin area.
type FooterFunc func(doc *Document, info PageInfo)

// SetHeader registers fn as the header callback.  It is called automatically
// each time AddPage is invoked.  Pass nil to remove a previously set header.
//
// Example – centred page title:
//
//	doc.SetHeader(func(d *pdf.Document, info pdf.PageInfo) {
//	    d.SetFont("regular", 9)
//	    d.SetTextColor(120, 120, 120)
//	    d.WriteLine("My Document", 50, 15)
//	})
func (d *Document) SetHeader(fn HeaderFunc) {
	d.header = fn
}

// SetFooter registers fn as the footer callback.  It is called automatically
// at the end of each page (when the next page begins, or at save time for the
// last page).  Pass nil to remove a previously set footer.
//
// Example – right-aligned "Page N of M":
//
//	doc.SetFooter(func(d *pdf.Document, info pdf.PageInfo) {
//	    d.SetFont("regular", 9)
//	    d.SetTextColor(120, 120, 120)
//	    label := fmt.Sprintf("Page %d of %d", info.Number, info.Total)
//	    d.WriteLine(label, 50, d.PageHeight()-20)
//	})
func (d *Document) SetFooter(fn FooterFunc) {
	d.footer = fn
}

// SetTotalPages explicitly sets the total page count used in PageInfo.Total.
// Call this when you know the page count upfront and do not need the two-pass
// Build approach.
//
// SetTotalPages and Build are mutually exclusive; Build overwrites the value
// set here.
func (d *Document) SetTotalPages(n int) {
	d.totalPages = n
}

// Build executes fn twice to support headers and footers that display the
// total page count ("Page N of M").
//
// First pass (counting): AddPage increments an internal counter but no PDF
// content is produced.  All drawing calls (WriteLine, WriteText, SetFont, …)
// are no-ops so fn must not rely on side-effects that persist between passes.
//
// Second pass (rendering): fn is called again with the total page count
// available in every PageInfo.Total field.
//
// Fonts must be registered with RegisterFont before calling Build (or inside
// fn — RegisterFont is always executed, even during the counting pass).
//
// Example:
//
//	doc.SetHeader(func(d *pdf.Document, info pdf.PageInfo) {
//	    d.SetFont("regular", 9)
//	    d.WriteLine(fmt.Sprintf("Page %d of %d", info.Number, info.Total), 50, 15)
//	})
//
//	doc.Build(func() {
//	    for _, section := range sections {
//	        doc.AddPage()
//	        doc.SetFont("regular", 12)
//	        doc.WriteText(section.Body, 50, 60, 495)
//	    }
//	})
//
//	doc.Save("report.pdf")
func (d *Document) Build(fn func()) {
	// ── Phase 1: counting pass ──────────────────────────────────────────
	d.countingMode = true
	d.pageCount = 0
	d.lastFooterPage = 0
	fn()
	// Finalize the last page's footer count even in counting mode so that
	// totalPages reflects the true number of pages.
	d.totalPages = d.pageCount

	// ── Phase 2: rendering pass ─────────────────────────────────────────
	d.countingMode = false
	d.pageCount = 0
	d.lastFooterPage = 0
	fn()
}
