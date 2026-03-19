// Package pdf provides a library for generating PDF documents in pure Go.
//
// The package supports standard paper formats (A3, A4, A5, Letter, Legal),
// custom TTF and OTF fonts, full Unicode text, and emoji rendering via PNG
// image substitution.
//
// # Quick start
//
//	resolver := &emoji.NotoResolver{Dir: "/path/to/noto-emoji-pngs"}
//
//	doc, err := pdf.New(pdf.Config{
//	    PageSize:     pdf.PageSizeA4,
//	    EmojiResolver: resolver,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	doc.AddPage()
//
//	if err := doc.RegisterFont("regular", "/path/to/font.ttf"); err != nil {
//	    log.Fatal(err)
//	}
//	if err := doc.SetFont("regular", 14); err != nil {
//	    log.Fatal(err)
//	}
//
//	if _, err := doc.WriteLine("Hello, World! 👋", 50, 100); err != nil {
//	    log.Fatal(err)
//	}
//
//	if err := doc.Save("output.pdf"); err != nil {
//	    log.Fatal(err)
//	}
//
// # Emoji support
//
// Because no pure-Go PDF library can render colour emoji glyphs from a font,
// emojis are substituted with PNG images at render time.  Supply a Resolver
// (see the emoji sub-package) that maps each emoji grapheme cluster to a PNG
// file path.  If no resolver is configured, or an emoji cannot be resolved,
// the emoji is silently skipped.
//
// Noto Emoji PNG files (Apache 2.0) from
// https://github.com/googlefonts/noto-emoji are a good free source.
//
// # Font support
//
// Both TTF and OTF fonts are supported.  Fonts must be registered before use
// with RegisterFont.  Multiple fonts (e.g. regular, bold, italic) can be
// registered under different names and switched at any time with SetFont.
package pdf
