package pdf

import (
	"fmt"
	"io"
	"time"

	"github.com/gvanbeck/nautilus/pdf/emoji"
	"github.com/signintech/gopdf"
)

// PageSize defines the width and height of a PDF page in points (1 pt = 1/72 inch).
type PageSize struct {
	Width  float64
	Height float64
}

// Standard ISO and North-American paper sizes in portrait orientation (points).
var (
	PageSizeA3     = PageSize{Width: 841.89, Height: 1190.55}
	PageSizeA4     = PageSize{Width: 595.28, Height: 841.89}
	PageSizeA5     = PageSize{Width: 419.53, Height: 595.28}
	PageSizeLetter = PageSize{Width: 612, Height: 792}
	PageSizeLegal  = PageSize{Width: 612, Height: 1008}
)

// Landscape variants of the standard paper sizes (width and height swapped).
var (
	PageSizeA3Landscape     = PageSize{Width: 1190.55, Height: 841.89}
	PageSizeA4Landscape     = PageSize{Width: 841.89, Height: 595.28}
	PageSizeA5Landscape     = PageSize{Width: 595.28, Height: 419.53}
	PageSizeLetterLandscape = PageSize{Width: 792, Height: 612}
	PageSizeLegalLandscape  = PageSize{Width: 1008, Height: 612}
)

// Margins defines the minimum white space between the page edge and the
// content area, in points (1 pt = 1/72 inch).
//
// Margins do not restrict where content is drawn — all write methods still
// accept explicit coordinates.  They serve as named constants that the
// caller can read back via Document.ContentX, ContentY, ContentWidth,
// ContentHeight, and ContentRightX instead of hard-coding numeric offsets.
type Margins struct {
	Top, Right, Bottom, Left float64
}

// UniformMargins returns a Margins with the same value on all four sides.
func UniformMargins(m float64) Margins {
	return Margins{Top: m, Right: m, Bottom: m, Left: m}
}

// Config holds all options for creating a new Document.
type Config struct {
	// PageSize sets the default page dimensions.
	// Defaults to PageSizeA4 when zero-valued.
	PageSize PageSize

	// Margins sets the page margins in points.
	// When zero-valued, all margins are 0 and ContentX/Y equal 0.
	Margins Margins

	// EmojiResolver maps emoji grapheme clusters to PNG image paths.
	// When nil, emoji characters are silently skipped during rendering.
	EmojiResolver emoji.Resolver

	// DefaultFontSize is the initial font size in points.
	// Defaults to 12 when zero-valued.
	DefaultFontSize float64

	// LineHeightFactor is multiplied by the current font size to obtain the
	// line height used by WriteText.  Defaults to 1.2 when zero-valued.
	LineHeightFactor float64
}

// Document represents a single PDF document.
//
// Call AddPage at least once before writing any content.  Register fonts with
// RegisterFont and activate them with SetFont before writing text.
//
// # Text rendering
//
// Use WriteLine for a single line and WriteText for word-wrapped paragraphs.
// Both return a continuation coordinate (X or Y) that can be used to chain
// further content.
//
// # Right-to-left text (Arabic and Hebrew)
//
// Use WriteLineRTL for a single pre-shaped RTL line and WriteTextRTL for
// word-wrapped RTL paragraphs.  Arabic text must first be shaped with the
// pdf/rtl package:
//
//	shaped := rtl.Shape("مرحبا بالعالم")       // single line
//	doc.WriteLineRTL(shaped, rightEdge, y)
//
//	doc.WriteTextRTL("مرحبا بالعالم", rightEdge, y, maxWidth)  // multi-line
//
// WriteTextRTL applies shaping and BiDi reordering internally per line so
// word order is preserved correctly across line breaks.
//
// # Headers and footers
//
// Register callbacks with SetHeader and SetFooter.  They are invoked
// automatically on every page.  Use Build when the footer must display the
// total page count ("Page N of M"); otherwise call SetTotalPages upfront or
// accept PageInfo.Total == 0 (unknown).
type Document struct {
	pdf              gopdf.GoPdf
	resolver         emoji.Resolver
	fontSize         float64
	lineHeightFactor float64
	pageSize         PageSize
	margins          Margins

	// header/footer state
	header         HeaderFunc
	footer         FooterFunc
	pageCount      int // number of AddPage calls so far
	totalPages     int // pre-set or computed by Build
	lastFooterPage int // page number for which footer was last rendered
	countingMode   bool

	// font tracking – updated by SetFont so other packages can restore state
	currentFont string
}

// New creates and returns a new Document using the supplied Config.
//
// Zero-valued fields in Config are replaced with sensible defaults:
// PageSizeA4, font size 12, line-height factor 1.2.
func New(cfg Config) (*Document, error) {
	if cfg.PageSize.Width == 0 || cfg.PageSize.Height == 0 {
		cfg.PageSize = PageSizeA4
	}
	if cfg.DefaultFontSize < 0 {
		return nil, fmt.Errorf("pdf: DefaultFontSize must not be negative, got %f", cfg.DefaultFontSize)
	}
	if cfg.DefaultFontSize == 0 {
		cfg.DefaultFontSize = 12
	}
	if cfg.LineHeightFactor < 0 {
		return nil, fmt.Errorf("pdf: LineHeightFactor must not be negative, got %f", cfg.LineHeightFactor)
	}
	if cfg.LineHeightFactor == 0 {
		cfg.LineHeightFactor = 1.2
	}
	if cfg.Margins.Top < 0 || cfg.Margins.Right < 0 || cfg.Margins.Bottom < 0 || cfg.Margins.Left < 0 {
		return nil, fmt.Errorf("pdf: margins must not be negative")
	}

	d := &Document{
		resolver:         cfg.EmojiResolver,
		fontSize:         cfg.DefaultFontSize,
		lineHeightFactor: cfg.LineHeightFactor,
		pageSize:         cfg.PageSize,
		margins:          cfg.Margins,
	}

	d.pdf.Start(gopdf.Config{
		PageSize: gopdf.Rect{
			W: cfg.PageSize.Width,
			H: cfg.PageSize.Height,
		},
	})

	return d, nil
}

