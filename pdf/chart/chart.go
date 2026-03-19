// Package chart defines the shared configuration types used by all Nautilus
// chart packages.  The API mirrors the Highcharts JSON configuration model so
// that charts can be built with a familiar, declarative style.
//
// Each chart renderer lives in its own sub-package so that a binary only pays
// for the renderers it imports:
//
//	pdf/chart/line        – line charts
//	pdf/chart/area        – area (filled line) charts
//	pdf/chart/column      – vertical bar (column) charts
//	pdf/chart/bar         – horizontal bar charts
//	pdf/chart/pie         – pie and donut charts
//	pdf/chart/polar       – polar (spider/radar) charts
//	pdf/chart/scatter     – scatter (X-Y point cloud) charts
//	pdf/chart/bubble      – bubble charts (scatter with z-sized circles)
//	pdf/chart/heatmap     – heatmap (color-coded grid) charts
//	pdf/chart/waterfall   – waterfall (running total) charts
//	pdf/chart/funnel      – funnel and pyramid charts
//	pdf/chart/gauge       – gauge and solid-gauge charts
//	pdf/chart/errorbar    – error bar charts
//	pdf/chart/boxplot     – box-and-whisker charts
//	pdf/chart/columnrange – column range (low-high bar) charts
//	pdf/chart/arearange   – area range (low-high band) charts
//	pdf/chart/bullet      – bullet charts
//	pdf/chart/dumbbell    – dumbbell (low-high range dot) charts
//	pdf/chart/lollipop    – lollipop (stick + dot) charts
//	pdf/chart/treemap     – treemap (hierarchical rectangle packing) charts
//
// All chart types implement the Drawable interface and can be embedded in a
// layout.Flowable via chart.NewFlowable:
//
//	lc := &line.LineChart{Options: chart.Options{...}}
//	story = append(story, chart.NewFlowable(lc, 0, 220))
package chart

import "github.com/gvanbeck/nautilus/pdf"

// Float returns a *float64 pointing to v.
// Use it to set optional float64 fields inline.
func Float(v float64) *float64 { return &v }

// Bool returns a *bool pointing to v.
// Use it to set optional bool fields inline.
func Bool(v bool) *bool { return &v }

// Drawable is the interface implemented by all chart types.
// x, y is the top-left corner of the bounding box in points.
type Drawable interface {
	Draw(doc *pdf.Document, x, y, width, height float64) error
}

// Options is the top-level chart configuration object, mirroring the
// Highcharts options hierarchy.
type Options struct {
	// Title is the main chart title.
	Title *Title

	// Subtitle is an optional secondary title rendered below Title.
	Subtitle *Title

	// XAxis configures the horizontal (category) axis.
	XAxis *Axis

	// YAxis configures the vertical (value) axis.
	YAxis *Axis

	// Series holds the data sets to be plotted.
	Series []Series

	// Legend configures the chart legend.
	Legend *Legend

	// PlotOptions provides per-chart-type rendering knobs.
	PlotOptions *PlotOptions

	// Colors overrides the default color palette for series.
	// When nil, DefaultColors is used.
	Colors []pdf.Color

	// Background fills the chart bounding box when set.
	Background *pdf.Color

	// FontName is the default registered font used for all chart text.
	// Must be registered on the Document before calling Draw.
	// Individual text elements can override this with their own FontName.
	FontName string

	// FontSize is the base font size in points.  Defaults to 9 when 0.
	FontSize float64
}

// Title configures title or subtitle text.
type Title struct {
	Text     string
	FontName string     // overrides Options.FontName when non-empty
	FontSize float64    // overrides Options.FontSize when > 0
	Color    *pdf.Color // defaults to black
}

// Axis configures one axis of a cartesian chart.
type Axis struct {
	// Title is an optional axis label drawn alongside the axis.
	Title *Title

	// Categories are the tick labels for discrete axes.
	// When set, data points map to their slice index.
	// When nil, numeric labels are auto-generated.
	Categories []string

	// Min and Max clamp the visible value range.
	// When nil they are derived automatically from the data.
	Min *float64
	Max *float64

	// TickInterval fixes the spacing between gridlines / tick labels.
	// When nil a "nice" interval is chosen automatically.
	TickInterval *float64

	// GridLineWidth is the stroke width of gridlines in points.
	// 0 uses the default (0.5 pt); negative hides gridlines entirely.
	GridLineWidth float64

	// GridLineColor overrides the default light-gray grid color.
	GridLineColor *pdf.Color

	// Labels configures tick label text.
	Labels *AxisLabels

	// Visible hides the axis when false.  Defaults to true.
	Visible *bool
}

