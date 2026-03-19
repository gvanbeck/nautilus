// Package emoji provides utilities for segmenting Unicode text into plain-text
// and emoji runs, and for resolving emoji grapheme clusters to PNG image paths.
//
// # Text segmentation
//
// Split breaks a string into a slice of Segment values, each labelled as
// either KindText or KindEmoji.  Multi-codepoint sequences (skin-tone
// modifiers, zero-width joiners, variation selectors) are kept together as a
// single grapheme cluster and treated as one emoji segment.
//
//	segments := emoji.Split("Hi 👋 there 🌍")
//	// → [{KindText "Hi "}, {KindEmoji "👋"}, {KindText " there "}, {KindEmoji "🌍"}]
//
// # Emoji resolution
//
// A Resolver maps an emoji grapheme cluster to the path of a PNG file that
// can be embedded in the PDF.  NotoResolver implements this interface using
// locally stored Noto Emoji PNG files.
//
// Download Noto Emoji PNGs (Apache 2.0) from:
//
//	https://github.com/googlefonts/noto-emoji/tree/main/png/128
//
// Then point NotoResolver.Dir at the directory containing the downloaded files:
//
//	r := &emoji.NotoResolver{Dir: "/path/to/noto/png/128"}
//	path, found := r.Resolve("😀")  // returns "…/emoji_u1f600.png", true
package emoji
