package pdf

import "fmt"

// Padding defines the inner spacing between a frame's edge and its content.
// All values are in points.
type Padding struct {
	Top, Right, Bottom, Left float64
}

// UniformPadding returns a Padding with the same value on all four sides.
func UniformPadding(p float64) Padding {
	return Padding{Top: p, Right: p, Bottom: p, Left: p}
}

// HorizontalPadding returns a Padding with h points on left and right and
// v points on top and bottom.
func HorizontalPadding(h, v float64) Padding {
	return Padding{Top: v, Right: h, Bottom: v, Left: h}
}

// FrameConfig configures the position, size, and visual appearance of a Frame.
type FrameConfig struct {
	// X and Y are the top-left corner of the frame in page coordinates (points).
	X, Y float64

	// Width is the outer width of the frame in points.
	Width float64

	// Height is the outer height of the frame in points.
	// Set to 0 for auto-height: the frame grows as content is added and the
	// border is drawn at Close time using the measured content height.
	// Background fill is only applied when Height > 0.
	Height float64

	// Border defines which sides to draw and their styles.
	// Use NewUniformBorder or construct a Border literal.
	// The border is drawn by Close.
	Border Border

	// Background fills the interior of the frame with a solid color before
	// any content is rendered.  Only applied when Height > 0.
	// For auto-height frames, set a fixed Height or draw the background
	// separately with DrawBorder / a manual rectangle.
	Background *Color

	// Padding is the inner spacing between the frame edge and the content area.
	Padding Padding
}

// Frame is a positioned rectangular content box on a PDF page, similar to a
// LaTeX box or minipage.
//
// Content is written sequentially from the top of the content area.  Each
// WriteText call advances the internal Y cursor so successive calls flow
// naturally downward.  WriteLine writes on the current Y line without
// advancing; use Advance to add vertical spacing.
//
// Create a Frame with Document.NewFrame and call Close when finished.
//
// # Layout
//
//	┌─────────────────── Frame (X, Y, Width, Height) ───────────────────┐
//	│  Padding.Top                                                       │
//	│  Padding.Left  ┌── content area ──┐  Padding.Right               │
//	│                │  (text flows     │                               │
//	│                │   here)          │                               │
//	│                └──────────────────┘                               │
//	│  Padding.Bottom                                                    │
//	└───────────────────────────────────────────────────────────────────┘
//
// # Background and border rendering order
//
// To ensure the background appears beneath text and the border on top:
//  1. Background fill is drawn in NewFrame (fixed-height only).
//  2. Content methods write text and images on top.
//  3. Close draws the border last so it overlays everything.
//
// # Example — two-column layout
//
//	left := doc.NewFrame(pdf.FrameConfig{
//	    X: 50, Y: 60, Width: 230,
//	    Border: pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorGray}),
//	    Padding: pdf.UniformPadding(8),
//	})
//	left.SetFont("regular", 11)
//	left.WriteText("Left column content…")
//	left.Close()
//
//	right := doc.NewFrame(pdf.FrameConfig{
//	    X: 315, Y: 60, Width: 230,
//	    Padding: pdf.UniformPadding(8),
//	})
//	right.SetFont("regular", 11)
//	right.WriteText("Right column content…")
//	right.Close()
type Frame struct {
	doc      *Document
	cfg      FrameConfig
	contentY float64 // absolute page Y of the next content line
	closed   bool
}

// NewFrame creates a Frame at the position and size specified in cfg.
//
// If cfg.Background is non-nil and cfg.Height > 0, the background is filled
// immediately so that subsequent text appears on top of the fill.
//
// NewFrame is a no-op during the counting pass of Build and returns a valid
// but inert Frame.
func (d *Document) NewFrame(cfg FrameConfig) *Frame {
	f := &Frame{
		doc:      d,
		cfg:      cfg,
		contentY: cfg.Y + cfg.Padding.Top,
	}

	if d.countingMode {
		return f
	}

	// Draw background fill before any content so text renders on top.
	if cfg.Background != nil && cfg.Height > 0 {
		d.pdf.SetFillColor(cfg.Background.R, cfg.Background.G, cfg.Background.B)
		d.pdf.RectFromUpperLeftWithStyle(cfg.X, cfg.Y, cfg.Width, cfg.Height, "F")
	}

	return f
}

// ── Content area helpers ───────────────────────────────────────────────────

// ContentX returns the absolute page X coordinate of the left edge of the
// content area (frame X + left padding).
func (f *Frame) ContentX() float64 {
	return f.cfg.X + f.cfg.Padding.Left
}

// ContentWidth returns the usable width of the content area in points
// (frame width minus horizontal padding).
func (f *Frame) ContentWidth() float64 {
	return f.cfg.Width - f.cfg.Padding.Left - f.cfg.Padding.Right
}

// CurrentY returns the absolute page Y coordinate where the next content
// will be written.
func (f *Frame) CurrentY() float64 {
	return f.contentY
}

// FrameHeight returns the current outer height of the frame.
//
// For fixed-height frames this is cfg.Height.
// For auto-height frames this is the distance from the frame top to the
// current content position plus the bottom padding.
func (f *Frame) FrameHeight() float64 {
	if f.cfg.Height > 0 {
		return f.cfg.Height
	}
	return f.contentY - f.cfg.Y + f.cfg.Padding.Bottom
}