// AxisLabels configures tick-label text on an axis.
type AxisLabels struct {
	// Enabled hides tick labels when false.  Defaults to true.
	Enabled *bool

	// Format is a template string where "{value}" is replaced with the
	// formatted number.  Example: "{value}%".
	Format string

	// FontName overrides Options.FontName.
	FontName string

	// FontSize overrides Options.FontSize when > 0.
	FontSize float64
}

// Series holds one data set to be plotted.
type Series struct {
	// Name is shown in the legend.
	Name string

	// Data contains the y-values, one per category.
	// Used by line, area, column, bar, and pie charts.
	Data []float64

	// Points holds rich data for chart types that need more than a single
	// y-value (scatter, bubble, heatmap, range charts, box plot, etc.).
	// When non-nil, the chart renderer uses Points instead of Data.
	Points []Point

	// Color overrides the automatic palette assignment for this series.
	Color *pdf.Color
}

// Point is a rich data point for chart types that require more than a
// single y-value.  Set only the fields meaningful for your chart type.
type Point struct {
	// X is the horizontal value (scatter, bubble, heatmap column index).
	X float64
	// Y is the primary value.
	Y float64
	// Z is the third dimension (bubble radius source, heatmap cell value).
	Z float64

	// Low is the lower bound (range charts, box plot, error bar, dumbbell).
	Low float64
	// Q1 is the first quartile (box plot only).
	Q1 float64
	// Median is the median value (box plot only).
	Median float64
	// Q3 is the third quartile (box plot only).
	Q3 float64
	// High is the upper bound (range charts, box plot, error bar, dumbbell).
	High float64

	// Target is a reference/target value (bullet chart).
	Target float64

	// Name labels this point (waterfall steps, funnel stages, treemap nodes).
	Name string
	// Color overrides the series color for this specific point.
	Color *pdf.Color

	// IsSum makes a waterfall bar show the cumulative total (Y is ignored).
	IsSum bool
	// IsIntermediateSum makes a waterfall bar a running subtotal.
	IsIntermediateSum bool
}

// Legend configures the chart legend.
type Legend struct {
	// Enabled hides the legend when false.  Defaults to true.
	Enabled *bool

	// Layout is "horizontal" (default) or "vertical".
	Layout string

	// Align is "left", "center" (default), or "right".
	Align string

	// VerticalAlign is "top", "middle", or "bottom" (default).
	VerticalAlign string

	// FontName overrides Options.FontName.
	FontName string

	// FontSize overrides Options.FontSize when > 0.
	FontSize float64
}

// PlotOptions contains per-chart-type rendering options.
type PlotOptions struct {
	Line        *LineOptions
	Area        *AreaOptions
	Column      *ColumnOptions
	Bar         *BarOptions
	Pie         *PieOptions
	Polar       *PolarOptions
	Scatter     *ScatterOptions
	Bubble      *BubbleOptions
	Heatmap     *HeatmapOptions
	Waterfall   *WaterfallOptions
	Funnel      *FunnelOptions
	Gauge       *GaugeOptions
	Errorbar    *ErrorbarOptions
	Boxplot     *BoxplotOptions
	ColumnRange *ColumnRangeOptions
	AreaRange   *AreaRangeOptions
	Bullet      *BulletOptions
	Dumbbell    *DumbbellOptions
	Lollipop    *LollipopOptions
	Treemap     *TreemapOptions
}

// LineOptions controls line chart rendering.
type LineOptions struct {
	DataLabels *DataLabels
	Marker     *Marker
	// LineWidth is the stroke width in points.  Defaults to 2.
	LineWidth float64
}

