package pdf

import (
	"os"
	"testing"
)

// systemFont returns the path of a TTF font present on the system, or skips
// the test when no candidate font can be found.  This keeps the test suite
// self-contained without bundling a font in the repository.
func systemFont(t *testing.T) string {
	t.Helper()

	candidates := []string{
		// macOS (common bundles)
		"/Library/Fonts/Arial.ttf",
		"/System/Library/Fonts/Supplemental/Arial.ttf",
		"/Library/Fonts/Lato-Medium.ttf",
		"/Library/Fonts/Helvetica.ttf",
		// Linux
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
		"/usr/share/fonts/truetype/freefont/FreeSans.ttf",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	t.Skip("no system TTF font found; set one of the candidate paths or add your own")
	return ""
}