// ── Writing methods ────────────────────────────────────────────────────────

// WriteLine renders text on the current Y line starting at the left edge of
// the content area.  The Y cursor is NOT advanced; use Advance or WriteText
// to move down.
//
// Returns the absolute X position after the last rendered element.
func (f *Frame) WriteLine(text string) (float64, error) {
	x, err := f.doc.WriteLine(text, f.ContentX(), f.contentY)
	if err != nil {
		return f.ContentX(), fmt.Errorf("frame.WriteLine: %w", err)
	}
	return x, nil
}

// WriteLineAt renders text on the current Y line at xOffset points from the
// left edge of the content area.  The Y cursor is NOT advanced.
//
// Returns the absolute X position after the last rendered element.
func (f *Frame) WriteLineAt(text string, xOffset float64) (float64, error) {
	x, err := f.doc.WriteLine(text, f.ContentX()+xOffset, f.contentY)
	if err != nil {
		return f.ContentX(), fmt.Errorf("frame.WriteLineAt: %w", err)
	}
	return x, nil
}

// WriteText renders text with automatic word wrapping within the content
// width.  The Y cursor is advanced by one line height per rendered line so
// successive WriteText calls flow downward.
//
// Returns the absolute page Y coordinate below the last rendered line.
func (f *Frame) WriteText(text string) (float64, error) {
	endY, err := f.doc.WriteText(text, f.ContentX(), f.contentY, f.ContentWidth())
	if err != nil {
		return f.contentY, fmt.Errorf("frame.WriteText: %w", err)
	}
	f.contentY = endY
	return endY, nil
}

// WriteLineRTL renders a single line of right-to-left text with its right
// edge aligned to the right edge of the frame's content area.
//
// The text must already be in visual order (use rtl.Shape first).
// The Y cursor is NOT advanced.
//
// Returns the absolute left-edge X of the rendered text.
func (f *Frame) WriteLineRTL(text string) (float64, error) {
	rightX := f.ContentX() + f.ContentWidth()
	x, err := f.doc.WriteLineRTL(text, rightX, f.contentY)
	if err != nil {
		return rightX, fmt.Errorf("frame.WriteLineRTL: %w", err)
	}
	return x, nil
}

// WriteTextRTL renders text with word wrapping for right-to-left scripts,
// right-aligned within the frame's content width.
//
// The text must be in its original (logical) form — shaping and BiDi
// reordering are applied internally per line.  The Y cursor is advanced
// by one line height per rendered line.
//
// Returns the absolute page Y coordinate below the last rendered line.
func (f *Frame) WriteTextRTL(text string) (float64, error) {
	rightX := f.ContentX() + f.ContentWidth()
	endY, err := f.doc.WriteTextRTL(text, rightX, f.contentY, f.ContentWidth())
	if err != nil {
		return f.contentY, fmt.Errorf("frame.WriteTextRTL: %w", err)
	}
	f.contentY = endY
	return endY, nil
}

// Advance moves the Y cursor down by n points without rendering anything.
// Use this to add vertical spacing between content elements.
func (f *Frame) Advance(n float64) {
	f.contentY += n
}

// NewLine advances the Y cursor by one line height of the currently active font.
func (f *Frame) NewLine() {
	f.contentY += f.doc.lineHeight()
}

// ── Delegated document methods ─────────────────────────────────────────────

// SetFont activates a registered font at the given size.
// Delegates to Document.SetFont.
func (f *Frame) SetFont(name string, size float64) error {
	return f.doc.SetFont(name, size)
}

// SetTextColor sets the RGB text color.
// Delegates to Document.SetTextColor.
func (f *Frame) SetTextColor(r, g, b uint8) {
	f.doc.SetTextColor(r, g, b)
}

// MeasureText returns the width in points of text in the current font.
// Delegates to Document.MeasureText.
func (f *Frame) MeasureText(text string) (float64, error) {
	return f.doc.MeasureText(text)
}

// DrawInnerBorder draws a Border at position (xOffset, yOffset) relative to
// the frame's top-left corner, with the given width and height.
//
// This is useful for drawing separator lines or nested boxes within a frame.
func (f *Frame) DrawInnerBorder(xOffset, yOffset, width, height float64, border Border) error {
	return f.doc.DrawBorder(
		f.cfg.X+xOffset,
		f.cfg.Y+yOffset,
		width, height,
		border,
	)
}

// ── Finalisation ───────────────────────────────────────────────────────────

// Close finalises the frame by drawing the configured border around it.
//
// For auto-height frames, the border height is calculated from the content
// written so far plus the bottom padding.
//
// Close is idempotent: calling it more than once has no effect.
// Close is a no-op during the counting pass of Build.
func (f *Frame) Close() error {
	if f.closed {
		return nil
	}
	f.closed = true

	if f.doc.countingMode {
		return nil
	}

	h := f.FrameHeight()
	return f.doc.DrawBorder(f.cfg.X, f.cfg.Y, f.cfg.Width, h, f.cfg.Border)
}