// AreaOptions controls area (filled line) chart rendering.
type AreaOptions struct {
	DataLabels *DataLabels
	Marker     *Marker
	// LineWidth is the stroke width in points.  Defaults to 2.
	LineWidth float64
	// FillAlpha controls how much the fill color is lightened towards white.
	// Range [0,1]: 0 = white fill, 1 = full series color.  Defaults to 0.3.
	FillAlpha float64
}

// ColumnOptions controls column (vertical bar) chart rendering.
type ColumnOptions struct {
	DataLabels *DataLabels
	// BorderWidth is the outline thickness around each bar.  Defaults to 0.
	BorderWidth float64
	// BorderColor is the bar outline color.  Defaults to white.
	BorderColor *pdf.Color
	// Stacking is "" (grouped), "normal" (stacked), or "percent".
	Stacking string
	// GroupPadding is the fraction of a category slot reserved as padding
	// between groups.  Defaults to 0.2.
	GroupPadding float64
	// PointPadding is the fraction of a per-series slot reserved as padding
	// on each side.  Defaults to 0.1.
	PointPadding float64
}

// BarOptions mirrors ColumnOptions for horizontal bar charts.
type BarOptions = ColumnOptions

// PieOptions controls pie and donut chart rendering.
type PieOptions struct {
	DataLabels *DataLabels
	// InnerSize is the inner-radius as a percentage string (e.g. "50%").
	// "" or "0%" renders a standard pie; any other value creates a donut.
	InnerSize string
	// StartAngle in degrees.  0 = 3 o'clock (east).
	// Defaults to -90 (12 o'clock / top).
	StartAngle float64
}

// PolarOptions controls polar (spider/radar) chart rendering.
type PolarOptions struct {
	// GridLineInterpolation is "polygon" (default, spider-web rings) or "circle".
	GridLineInterpolation string

	// FillAlpha controls the fill opacity of each series polygon.
	// Range [0,1]: 0 = white fill, 1 = full series color. Defaults to 0.3.
	FillAlpha float64

	// LineWidth is the stroke width in points. Defaults to 2.
	LineWidth float64

	// Marker controls the symbol drawn at each data point.
	Marker *Marker

	// DataLabels configures value labels at each data point.
	DataLabels *DataLabels
}

// ScatterOptions controls scatter chart rendering.
type ScatterOptions struct {
	Marker     *Marker
	DataLabels *DataLabels
}

// BubbleOptions controls bubble chart rendering.
type BubbleOptions struct {
	// MinSize is the minimum bubble radius in points. Defaults to 4.
	MinSize float64
	// MaxSize is the maximum bubble radius in points. Defaults to 30.
	MaxSize float64
	// ZMin overrides the minimum z-value for radius scaling. Derived from data when 0.
	ZMin float64
	// ZMax overrides the maximum z-value for radius scaling. Derived from data when 0.
	ZMax    float64
	DataLabels *DataLabels
}

// HeatmapOptions controls heatmap chart rendering.
type HeatmapOptions struct {
	// BorderWidth is the gap between cells in points. Defaults to 1.
	BorderWidth float64
	// BorderColor is the cell border color. Defaults to white.
	BorderColor *pdf.Color
	// MinColor is the fill color for the lowest cell values. Defaults to white.
	MinColor *pdf.Color
	// MaxColor is the fill color for the highest cell values. Defaults to the series color.
	MaxColor   *pdf.Color
	DataLabels *DataLabels
}

// WaterfallOptions controls waterfall chart rendering.
type WaterfallOptions struct {
	// UpColor is the fill for positive increment bars. Defaults to the series color.
	UpColor *pdf.Color
	// NegativeColor is the fill for negative increment bars. Defaults to red.
	NegativeColor *pdf.Color
	// LineWidth is the bar border width. Defaults to 0.
	LineWidth  float64
	DataLabels *DataLabels
}

// FunnelOptions controls funnel and pyramid chart rendering.
type FunnelOptions struct {
	// NeckWidth is the funnel neck width as a percentage string. Defaults to "30%".
	NeckWidth string
	// NeckHeight is the funnel neck height as a percentage string. Defaults to "25%".
	NeckHeight string
	// Width is the widest section as a percentage string. Defaults to "80%".
	Width string
	// Reversed renders as a pyramid (wide at top) when true.
	Reversed   bool
	DataLabels *DataLabels
}

