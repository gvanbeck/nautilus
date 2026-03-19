package chart

import "github.com/gvanbeck/nautilus/pdf"

// DefaultColors is the default Highcharts series color palette.
var DefaultColors = []pdf.Color{
	{R: 124, G: 181, B: 236}, // #7cb5ec  steel blue
	{R: 67, G: 67, B: 72},    // #434348  charcoal
	{R: 144, G: 237, B: 125}, // #90ed7d  lime green
	{R: 247, G: 163, B: 92},  // #f7a35c  sandy orange
	{R: 128, G: 133, B: 233}, // #8085e9  periwinkle
	{R: 241, G: 92, B: 128},  // #f15c80  rose
	{R: 228, G: 211, B: 84},  // #e4d354  golden yellow
	{R: 43, G: 144, B: 143},  // #2b908f  teal
	{R: 244, G: 91, B: 91},   // #f45b5b  coral red
	{R: 145, G: 232, B: 225}, // #91e8e1  pale cyan
}

// SeriesColor returns the color for series index i, cycling through the
// configured palette (opts.Colors) or DefaultColors when opts.Colors is nil.
func SeriesColor(opts Options, i int) pdf.Color {
	palette := opts.Colors
	if len(palette) == 0 {
		palette = DefaultColors
	}
	return palette[i%len(palette)]
}
