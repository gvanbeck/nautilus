# Nautilus Charts — Developer Guide

The `pdf/chart` package provides 20 declarative chart types that can be drawn
directly onto a `pdf.Document` or embedded in a layout story via
`chart.NewFlowable`.  The API mirrors the Highcharts JSON configuration model,
so building a chart is a matter of filling in a `chart.Options` struct.

Each chart renderer lives in its own sub-package so that a binary only pays
for the chart types it imports.

---

## Table of Contents

1. [Quick start](#1-quick-start)
2. [Options — top-level configuration](#2-options--top-level-configuration)
3. [Axes](#3-axes)
4. [Series and data](#4-series-and-data)
5. [Legend](#5-legend)
6. [DataLabels and Markers](#6-datalabels-and-markers)
7. [Chart types](#7-chart-types)
   - [line](#71-line)
   - [area](#72-area)
   - [column](#73-column)
   - [bar](#74-bar)
   - [pie / donut](#75-pie--donut)
   - [polar / spider / radar](#76-polar--spider--radar)
   - [scatter](#77-scatter)
   - [bubble](#78-bubble)
   - [heatmap](#79-heatmap)
   - [waterfall](#710-waterfall)
   - [funnel / pyramid](#711-funnel--pyramid)
   - [gauge / solid-gauge](#712-gauge--solid-gauge)
   - [errorbar](#713-errorbar)
   - [boxplot](#714-boxplot)
   - [columnrange](#715-columnrange)
   - [arearange](#716-arearange)
   - [bullet](#717-bullet)
   - [dumbbell](#718-dumbbell)
   - [lollipop](#719-lollipop)
   - [treemap](#720-treemap)
8. [Embedding in a layout story](#8-embedding-in-a-layout-story)
9. [Drawing directly onto a Document](#9-drawing-directly-onto-a-document)
10. [Colors](#10-colors)

---

## 1. Quick start

```go
package main

import (
    "log"

    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/chart"
    "github.com/gvanbeck/nautilus/pdf/chart/line"
)

func main() {
    doc, err := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
    if err != nil {
        log.Fatal(err)
    }
    if err := doc.RegisterFont("regular", "/path/to/NotoSans-Regular.ttf"); err != nil {
        log.Fatal(err)
    }
    doc.AddPage()

    lc := &line.LineChart{
        Options: chart.Options{
            FontName: "regular",
            Title:    &chart.Title{Text: "Monthly Revenue"},
            XAxis:    &chart.Axis{Categories: []string{"Jan", "Feb", "Mar", "Apr"}},
            YAxis:    &chart.Axis{},
            Series: []chart.Series{
                {Name: "2023", Data: []float64{120, 150, 130, 180}},
                {Name: "2024", Data: []float64{140, 160, 175, 210}},
            },
        },
    }

    // x=50, y=80 (top-left of chart), width=495, height=200 — all in points
    if err := lc.Draw(doc, 50, 80, 495, 200); err != nil {
        log.Fatal(err)
    }

    if err := doc.Save("chart.pdf"); err != nil {
        log.Fatal(err)
    }
}
```

---

## 2. Options — top-level configuration

`chart.Options` is shared by every chart type.

```go
type Options struct {
    Title       *Title
    Subtitle    *Title
    XAxis       *Axis
    YAxis       *Axis
    Series      []Series
    Legend      *Legend
    PlotOptions *PlotOptions
    Colors      []pdf.Color   // overrides default palette
    Background  *pdf.Color    // fills the bounding box when set
    FontName    string        // registered font for all chart text
    FontSize    float64       // base size in points; defaults to 9
}
```

### Title / Subtitle

```go
Title: &chart.Title{
    Text:     "Sales 2024",
    FontName: "bold",      // overrides Options.FontName
    FontSize: 14,          // overrides Options.FontSize
    Color:    &pdf.Color{R: 0.2, G: 0.2, B: 0.2},
},
Subtitle: &chart.Title{Text: "All regions combined"},
```

### Helper functions

```go
chart.Float(1.0)  // returns *float64 — for optional float fields
chart.Bool(true)  // returns *bool  — for optional bool fields
```

---

## 3. Axes

```go
XAxis: &chart.Axis{
    Title:         &chart.Title{Text: "Month"},
    Categories:    []string{"Jan", "Feb", "Mar"},
    Min:           chart.Float(0),
    Max:           chart.Float(500),
    TickInterval:  chart.Float(100),
    GridLineWidth: -1,   // negative = hide gridlines
    GridLineColor: &pdf.Color{R: 0.8, G: 0.8, B: 0.8},
    Labels: &chart.AxisLabels{
        Enabled:  chart.Bool(true),
        Format:   "{value} €",
        FontName: "regular",
        FontSize: 8,
    },
    Visible: chart.Bool(true),
},
```

| Field           | Default            | Description                                           |
|-----------------|--------------------|-------------------------------------------------------|
| `Title`         | —                  | Optional axis label                                   |
| `Categories`    | —                  | Discrete tick labels; numeric labels when nil          |
| `Min` / `Max`   | from data          | Clamp the visible value range                         |
| `TickInterval`  | auto               | Fixed spacing between gridlines / tick labels         |
| `GridLineWidth` | 0.5                | Stroke width in pt; **negative** hides gridlines      |
| `GridLineColor` | light gray         | Gridline color                                        |
| `Labels`        | —                  | Tick label configuration                              |
| `Visible`       | true               | Hide the axis when false                              |

---

## 4. Series and data

### Simple data (line, area, column, bar, pie)

```go
Series: []chart.Series{
    {Name: "Revenue", Data: []float64{120, 150, 130, 180}},
    {Name: "Cost",    Data: []float64{80, 90, 95, 100}, Color: &pdf.Color{R: 1}},
},
```

### Rich data points (scatter, bubble, heatmap, range charts, …)

Use `Points` instead of `Data` when a chart type needs more than one value per
data point.

```go
// Scatter
Series: []chart.Series{{
    Name: "Measurements",
    Points: []chart.Point{
        {X: 10, Y: 55},
        {X: 20, Y: 72},
        {X: 35, Y: 48},
    },
}},
```

### Point fields

| Field               | Used by                                         |
|---------------------|-------------------------------------------------|
| `X`, `Y`            | scatter, bubble, heatmap (column/row index)     |
| `Z`                 | bubble (radius), heatmap (cell value)           |
| `Low`, `High`       | columnrange, arearange, errorbar, dumbbell      |
| `Q1`, `Median`, `Q3`| boxplot                                         |
| `Target`            | bullet                                          |
| `Name`              | waterfall steps, funnel stages, treemap nodes   |
| `Color`             | per-point color override                        |
| `IsSum`             | waterfall — cumulative total bar                |
| `IsIntermediateSum` | waterfall — running subtotal bar                |

---

## 5. Legend

```go
Legend: &chart.Legend{
    Enabled:       chart.Bool(true),
    Layout:        "horizontal",  // "horizontal" (default) or "vertical"
    Align:         "center",      // "left", "center" (default), "right"
    VerticalAlign: "bottom",      // "top", "middle", "bottom" (default)
    FontName:      "regular",
    FontSize:      9,
},
```

Set `Enabled: chart.Bool(false)` to hide the legend entirely.

---

## 6. DataLabels and Markers

### DataLabels

Show value labels next to bars, slices, or data points.

```go
PlotOptions: &chart.PlotOptions{
    Column: &chart.ColumnOptions{
        DataLabels: &chart.DataLabels{
            Enabled:  chart.Bool(true),
            Format:   "{y} €",   // {y} or {value} is replaced by the data value
            FontName: "regular",
            FontSize: 8,
            Color:    &pdf.Color{R: 0.2, G: 0.2, B: 0.2},
        },
    },
},
```

### Marker

Controls the symbol drawn at each data point on line, area, polar, scatter, and
similar charts.

```go
PlotOptions: &chart.PlotOptions{
    Line: &chart.LineOptions{
        Marker: &chart.Marker{
            Enabled: chart.Bool(true),
            Symbol:  "circle",  // "circle" (default), "square", "diamond"
            Radius:  4,         // radius in points
        },
    },
},
```

---

## 7. Chart types

### 7.1 line

Package: `pdf/chart/line`

Multi-series line chart. Uses `Series.Data` (one value per category).

```go
import "github.com/gvanbeck/nautilus/pdf/chart/line"

lc := &line.LineChart{
    Options: chart.Options{
        FontName: "regular",
        XAxis:    &chart.Axis{Categories: []string{"Q1", "Q2", "Q3", "Q4"}},
        YAxis:    &chart.Axis{},
        Series: []chart.Series{
            {Name: "Product A", Data: []float64{42, 55, 61, 78}},
            {Name: "Product B", Data: []float64{30, 38, 45, 52}},
        },
        PlotOptions: &chart.PlotOptions{
            Line: &chart.LineOptions{
                LineWidth: 2,
                Marker:    &chart.Marker{Symbol: "circle", Radius: 3},
            },
        },
    },
}
lc.Draw(doc, x, y, width, height)
```

### 7.2 area

Package: `pdf/chart/area`

Filled area chart. Same data as line. Use `FillAlpha` to control fill opacity.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/area"

ac := &area.AreaChart{
    Options: chart.Options{
        /* ... same as line ... */
        PlotOptions: &chart.PlotOptions{
            Area: &chart.AreaOptions{
                FillAlpha: 0.3,  // 0 = white fill, 1 = full color
                LineWidth: 2,
            },
        },
    },
}
```

### 7.3 column

Package: `pdf/chart/column`

Vertical bar chart. Supports grouped, stacked, and percent stacking.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/column"

cc := &column.ColumnChart{
    Options: chart.Options{
        XAxis:  &chart.Axis{Categories: []string{"Jan", "Feb", "Mar"}},
        YAxis:  &chart.Axis{},
        Series: []chart.Series{
            {Name: "North", Data: []float64{100, 120, 90}},
            {Name: "South", Data: []float64{80,  95, 110}},
        },
        PlotOptions: &chart.PlotOptions{
            Column: &chart.ColumnOptions{
                Stacking:     "normal",  // "" (grouped), "normal", "percent"
                GroupPadding: 0.2,       // fraction of category slot
                PointPadding: 0.1,       // fraction of per-series slot
                BorderWidth:  1,
            },
        },
    },
}
```

### 7.4 bar

Package: `pdf/chart/bar`

Horizontal bar chart. Uses the same `ColumnOptions` (via type alias).

```go
import "github.com/gvanbeck/nautilus/pdf/chart/bar"

bc := &bar.BarChart{Options: chart.Options{ /* same as column */ }}
```

### 7.5 pie / donut

Package: `pdf/chart/pie`

Pie chart. To create a **donut**, set `InnerSize` to a non-zero percentage.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/pie"

pc := &pie.PieChart{
    Options: chart.Options{
        Series: []chart.Series{{
            Data: []float64{35, 25, 20, 20},
            Points: []chart.Point{
                {Name: "Rent", Y: 35},
                {Name: "Staff", Y: 25},
                {Name: "Marketing", Y: 20},
                {Name: "Other", Y: 20},
            },
        }},
        PlotOptions: &chart.PlotOptions{
            Pie: &chart.PieOptions{
                InnerSize:  "50%",    // "" for pie, "50%" for donut
                StartAngle: -90,      // 0 = east, -90 = top (12 o'clock)
                DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
            },
        },
    },
}
```

> **Note:** For pie charts, set `Name` in `Points` to label each slice.
> Use `Points` alongside `Data`, or use only `Points` with `Y` filled in.

### 7.6 polar / spider / radar

Package: `pdf/chart/polar`

Radar chart. Data follows the same convention as line/area charts.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/polar"

pr := &polar.PolarChart{
    Options: chart.Options{
        XAxis: &chart.Axis{
            Categories: []string{"Speed", "Power", "Agility", "Defense", "Stamina"},
        },
        YAxis:  &chart.Axis{Min: chart.Float(0), Max: chart.Float(100)},
        Series: []chart.Series{
            {Name: "Player A", Data: []float64{80, 60, 90, 70, 75}},
            {Name: "Player B", Data: []float64{65, 85, 55, 90, 60}},
        },
        PlotOptions: &chart.PlotOptions{
            Polar: &chart.PolarOptions{
                GridLineInterpolation: "polygon", // "polygon" or "circle"
                FillAlpha:             0.3,
                LineWidth:             2,
            },
        },
    },
}
```

### 7.7 scatter

Package: `pdf/chart/scatter`

X-Y point cloud. Use `Series.Points` with `X` and `Y` set.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/scatter"

sc := &scatter.ScatterChart{
    Options: chart.Options{
        XAxis: &chart.Axis{},
        YAxis: &chart.Axis{},
        Series: []chart.Series{{
            Name: "Sample A",
            Points: []chart.Point{
                {X: 10.5, Y: 23.1},
                {X: 18.2, Y: 41.7},
                {X: 27.0, Y: 35.9},
            },
        }},
        PlotOptions: &chart.PlotOptions{
            Scatter: &chart.ScatterOptions{
                Marker: &chart.Marker{Symbol: "circle", Radius: 4},
            },
        },
    },
}
```

### 7.8 bubble

Package: `pdf/chart/bubble`

Scatter chart where `Z` controls the bubble radius.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/bubble"

bc := &bubble.BubbleChart{
    Options: chart.Options{
        XAxis: &chart.Axis{},
        YAxis: &chart.Axis{},
        Series: []chart.Series{{
            Points: []chart.Point{
                {X: 10, Y: 30, Z: 5},
                {X: 20, Y: 50, Z: 20},
                {X: 35, Y: 20, Z: 10},
            },
        }},
        PlotOptions: &chart.PlotOptions{
            Bubble: &chart.BubbleOptions{
                MinSize: 4,   // minimum radius in points
                MaxSize: 30,  // maximum radius in points
            },
        },
    },
}
```

### 7.9 heatmap

Package: `pdf/chart/heatmap`

Color-coded grid. Use `Points` with `X` (column index), `Y` (row index),
and `Z` (cell value).

```go
import "github.com/gvanbeck/nautilus/pdf/chart/heatmap"

hm := &heatmap.HeatmapChart{
    Options: chart.Options{
        XAxis: &chart.Axis{Categories: []string{"Mon", "Tue", "Wed", "Thu", "Fri"}},
        YAxis: &chart.Axis{Categories: []string{"Morning", "Afternoon", "Evening"}},
        Series: []chart.Series{{
            Points: []chart.Point{
                {X: 0, Y: 0, Z: 12}, {X: 1, Y: 0, Z: 8},
                {X: 0, Y: 1, Z: 25}, {X: 1, Y: 1, Z: 31},
                // …
            },
        }},
        PlotOptions: &chart.PlotOptions{
            Heatmap: &chart.HeatmapOptions{
                MinColor: &pdf.Color{R: 1, G: 1, B: 1},            // white = low
                MaxColor: &pdf.Color{R: 0.07, G: 0.44, B: 0.73},   // blue  = high
                BorderWidth: 2,
            },
        },
    },
}
```

### 7.10 waterfall

Package: `pdf/chart/waterfall`

Running total / cascade chart. Mark summary bars with `IsSum` or
`IsIntermediateSum`.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/waterfall"

wf := &waterfall.WaterfallChart{
    Options: chart.Options{
        XAxis: &chart.Axis{},
        YAxis: &chart.Axis{},
        Series: []chart.Series{{
            Points: []chart.Point{
                {Name: "Start",      Y: 1000},
                {Name: "Sales",      Y: 350},
                {Name: "Refunds",    Y: -80},
                {Name: "Costs",      Y: -200},
                {Name: "Total",      IsSum: true},
            },
        }},
        PlotOptions: &chart.PlotOptions{
            Waterfall: &chart.WaterfallOptions{
                NegativeColor: &pdf.Color{R: 0.8, G: 0.1, B: 0.1},
                DataLabels:    &chart.DataLabels{Enabled: chart.Bool(true)},
            },
        },
    },
}
```

### 7.11 funnel / pyramid

Package: `pdf/chart/funnel`

Funnel chart. Set `Reversed: true` to render as a pyramid (wide at top).

```go
import "github.com/gvanbeck/nautilus/pdf/chart/funnel"

fc := &funnel.FunnelChart{
    Options: chart.Options{
        Series: []chart.Series{{
            Points: []chart.Point{
                {Name: "Leads",      Y: 5000},
                {Name: "Prospects",  Y: 2500},
                {Name: "Qualified",  Y: 1200},
                {Name: "Closed",     Y: 400},
            },
        }},
        PlotOptions: &chart.PlotOptions{
            Funnel: &chart.FunnelOptions{
                NeckWidth:  "30%",
                NeckHeight: "25%",
                Width:      "80%",
                Reversed:   false,
                DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
            },
        },
    },
}
```

### 7.12 gauge / solid-gauge

Package: `pdf/chart/gauge`

Needle gauge or solid filled arc. Use `PlotBands` for colored zones and set
`Solid: true` for a solid-gauge.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/gauge"

gc := &gauge.GaugeChart{
    Options: chart.Options{
        YAxis: &chart.Axis{Min: chart.Float(0), Max: chart.Float(100)},
        Series: []chart.Series{{Data: []float64{68}}},
        PlotOptions: &chart.PlotOptions{
            Gauge: &chart.GaugeOptions{
                PaneStartAngle: -150,  // degrees; 0 = east
                PaneEndAngle:   150,
                Solid:          false, // true = solid-gauge (filled arc)
                PlotBands: []chart.GaugePlotBand{
                    {From: 0,  To: 33,  Color: pdf.Color{R: 0.2, G: 0.7, B: 0.2}, Thickness: 12},
                    {From: 33, To: 66,  Color: pdf.Color{R: 1,   G: 0.8, B: 0},   Thickness: 12},
                    {From: 66, To: 100, Color: pdf.Color{R: 0.8, G: 0.1, B: 0.1}, Thickness: 12},
                },
                DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
            },
        },
    },
}
```

### 7.13 errorbar

Package: `pdf/chart/errorbar`

Error bar chart with whisker caps. Use `Points` with `Low` and `High`.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/errorbar"

eb := &errorbar.ErrorbarChart{
    Options: chart.Options{
        XAxis: &chart.Axis{Categories: []string{"A", "B", "C"}},
        YAxis: &chart.Axis{},
        Series: []chart.Series{{
            Points: []chart.Point{
                {Y: 42, Low: 35, High: 49},
                {Y: 55, Low: 47, High: 63},
                {Y: 38, Low: 30, High: 44},
            },
        }},
        PlotOptions: &chart.PlotOptions{
            Errorbar: &chart.ErrorbarOptions{
                LineWidth:     1.5,
                WhiskerLength: 0.25, // fraction of bucket width
            },
        },
    },
}
```

### 7.14 boxplot

Package: `pdf/chart/boxplot`

Box-and-whisker plot. Use `Points` with `Low`, `Q1`, `Median`, `Q3`, `High`.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/boxplot"

bp := &boxplot.BoxplotChart{
    Options: chart.Options{
        XAxis: &chart.Axis{Categories: []string{"Group 1", "Group 2"}},
        YAxis: &chart.Axis{},
        Series: []chart.Series{{
            Points: []chart.Point{
                {Low: 10, Q1: 20, Median: 30, Q3: 40, High: 55},
                {Low: 15, Q1: 25, Median: 38, Q3: 48, High: 60},
            },
        }},
        PlotOptions: &chart.PlotOptions{
            Boxplot: &chart.BoxplotOptions{
                LineWidth:     1.5,
                WhiskerLength: 0.25,
                FillColor:     &pdf.Color{R: 0.9, G: 0.95, B: 1},
            },
        },
    },
}
```

### 7.15 columnrange

Package: `pdf/chart/columnrange`

Column range chart (low-to-high bars). Use `Points` with `Low` and `High`.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/columnrange"

cr := &columnrange.ColumnRangeChart{
    Options: chart.Options{
        XAxis: &chart.Axis{Categories: []string{"Jan", "Feb", "Mar"}},
        YAxis: &chart.Axis{},
        Series: []chart.Series{{
            Name: "Temperature range",
            Points: []chart.Point{
                {Low: -3, High: 8},
                {Low: -1, High: 11},
                {Low: 4,  High: 17},
            },
        }},
    },
}
```

### 7.16 arearange

Package: `pdf/chart/arearange`

Area range chart (filled low-to-high band). Use `Points` with `Low` and `High`.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/arearange"

ar := &arearange.AreaRangeChart{
    Options: chart.Options{
        XAxis: &chart.Axis{Categories: []string{"Jan", "Feb", "Mar"}},
        YAxis: &chart.Axis{},
        Series: []chart.Series{{
            Points: []chart.Point{
                {Low: 2, High: 9},
                {Low: 4, High: 14},
                {Low: 1, High: 10},
            },
        }},
        PlotOptions: &chart.PlotOptions{
            AreaRange: &chart.AreaRangeOptions{FillAlpha: 0.3},
        },
    },
}
```

### 7.17 bullet

Package: `pdf/chart/bullet`

Bullet chart with a value bar and a target marker. Use `PlotBands` for
qualitative background bands.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/bullet"

bc := &bullet.BulletChart{
    Options: chart.Options{
        XAxis: &chart.Axis{Categories: []string{"Revenue", "Profit"}},
        YAxis: &chart.Axis{Min: chart.Float(0), Max: chart.Float(200)},
        Series: []chart.Series{{
            Points: []chart.Point{
                {Y: 135, Target: 150},
                {Y: 52,  Target: 60},
            },
        }},
        PlotOptions: &chart.PlotOptions{
            Bullet: &chart.BulletOptions{
                PlotBands: []chart.GaugePlotBand{
                    {From: 0,   To: 100, Color: pdf.Color{R: 0.85, G: 0.85, B: 0.85}},
                    {From: 100, To: 150, Color: pdf.Color{R: 0.7,  G: 0.7,  B: 0.7}},
                    {From: 150, To: 200, Color: pdf.Color{R: 0.55, G: 0.55, B: 0.55}},
                },
                TargetWidth: 0.15,
            },
        },
    },
}
```

### 7.18 dumbbell

Package: `pdf/chart/dumbbell`

Low-high range dots connected by a line.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/dumbbell"

db := &dumbbell.DumbbellChart{
    Options: chart.Options{
        XAxis: &chart.Axis{Categories: []string{"2020", "2021", "2022"}},
        YAxis: &chart.Axis{},
        Series: []chart.Series{{
            Points: []chart.Point{
                {Low: 20, High: 60},
                {Low: 25, High: 70},
                {Low: 30, High: 75},
            },
        }},
        PlotOptions: &chart.PlotOptions{
            Dumbbell: &chart.DumbbellOptions{LineWidth: 2},
        },
    },
}
```

### 7.19 lollipop

Package: `pdf/chart/lollipop`

Stick + dot chart. Uses `Series.Data` like a column chart.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/lollipop"

lp := &lollipop.LollipopChart{
    Options: chart.Options{
        XAxis: &chart.Axis{Categories: []string{"A", "B", "C", "D"}},
        YAxis: &chart.Axis{},
        Series: []chart.Series{
            {Name: "Score", Data: []float64{42, 67, 53, 81}},
        },
        PlotOptions: &chart.PlotOptions{
            Lollipop: &chart.LollipopOptions{
                LineWidth: 1.5,
                Marker:    &chart.Marker{Symbol: "circle", Radius: 5},
            },
        },
    },
}
```

### 7.20 treemap

Package: `pdf/chart/treemap`

Hierarchical rectangle packing. Use `Points` with `Name` and `Y` (size).

```go
import "github.com/gvanbeck/nautilus/pdf/chart/treemap"

tm := &treemap.TreemapChart{
    Options: chart.Options{
        Series: []chart.Series{{
            Points: []chart.Point{
                {Name: "Engineering", Y: 45},
                {Name: "Sales",       Y: 30},
                {Name: "Marketing",   Y: 15},
                {Name: "Support",     Y: 10},
            },
        }},
        PlotOptions: &chart.PlotOptions{
            Treemap: &chart.TreemapOptions{
                ColorByPoint: true,
                BorderWidth:  2,
                DataLabels:   &chart.DataLabels{Enabled: chart.Bool(true)},
            },
        },
    },
}
```

---

## 8. Embedding in a layout story

All chart types implement `chart.Drawable`. Wrap any chart in a
`layout.Flowable` with `chart.NewFlowable` to place it in a
`DocTemplate` story.

```go
import (
    "github.com/gvanbeck/nautilus/pdf/chart"
    "github.com/gvanbeck/nautilus/pdf/chart/line"
    "github.com/gvanbeck/nautilus/pdf/layout"
)

lc := &line.LineChart{Options: chart.Options{ /* … */ }}

story := []layout.Flowable{
    &layout.Paragraph{Text: "Monthly Revenue", Style: headingStyle},
    &layout.Spacer{Height: 8},

    // width=0 fills the full frame width; height=200 is fixed in points
    chart.NewFlowable(lc, 0, 200),

    &layout.Spacer{Height: 16},
    &layout.Paragraph{Text: "See table below for details.", Style: bodyStyle},
}
```

`chart.NewFlowable(drawable, width, height float64) layout.Flowable`

| Parameter | Meaning                                                |
|-----------|--------------------------------------------------------|
| `drawable`| Any `chart.Drawable` (e.g. `*line.LineChart`)          |
| `width`   | Desired width in pt; **0** fills the available frame   |
| `height`  | Fixed chart height in pt                               |

Charts cannot be split across frames or pages. If a chart is taller than the
remaining space, the engine advances to the next frame before drawing it.

---

## 9. Drawing directly onto a Document

When you don't use the layout engine you can call `Draw` on any chart type
directly. The coordinate system has its origin at the **top-left** of the page;
Y increases downward.

```go
doc.AddPage()

// Draw two charts side by side on the same page.
left  := &column.ColumnChart{Options: chart.Options{ /* … */ }}
right := &pie.PieChart{Options: chart.Options{ /* … */ }}

pageW := doc.PageWidth()
margin := 40.0
gap    := 20.0
halfW  := (pageW - 2*margin - gap) / 2

left.Draw(doc,  margin,          60, halfW, 180)
right.Draw(doc, margin+halfW+gap, 60, halfW, 180)
```

---

## 10. Colors

The default color palette is defined in `pdf/chart/colors.go`. Override it
per chart with `Options.Colors`:

```go
Options: chart.Options{
    Colors: []pdf.Color{
        {R: 0.18, G: 0.55, B: 0.84},
        {R: 0.96, G: 0.60, B: 0.07},
        {R: 0.20, G: 0.72, B: 0.40},
    },
},
```

Individual series or points can override the palette:

```go
// Series-level override
chart.Series{Name: "Special", Data: []float64{1, 2, 3},
    Color: &pdf.Color{R: 0.8, G: 0.1, B: 0.1}}

// Point-level override
chart.Point{Name: "Outlier", Y: 99,
    Color: &pdf.Color{R: 1, G: 0, B: 0}}
```

Colors are `pdf.Color` structs with `R`, `G`, `B` fields in the range `[0, 1]`.
