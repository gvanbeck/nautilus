package emoji_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gvanbeck/nautilus/pdf/emoji"
)

// TestClusterToFilename verifies the Noto Emoji filename convention.
func TestClusterToFilename(t *testing.T) {
	tests := []struct {
		name    string
		cluster string
		want    string
	}{
		{
			name:    "simple emoji",
			cluster: "😀", // U+1F600
			want:    "emoji_u1f600.png",
		},
		{
			name:    "heart with variation selector FE0F stripped",
			cluster: "❤️", // U+2764 U+FE0F
			want:    "emoji_u2764.png",
		},
		{
			name:    "thumbs up with skin-tone modifier",
			cluster: "👍🏼", // U+1F44D U+1F3FC
			want:    "emoji_u1f44d_1f3fc.png",
		},
		{
			name:    "family ZWJ sequence",
			cluster: "👨‍👩‍👧", // U+1F468 U+200D U+1F469 U+200D U+1F467
			want:    "emoji_u1f468_200d_1f469_200d_1f467.png",
		},
		{
			name:    "globe",
			cluster: "🌍", // U+1F30D
			want:    "emoji_u1f30d.png",
		},
		{
			name:    "wave hand",
			cluster: "👋", // U+1F44B
			want:    "emoji_u1f44b.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := emoji.ClusterToFilename(tt.cluster)
			if got != tt.want {
				t.Errorf("ClusterToFilename(%q) = %q, want %q", tt.cluster, got, tt.want)
			}
		})
	}
}

// TestNotoResolver_found verifies that NotoResolver returns the correct path
// when the PNG file exists.
func TestNotoResolver_found(t *testing.T) {
	// Create a temporary directory with a fake emoji PNG.
	dir := t.TempDir()
	fakeFile := filepath.Join(dir, "emoji_u1f600.png")
	if err := os.WriteFile(fakeFile, []byte("fake png data"), 0644); err != nil {
		t.Fatalf("setup: write fake file: %v", err)
	}

	r := &emoji.NotoResolver{Dir: dir}
	path, found := r.Resolve("😀")
	if !found {
		t.Fatal("Resolve returned found=false, want true")
	}
	if path != fakeFile {
		t.Errorf("Resolve path = %q, want %q", path, fakeFile)
	}
}

// TestNotoResolver_notFound verifies that NotoResolver returns ("", false) when
// the PNG file is absent.
func TestNotoResolver_notFound(t *testing.T) {
	r := &emoji.NotoResolver{Dir: t.TempDir()} // empty directory
	path, found := r.Resolve("😀")
	if found {
		t.Errorf("Resolve returned found=true with path %q, want found=false", path)
	}
}
