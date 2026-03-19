// Command chart demonstrates all Nautilus chart types rendered directly into a PDF.
//
// Usage:
//
//	go run ./examples/chart -font /path/to/regular.ttf -bold /path/to/bold.ttf -out output.pdf
package main

import (
	"flag"
	"log"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/chart/area"
	"github.com/gvanbeck/nautilus/pdf/chart/arearange"
	"github.com/gvanbeck/nautilus/pdf/chart/bar"
	"github.com/gvanbeck/nautilus/pdf/chart/boxplot"
	"github.com/gvanbeck/nautilus/pdf/chart/bubble"
	"github.com/gvanbeck/nautilus/pdf/chart/bullet"
	"github.com/gvanbeck/nautilus/pdf/chart/column"
	"github.com/gvanbeck/nautilus/pdf/chart/columnrange"
	"github.com/gvanbeck/nautilus/pdf/chart/dumbbell"
	"github.com/gvanbeck/nautilus/pdf/chart/errorbar"
	"github.com/gvanbeck/nautilus/pdf/chart/funnel"
	"github.com/gvanbeck/nautilus/pdf/chart/gauge"
	"github.com/gvanbeck/nautilus/pdf/chart/heatmap"
	"github.com/gvanbeck/nautilus/pdf/chart/line"
	"github.com/gvanbeck/nautilus/pdf/chart/lollipop"
	"github.com/gvanbeck/nautilus/pdf/chart/pie"
	"github.com/gvanbeck/nautilus/pdf/chart/polar"
	"github.com/gvanbeck/nautilus/pdf/chart/scatter"
	"github.com/gvanbeck/nautilus/pdf/chart/treemap"
	"github.com/gvanbeck/nautilus/pdf/chart/waterfall"
	"github.com/gvanbeck/nautilus/pdf/layout"
)

