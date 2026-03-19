package chart_test

import (
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
)

// ── Float / Bool helpers ──────────────────────────────────────────────────────

func TestFloat(t *testing.T) {
	v := chart.Float(3.14)
	if v == nil {
		t.Fatal("Float returned nil")
	}
	if *v != 3.14 {
		t.Errorf("*Float(3.14) = %v, want 3.14", *v)
	}
}

func TestBool(t *testing.T) {
	tr := chart.Bool(true)
	if tr == nil || !*tr {
		t.Error("Bool(true) failed")
	}
	fa := chart.Bool(false)
	if fa == nil || *fa {
		t.Error("Bool(false) failed")
	}
}

// ── SeriesColor ───────────────────────────────────────────────────────────────

func TestSeriesColor_cyclesPalette(t *testing.T) {
	opts := chart.Options{}
	n := 20
	seen := make(map[pdf.Color]bool)
	for i := 0; i < n; i++ {
		c := chart.SeriesColor(opts, i)
		seen[c] = true
	}
	// palette should have several distinct colors
	if len(seen) < 2 {
		t.Errorf("SeriesColor returned only %d distinct color(s) for %d indices", len(seen), n)
	}
}

func TestSeriesColor_customPalette(t *testing.T) {
	red := pdf.Color{R: 255}
	blue := pdf.Color{B: 255}
	opts := chart.Options{Colors: []pdf.Color{red, blue}}

	if got := chart.SeriesColor(opts, 0); got != red {
		t.Errorf("index 0: got %v, want %v", got, red)
	}
	if got := chart.SeriesColor(opts, 1); got != blue {
		t.Errorf("index 1: got %v, want %v", got, blue)
	}
	// wraps around
	if got := chart.SeriesColor(opts, 2); got != red {
		t.Errorf("index 2 (wrap): got %v, want %v", got, red)
	}
}

// ── Point zero value ──────────────────────────────────────────────────────────

func TestPoint_zeroValue(t *testing.T) {
	var p chart.Point
	if p.X != 0 || p.Y != 0 || p.Z != 0 {
		t.Error("zero Point has unexpected non-zero fields")
	}
	if p.IsSum || p.IsIntermediateSum {
		t.Error("zero Point has unexpected IsSum flags")
	}
}
