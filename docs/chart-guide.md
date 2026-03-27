# Nautilus Charts — Ontwikkelaarsgids

Het pakket `pdf/chart` biedt 20 declaratieve grafiektypen die rechtstreeks op
een `pdf.Document` getekend kunnen worden of via `chart.NewFlowable` in een
layout-story opgenomen kunnen worden.  De API volgt het Highcharts-configuratiemodel,
zodat een grafiek samenstellen neerkomt op het invullen van een `chart.Options`-struct.

Elk grafiektype leeft in een eigen sub-pakket, zodat een binary alleen betaalt
voor de typen die het importeert.

---

## Inhoudsopgave

1. [Snelle start](#1-snelle-start)
2. [Options — top-level configuratie](#2-options--top-level-configuratie)
3. [Assen](#3-assen)
4. [Series en data](#4-series-en-data)
5. [Legenda](#5-legenda)
6. [DataLabels en Markers](#6-datalabels-en-markers)
7. [Grafiektypen](#7-grafiektypen)
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
8. [Inbedden in een layout-story](#8-inbedden-in-een-layout-story)
9. [Rechtstreeks tekenen op een Document](#9-rechtstreeks-tekenen-op-een-document)
10. [Kleuren](#10-kleuren)

---

## 1. Snelle start

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
    if err := doc.RegisterFont("regular", "/pad/naar/NotoSans-Regular.ttf"); err != nil {
        log.Fatal(err)
    }
    doc.AddPage()

    lc := &line.LineChart{
        Options: chart.Options{
            FontName: "regular",
            Title:    &chart.Title{Text: "Maandelijkse omzet"},
            XAxis:    &chart.Axis{Categories: []string{"Jan", "Feb", "Mar", "Apr"}},
            YAxis:    &chart.Axis{},
            Series: []chart.Series{
                {Name: "2023", Data: []float64{120, 150, 130, 180}},
                {Name: "2024", Data: []float64{140, 160, 175, 210}},
            },
        },
    }

    // x=50, y=80 (linksboven van de grafiek), breedte=495, hoogte=200 — alles in punten
    if err := lc.Draw(doc, 50, 80, 495, 200); err != nil {
        log.Fatal(err)
    }

    if err := doc.Save("grafiek.pdf"); err != nil {
        log.Fatal(err)
    }
}
```

---

## 2. Options — top-level configuratie

`chart.Options` wordt gedeeld door elk grafiektype.

```go
type Options struct {
    Title       *Title
    Subtitle    *Title
    XAxis       *Axis
    YAxis       *Axis
    Series      []Series
    Legend      *Legend
    PlotOptions *PlotOptions
    Colors      []pdf.Color   // overschrijft het standaard kleurpalet
    Background  *pdf.Color    // vult het kader wanneer ingesteld
    FontName    string        // geregistreerd lettertype voor alle grafiektekst
    FontSize    float64       // basisgrootte in punten; standaard 9
}
```

### Title / Subtitle

```go
Title: &chart.Title{
    Text:     "Verkoop 2024",
    FontName: "bold",
    FontSize: 14,
    Color:    &pdf.Color{R: 0.2, G: 0.2, B: 0.2},
},
Subtitle: &chart.Title{Text: "Alle regio's samen"},
```

### Hulpfuncties

```go
chart.Float(1.0)  // geeft *float64 terug — voor optionele float-velden
chart.Bool(true)  // geeft *bool terug   — voor optionele bool-velden
```

---

## 3. Assen

```go
XAxis: &chart.Axis{
    Title:         &chart.Title{Text: "Maand"},
    Categories:    []string{"Jan", "Feb", "Mar"},
    Min:           chart.Float(0),
    Max:           chart.Float(500),
    TickInterval:  chart.Float(100),
    GridLineWidth: -1,   // negatief = rasterlijnen verbergen
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

| Veld            | Standaard          | Beschrijving                                               |
|-----------------|--------------------|------------------------------------------------------------|
| `Title`         | —                  | Optioneel aslabel                                          |
| `Categories`    | —                  | Discrete tikkenopschriften; numeriek wanneer nil           |
| `Min` / `Max`   | van data           | Begrens het zichtbare waardebereik                         |
| `TickInterval`  | automatisch        | Vaste afstand tussen rasterlijnen / tikkenopschriften      |
| `GridLineWidth` | 0,5                | Lijndikte in pt; **negatief** verbergt rasterlijnen        |
| `GridLineColor` | lichtgrijs         | Kleur van rasterlijnen                                     |
| `Labels`        | —                  | Configuratie van tikkenopschriften                         |
| `Visible`       | true               | Verberg de as wanneer false                                |

---

## 4. Series en data

### Eenvoudige data (line, area, column, bar, pie)

```go
Series: []chart.Series{
    {Name: "Omzet", Data: []float64{120, 150, 130, 180}},
    {Name: "Kosten", Data: []float64{80, 90, 95, 100}, Color: &pdf.Color{R: 1}},
},
```

### Rijke datapunten (scatter, bubble, heatmap, bereikgrafieken, …)

Gebruik `Points` in plaats van `Data` wanneer een grafiektype meer dan één
waarde per datapunt nodig heeft.

```go
// Scatter
Series: []chart.Series{{
    Name: "Metingen",
    Points: []chart.Point{
        {X: 10, Y: 55},
        {X: 20, Y: 72},
        {X: 35, Y: 48},
    },
}},
```

### Point-velden

| Veld                | Gebruikt door                                         |
|---------------------|-------------------------------------------------------|
| `X`, `Y`            | scatter, bubble, heatmap (kolom-/rij-index)           |
| `Z`                 | bubble (straal), heatmap (celwaarde)                  |
| `Low`, `High`       | columnrange, arearange, errorbar, dumbbell            |
| `Q1`, `Median`, `Q3`| boxplot                                               |
| `Target`            | bullet                                                |
| `Name`              | waterfall-stappen, trechter-fasen, treemap-knooppunten|
| `Color`             | kleuroverride per punt                                |
| `IsSum`             | waterfall — cumulatieve totaalbalk                    |
| `IsIntermediateSum` | waterfall — lopend subtotaal                          |

---

## 5. Legenda

```go
Legend: &chart.Legend{
    Enabled:       chart.Bool(true),
    Layout:        "horizontal",  // "horizontal" (standaard) of "vertical"
    Align:         "center",      // "left", "center" (standaard), "right"
    VerticalAlign: "bottom",      // "top", "middle", "bottom" (standaard)
    FontName:      "regular",
    FontSize:      9,
},
```

Stel `Enabled: chart.Bool(false)` in om de legenda volledig te verbergen.

---

## 6. DataLabels en Markers

### DataLabels

Toont waardelabels naast balken, schijven of datapunten.

```go
PlotOptions: &chart.PlotOptions{
    Column: &chart.ColumnOptions{
        DataLabels: &chart.DataLabels{
            Enabled:  chart.Bool(true),
            Format:   "{y} €",   // {y} of {value} wordt vervangen door de waarde
            FontName: "regular",
            FontSize: 8,
            Color:    &pdf.Color{R: 0.2, G: 0.2, B: 0.2},
        },
    },
},
```

### Marker

Bepaalt het symbool dat bij elk datapunt getekend wordt op lijn-, vlak-,
polaire, scatter- en vergelijkbare grafieken.

```go
PlotOptions: &chart.PlotOptions{
    Line: &chart.LineOptions{
        Marker: &chart.Marker{
            Enabled: chart.Bool(true),
            Symbol:  "circle",  // "circle" (standaard), "square", "diamond"
            Radius:  4,         // straal in punten
        },
    },
},
```

---

## 7. Grafiektypen

### 7.1 line

Pakket: `pdf/chart/line`

Multi-serie lijngrafiek. Gebruikt `Series.Data` (één waarde per categorie).

```go
import "github.com/gvanbeck/nautilus/pdf/chart/line"

lc := &line.LineChart{
    Options: chart.Options{
        FontName: "regular",
        XAxis:    &chart.Axis{Categories: []string{"K1", "K2", "K3", "K4"}},
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
lc.Draw(doc, x, y, breedte, hoogte)
```

### 7.2 area

Pakket: `pdf/chart/area`

Gevulde vlakgrafiek. Zelfde data als lijn. Gebruik `FillAlpha` voor de
transparantie van de vulkleur.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/area"

ac := &area.AreaChart{
    Options: chart.Options{
        /* … zelfde als line … */
        PlotOptions: &chart.PlotOptions{
            Area: &chart.AreaOptions{
                FillAlpha: 0.3,  // 0 = witte vulling, 1 = volledige kleur
                LineWidth: 2,
            },
        },
    },
}
```

### 7.3 column

Pakket: `pdf/chart/column`

Verticale staafgrafiek. Ondersteunt gegroepeerd, gestapeld en procentueel stapelen.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/column"

cc := &column.ColumnChart{
    Options: chart.Options{
        XAxis:  &chart.Axis{Categories: []string{"Jan", "Feb", "Mar"}},
        YAxis:  &chart.Axis{},
        Series: []chart.Series{
            {Name: "Noord", Data: []float64{100, 120, 90}},
            {Name: "Zuid",  Data: []float64{80,  95, 110}},
        },
        PlotOptions: &chart.PlotOptions{
            Column: &chart.ColumnOptions{
                Stacking:     "normal",  // "" (gegroepeerd), "normal", "percent"
                GroupPadding: 0.2,
                PointPadding: 0.1,
                BorderWidth:  1,
            },
        },
    },
}
```

### 7.4 bar

Pakket: `pdf/chart/bar`

Horizontale staafgrafiek. Gebruikt dezelfde `ColumnOptions` (via type-alias).

```go
import "github.com/gvanbeck/nautilus/pdf/chart/bar"

bc := &bar.BarChart{Options: chart.Options{ /* zelfde als column */ }}
```

### 7.5 pie / donut

Pakket: `pdf/chart/pie`

Taartgrafiek. Maak een **donut** door `InnerSize` in te stellen op een
percentage.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/pie"

pc := &pie.PieChart{
    Options: chart.Options{
        Series: []chart.Series{{
            Points: []chart.Point{
                {Name: "Huur",      Y: 35},
                {Name: "Personeel", Y: 25},
                {Name: "Marketing", Y: 20},
                {Name: "Overig",    Y: 20},
            },
        }},
        PlotOptions: &chart.PlotOptions{
            Pie: &chart.PieOptions{
                InnerSize:  "50%",   // "" voor taart, "50%" voor donut
                StartAngle: -90,     // 0 = oost, -90 = boven (12 uur)
                DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
            },
        },
    },
}
```

### 7.6 polar / spider / radar

Pakket: `pdf/chart/polar`

Radargrafiek. Data volgt dezelfde conventie als lijn/vlak.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/polar"

pr := &polar.PolarChart{
    Options: chart.Options{
        XAxis: &chart.Axis{
            Categories: []string{"Snelheid", "Kracht", "Wendbaarheid", "Verdediging", "Uithoudingsvermogen"},
        },
        YAxis:  &chart.Axis{Min: chart.Float(0), Max: chart.Float(100)},
        Series: []chart.Series{
            {Name: "Speler A", Data: []float64{80, 60, 90, 70, 75}},
            {Name: "Speler B", Data: []float64{65, 85, 55, 90, 60}},
        },
        PlotOptions: &chart.PlotOptions{
            Polar: &chart.PolarOptions{
                GridLineInterpolation: "polygon", // "polygon" of "circle"
                FillAlpha:             0.3,
                LineWidth:             2,
            },
        },
    },
}
```

### 7.7 scatter

Pakket: `pdf/chart/scatter`

X-Y puntenwolk. Gebruik `Series.Points` met `X` en `Y`.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/scatter"

sc := &scatter.ScatterChart{
    Options: chart.Options{
        XAxis: &chart.Axis{},
        YAxis: &chart.Axis{},
        Series: []chart.Series{{
            Name: "Steekproef A",
            Points: []chart.Point{
                {X: 10.5, Y: 23.1},
                {X: 18.2, Y: 41.7},
                {X: 27.0, Y: 35.9},
            },
        }},
    },
}
```

### 7.8 bubble

Pakket: `pdf/chart/bubble`

Scatter waarbij `Z` de belstraal bepaalt.

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
                MinSize: 4,
                MaxSize: 30,
            },
        },
    },
}
```

### 7.9 heatmap

Pakket: `pdf/chart/heatmap`

Kleurgecodeerd raster. Gebruik `Points` met `X` (kolomindex), `Y` (rijindex)
en `Z` (celwaarde).

```go
import "github.com/gvanbeck/nautilus/pdf/chart/heatmap"

hm := &heatmap.HeatmapChart{
    Options: chart.Options{
        XAxis: &chart.Axis{Categories: []string{"Ma", "Di", "Wo", "Do", "Vr"}},
        YAxis: &chart.Axis{Categories: []string{"Ochtend", "Middag", "Avond"}},
        Series: []chart.Series{{
            Points: []chart.Point{
                {X: 0, Y: 0, Z: 12}, {X: 1, Y: 0, Z: 8},
                {X: 0, Y: 1, Z: 25}, {X: 1, Y: 1, Z: 31},
            },
        }},
        PlotOptions: &chart.PlotOptions{
            Heatmap: &chart.HeatmapOptions{
                MinColor: &pdf.Color{R: 1, G: 1, B: 1},
                MaxColor: &pdf.Color{R: 0.07, G: 0.44, B: 0.73},
                BorderWidth: 2,
            },
        },
    },
}
```

### 7.10 waterfall

Pakket: `pdf/chart/waterfall`

Waterval-/cascadegrafiek. Markeer samenvattingsbalken met `IsSum` of
`IsIntermediateSum`.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/waterfall"

wf := &waterfall.WaterfallChart{
    Options: chart.Options{
        XAxis: &chart.Axis{},
        YAxis: &chart.Axis{},
        Series: []chart.Series{{
            Points: []chart.Point{
                {Name: "Begin",    Y: 1000},
                {Name: "Verkoop",  Y: 350},
                {Name: "Retouren", Y: -80},
                {Name: "Kosten",   Y: -200},
                {Name: "Totaal",   IsSum: true},
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

Pakket: `pdf/chart/funnel`

Trechtergrafiek. Stel `Reversed: true` in voor een piramide (breed bovenaan).

```go
import "github.com/gvanbeck/nautilus/pdf/chart/funnel"

fc := &funnel.FunnelChart{
    Options: chart.Options{
        Series: []chart.Series{{
            Points: []chart.Point{
                {Name: "Leads",       Y: 5000},
                {Name: "Prospects",   Y: 2500},
                {Name: "Gekwalif.",   Y: 1200},
                {Name: "Gesloten",    Y: 400},
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

Pakket: `pdf/chart/gauge`

Naaldmeter of gevulde boog. Gebruik `PlotBands` voor gekleurde zones en stel
`Solid: true` in voor een solid-gauge.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/gauge"

gc := &gauge.GaugeChart{
    Options: chart.Options{
        YAxis: &chart.Axis{Min: chart.Float(0), Max: chart.Float(100)},
        Series: []chart.Series{{Data: []float64{68}}},
        PlotOptions: &chart.PlotOptions{
            Gauge: &chart.GaugeOptions{
                PaneStartAngle: -150,
                PaneEndAngle:   150,
                Solid:          false,
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

Pakket: `pdf/chart/errorbar`

Foutbalkgrafiek met snorbaard. Gebruik `Points` met `Low` en `High`.

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
                WhiskerLength: 0.25,
            },
        },
    },
}
```

### 7.14 boxplot

Pakket: `pdf/chart/boxplot`

Box-and-whisker plot. Gebruik `Points` met `Low`, `Q1`, `Median`, `Q3`, `High`.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/boxplot"

bp := &boxplot.BoxplotChart{
    Options: chart.Options{
        XAxis: &chart.Axis{Categories: []string{"Groep 1", "Groep 2"}},
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

Pakket: `pdf/chart/columnrange`

Kolombereikgrafiek (laag-naar-hoog balken). Gebruik `Points` met `Low` en `High`.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/columnrange"

cr := &columnrange.ColumnRangeChart{
    Options: chart.Options{
        XAxis: &chart.Axis{Categories: []string{"Jan", "Feb", "Mar"}},
        YAxis: &chart.Axis{},
        Series: []chart.Series{{
            Name: "Temperatuurbereik",
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

Pakket: `pdf/chart/arearange`

Vlakbereikgrafiek (gevuld laag-naar-hoog band). Gebruik `Points` met `Low` en `High`.

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

Pakket: `pdf/chart/bullet`

Bulletgrafiek met een waardebalk en een doelmarkering. Gebruik `PlotBands` voor
kwalitatieve achtergrondstroken.

```go
import "github.com/gvanbeck/nautilus/pdf/chart/bullet"

bc := &bullet.BulletChart{
    Options: chart.Options{
        XAxis: &chart.Axis{Categories: []string{"Omzet", "Winst"}},
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

Pakket: `pdf/chart/dumbbell`

Laag-hoog bereikpunten verbonden door een lijn.

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

Pakket: `pdf/chart/lollipop`

Stok + punt grafiek. Gebruikt `Series.Data` zoals een kolomgrafiek.

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

Pakket: `pdf/chart/treemap`

Hiërarchische rechthoekopvulling. Gebruik `Points` met `Name` en `Y` (grootte).

```go
import "github.com/gvanbeck/nautilus/pdf/chart/treemap"

tm := &treemap.TreemapChart{
    Options: chart.Options{
        Series: []chart.Series{{
            Points: []chart.Point{
                {Name: "Engineering", Y: 45},
                {Name: "Verkoop",     Y: 30},
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

## 8. Inbedden in een layout-story

Alle grafiektypen implementeren `chart.Drawable`. Wikkel elke grafiek in een
`layout.Flowable` met `chart.NewFlowable` om hem in een `DocTemplate`-story
te plaatsen.

```go
import (
    "github.com/gvanbeck/nautilus/pdf/chart"
    "github.com/gvanbeck/nautilus/pdf/chart/line"
    "github.com/gvanbeck/nautilus/pdf/layout"
)

lc := &line.LineChart{Options: chart.Options{ /* … */ }}

story := []layout.Flowable{
    &layout.Paragraph{Text: "Maandelijkse omzet", Style: kopregel},
    &layout.Spacer{Height: 8},

    // breedte=0 vult de volledige kadersbreedte; hoogte=200 is vast in punten
    chart.NewFlowable(lc, 0, 200),

    &layout.Spacer{Height: 16},
    &layout.Paragraph{Text: "Zie onderstaande tabel voor details.", Style: body},
}
```

`chart.NewFlowable(drawable, breedte, hoogte float64) layout.Flowable`

| Parameter  | Betekenis                                                          |
|------------|--------------------------------------------------------------------|
| `drawable` | Elke `chart.Drawable` (bijv. `*line.LineChart`)                    |
| `breedte`  | Gewenste breedte in pt; **0** vult de beschikbare kadersbreedte    |
| `hoogte`   | Vaste grafiakhoogte in pt                                          |

Grafieken kunnen niet over kaders of pagina's worden gesplitst. Als een
grafiek hoger is dan de resterende ruimte, springt de engine naar het volgende
kader voordat hij tekent.

---

## 9. Rechtstreeks tekenen op een Document

Wanneer u de layout-engine niet gebruikt, kunt u `Draw` op elk grafiektype
rechtstreeks aanroepen. Het coördinatenstelsel heeft zijn oorsprong
linksboven op de pagina; Y loopt naar beneden.

```go
doc.AddPage()

links  := &column.ColumnChart{Options: chart.Options{ /* … */ }}
rechts := &pie.PieChart{Options: chart.Options{ /* … */ }}

paginaB := doc.PageWidth()
marge   := 40.0
tussenr := 20.0
helftB  := (paginaB - 2*marge - tussenr) / 2

links.Draw(doc,  marge,              60, helftB, 180)
rechts.Draw(doc, marge+helftB+tussenr, 60, helftB, 180)
```

---

## 10. Kleuren

Het standaard kleurpalet is gedefinieerd in `pdf/chart/colors.go`. Overschrijf
het per grafiek met `Options.Colors`:

```go
Options: chart.Options{
    Colors: []pdf.Color{
        {R: 0.18, G: 0.55, B: 0.84},
        {R: 0.96, G: 0.60, B: 0.07},
        {R: 0.20, G: 0.72, B: 0.40},
    },
},
```

Afzonderlijke series of punten kunnen het palet overschrijven:

```go
// Override op serie-niveau
chart.Series{Name: "Speciaal", Data: []float64{1, 2, 3},
    Color: &pdf.Color{R: 0.8, G: 0.1, B: 0.1}}

// Override op punt-niveau
chart.Point{Name: "Uitbijter", Y: 99,
    Color: &pdf.Color{R: 1, G: 0, B: 0}}
```

Kleuren zijn `pdf.Color`-structs met `R`, `G`, `B`-velden in het bereik `[0, 1]`.
