package render_test

import (
	"math"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/internal/render"
)

// ── BoolVal ───────────────────────────────────────────────────────────────────

func TestBoolVal_nil(t *testing.T) {
	if got := render.BoolVal(nil, true); got != true {
		t.Errorf("BoolVal(nil, true) = %v, want true", got)
	}
	if got := render.BoolVal(nil, false); got != false {
		t.Errorf("BoolVal(nil, false) = %v, want false", got)
	}
}

func TestBoolVal_nonNil(t *testing.T) {
	tr, fa := true, false
	if got := render.BoolVal(&tr, false); got != true {
		t.Errorf("BoolVal(&true, false) = %v, want true", got)
	}
	if got := render.BoolVal(&fa, true); got != false {
		t.Errorf("BoolVal(&false, true) = %v, want false", got)
	}
}

// ── FormatFloat ───────────────────────────────────────────────────────────────

func TestFormatFloat(t *testing.T) {
	cases := []struct {
		v    float64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{-3, "-3"},
		{3.14, "3.14"},
		{3.10, "3.1"},
		{3.00, "3"},
		{100, "100"},
		{0.5, "0.5"},
	}
	for _, c := range cases {
		if got := render.FormatFloat(c.v); got != c.want {
			t.Errorf("FormatFloat(%v) = %q, want %q", c.v, got, c.want)
		}
	}
}

// ── FormatAxisValue ───────────────────────────────────────────────────────────

func TestFormatAxisValue_noFormat(t *testing.T) {
	if got := render.FormatAxisValue(42, nil); got != "42" {
		t.Errorf("got %q, want %q", got, "42")
	}
}

func TestFormatAxisValue_withFormat(t *testing.T) {
	ax := &chart.Axis{Labels: &chart.AxisLabels{Format: "{value}%"}}
	if got := render.FormatAxisValue(75, ax); got != "75%" {
		t.Errorf("got %q, want %q", got, "75%")
	}
}

// ── AutoCategories ────────────────────────────────────────────────────────────

func TestAutoCategories(t *testing.T) {
	cats := render.AutoCategories(3)
	if len(cats) != 3 {
		t.Fatalf("len = %d, want 3", len(cats))
	}
	for i, want := range []string{"1", "2", "3"} {
		if cats[i] != want {
			t.Errorf("[%d] = %q, want %q", i, cats[i], want)
		}
	}
}

func TestAutoCategories_zero(t *testing.T) {
	if cats := render.AutoCategories(0); len(cats) != 0 {
		t.Errorf("len = %d, want 0", len(cats))
	}
}

// ── CategoriesFor ─────────────────────────────────────────────────────────────

func TestCategoriesFor_fromXAxis(t *testing.T) {
	opts := chart.Options{
		XAxis: &chart.Axis{Categories: []string{"A", "B", "C"}},
	}
	cats := render.CategoriesFor(opts)
	if len(cats) != 3 || cats[0] != "A" {
		t.Errorf("unexpected categories: %v", cats)
	}
}

func TestCategoriesFor_autoFromData(t *testing.T) {
	opts := chart.Options{
		Series: []chart.Series{{Data: []float64{10, 20, 30}}},
	}
	cats := render.CategoriesFor(opts)
	if len(cats) != 3 {
		t.Errorf("len = %d, want 3", len(cats))
	}
}

func TestCategoriesFor_empty(t *testing.T) {
	if cats := render.CategoriesFor(chart.Options{}); cats != nil {
		t.Errorf("expected nil, got %v", cats)
	}
}

// ── DataRange ─────────────────────────────────────────────────────────────────

func TestDataRange_basic(t *testing.T) {
	series := []chart.Series{
		{Data: []float64{10, 5, 20}},
		{Data: []float64{3, 15, 8}},
	}
	min, max := render.DataRange(series)
	if min != 3 {
		t.Errorf("min = %v, want 3", min)
	}
	if max != 20 {
		t.Errorf("max = %v, want 20", max)
	}
}

func TestDataRange_empty(t *testing.T) {
	min, max := render.DataRange(nil)
	if min != 0 || max != 0 {
		t.Errorf("empty DataRange = (%v, %v), want (0, 0)", min, max)
	}
}

// ── XDataRange / ZDataRange ───────────────────────────────────────────────────

func TestXDataRange(t *testing.T) {
	series := []chart.Series{{
		Points: []chart.Point{{X: 5}, {X: 1}, {X: 9}},
	}}
	min, max := render.XDataRange(series)
	if min != 1 || max != 9 {
		t.Errorf("got (%v, %v), want (1, 9)", min, max)
	}
}

func TestZDataRange(t *testing.T) {
	series := []chart.Series{{
		Points: []chart.Point{{Z: 3}, {Z: 10}, {Z: 7}},
	}}
	min, max := render.ZDataRange(series)
	if min != 3 || max != 10 {
		t.Errorf("got (%v, %v), want (3, 10)", min, max)
	}
}

// ── NiceRange ─────────────────────────────────────────────────────────────────