// AddPage appends a new page to the document and makes it the active page.
//
// If a footer callback is set, it is rendered on the previous page before
// the new page is created.  If a header callback is set, it is rendered
// immediately on the new page.
//
// In the counting pass of Build, AddPage only increments the internal page
// counter; no PDF content is produced.
func (d *Document) AddPage() {
	if d.countingMode {
		d.pageCount++
		return
	}

	// Render the footer for the page we are about to leave.
	d.renderFooterIfNeeded()

	d.pageCount++
	d.pdf.AddPage()

	// Render the header at the top of the new page.
	if d.header != nil {
		d.header(d, d.pageInfo())
	}
}

// PageWidth returns the width of the current page in points.
func (d *Document) PageWidth() float64 {
	return d.pageSize.Width
}

// PageHeight returns the height of the current page in points.
func (d *Document) PageHeight() float64 {
	return d.pageSize.Height
}

// PageCount returns the number of pages added so far.
func (d *Document) PageCount() int {
	return d.pageCount
}

// ContentX returns the X coordinate of the left edge of the content area
// (page left edge + left margin).
func (d *Document) ContentX() float64 {
	return d.margins.Left
}

// ContentY returns the Y coordinate of the top edge of the content area
// (page top edge + top margin).
func (d *Document) ContentY() float64 {
	return d.margins.Top
}

// ContentWidth returns the usable width of the content area
// (page width − left margin − right margin).
func (d *Document) ContentWidth() float64 {
	return d.pageSize.Width - d.margins.Left - d.margins.Right
}

// ContentHeight returns the usable height of the content area
// (page height − top margin − bottom margin).
func (d *Document) ContentHeight() float64 {
	return d.pageSize.Height - d.margins.Top - d.margins.Bottom
}

// ContentRightX returns the X coordinate of the right edge of the content area
// (page width − right margin).  Useful as the anchor for right-to-left text.
func (d *Document) ContentRightX() float64 {
	return d.pageSize.Width - d.margins.Right
}

// ContentBottomY returns the Y coordinate of the bottom edge of the content
// area (page height − bottom margin).  Useful as the overflow threshold for
// tables and frames.
func (d *Document) ContentBottomY() float64 {
	return d.pageSize.Height - d.margins.Bottom
}

// SetLineHeightFactor sets the multiplier used when advancing between lines in
// WriteText.  A value of 1.0 produces single spacing; 1.5 produces
// one-and-a-half spacing.
func (d *Document) SetLineHeightFactor(factor float64) error {
	if factor <= 0 {
		return fmt.Errorf("pdf: line height factor must be positive, got %f", factor)
	}
	d.lineHeightFactor = factor
	return nil
}

// lineHeight returns the current line height in points.
func (d *Document) lineHeight() float64 {
	return d.fontSize * d.lineHeightFactor
}

// SetTextColor sets the RGB colour used for subsequent text rendering.
// Each component must be in the range [0, 255].
func (d *Document) SetTextColor(r, g, b uint8) {
	if d.countingMode {
		return
	}
	d.pdf.SetTextColor(r, g, b)
}

// GetX returns the current horizontal cursor position in points.
func (d *Document) GetX() float64 {
	return d.pdf.GetX()
}

// GetY returns the current vertical cursor position in points.
func (d *Document) GetY() float64 {
	return d.pdf.GetY()
}

// SetInfo sets the PDF document information dictionary (metadata).
// Fields left empty are omitted from the PDF.
func (d *Document) SetInfo(title, author, subject, creator string) {
	d.pdf.SetInfo(gopdf.PdfInfo{
		Title:        title,
		Author:       author,
		Subject:      subject,
		Creator:      creator,
		Producer:     "nautilus",
		CreationDate: time.Now(),
	})
}

// Save writes the complete PDF to the file at path, creating or truncating it.
// The footer of the last page is rendered before writing.
func (d *Document) Save(path string) error {
	d.renderFooterIfNeeded()
	if err := d.pdf.WritePdf(path); err != nil {
		return fmt.Errorf("pdf: save to %q: %w", path, err)
	}
	return nil
}

// Output writes the complete PDF to w.
// The footer of the last page is rendered before writing.
func (d *Document) Output(w io.Writer) error {
	d.renderFooterIfNeeded()
	if _, err := d.pdf.WriteTo(w); err != nil {
		return fmt.Errorf("pdf: write to writer: %w", err)
	}
	return nil
}

// pageInfo returns a PageInfo for the current page.
func (d *Document) pageInfo() PageInfo {
	return PageInfo{
		Number: d.pageCount,
		Total:  d.totalPages,
	}
}

// renderFooterIfNeeded renders the footer for the current page if it has not
// been rendered yet.  It is a no-op when no footer is set, during the counting
// pass, or when no pages have been added.
func (d *Document) renderFooterIfNeeded() {
	if d.countingMode || d.footer == nil || d.pageCount == 0 {
		return
	}
	if d.lastFooterPage >= d.pageCount {
		return // already rendered for this page
	}
	d.lastFooterPage = d.pageCount
	d.footer(d, d.pageInfo())
}