func main() {
	fontPath := flag.String("font", "", "Path to regular TTF/OTF font (required)")
	boldPath := flag.String("bold", "", "Path to bold TTF/OTF font (optional)")
	outPath := flag.String("out", "chart_output.pdf", "Output PDF path")
	flag.Parse()

	if *fontPath == "" {
		log.Fatal("usage: -font <path/to/font.ttf> [-bold <path/to/bold.ttf>] [-out output.pdf]")
	}

	doc, err := pdf.New(pdf.Config{
		PageSize: pdf.PageSizeA4,
		Margins:  pdf.UniformMargins(40),
	})
	if err != nil {
		log.Fatalf("create document: %v", err)
	}

	if err := doc.RegisterFont("regular", *fontPath); err != nil {
		log.Fatalf("register font: %v", err)
	}
	boldFont := "regular"
	if *boldPath != "" {
		if err := doc.RegisterFont("bold", *boldPath); err != nil {
			log.Fatalf("register bold font: %v", err)
		}
		boldFont = "bold"
	}
	doc.SetFont("regular", 11) //nolint:errcheck

	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}
	baseOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		XAxis:    &chart.Axis{Categories: months},
		YAxis:    &chart.Axis{},
		Legend:   &chart.Legend{},
	}

	// ── Line chart ──────────────────────────────────────────────────────────
	lineOpts := baseOpts
	lineOpts.Title = &chart.Title{Text: "Monthly Revenue", FontName: boldFont, FontSize: 11}
	lineOpts.Subtitle = &chart.Title{Text: "2023 vs 2024"}
	lineOpts.Series = []chart.Series{
		{Name: "2023", Data: []float64{120, 150, 130, 180, 160, 200}},
		{Name: "2024", Data: []float64{140, 165, 175, 195, 210, 240}},
	}
	lc := &line.LineChart{Options: lineOpts}

	// ── Area chart ───────────────────────────────────────────────────────────
	areaOpts := baseOpts
	areaOpts.Title = &chart.Title{Text: "Website Visitors", FontName: boldFont, FontSize: 11}
	areaOpts.Series = []chart.Series{
		{Name: "Mobile",  Data: []float64{800, 950, 1100, 1050, 1200, 1400}},
		{Name: "Desktop", Data: []float64{500, 580, 620,  700,  680,  750}},
	}
	areaOpts.PlotOptions = &chart.PlotOptions{
		Area: &chart.AreaOptions{FillAlpha: 0.25},
	}
	ac := &area.AreaChart{Options: areaOpts}

	// ── Column chart ─────────────────────────────────────────────────────────
	colOpts := baseOpts
	colOpts.Title = &chart.Title{Text: "Quarterly Sales by Region", FontName: boldFont, FontSize: 11}
	colOpts.XAxis = &chart.Axis{Categories: []string{"Q1", "Q2", "Q3", "Q4"}}
	colOpts.Series = []chart.Series{
		{Name: "North", Data: []float64{43, 55, 57, 60}},
		{Name: "South", Data: []float64{23, 35, 41, 47}},
		{Name: "West",  Data: []float64{31, 28, 38, 44}},
	}
	cc := &column.ColumnChart{Options: colOpts}

	// ── Stacked column chart ─────────────────────────────────────────────────
	stackedOpts := colOpts
	stackedOpts.Title = &chart.Title{Text: "Stacked Regional Sales", FontName: boldFont, FontSize: 11}
	stackedOpts.PlotOptions = &chart.PlotOptions{
		Column: &chart.ColumnOptions{Stacking: "normal"},
	}
	sc := &column.ColumnChart{Options: stackedOpts}

	// ── Horizontal bar chart ─────────────────────────────────────────────────
	barOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Top 5 Products", FontName: boldFont, FontSize: 11},
		XAxis:    &chart.Axis{},
		YAxis: &chart.Axis{
			Categories: []string{"Product A", "Product B", "Product C", "Product D", "Product E"},
		},
		Series: []chart.Series{
			{Name: "Units Sold", Data: []float64{143, 112, 98, 87, 76}},
		},
		Legend: &chart.Legend{Enabled: chart.Bool(false)},
	}
	bc := &bar.BarChart{Options: barOpts}

	// ── Pie chart ────────────────────────────────────────────────────────────
	pieOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Browser Market Share", FontName: boldFont, FontSize: 11},
		Series: []chart.Series{
			{Name: "Chrome",  Data: []float64{65}},
			{Name: "Firefox", Data: []float64{15}},
			{Name: "Safari",  Data: []float64{12}},
			{Name: "Other",   Data: []float64{8}},
		},
		Legend: &chart.Legend{},
		PlotOptions: &chart.PlotOptions{
			Pie: &chart.PieOptions{
				DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
			},
		},
	}
	pc := &pie.PieChart{Options: pieOpts}

	// ── Donut chart ──────────────────────────────────────────────────────────
	donutOpts := pieOpts
	donutOpts.Title = &chart.Title{Text: "Browser Market Share (Donut)", FontName: boldFont, FontSize: 11}
	donutOpts.PlotOptions = &chart.PlotOptions{
		Pie: &chart.PieOptions{
			InnerSize:  "50%",
			DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
		},
	}
	dc := &pie.PieChart{Options: donutOpts}

	// ── Scatter chart ────────────────────────────────────────────────────────
	scatterOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Height vs Weight", FontName: boldFont, FontSize: 11},
		XAxis:    &chart.Axis{},
		YAxis:    &chart.Axis{},
		Legend:   &chart.Legend{},
		Series: []chart.Series{
			{Name: "Group A", Points: []chart.Point{
				{X: 161, Y: 51}, {X: 167, Y: 59}, {X: 159, Y: 49}, {X: 175, Y: 72},
				{X: 180, Y: 81}, {X: 166, Y: 63}, {X: 172, Y: 68}, {X: 155, Y: 46},
			}},
			{Name: "Group B", Points: []chart.Point{
				{X: 183, Y: 88}, {X: 170, Y: 77}, {X: 177, Y: 79}, {X: 164, Y: 65},
				{X: 178, Y: 84}, {X: 169, Y: 70}, {X: 185, Y: 90}, {X: 156, Y: 53},
			}},
		},
	}
	scatterC := &scatter.ScatterChart{Options: scatterOpts}

	// ── Bubble chart ─────────────────────────────────────────────────────────
	bubbleOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Countries — GDP, Life Expectancy, Population", FontName: boldFont, FontSize: 11},
		XAxis:    &chart.Axis{},
		YAxis:    &chart.Axis{},
		Legend:   &chart.Legend{},
		Series: []chart.Series{
			{Name: "Europe", Points: []chart.Point{
				{X: 54, Y: 78, Z: 66, Name: "Germany"},
				{X: 44, Y: 77, Z: 55, Name: "France"},
				{X: 37, Y: 77, Z: 27, Name: "Netherlands"},
				{X: 31, Y: 78, Z: 12, Name: "Belgium"},
			}},
			{Name: "Asia", Points: []chart.Point{
				{X: 16, Y: 74, Z: 138, Name: "China"},
				{X: 3, Y: 69, Z: 136, Name: "India"},
				{X: 42, Y: 82, Z: 13, Name: "Japan"},
			}},
		},
	}
	bubbleC := &bubble.BubbleChart{Options: bubbleOpts}

	// ── Heatmap ───────────────────────────────────────────────────────────────
	salesPeople := []string{"Alexander", "Marie", "Maximilian", "Sophia", "Lukas"}
	weekdays := []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
	heatData := []chart.Point{}
	rawHeat := [][]float64{
		{0, 1, 0, 5, 1, 5},
		{1, 0, 0, 2, 4, 3},
		{0, 1, 2, 0, 5, 0},
		{3, 1, 0, 0, 2, 3},
		{0, 0, 4, 1, 0, 4},
	}
	for row, vals := range rawHeat {
		for col, v := range vals {
			if col < len(salesPeople) {
				heatData = append(heatData, chart.Point{X: float64(col), Y: float64(row), Z: v})
			}
		}
	}
	heatOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Sales per Employee per Weekday", FontName: boldFont, FontSize: 11},
		XAxis:    &chart.Axis{Categories: salesPeople},
		YAxis:    &chart.Axis{Categories: weekdays},
		Series:   []chart.Series{{Name: "Sales", Points: heatData}},
		PlotOptions: &chart.PlotOptions{
			Heatmap: &chart.HeatmapOptions{
				MaxColor:   &pdf.Color{R: 7, G: 75, B: 154},
				DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
			},
		},
	}
	heatC := &heatmap.HeatmapChart{Options: heatOpts}

	// ── Waterfall chart ───────────────────────────────────────────────────────
	waterfallOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Company Financials", FontName: boldFont, FontSize: 11},
		YAxis:    &chart.Axis{},
		Series: []chart.Series{{
			Points: []chart.Point{
				{Name: "Start",      Y: 120000},
				{Name: "Revenue",    Y: 569000},
				{Name: "Costs",      Y: -342000},
				{Name: "Subtotal",   IsIntermediateSum: true},
				{Name: "More costs", Y: -233000},
				{Name: "Balance",    IsSum: true},
			},
		}},
	}
	waterfallC := &waterfall.WaterfallChart{Options: waterfallOpts}

	// ── Funnel chart ──────────────────────────────────────────────────────────
	funnelOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Sales Funnel", FontName: boldFont, FontSize: 11},
		Series: []chart.Series{{
			Points: []chart.Point{
				{Name: "Website visits",       Y: 15654},
				{Name: "Downloads",            Y: 4064},
				{Name: "Price list requested", Y: 1987},
				{Name: "Invoice sent",         Y: 976},
				{Name: "Finalized",            Y: 846},
			},
		}},
		Legend: &chart.Legend{},
		PlotOptions: &chart.PlotOptions{
			Funnel: &chart.FunnelOptions{
				DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
			},
		},
	}
	funnelC := &funnel.FunnelChart{Options: funnelOpts}

	// ── Gauge chart ───────────────────────────────────────────────────────────
	gaugeOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Speedometer", FontName: boldFont, FontSize: 11},
		YAxis:    &chart.Axis{Min: chart.Float(0), Max: chart.Float(200)},
		Series:   []chart.Series{{Name: "Speed km/h", Data: []float64{120}}},
		PlotOptions: &chart.PlotOptions{Gauge: &chart.GaugeOptions{
			PaneStartAngle: -150,
			PaneEndAngle:   150,
			PlotBands: []chart.GaugePlotBand{
				{From: 0, To: 80, Color: pdf.Color{R: 85, G: 191, B: 59}, Thickness: 12},
				{From: 80, To: 140, Color: pdf.Color{R: 221, G: 223, B: 13}, Thickness: 12},
				{From: 140, To: 200, Color: pdf.Color{R: 223, G: 83, B: 83}, Thickness: 12},
			},
			DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
		}},
	}
	gaugeC := &gauge.GaugeChart{Options: gaugeOpts}

	// ── Solid gauge ───────────────────────────────────────────────────────────
	solidGaugeOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Activity — 84%", FontName: boldFont, FontSize: 11},
		YAxis:    &chart.Axis{Min: chart.Float(0), Max: chart.Float(100)},
		Series:   []chart.Series{{Name: "Activity", Data: []float64{84}}},
		PlotOptions: &chart.PlotOptions{Gauge: &chart.GaugeOptions{
			PaneStartAngle: -90,
			PaneEndAngle:   90,
			Solid:          true,
			PlotBands: []chart.GaugePlotBand{
				{From: 0, To: 100, Color: pdf.Color{R: 230, G: 230, B: 230}, Thickness: 20},
			},
			DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
		}},
	}
	solidGaugeC := &gauge.GaugeChart{Options: solidGaugeOpts}

	// ── Error bar chart ───────────────────────────────────────────────────────
	errorbarOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Rainfall with Error Bars", FontName: boldFont, FontSize: 11},
		XAxis:    &chart.Axis{Categories: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}},
		YAxis:    &chart.Axis{},
		Legend:   &chart.Legend{},
		Series: []chart.Series{{
			Name: "Error range",
			Points: []chart.Point{
				{Low: 48, High: 51}, {Low: 68, High: 73},
				{Low: 92, High: 110}, {Low: 178, High: 220},
				{Low: 168, High: 200}, {Low: 140, High: 162},
			},
		}},
	}
	errorbarC := &errorbar.ErrorbarChart{Options: errorbarOpts}

	// ── Box plot ──────────────────────────────────────────────────────────────
	boxOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Box Plot — Observations", FontName: boldFont, FontSize: 11},
		XAxis:    &chart.Axis{Categories: []string{"Loc A", "Loc B", "Loc C", "Loc D", "Loc E"}},
		YAxis:    &chart.Axis{},
		Legend:   &chart.Legend{},
		Series: []chart.Series{{
			Name: "Observations",
			Points: []chart.Point{
				{Low: 760, Q1: 801, Median: 848, Q3: 895, High: 965},
				{Low: 733, Q1: 853, Median: 939, Q3: 980, High: 1080},
				{Low: 714, Q1: 762, Median: 817, Q3: 870, High: 918},
				{Low: 724, Q1: 802, Median: 806, Q3: 871, High: 950},
				{Low: 747, Q1: 835, Median: 882, Q3: 910, High: 980},
			},
		}},
	}
	boxC := &boxplot.BoxplotChart{Options: boxOpts}

	// ── Column range ──────────────────────────────────────────────────────────
	colRangeOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Temperature Range per Month", FontName: boldFont, FontSize: 11},
		XAxis:    &chart.Axis{Categories: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}},
		YAxis:    &chart.Axis{},
		Legend:   &chart.Legend{},
		Series: []chart.Series{{
			Name: "Temp °C",
			Points: []chart.Point{
				{Low: -9.5, High: 8.0}, {Low: -7.8, High: 8.3},
				{Low: -4.1, High: 13.0}, {Low: 0.4, High: 18.2},
				{Low: 4.6, High: 22.7}, {Low: 9.0, High: 26.4},
			},
		}},
	}
	colRangeC := &columnrange.ColumnRangeChart{Options: colRangeOpts}

	// ── Area range ────────────────────────────────────────────────────────────
	areaRangeOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Temperature Band", FontName: boldFont, FontSize: 11},
		XAxis:    &chart.Axis{Categories: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}},
		YAxis:    &chart.Axis{},
		Legend:   &chart.Legend{},
		Series: []chart.Series{{
			Name: "Temp range",
			Points: []chart.Point{
				{Low: -9.5, High: 8.0}, {Low: -7.8, High: 8.3},
				{Low: -4.1, High: 13.0}, {Low: 0.4, High: 18.2},
				{Low: 4.6, High: 22.7}, {Low: 9.0, High: 26.4},
			},
		}},
	}
	areaRangeC := &arearange.AreaRangeChart{Options: areaRangeOpts}

	// ── Bullet chart ──────────────────────────────────────────────────────────
	bulletOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Quarterly Performance", FontName: boldFont, FontSize: 11},
		YAxis:    &chart.Axis{Min: chart.Float(0), Max: chart.Float(300)},
		Series: []chart.Series{{
			Points: []chart.Point{
				{Name: "Q1", Y: 180, Target: 220},
				{Name: "Q2", Y: 210, Target: 200},
				{Name: "Q3", Y: 150, Target: 240},
				{Name: "Q4", Y: 260, Target: 250},
			},
		}},
		PlotOptions: &chart.PlotOptions{Bullet: &chart.BulletOptions{
			PlotBands: []chart.GaugePlotBand{
				{From: 0, To: 150, Color: pdf.Color{R: 200, G: 200, B: 200}},
				{From: 150, To: 225, Color: pdf.Color{R: 175, G: 175, B: 175}},
				{From: 225, To: 300, Color: pdf.Color{R: 155, G: 155, B: 155}},
			},
		}},
	}
	bulletC := &bullet.BulletChart{Options: bulletOpts}

	// ── Dumbbell chart ────────────────────────────────────────────────────────
	dumbbellOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Life Expectancy 1990 vs 2020", FontName: boldFont, FontSize: 11},
		XAxis:    &chart.Axis{Categories: []string{"Austria", "Belgium", "Germany", "France", "Netherlands"}},
		YAxis:    &chart.Axis{},
		Legend:   &chart.Legend{},
		Series: []chart.Series{{
			Points: []chart.Point{
				{Low: 70.1, High: 81.3},
				{Low: 71.0, High: 81.9},
				{Low: 70.8, High: 81.2},
				{Low: 70.5, High: 82.3},
				{Low: 71.3, High: 81.7},
			},
		}},
	}
	dumbbellC := &dumbbell.DumbbellChart{Options: dumbbellOpts}

	// ── Lollipop chart ────────────────────────────────────────────────────────
	lollipopOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Top 5 Products by Units Sold", FontName: boldFont, FontSize: 11},
		XAxis:    &chart.Axis{Categories: []string{"Product A", "Product B", "Product C", "Product D", "Product E"}},
		YAxis:    &chart.Axis{},
		Series:   []chart.Series{{Name: "Units", Data: []float64{143, 112, 98, 87, 76}}},
		Legend:   &chart.Legend{Enabled: chart.Bool(false)},
	}
	lollipopC := &lollipop.LollipopChart{Options: lollipopOpts}

	// ── Treemap ───────────────────────────────────────────────────────────────
	treemapOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Global Revenue by Region", FontName: boldFont, FontSize: 11},
		Series: []chart.Series{{
			Points: []chart.Point{
				{Name: "North America", Y: 42},
				{Name: "Europe",        Y: 35},
				{Name: "Asia Pacific",  Y: 18},
				{Name: "Latin America", Y: 3},
				{Name: "Middle East",   Y: 1.5},
				{Name: "Africa",        Y: 0.5},
			},
		}},
	}
	treemapC := &treemap.TreemapChart{Options: treemapOpts}

	// ── Polar (spider/radar) chart ──────────────────────────────────────────
	// Mirrors the Highcharts "Budget vs spending" polar demo.
	polarOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		Title:    &chart.Title{Text: "Budget vs Spending", FontName: boldFont, FontSize: 11},
		XAxis: &chart.Axis{
			Categories: []string{
				"Sales", "Marketing", "Development",
				"Customer Support", "IT", "Administration",
			},
		},
		YAxis:  &chart.Axis{Min: chart.Float(0)},
		Legend: &chart.Legend{},
		Series: []chart.Series{
			{Name: "Allocated Budget", Data: []float64{43000, 19000, 60000, 35000, 17000, 10000}},
			{Name: "Actual Spending",  Data: []float64{50000, 39000, 42000, 31000, 26000, 14000}},
		},
		PlotOptions: &chart.PlotOptions{
			Polar: &chart.PolarOptions{GridLineInterpolation: "polygon"},
		},
	}
	polarC := &polar.PolarChart{Options: polarOpts}

	// ── Build the story ──────────────────────────────────────────────────────
	chartH := 220.0
	halfH := 200.0
	story := []layout.Flowable{
		// Page 1: line + area
		chart.NewFlowable(lc, 0, chartH),
		&layout.Spacer{Height: 20},
		chart.NewFlowable(ac, 0, chartH),

		// Page 2: column + stacked column
		&layout.PageBreak{},
		chart.NewFlowable(cc, 0, chartH),
		&layout.Spacer{Height: 20},
		chart.NewFlowable(sc, 0, chartH),

		// Page 3: bar + pie + donut
		&layout.PageBreak{},
		chart.NewFlowable(bc, 0, halfH),
		&layout.Spacer{Height: 20},
		chart.NewFlowable(pc, 0, halfH),
		&layout.Spacer{Height: 20},
		chart.NewFlowable(dc, 0, halfH),

		// Page 4: polar / spider chart
		&layout.PageBreak{},
		chart.NewFlowable(polarC, 0, chartH),

		// Page 5: scatter + bubble
		&layout.PageBreak{},
		chart.NewFlowable(scatterC, 0, chartH),
		&layout.Spacer{Height: 20},
		chart.NewFlowable(bubbleC, 0, chartH),

		// Page 6: heatmap + waterfall
		&layout.PageBreak{},
		chart.NewFlowable(heatC, 0, chartH),
		&layout.Spacer{Height: 20},
		chart.NewFlowable(waterfallC, 0, chartH),

		// Page 7: funnel + gauge + solid gauge
		&layout.PageBreak{},
		chart.NewFlowable(funnelC, 0, chartH),
		&layout.Spacer{Height: 20},
		chart.NewFlowable(gaugeC, 0, halfH),
		&layout.Spacer{Height: 10},
		chart.NewFlowable(solidGaugeC, 0, halfH),

		// Page 8: error bar + box plot
		&layout.PageBreak{},
		chart.NewFlowable(errorbarC, 0, chartH),
		&layout.Spacer{Height: 20},
		chart.NewFlowable(boxC, 0, chartH),

		// Page 9: column range + area range
		&layout.PageBreak{},
		chart.NewFlowable(colRangeC, 0, chartH),
		&layout.Spacer{Height: 20},
		chart.NewFlowable(areaRangeC, 0, chartH),

		// Page 10: bullet + dumbbell + lollipop
		&layout.PageBreak{},
		chart.NewFlowable(bulletC, 0, halfH),
		&layout.Spacer{Height: 20},
		chart.NewFlowable(dumbbellC, 0, halfH),
		&layout.Spacer{Height: 20},
		chart.NewFlowable(lollipopC, 0, halfH),

		// Page 11: treemap
		&layout.PageBreak{},
		chart.NewFlowable(treemapC, 0, chartH),
	}

	frame := &layout.LayoutFrame{
		X:      doc.ContentX(),
		Y:      doc.ContentY(),
		Width:  doc.ContentWidth(),
		Height: doc.ContentHeight(),
	}
	tmpl := &layout.PageTemplate{ID: "main", Frames: []*layout.LayoutFrame{frame}}
	dt := layout.NewDocTemplate(doc)
	dt.AddPageTemplate(tmpl)

	if err := dt.Build(story); err != nil {
		log.Fatalf("build document: %v", err)
	}

	if err := doc.Save(*outPath); err != nil {
		log.Fatalf("save: %v", err)
	}
	log.Printf("saved %s", *outPath)
}