// GaugePlotBand is a colored arc band on a gauge dial.
type GaugePlotBand struct {
	From  float64
	To    float64
	Color pdf.Color
	// Thickness is the arc width in points. Defaults to 10.
	Thickness float64
}

// GaugeOptions controls gauge and solid-gauge chart rendering.
type GaugeOptions struct {
	// PaneStartAngle is the start of the gauge arc in degrees.
	// 0 = east, counterclockwise positive. Default -150 (10 o'clock position).
	PaneStartAngle float64
	// PaneEndAngle is the end of the gauge arc in degrees. Default 150 (2 o'clock).
	PaneEndAngle float64
	// PlotBands are colored arc sections on the gauge scale.
	PlotBands []GaugePlotBand
	// Solid, when true, draws a solid-gauge (filled arc) instead of a needle gauge.
	Solid      bool
	DataLabels *DataLabels
}

// ErrorbarOptions controls error bar chart rendering.
type ErrorbarOptions struct {
	// LineWidth is the stroke width. Defaults to 1.5.
	LineWidth float64
	// WhiskerLength is the endcap half-width as a fraction of the bucket width.
	// Defaults to 0.25.
	WhiskerLength float64
	// Color overrides the error bar color.
	Color *pdf.Color
}

// BoxplotOptions controls box-and-whisker chart rendering.
type BoxplotOptions struct {
	// FillColor is the box interior color. Defaults to white.
	FillColor *pdf.Color
	// LineWidth is the stroke width. Defaults to 1.5.
	LineWidth float64
	// WhiskerLength is the endcap half-width as a fraction of the box width.
	// Defaults to 0.25.
	WhiskerLength float64
}

// ColumnRangeOptions controls column range chart rendering.
type ColumnRangeOptions struct {
	DataLabels   *DataLabels
	BorderWidth  float64
	BorderColor  *pdf.Color
	GroupPadding float64
	PointPadding float64
}

// AreaRangeOptions controls area range chart rendering.
type AreaRangeOptions struct {
	// FillAlpha controls fill opacity. Defaults to 0.3.
	FillAlpha  float64
	LineWidth  float64
	DataLabels *DataLabels
}

// BulletOptions controls bullet chart rendering.
type BulletOptions struct {
	// PlotBands are qualitative background bands on the value axis.
	PlotBands []GaugePlotBand
	// TargetWidth is the target marker's half-width as a fraction of bar width.
	// Defaults to 0.15.
	TargetWidth float64
	// TargetColor is the target marker color. Defaults to dark gray.
	TargetColor *pdf.Color
	DataLabels  *DataLabels
}

// DumbbellOptions controls dumbbell chart rendering.
type DumbbellOptions struct {
	// LineWidth is the connecting line stroke width. Defaults to 2.
	LineWidth  float64
	Marker     *Marker
	DataLabels *DataLabels
}

// LollipopOptions controls lollipop chart rendering.
type LollipopOptions struct {
	// LineWidth is the stick stroke width. Defaults to 1.5.
	LineWidth  float64
	Marker     *Marker
	DataLabels *DataLabels
}

// TreemapOptions controls treemap chart rendering.
type TreemapOptions struct {
	// ColorByPoint assigns a distinct palette color to each cell. Defaults to true.
	ColorByPoint bool
	// BorderWidth is the gap between cells in points. Defaults to 1.
	BorderWidth float64
	// BorderColor is the cell border color. Defaults to white.
	BorderColor *pdf.Color
	DataLabels  *DataLabels
}

// DataLabels configures value labels rendered next to data points or bars.
type DataLabels struct {
	// Enabled shows labels when true.  Defaults to false.
	Enabled *bool
	// Format is a template where "{y}" is replaced with the value.
	// Defaults to "{y}".
	Format string
	// FontName overrides Options.FontName.
	FontName string
	// FontSize overrides Options.FontSize when > 0.
	FontSize float64
	// Color overrides the label text color.
	Color *pdf.Color
}

// Marker controls the symbol drawn at each data point on line/area charts.
type Marker struct {
	// Enabled shows markers when true.  Defaults to true.
	Enabled *bool
	// Symbol is "circle" (default), "square", or "diamond".
	Symbol string
	// Radius is the marker radius in points.  Defaults to 3.
	Radius float64
}