func TestNiceRange_positive(t *testing.T) {
	min, max, step := render.NiceRange(0, 100, nil)
	if min != 0 {
		t.Errorf("min = %v, want 0", min)
	}
	if max < 100 {
		t.Errorf("max = %v, want >= 100", max)
	}
	if step <= 0 {
		t.Errorf("step = %v, want > 0", step)
	}
	// max should be reachable in steps from min
	if math.Mod(max-min, step) > 1e-9 {
		t.Errorf("max-min (%v) is not a multiple of step (%v)", max-min, step)
	}
}

func TestNiceRange_negative(t *testing.T) {
	min, max, step := render.NiceRange(-50, 50, nil)
	if min > -50 {
		t.Errorf("min = %v should be <= -50", min)
	}
	if max < 50 {
		t.Errorf("max = %v should be >= 50", max)
	}
	if step <= 0 {
		t.Errorf("step = %v, want > 0", step)
	}
}

func TestNiceRange_clampedByAxis(t *testing.T) {
	ax := &chart.Axis{Min: chart.Float(0), Max: chart.Float(200)}
	min, max, _ := render.NiceRange(10, 100, ax)
	if min != 0 {
		t.Errorf("min = %v, want 0 (axis min)", min)
	}
	if max != 200 {
		t.Errorf("max = %v, want 200 (axis max)", max)
	}
}

func TestNiceRange_fixedTickInterval(t *testing.T) {
	ax := &chart.Axis{TickInterval: chart.Float(25)}
	_, _, step := render.NiceRange(0, 100, ax)
	if step != 25 {
		t.Errorf("step = %v, want 25", step)
	}
}

// ── ValueToY / ValueToX ───────────────────────────────────────────────────────

func TestValueToY(t *testing.T) {
	plot := render.Area{X: 0, Y: 0, W: 100, H: 200}
	// min maps to bottom
	if got := render.ValueToY(0, 0, 100, plot); got != 200 {
		t.Errorf("ValueToY(0) = %v, want 200 (bottom)", got)
	}
	// max maps to top
	if got := render.ValueToY(100, 0, 100, plot); got != 0 {
		t.Errorf("ValueToY(100) = %v, want 0 (top)", got)
	}
	// midpoint
	if got := render.ValueToY(50, 0, 100, plot); got != 100 {
		t.Errorf("ValueToY(50) = %v, want 100 (mid)", got)
	}
}

func TestValueToX(t *testing.T) {
	plot := render.Area{X: 10, Y: 0, W: 100, H: 50}
	// min maps to left edge
	if got := render.ValueToX(0, 0, 100, plot); got != 10 {
		t.Errorf("ValueToX(0) = %v, want 10 (left)", got)
	}
	// max maps to right edge
	if got := render.ValueToX(100, 0, 100, plot); got != 110 {
		t.Errorf("ValueToX(100) = %v, want 110 (right)", got)
	}
}

func TestValueToY_equalMinMax(t *testing.T) {
	plot := render.Area{X: 0, Y: 0, W: 100, H: 200}
	got := render.ValueToY(50, 50, 50, plot)
	if got != 100 { // midpoint of plot height
		t.Errorf("ValueToY equal range = %v, want 100", got)
	}
}

// ── CategoryCenterX / CategoryLeftX ──────────────────────────────────────────

func TestCategoryCenterX(t *testing.T) {
	plot := render.Area{X: 0, Y: 0, W: 300, H: 100}
	// 3 categories: each slot = 100 wide, centres at 50, 150, 250
	cases := []struct{ i, n int; want float64 }{
		{0, 3, 50},
		{1, 3, 150},
		{2, 3, 250},
	}
	for _, c := range cases {
		if got := render.CategoryCenterX(c.i, c.n, plot); got != c.want {
			t.Errorf("CategoryCenterX(%d, %d) = %v, want %v", c.i, c.n, got, c.want)
		}
	}
}

func TestCategoryLeftX(t *testing.T) {
	plot := render.Area{X: 0, Y: 0, W: 300, H: 100}
	if got := render.CategoryLeftX(0, 3, plot); got != 0 {
		t.Errorf("CategoryLeftX(0, 3) = %v, want 0", got)
	}
	if got := render.CategoryLeftX(2, 3, plot); got != 200 {
		t.Errorf("CategoryLeftX(2, 3) = %v, want 200", got)
	}
}

// ── LightenColor / BlendColor ─────────────────────────────────────────────────

func TestLightenColor_alpha1(t *testing.T) {
	c := pdf.Color{R: 100, G: 150, B: 200}
	got := render.LightenColor(c, 1.0)
	if got.R != c.R || got.G != c.G || got.B != c.B {
		t.Errorf("LightenColor(alpha=1) changed color: %v → %v", c, got)
	}
}

func TestLightenColor_alpha0(t *testing.T) {
	c := pdf.Color{R: 100, G: 150, B: 200}
	got := render.LightenColor(c, 0)
	if got.R != 255 || got.G != 255 || got.B != 255 {
		t.Errorf("LightenColor(alpha=0) = %v, want white", got)
	}
}

