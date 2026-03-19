package emoji

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Resolver maps an emoji grapheme cluster to the path of a PNG image file
// that can be embedded in the PDF.
//
// Implementations should return ("", false) when the emoji cannot be resolved
// so that the caller can skip rendering gracefully rather than returning an
// error.
type Resolver interface {
	// Resolve returns the filesystem path of a PNG for the given emoji
	// grapheme cluster.  The second return value is false when no image is
	// available for the cluster.
	Resolve(cluster string) (path string, found bool)
}

// ClusterToFilename converts an emoji grapheme cluster to the filename used
// by the Noto Emoji PNG set.
//
// The convention is:
//
//	emoji_u{codepoint1}_{codepoint2}_….png
//
// where each codepoint is a lowercase hexadecimal value.
// U+FE0F (variation selector-16) is omitted because Noto Emoji filenames
// do not include it.
//
// Examples:
//
//	"😀"     → "emoji_u1f600.png"
//	"❤️"     → "emoji_u2764.png"      (FE0F stripped)
//	"👍🏼"   → "emoji_u1f44d_1f3fc.png"
//	"👨‍👩‍👧" → "emoji_u1f468_200d_1f469_200d_1f467.png"
func ClusterToFilename(cluster string) string {
	var parts []string
	for _, r := range cluster {
		// U+FE0F is the variation selector-16 ("emoji presentation").
		// Noto Emoji filenames omit it.
		if r == '\uFE0F' {
			continue
		}
		parts = append(parts, fmt.Sprintf("%x", r))
	}
	return "emoji_u" + strings.Join(parts, "_") + ".png"
}

// NotoResolver resolves emoji grapheme clusters to Noto Emoji PNG files stored
// in a local directory.
//
// Download the PNG files from:
//
//	https://github.com/googlefonts/noto-emoji/tree/main/png/128  (128 px)
//	https://github.com/googlefonts/noto-emoji/tree/main/png/72   (72 px)
//
// The files are licensed under Apache 2.0.
type NotoResolver struct {
	// Dir is the directory that contains the Noto Emoji PNG files.
	// The files must follow the naming convention produced by ClusterToFilename.
	Dir string
}

// Resolve returns the path to the Noto Emoji PNG for the given cluster.
// It returns ("", false) when the file does not exist in Dir.
func (r *NotoResolver) Resolve(cluster string) (string, bool) {
	name := ClusterToFilename(cluster)
	path := filepath.Join(r.Dir, name)
	if _, err := os.Stat(path); err == nil {
		return path, true
	}
	return "", false
}