func TestBlendColor(t *testing.T) {
	a := pdf.Color{R: 0, G: 0, B: 0}
	b := pdf.Color{R: 200, G: 100, B: 50}
	mid := render.BlendColor(a, b, 0.5)
	if mid.R != 100 || mid.G != 50 || mid.B != 25 {
		t.Errorf("BlendColor midpoint = %v, want {100 50 25}", mid)
	}
	// t=0 → a
	got0 := render.BlendColor(a, b, 0)
	if got0.R != 0 || got0.G != 0 || got0.B != 0 {
		t.Errorf("BlendColor(t=0) = %v, want a={0 0 0}", got0)
	}
	// t=1 → b
	got1 := render.BlendColor(a, b, 1)
	if got1.R != b.R || got1.G != b.G || got1.B != b.B {
		t.Errorf("BlendColor(t=1) = %v, want b=%v", got1, b)
	}
}

// ── ParsePercent ──────────────────────────────────────────────────────────────

func TestParsePercent(t *testing.T) {
	cases := []struct {
		s    string
		want float64
	}{
		{"50%", 0.5},
		{"100%", 1.0},
		{"0%", 0.0},
		{"30%", 0.3},
		{"", 0},
		{"invalid", 0},
	}
	for _, c := range cases {
		got := render.ParsePercent(c.s)
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("ParsePercent(%q) = %v, want %v", c.s, got, c.want)
		}
	}
}

// ── PieSlicePolygon / DonutSlicePolygon ───────────────────────────────────────

func TestPieSlicePolygon_includesCenter(t *testing.T) {
	pts := render.PieSlicePolygon(100, 100, 50, 0, math.Pi/2)
	// First point should be the center.
	if pts[0].X != 100 || pts[0].Y != 100 {
		t.Errorf("first point = %v, want center (100,100)", pts[0])
	}
	if len(pts) < 3 {
		t.Errorf("too few points: %d", len(pts))
	}
}

func TestPieSlicePolygon_fullCircle(t *testing.T) {
	pts := render.PieSlicePolygon(0, 0, 10, 0, 2*math.Pi)
	// Should have many arc points plus center.
	if len(pts) < 10 {
		t.Errorf("full circle has too few points: %d", len(pts))
	}
}

func TestDonutSlicePolygon_noCenter(t *testing.T) {
	pts := render.DonutSlicePolygon(100, 100, 50, 25, 0, math.Pi/2)
	// Donut has no center point; all points lie on the two arcs.
	for _, p := range pts {
		d := math.Sqrt((p.X-100)*(p.X-100) + (p.Y-100)*(p.Y-100))
		if d < 24 || d > 51 {
			t.Errorf("point %v at distance %v is outside donut radii (25–50)", p, d)
		}
	}
}

// ── ComputeLayout ─────────────────────────────────────────────────────────────

func TestComputeLayout_plotInsideBounds(t *testing.T) {
	opts := chart.Options{
		FontSize: 10,
		Title:    &chart.Title{Text: "T"},
		XAxis:    &chart.Axis{},
		YAxis:    &chart.Axis{},
		Legend:   &chart.Legend{},
		Series:   []chart.Series{{Name: "S", Data: []float64{1}}},
	}
	l := render.ComputeLayout(opts, 0, 0, 500, 400)
	if l.Plot.X < 0 || l.Plot.Y < 0 {
		t.Errorf("plot origin negative: %v", l.Plot)
	}
	if l.Plot.W <= 0 || l.Plot.H <= 0 {
		t.Errorf("plot has non-positive dimensions: %v", l.Plot)
	}
	if l.Plot.Right() > 500 {
		t.Errorf("plot right (%v) exceeds bounding width 500", l.Plot.Right())
	}
	if l.Plot.Bottom() > 400 {
		t.Errorf("plot bottom (%v) exceeds bounding height 400", l.Plot.Bottom())
	}
}

func TestComputeLayout_hiddenAxes(t *testing.T) {
	opts := chart.Options{
		FontSize: 10,
		YAxis:    &chart.Axis{Visible: chart.Bool(false)},
		XAxis:    &chart.Axis{Visible: chart.Bool(false)},
	}
	l := render.ComputeLayout(opts, 0, 0, 400, 300)
	// With hidden axes the plot should be wider/taller.
	lVisible := render.ComputeLayout(chart.Options{
		FontSize: 10,
		YAxis:    &chart.Axis{},
		XAxis:    &chart.Axis{},
	}, 0, 0, 400, 300)
	if l.Plot.W <= lVisible.Plot.W {
		t.Errorf("hidden axes: plot.W (%v) should be wider than with visible axes (%v)", l.Plot.W, lVisible.Plot.W)
	}
}

// ── Area helpers ──────────────────────────────────────────────────────────────

func TestArea_RightBottom(t *testing.T) {
	a := render.Area{X: 10, Y: 20, W: 100, H: 50}
	if got := a.Right(); got != 110 {
		t.Errorf("Right() = %v, want 110", got)
	}
	if got := a.Bottom(); got != 70 {
		t.Errorf("Bottom() = %v, want 70", got)
	}
}
