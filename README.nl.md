# Nautilus

Een pure-Go PDF-generatiebibliotheek met ondersteuning voor Unicode-tekst, emoji-weergave,
tabellen, frames, randen, kop- en voetteksten, inline HTML-opmaak, tekst van rechts naar links
(Arabisch en Hebreeuws), een high-level layout-engine en 20 grafiektypen.

Gebouwd op basis van [gopdf](https://github.com/signintech/gopdf).

## Functies

- **Papierformaten** — A3, A4, A5, Letter, Legal (staand); aangepaste afmetingen worden ondersteund
- **Marges** — per-zijde instelbare paginamarges met inhoudsgebied-accessors (`ContentX`, `ContentY`, `ContentWidth`, `ContentHeight`, `ContentRightX`, `ContentBottomY`)
- **Lettertype-ondersteuning** — TTF- en OTF-lettertypen; registreer meerdere lettertypen (regulier, vet, cursief) en wissel er vrij tussen
- **Volledig Unicode** — Latijns uitgebreid, CJK, Cyrillisch, Grieks, Arabisch, Hebreeuws en meer
- **Emoji** — inline PNG-vervanging via grafeem-clusterresolutie (compatibel met Noto Emoji)
- **Tekstweergave** — één regel (`WriteLine`) en afgebroken tekst (`WriteText`) met instelbare regelafstand
- **Tekst van rechts naar links** — Arabische contextuele lettertekenvorm (presentatievormen + lam-alef-ligaturen) en Unicode BiDi-herordening voor Arabisch en Hebreeuws
- **Inline HTML-opmaak** — zet `<b>`, `<i>`, `<u>` en inline tags met klassen om naar opgemaakte tekst-spans
- **Tabellen** — kolomspannen, rijspannen, per-cel opmaak, horizontale/verticale uitlijning, automatische paginaoverloop
- **Frames** — gepositioneerde rechthoekige inhoudsboxen met opvulling, randen en achtergrondvulling; nestbaar
- **Randen** — per-zijde ingesteld met doorgetrokken, gestreepte, gestippelde, streep-punt en aangepaste patronen
- **Kop- en voetteksten** — callbacks die op elke pagina worden aangeroepen met paginanummercontext
- **Tweefasige Build** — maakt "Pagina N van M" voetteksten mogelijk door pagina's te tellen vóór de weergave
- **Tekenprimatieven** — lijnen, veelhoeken, cirkels en rechthoeken direct op de pagina getekend
- **Afbeeldingen** — PNG en JPEG inline via `DrawImage`
- **Layout-engine** — `DocTemplate`, `PageTemplate`, `LayoutFrame` en een op `Flowable` gebaseerd story-compositiesysteem geïnspireerd door ReportLab/Platypus
- **20 grafiektypen** — lijn, gebied, kolom, balk, taart, polair, scatter, bubbel, heatmap, waterval, trechter, meter, foutbalk, boxplot, kolombereik, gebiedsbereik, bullet, dumbbell, lollipop, treemap
- **Uitvoer** — sla op in een bestand of schrijf naar elke `io.Writer`

## Installatie

```sh
go get github.com/gvanbeck/nautilus
```

## Snel starten

```go
package main

import (
    "log"

    "github.com/gvanbeck/nautilus/pdf"
)

func main() {
    doc, err := pdf.New(pdf.Config{
        PageSize:        pdf.PageSizeA4,
        DefaultFontSize: 12,
    })
    if err != nil {
        log.Fatal(err)
    }

    doc.AddPage()

    if err := doc.RegisterFont("regular", "/pad/naar/lettertype.ttf"); err != nil {
        log.Fatal(err)
    }
    if err := doc.SetFont("regular", 14); err != nil {
        log.Fatal(err)
    }

    if _, err := doc.WriteLine("Hallo, wereld!", 50, 100); err != nil {
        log.Fatal(err)
    }

    if err := doc.Save("hallo.pdf"); err != nil {
        log.Fatal(err)
    }
}
```

## API-referentie

### Document aanmaken

```go
// Maak een document met standaardinstellingen (A4, 12 pt, 1,2× regelafstand).
doc, err := pdf.New(pdf.Config{})

// Maak aan met expliciete instellingen.
doc, err := pdf.New(pdf.Config{
    PageSize:         pdf.PageSizeA4,      // of PageSizeA3, A5, Letter, Legal
    EmojiResolver:    resolver,            // optionele emoji.Resolver
    DefaultFontSize:  12,
    LineHeightFactor: 1.4,
})
```

**Beschikbare paginaformaten:**

| Constante         | Breedte (pt) | Hoogte (pt) |
|-------------------|-------------|-------------|
| `PageSizeA3`      | 841,89      | 1190,55     |
| `PageSizeA4`      | 595,28      | 841,89      |
| `PageSizeA5`      | 419,53      | 595,28      |
| `PageSizeLetter`  | 612         | 792         |
| `PageSizeLegal`   | 612         | 1008        |

### Pagina's

```go
doc.AddPage()                   // voeg een nieuwe pagina toe en maak deze actief
doc.PageWidth()                 // paginabreedte in punten
doc.PageHeight()                // paginahoogte in punten
doc.PageCount()                 // aantal toegevoegde pagina's tot nu toe
```

### Marges

Stel marges eenmalig in via `Config`. Alle schrijfmethoden accepteren nog steeds expliciete
coördinaten; de marge-accessors geven benoemde verwijzingen naar het inhoudsgebied,
zodat je nooit numerieke offsets hard hoeft te coderen.

```go
doc, err := pdf.New(pdf.Config{
    PageSize: pdf.PageSizeA4,
    Margins:  pdf.UniformMargins(50),           // 50 pt aan alle zijden
    // of:
    Margins:  pdf.Margins{Top: 60, Right: 50, Bottom: 60, Left: 50},
})

// Inhoudsgebied-accessors
doc.ContentX()        // linker rand  = margins.Left
doc.ContentY()        // bovenste rand   = margins.Top
doc.ContentWidth()    // bruikbare breedte  = paginabreedte  − links − rechts marge
doc.ContentHeight()   // bruikbare hoogte = paginahoogte − boven  − onder marge
doc.ContentRightX()   // rechter rand = paginabreedte − margins.Right  (RTL-ankerpunt)
doc.ContentBottomY()  // onderste rand = paginahoogte − margins.Bottom

// Gebruik bij het schrijven van inhoud
doc.WriteText(text, doc.ContentX(), doc.ContentY(), doc.ContentWidth())

// Tekst van rechts naar links
shaped := rtl.Shape("مرحبا بالعالم")
doc.WriteLineRTL(shaped, doc.ContentRightX(), y)

// Tabeloverloopdrempel
tbl := doc.NewTable(pdf.TableConfig{
    X:             doc.ContentX(),
    Y:             startY,
    ColWidths:     []float64{...},
    PageBottom:    doc.ContentBottomY(),
    ContinuationY: doc.ContentY(),
})
```

### Lettertypen

Zowel `.ttf`- als `.otf`-bestanden worden ondersteund.

```go
// Registreer lettertypen onder benoemde aliassen.
doc.RegisterFont("regular", "/pad/naar/NotoSans-Regular.ttf")
doc.RegisterFont("bold",    "/pad/naar/NotoSans-Bold.ttf")
doc.RegisterFont("italic",  "/pad/naar/NotoSans-Italic.otf")

// Activeer een lettertype op een bepaalde grootte (punten).
doc.SetFont("regular", 12)
doc.SetFont("bold", 14)

// Meet de tekstbreedte in het huidige lettertype.
width, err := doc.MeasureText("Hallo")
```

### Tekstweergave

```go
// Schrijf één regel op (x, y). Geeft de X na het laatste teken terug.
endX, err := doc.WriteLine("Hallo, wereld! 👋", 50, 100)

// Schrijf afgebroken tekst binnen maxWidth. Geeft de Y onder de laatste regel terug.
endY, err := doc.WriteText(langeTekst, 50, 100, 495)

// Stel tekstkleur in (RGB, 0–255).
doc.SetTextColor(60, 60, 60)

// Pas de regelafstandvermenigvuldiger aan (standaard 1,2).
doc.SetLineHeightFactor(1.5)
```

`WriteText` respecteert expliciete `\n`-regeleinden en breekt lange regels op woordgrenzen.

### Cursorpositie

```go
x := doc.GetX()  // huidige horizontale positie
y := doc.GetY()  // huidige verticale positie
```

### Tekst van rechts naar links (Arabisch en Hebreeuws)

Het pakket `pdf/rtl` bereidt RTL-tekst voor op weergave door Arabische contextuele
lettertekenvorming en Unicode Bidirectioneel Algoritme (BiDi)-herordening toe te passen.
Het resultaat is een tekenreeks in visuele (links-naar-rechts glyph) volgorde die kan worden
doorgegeven aan de RTL-weergavemethoden.

```go
import "github.com/gvanbeck/nautilus/pdf/rtl"
```

**Enkele-regel RTL-weergave:**

```go
// 1. Vorm en herorden de tekst.
shaped := rtl.Shape("مرحبا بالعالم")

// 2. Geef weer met de rechterrand op rightX.
leftX, err := doc.WriteLineRTL(shaped, rightEdge, y)

// Hebreeuws (geen Arabische vorming nodig — Shape past nog steeds BiDi-herordening toe).
shaped := rtl.Shape("שלום עולם")
leftX, err := doc.WriteLineRTL(shaped, rightEdge, y)
```

**Meerdere regels (afgebroken) RTL-weergave:**

`WriteTextRTL` verwerkt intern vorming, afbreken en per-regel BiDi-herordening,
zodat de woordvolgorde correct behouden blijft over regelbreuken.
Geef de originele tekst in logische volgorde direct door.

```go
// Expliciete regeleinden (\n) worden behandeld als alinea-einden.
endY, err := doc.WriteTextRTL("مرحبا بالعالم\nكيف حالك", rightEdge, y, maxWidth)
```

**RTL binnen een Frame:**

```go
f := doc.NewFrame(pdf.FrameConfig{
    X: 50, Y: 200, Width: 495,
    Padding: pdf.UniformPadding(8),
})
f.SetFont("arabic", 12)

// Enkele regel — tekst moet vooraf gevormd zijn.
shaped := rtl.Shape("مرحبا")
f.WriteLineRTL(shaped)

// Meerdere regels — geef originele tekst door, vorming wordt intern toegepast.
f.WriteTextRTL("مرحبا بالعالم\nكيف حالك")

f.Close()
```

**Functies in het `rtl`-pakket:**

| Functie | Beschrijving |
|----------|-------------|
| `rtl.Shape(text)` | Arabische vorming + BiDi-herordening → gebruik voor enkele regels. |
| `rtl.ShapeOnly(text)` | Alleen Arabische vorming, logische volgorde bewaard → zelden direct nodig. |
| `rtl.Reorder(text)` | Alleen BiDi-herordening, geen Arabische vorming → geschikt voor Hebreeuws. |

**Lettertypevereisten voor Arabisch:**

Het lettertype moet het **Unicode Arabic Presentation Forms-B** blok
(U+FE70–U+FEFF) bevatten. Aanbevolen lettertypen: [Noto Naskh Arabic](https://fonts.google.com/noto/specimen/Noto+Naskh+Arabic),
[Amiri](https://fonts.google.com/specimen/Amiri).
Voor Hebreeuws is elk lettertype dat het Hebreeuws blok (U+0590–U+05FF) dekt voldoende.

**Ondersteunde Arabische letters:**

Alle letters van het basisarabische alfabet (U+0621–U+064A), inclusief alef-varianten
(آ أ إ ا), ta marbuta (ة), waw (و), ya (ي) en de verplichte lam-alef-ligaturen
(لا لأ لإ لآ). Arabische diakritische tekens (harakat) worden behandeld als transparant
tijdens verbinding en worden ongewijzigd doorgegeven.

### Inline HTML-opmaak

Het pakket `pdf/html` converteert een tekenreeks van inline HTML naar een slice van `Span`-waarden,
elk met tekst, opmaakvlaggen en een optionele CSS-klassenaam.

```go
import "github.com/gvanbeck/nautilus/pdf/html"
```

**Ondersteunde tags:** `<b>`, `<strong>`, `<i>`, `<em>`, `<u>` en elke inline tag
met een `class`-attribuut. Tags mogen vrij worden genest.

```go
spans, err := html.Parse("<b>vet</b> en <i>cursief</i>", nil)
```

Elke `Span` bevat:

```go
type Span struct {
    Text  string      // platte tekstinhoud
    Style html.Style  // vlaggen Bold, Italic, Underline
    Class string      // CSS-klassenaam (binnenste), indien aanwezig
}
```

**Op klassen gebaseerde stijlen:**

Geef een `ClassStyle`-map door om extra stijlvlaggen samen te voegen op spans waarvan de
tag een overeenkomend `class`-attribuut heeft. De klassenaam wordt altijd bewaard in
`Span.Class`, ongeacht de rest.

```go
cs := html.ClassStyle{
    "highlight": {Bold: true},
    "note":      {Italic: true},
    "important": {Bold: true, Underline: true},
}
spans, err := html.Parse(`<span class="highlight">tekst</span>`, cs)
// spans[0].Style.Bold == true
// spans[0].Class == "highlight"
```

**Spans weergeven:**

Wissel van lettertype op basis van `Span.Style` en render elke span achtereenvolgens,
waarbij `x` wordt bijgewerkt met de terugkeerwaarde van `WriteLine`:

```go
fontFor := func(s html.Style) string {
    switch {
    case s.Bold:
        return "bold"
    case s.Italic:
        return "italic"
    default:
        return "regular"
    }
}

x := startX
for _, span := range spans {
    doc.SetFont(fontFor(span.Style), fontSize)
    x, _ = doc.WriteLine(span.Text, x, y)
}
```

### Emoji-ondersteuning

Emoji worden weergegeven als inline PNG-afbeeldingen op grootte van het huidige lettertype.
Geef een `Resolver` op die elk emoji-grafeem-cluster aan een PNG-bestandspad koppelt.

```go
import "github.com/gvanbeck/nautilus/pdf/emoji"

// Gebruik de ingebouwde Noto Emoji-resolver.
resolver := &emoji.NotoResolver{Dir: "/pad/naar/noto-emoji/png/128"}

doc, _ := pdf.New(pdf.Config{
    EmojiResolver: resolver,
})
```

Download Noto Emoji PNG's (Apache 2.0) van
[googlefonts/noto-emoji](https://github.com/googlefonts/noto-emoji/tree/main/png/128).

**Emoji-segmentatie** — het `emoji`-pakket biedt ook tekstsegmentatie:

```go
segments := emoji.Split("Hoi 👋 daar 🌍")
// → [{KindText "Hoi "}, {KindEmoji "👋"}, {KindText " daar "}, {KindEmoji "🌍"}]

// Zet een cluster om naar een Noto-stijl bestandsnaam.
emoji.ClusterToFilename("👨‍👩‍👧")
// → "emoji_u1f468_200d_1f469_200d_1f467.png"
```

**Aangepaste resolver** — implementeer de interface `emoji.Resolver`:

```go
type Resolver interface {
    Resolve(cluster string) (path string, found bool)
}
```

### Randen

Randen kunnen worden getekend rondom elke rechthoek. Elke zijde is onafhankelijk
instelbaar met eigen dikte, kleur en lijnpatroon.

```go
// Uniforme rand — alle vier zijden identiek.
border := pdf.NewUniformBorder(pdf.BorderSpec{
    Thickness: 1.5,
    Color:     pdf.ColorNavy,
    Pattern:   pdf.PatternSolid,
})
doc.DrawBorder(50, 100, 495, 40, border)

// Per-zijde rand — alleen boven en onder.
doc.DrawBorder(50, 100, 495, 40, pdf.Border{
    Top:    &pdf.BorderSpec{Thickness: 2, Color: pdf.ColorNavy, Pattern: pdf.PatternSolid},
    Bottom: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorGray, Pattern: pdf.PatternDashed},
})
```

**Randpatronen:**

| Constante         | Beschrijving                                       |
|-------------------|----------------------------------------------------|
| `PatternSolid`    | Doorlopende ononderbroken lijn                     |
| `PatternDashed`   | Lang-streep / tussenruimte patroon                 |
| `PatternDotted`   | Korte-punt / tussenruimte patroon                  |
| `PatternDashDot`  | Afwisselend lange streep en korte punt             |
| `PatternCustom`   | Aangepast streeppatroon via het veld `DashArray`   |

```go
// Aangepast streeppatroon.
spec := pdf.BorderSpec{
    Thickness: 2,
    Color:     pdf.ColorRed,
    Pattern:   pdf.PatternCustom,
    DashArray: []float64{12, 4, 4, 4},
    DashPhase: 0,
}
```

**Voorgedefinieerde kleuren:**

`ColorBlack`, `ColorWhite`, `ColorLightGray`, `ColorGray`, `ColorDarkGray`,
`ColorRed`, `ColorGreen`, `ColorBlue`, `ColorNavy`, `ColorOrange`

```go
custom := pdf.Color{R: 235, G: 245, B: 255}
```

### Frames

Frames zijn gepositioneerde rechthoekige inhoudsboxen — vergelijkbaar met LaTeX-minipages.
Inhoud stroomt automatisch omlaag binnen het frame.

```go
// Frame met vaste hoogte, achtergrondvulling en accentrand.
f := doc.NewFrame(pdf.FrameConfig{
    X: 50, Y: 200, Width: 495, Height: 80,
    Background: &pdf.Color{R: 235, G: 245, B: 255},
    Border: pdf.Border{
        Left: &pdf.BorderSpec{Thickness: 4, Color: pdf.ColorNavy},
    },
    Padding: pdf.Padding{Top: 8, Right: 12, Bottom: 8, Left: 16},
})
f.SetFont("regular", 10)
f.SetTextColor(20, 20, 80)
f.WriteText("Deze tekst stroomt binnen het frame.")
f.Close() // tekent de rand

// Frame met automatische hoogte (Height: 0) — rand past zich aan de inhoud aan.
f := doc.NewFrame(pdf.FrameConfig{
    X: 50, Y: 300, Width: 230,
    Border:  pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}),
    Padding: pdf.UniformPadding(8),
})
f.SetFont("regular", 10)
f.WriteText("De framehoogte past zich aan om deze inhoud te bevatten.")
f.Close()
```

**Frame-methoden:**

```go
f.ContentX()       // linker rand van inhoudsgebied (X + opvulling links)
f.ContentWidth()   // bruikbare breedte (framebreedte − horizontale opvulling)
f.CurrentY()       // Y-positie van de volgende inhoudsregel
f.FrameHeight()    // huidige buitenhoogte (vast of berekend)

f.WriteLine(text)                // geef weer op huidige regel (geen Y-vooruitgang)
f.WriteLineAt(text, xOffset)    // geef weer op offset van linker rand inhoudsgebied
f.WriteText(text)                // afgebroken tekst, Y gaat vooruit
f.WriteLineRTL(shaped)          // RTL enkele regel, rechts uitgelijn (vooraf gevormd)
f.WriteTextRTL(text)            // RTL afgebroken, vorming intern toegepast
f.Advance(n)                     // verplaats Y n punten omlaag
f.NewLine()                      // verplaats Y één regelafstand omlaag

f.SetFont(name, size)            // delegeert naar Document.SetFont
f.SetTextColor(r, g, b)         // delegeert naar Document.SetTextColor
f.MeasureText(text)              // delegeert naar Document.MeasureText

// Teken een rand binnen het frame op een relatieve offset.
f.DrawInnerBorder(xOffset, yOffset, width, height, border)

f.Close()  // afronden: teken buitenrand (idempotent)
```

**Opvullinghulpfuncties:**

```go
pdf.UniformPadding(8)           // 8 pt aan alle zijden
pdf.HorizontalPadding(12, 6)   // 12 pt links/rechts, 6 pt boven/onder
pdf.Padding{Top: 5, Right: 8, Bottom: 5, Left: 8}  // expliciet
```

### Tabellen

Tabellen bieden rastergebaseerde opmaak met kolomspannen, rijspannen, per-cel
opmaak en automatische paginaoverloop.

```go
tbl := doc.NewTable(pdf.TableConfig{
    X: 50, Y: 100,
    ColWidths: []float64{120, 260, 115},     // expliciete kolombreedten
    PageBottom:    doc.PageHeight() - 60,     // overloopdrempel
    ContinuationY: 60,                       // Y op vervolgpagina's
    Border: pdf.NewUniformBorder(pdf.BorderSpec{
        Thickness: 1.5, Color: pdf.ColorNavy,
    }),
    DefaultCellStyle: pdf.CellStyle{
        Padding:  pdf.Padding{Top: 5, Right: 8, Bottom: 5, Left: 8},
        Border:   pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}),
        FontName: "regular",
        FontSize: 10,
    },
})
```

**Rijen toevoegen:**

```go
// Enkele rij.
tbl.AddRow(pdf.Row{
    Height: 24,  // vaste hoogte; 0 = automatische hoogte op basis van inhoud
    Cells: []pdf.Cell{
        {Text: "Naam"},
        {Text: "Beschrijving"},
        {Text: "Waarde"},
    },
})

// Meerdere rijen tegelijk.
tbl.AddRows(rij1, rij2, rij3)
```

**Kolomspan:**

```go
tbl.AddRow(pdf.Row{
    Cells: []pdf.Cell{
        {Text: "Overspant alle kolommen", ColSpan: 3},
    },
})
```

**Rijspan:**

```go
tbl.AddRow(pdf.Row{
    Cells: []pdf.Cell{
        {Text: "Overspant 2 rijen", RowSpan: 2},
        {Text: "Rij 1, Kolom 2"},
        {Text: "Rij 1, Kolom 3"},
    },
})
tbl.AddRow(pdf.Row{
    Cells: []pdf.Cell{
        // Kolom 1 bezet door rijspan — weglaten.
        {Text: "Rij 2, Kolom 2"},
        {Text: "Rij 2, Kolom 3"},
    },
})
```

**Per-cel opmaak:**

```go
navy := pdf.ColorNavy
white := pdf.ColorWhite

tbl.AddRow(pdf.Row{
    Cells: []pdf.Cell{
        {Text: "Koptekst", Style: pdf.CellStyle{
            Background: &navy,
            TextColor:  &white,
            FontName:   "bold",
            FontSize:   11,
            HAlign:     pdf.HAlignCenter,
            VAlign:     pdf.VAlignMiddle,
            Padding:    pdf.UniformPadding(6),
            Border:     pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 1, Color: pdf.ColorNavy}),
        }},
    },
})
```

**Rijachtergrond:**

```go
bg := pdf.Color{R: 230, G: 240, B: 255}
tbl.AddRow(pdf.Row{
    Background: &bg,
    Cells:      []pdf.Cell{{Text: "Gearceerde rij"}},
})
```

**Celuitlijning:**

| Horizontaal         | Verticaal           |
|---------------------|---------------------|
| `HAlignDefault`     | `VAlignDefault`     |
| `HAlignLeft`        | `VAlignTop`         |
| `HAlignCenter`      | `VAlignMiddle`      |
| `HAlignRight`       | `VAlignBottom`      |

**De tabel tekenen:**

```go
if err := tbl.Draw(); err != nil {
    log.Fatal(err)
}
```

Tabellen roepen automatisch `doc.AddPage()` aan wanneer een rijgroep de resterende
ruimte op de huidige pagina overschrijdt. Rijen samengevoegd via een `RowSpan` worden
bij elkaar gehouden en nooit over een pagina-einde gesplitst.

### Kop- en voetteksten

Registreer callbacks die automatisch op elke pagina worden aangeroepen.

```go
doc.SetHeader(func(d *pdf.Document, info pdf.PageInfo) {
    d.SetFont("regular", 8)
    d.SetTextColor(100, 100, 100)
    d.WriteLine("Mijn Document", 50, 15)
})

doc.SetFooter(func(d *pdf.Document, info pdf.PageInfo) {
    d.SetFont("regular", 8)
    d.SetTextColor(120, 120, 120)
    label := fmt.Sprintf("Pagina %d van %d", info.Number, info.Total)
    w, _ := d.MeasureText(label)
    d.WriteLine(label, (d.PageWidth()-w)/2, d.PageHeight()-20)
})
```

`PageInfo` biedt:

- `Number` — op 1 gebaseerde index van de huidige pagina
- `Total` — totaal aantal pagina's (0 indien onbekend)

**Bekend totaal aantal pagina's** — als je het aantal van tevoren weet:

```go
doc.SetTotalPages(10)
```

### Tweefasige Build

Gebruik `Build` wanneer voetteksten het totale aantal pagina's moeten weergeven maar het
aantal niet van tevoren bekend is. Build voert de callback tweemaal uit: eerst om pagina's
te tellen, dan om te renderen met het totaal beschikbaar.

```go
doc.SetFooter(func(d *pdf.Document, info pdf.PageInfo) {
    d.SetFont("regular", 8)
    label := fmt.Sprintf("Pagina %d van %d", info.Number, info.Total)
    d.WriteLine(label, 50, d.PageHeight()-20)
})

doc.Build(func() {
    doc.AddPage()
    doc.SetFont("regular", 12)
    doc.WriteText("Inhoud van de eerste pagina…", 50, 60, 495)

    doc.AddPage()
    doc.SetFont("regular", 12)
    doc.WriteText("Inhoud van de tweede pagina…", 50, 60, 495)
})

doc.Save("rapport.pdf")
```

Tijdens de telfase:
- `AddPage` verhoogt de teller maar produceert geen PDF-inhoud
- `SetFont` houdt de lettertypestatus bij maar roept gopdf niet aan
- `WriteLine`, `WriteText`, `WriteLineRTL`, `WriteTextRTL`, `SetTextColor`, `DrawBorder` zijn no-ops
- `RegisterFont` wordt altijd uitgevoerd (lettertypen moeten beschikbaar zijn voor beide fasen)

### Uitvoer

```go
// Sla op in een bestand.
doc.Save("uitvoer.pdf")

// Schrijf naar elke io.Writer.
var buf bytes.Buffer
doc.Output(&buf)
```

## Tekenprimatieven

Het coördinatenstelsel voor alle tekenmethoden plaatst de oorsprong in de
linkerbovenhoek van de pagina, waarbij X naar rechts toeneemt en Y naar
beneden toeneemt. Alle afmetingen zijn in punten (1 pt = 1/72 inch).

Alle tekenprimatieven zijn no-ops tijdens de telfase van `Build`.

### Lijnen

```go
// Teken een rechte lijn van (x1,y1) naar (x2,y2).
doc.DrawLine(x1, y1, x2, y2, lineWidth float64, color pdf.Color)
```

```go
// Teken een diagonale scheidingslijn.
doc.DrawLine(50, 100, 545, 100, 0.5, pdf.ColorGray)
```

### Veelhoeken

```go
// Teken een gevulde veelhoek. Vereist minimaal 3 punten.
// De veelhoek wordt automatisch gesloten (laatste punt verbindt met het eerste).
doc.FillPolygon(points []pdf.Point, color pdf.Color)

// Teken een gevulde veelhoek met een omrand omtrek.
doc.FillAndStrokePolygon(points []pdf.Point, fillColor pdf.Color, lineWidth float64, strokeColor pdf.Color)
```

```go
// Teken een gevulde driehoek.
doc.FillPolygon([]pdf.Point{
    {X: 100, Y: 200},
    {X: 150, Y: 120},
    {X: 200, Y: 200},
}, pdf.Color{R: 70, G: 130, B: 180})

// Teken een gevulde driehoek met een donkere omtrek.
doc.FillAndStrokePolygon(
    []pdf.Point{{X: 100, Y: 200}, {X: 150, Y: 120}, {X: 200, Y: 200}},
    pdf.Color{R: 70, G: 130, B: 180},
    1.5,
    pdf.ColorNavy,
)
```

### Cirkels

```go
// Teken een gevulde cirkel gecentreerd op (cx, cy) met straal r.
doc.FillCircle(cx, cy, r float64, color pdf.Color)

// Teken de omtrek van een cirkel.
doc.StrokeCircle(cx, cy, r, lineWidth float64, color pdf.Color)
```

```go
// Gevulde cirkel.
doc.FillCircle(150, 200, 30, pdf.Color{R: 255, G: 100, B: 100})

// Omrandde cirkel.
doc.StrokeCircle(150, 200, 30, 1.5, pdf.ColorNavy)
```

Cirkels worden benaderd met 32 veelhoekhoekpunten, wat op normale documentresoluties
niet te onderscheiden is van een echte cirkel.

### Rechthoeken

```go
// Teken een gevulde rechthoek. Oorsprong is linksboven; Y neemt naar beneden toe.
doc.FillRect(x, y, w, h float64, color pdf.Color)
```

```go
// Teken een marine koptekstband over de volledige inhoudsbreedte.
doc.FillRect(doc.ContentX(), 40, doc.ContentWidth(), 24, pdf.ColorNavy)
```

### Afbeeldingen

```go
// Geef een PNG- of JPEG-afbeelding weer op (x, y) geschaald naar de opgegeven breedte en hoogte.
err := doc.DrawImage(path string, x, y, width, height float64) error
```

```go
// Plaats een logo in de linkerbovenhoek van het inhoudsgebied.
if err := doc.DrawImage("logo.png", doc.ContentX(), doc.ContentY(), 80, 40); err != nil {
    log.Fatal(err)
}
```

### Grafische toestand

De grafische toestandsstapel stelt je in staat tekening-attributen op te slaan en te herstellen
(lijnbreedte, streepkleur, vulkleur), zodat grafiek- en tekeningcode elkaar niet kunnen
beïnvloeden.

```go
doc.SaveGraphicsState()    // duw huidige toestand op de stapel
// ... teken met tijdelijke instellingen ...
doc.RestoreGraphicsState() // haal vorige toestand terug van de stapel
```

Aanroepen moeten gebalanceerd zijn. Beide methoden zijn no-ops tijdens de telfase van `Build`.

## Layout-engine (`pdf/layout`)

De layout-engine biedt automatische paginastroom op hoog niveau. In plaats van
posities handmatig te berekenen, bouw je een vlakke lijst van `Flowable`-elementen
(de _story_) en laat je de engine ze verdelen over frames en pagina's.

```go
import "github.com/gvanbeck/nautilus/pdf/layout"
```

### Architectuur

```
Story  ([]Flowable)
    ↓ verwerkt door
DocTemplate          — pagina/frame-planner
    ↓ beheert
PageTemplate         — paginageometrie (geordende LayoutFrames + decorators)
    ↓ bevat
LayoutFrame          — rechthoekig gebied met neerwaartse Y-cursor
    ↓ tekent in
pdf.Document         — onderliggend PDF-canvas
```

### De Flowable-interface

Elk inhoudselement implementeert `Flowable`:

```go
type Flowable interface {
    // Wrap meet het flowable binnen de beschikbare ruimte.
    // Geeft de werkelijke (breedte, hoogte) terug die het flowable zal innemen.
    // Een teruggegeven hoogte groter dan availHeight geeft aan dat het flowable
    // niet past en gesplitst of naar het volgende frame verplaatst moet worden.
    Wrap(doc *pdf.Document, availWidth, availHeight float64) (float64, float64)

    // Draw geeft het flowable weer met de linkerbovenhoek op (x, y).
    // Altijd aangeroepen na een succesvolle Wrap.
    Draw(doc *pdf.Document, x, y float64) error

    // Split verdeelt het flowable zodat het eerste deel past binnen availHeight.
    // Geeft nil terug als splitsen niet mogelijk is; de engine verplaatst het
    // flowable naar het volgende frame. Teruggegeven delen moeten alle inhoud reproduceren.
    Split(doc *pdf.Document, availWidth, availHeight float64) []Flowable

    SpaceBefore() float64  // extra witruimte boven dit flowable
    SpaceAfter() float64   // extra witruimte onder dit flowable
    KeepWithNext() bool    // voorkom een breuk tussen dit en het volgende flowable
    MinWidth() float64     // minimale vereiste breedte
}
```

Belangrijke invarianten:
- `Wrap` wordt altijd aangeroepen vóór `Draw` of `Split`.
- Voorloopruimte (`SpaceBefore`) wordt onderdrukt bovenaan een nieuw frame.
- `DocTemplate.Build` heeft lusdetectie: het geeft een fout na 10 opeenvolgende
  mislukte plaatsingen.

### Ingebouwde flowables

#### Paragraph

Geeft afgebroken tekst weer met per-alinea lettertype, kleur, uitlijning en
afstandsregeling.

```go
style := layout.ParagraphStyle{
    FontName:         "regular",    // geregistreerde lettertypenaam
    FontSize:         12,           // punten; 0 gebruikt documentstandaard
    Leading:          16,           // regelafstand; 0 standaard naar FontSize × 1,2
    Alignment:        layout.AlignLeft, // AlignLeft, AlignCenter, AlignRight
    SpaceBefore:      8,            // extra ruimte boven de alinea
    SpaceAfter:       6,            // extra ruimte onder de alinea
    KeepWithNextPara: true,         // voorkom breuk vóór het volgende flowable
    LeftIndent:       20,           // verminder bruikbare breedte vanaf links
    RightIndent:      20,           // verminder bruikbare breedte vanaf rechts
    TextColor:        &pdf.Color{R: 40, G: 40, B: 40},
}

p := &layout.Paragraph{Text: "Hallo, layout-engine!", Style: style}
```

Lange alinea's worden automatisch gesplitst over frames op regelgrenzen.

#### Spacer

Reserveert een vaste hoeveelheid verticale ruimte zonder iets weer te geven.

```go
&layout.Spacer{Height: 12}             // 12 pt tussenruimte
&layout.Spacer{Width: 80, Height: 12}  // 80 pt breed, 12 pt hoog
```

#### HRFlowable

Tekent een horizontale lijn als een gevulde balk.

```go
&layout.HRFlowable{
    Width:     0.8,              // fractie van beschikbare breedte (0..1) of absolute pt (>1)
    Thickness: 1,                // balkhoogte in punten; standaard 1
    Color:     pdf.ColorGray,
    Align:     layout.AlignCenter, // AlignLeft, AlignCenter, AlignRight
    Before:    6,                // ruimte boven de lijn
    After:     6,                // ruimte onder de lijn
}
```

#### KeepTogether

Voorkomt dat een groep flowables over frames wordt gesplitst. Als de groep niet
past in de resterende frameruimte, voegt de engine een `FrameBreak` in en probeert
het opnieuw op het volgende frame. Als de groep groter is dan een volledig frame,
worden individuele flowables teruggegeven voor onafhankelijke splitsing.

```go
// Houd een koptekst bij zijn eerste bodytekst-alinea.
&layout.KeepTogether{
    Flowables: []layout.Flowable{
        &layout.Paragraph{Text: "Sectiekoptekst", Style: h1Style},
        &layout.Paragraph{Text: "Eerste alinea…", Style: bodyStyle},
    },
}
```

### Actie-flowables

Actie-flowables zijn nulhoogte-elementen die de engine aansturen in plaats van
zichtbare inhoud weer te geven.

#### PageBreak

Dwingt een onmiddellijk pagina-einde af.

```go
// Eenvoudig pagina-einde.
&layout.PageBreak{}

// Pagina-einde met onmiddellijke sjabloonwissel.
&layout.PageBreak{NextTemplate: "TweeKolommen"}
```

#### FrameBreak

Brengt de engine naar het volgende frame (of de volgende pagina als er geen
frames meer zijn op de huidige pagina).

```go
&layout.FrameBreak{}
```

#### CondPageBreak

Voegt een pagina-einde in alleen wanneer minder dan `MinHeight` punten resten in
het huidige frame.

```go
// Breek als minder dan 72 pt (één inch) resteert.
&layout.CondPageBreak{MinHeight: 72}
```

#### NextPageTemplate

Plant een sjabloonwissel die van kracht wordt op het volgende pagina-einde. De
huidige pagina blijft het bestaande sjabloon gebruiken.

```go
story = append(story,
    titelInhoud...,
    &layout.NextPageTemplate{TemplateID: "TweeKolommen"},
    &layout.PageBreak{},
    bodyInhoud...,
)
```

### LayoutFrame

Een `LayoutFrame` is een rechthoekig gebied dat flowables ontvangt. Het frame
behoudt een interne Y-cursor die omlaag beweegt naarmate inhoud wordt toegevoegd.

```go
frame := &layout.LayoutFrame{
    X:            50,    // linkerbovenhoek X in paginacoördinaten (punten)
    Y:            80,    // linkerbovenhoek Y in paginacoördinaten (punten)
    Width:        495,   // buitenbreedte in punten
    Height:       700,   // buitenhoogte in punten
    Padding:      pdf.Padding{Top: 8, Right: 8, Bottom: 8, Left: 8},
    ID:           "main",    // optionele naam voor foutopsporing
    ShowBoundary: false,     // teken een dunne omtrek als true (handig tijdens ontwikkeling)
}
```

### PageTemplate

Een `PageTemplate` groepeert één of meer `LayoutFrame`s met optionele decorators.

```go
tmpl := &layout.PageTemplate{
    ID:               "enkel",           // gerefereerd door NextPageTemplate / PageBreak
    Frames:           []*layout.LayoutFrame{frame},
    OnPage:           kopVoetTekstFunc,  // aangeroepen na AddPage (teken kopteksten, watermerken)
    OnPageEnd:        func(doc *pdf.Document, pageNum int) { /* ... */ },
    AutoNextTemplate: "enkel",           // schakel naar dit sjabloon na elke pagina
}
```

Handtekening van `PageDecorator`:

```go
type PageDecorator func(doc *pdf.Document, pageNum int)
```

### DocTemplate

`DocTemplate` is de engine die de story verwerkt.

```go
dt := layout.NewDocTemplate(doc)
dt.AddPageTemplate(enkelSjabloon)
dt.AddPageTemplate(tweeKolommenSjabloon)

if err := dt.Build(story); err != nil {
    log.Fatal(err)
}
```

- `NewDocTemplate(doc)` — maak een engine aan voor het opgegeven `pdf.Document`.
- `AddPageTemplate(pt)` — registreer een paginasjabloon. Het eerste geregistreerde
  sjabloon wordt gebruikt voor de eerste pagina.
- `Build(story)` — verspreid alle flowables over frames en pagina's. Geeft een
  fout terug als er geen sjablonen zijn geregistreerd of als een flowable na 10
  pogingen niet in een frame past.

### Minimaal voorbeeld

```go
package main

import (
    "log"

    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/layout"
)

func main() {
    doc, _ := pdf.New(pdf.Config{
        PageSize: pdf.PageSizeA4,
        Margins:  pdf.UniformMargins(50),
    })
    doc.RegisterFont("regular", "/pad/naar/lettertype.ttf")
    doc.SetFont("regular", 12)

    style := layout.ParagraphStyle{FontName: "regular", FontSize: 12}
    story := []layout.Flowable{
        &layout.Paragraph{Text: "Hallo, Nautilus!", Style: style},
        &layout.Spacer{Height: 12},
        &layout.Paragraph{Text: "Tweede alinea.", Style: style},
    }

    frame := &layout.LayoutFrame{
        X: doc.ContentX(), Y: doc.ContentY(),
        Width: doc.ContentWidth(), Height: doc.ContentHeight(),
    }
    tmpl := &layout.PageTemplate{ID: "main", Frames: []*layout.LayoutFrame{frame}}

    dt := layout.NewDocTemplate(doc)
    dt.AddPageTemplate(tmpl)
    if err := dt.Build(story); err != nil {
        log.Fatal(err)
    }
    doc.Save("uitvoer.pdf")
}
```

### Meerkolomsopmaak

Geef twee `LayoutFrame`s per `PageTemplate` op. De engine vult eerst het linker
frame, dan het rechter frame, en start daarna een nieuwe pagina.

```go
const (
    margin  = 50.0
    gutter  = 12.0
    headerH = 40.0
    footerH = 36.0
)

pageW := doc.PageWidth()
pageH := doc.PageHeight()
contentX := margin
contentY := margin + headerH
contentW := pageW - 2*margin
contentH := pageH - margin - headerH - footerH

colW := (contentW - gutter) / 2

leftFrame := &layout.LayoutFrame{
    X: contentX,           Y: contentY,
    Width: colW,           Height: contentH,
    ShowBoundary: true,    // toon omtrek tijdens ontwikkeling
}
rightFrame := &layout.LayoutFrame{
    X: contentX + colW + gutter, Y: contentY,
    Width: colW,                 Height: contentH,
    ShowBoundary: true,
}

pageDecorator := func(d *pdf.Document, pageNum int) {
    d.SetFont("regular", 8)
    d.WriteLine("Mijn Document", margin, margin+10)
}

twoColTemplate := &layout.PageTemplate{
    ID:               "two-column",
    Frames:           []*layout.LayoutFrame{leftFrame, rightFrame},
    OnPage:           pageDecorator,
    AutoNextTemplate: "two-column",
}
```

### Sjabloonwisseling

Wissel van sjabloon op pagina-einden om eerste-pagina versus body-pagina
opmaak te implementeren:

```go
singleTemplate := &layout.PageTemplate{ID: "single", Frames: []*layout.LayoutFrame{singleFrame}}
twoColTemplate  := &layout.PageTemplate{ID: "two-column", Frames: []*layout.LayoutFrame{leftFrame, rightFrame}}

dt := layout.NewDocTemplate(doc)
dt.AddPageTemplate(singleTemplate)
dt.AddPageTemplate(twoColTemplate)

story := []layout.Flowable{
    // ... inhoud titelblad ...
    &layout.NextPageTemplate{TemplateID: "two-column"},
    &layout.PageBreak{},
    // ... bodytekst stroomt in twee kolommen ...
    &layout.NextPageTemplate{TemplateID: "single"},
    &layout.PageBreak{},
    // ... terug naar één kolom ...
}

dt.Build(story)
```

## Grafieken (`pdf/chart`)

Nautilus bevat 20 grafiektypen met een declaratieve API die het Highcharts
JSON-configuratiemodel weerspiegelt. Grafieken tekenen direct op `pdf.Document`
en integreren naadloos met de layout-engine via `chart.NewFlowable`.

```go
import (
    "github.com/gvanbeck/nautilus/pdf/chart"
    "github.com/gvanbeck/nautilus/pdf/chart/line"
)
```

### Grafiek-subpakketten

Elk grafiektype leeft in zijn eigen importeerbaar subpakket, zodat binaire bestanden
alleen betalen voor de renderers die ze gebruiken.

| Pakket | Grafiektype |
|---------|------------|
| `pdf/chart/line` | Lijndiagram — X/Y lijnen met optionele markeringen |
| `pdf/chart/area` | Gebiedsdiagram — gevuld lijndiagram |
| `pdf/chart/column` | Kolomdiagram — verticale balken, gegroepeerd of gestapeld |
| `pdf/chart/bar` | Balkdiagram — horizontale balken |
| `pdf/chart/pie` | Taart- en donutdiagram |
| `pdf/chart/polar` | Polair / spin / radardiagram |
| `pdf/chart/scatter` | Spreidingsdiagram — X/Y puntenwolk |
| `pdf/chart/bubble` | Bellendiagram — scatter met Z-grootte cirkels |
| `pdf/chart/heatmap` | Heatmap — kleurgecodeerd raster |
| `pdf/chart/waterfall` | Watervaldiagram — lopend-totaal balkdiagram |
| `pdf/chart/funnel` | Trechter- en pyramidediagram |
| `pdf/chart/gauge` | Meter- en solid-gauge-diagram |
| `pdf/chart/errorbar` | Foutbalk-diagram |
| `pdf/chart/boxplot` | Box-en-whisker-diagram |
| `pdf/chart/columnrange` | Kolombereiksdiagram — laag/hoog verticale balken |
| `pdf/chart/arearange` | Gebiedsbereikdiagram — laag/hoog gevulde band |
| `pdf/chart/bullet` | Bulletdiagram — balk met doelmarker en kwalitatieve banden |
| `pdf/chart/dumbbell` | Dumbbelldiagram — laag/hoog bereikpunten verbonden door een lijn |
| `pdf/chart/lollipop` | Lollipop-diagram — stok met eindpunt |
| `pdf/chart/treemap` | Treemap — hiërarchische rechthoekpakking |

### De Drawable-interface

Alle grafiektypen implementeren `Drawable`:

```go
type Drawable interface {
    Draw(doc *pdf.Document, x, y, width, height float64) error
}
```

`x, y` is de linkerbovenhoek van de begrenzende box in punten.

### Grafieken insluiten in een layout-story

Gebruik `chart.NewFlowable` om elke `Drawable` in te pakken als een `layout.Flowable`:

```go
// width: 0 vult de beschikbare framebreedte; height is vast.
story = append(story, chart.NewFlowable(mijnGrafiek, 0, 220))
```

### chart.Options

`chart.Options` is het configuratie-object op het hoogste niveau.

```go
opts := chart.Options{
    FontName:   "regular",             // geregistreerde lettertypenaam; moet op het Document staan
    FontSize:   9,                     // basislettergrootte in punten; standaard 9
    Title:      &chart.Title{Text: "Omzet per kwartaal"},
    Subtitle:   &chart.Title{Text: "2023 vs 2024"},
    XAxis:      &chart.Axis{Categories: []string{"K1", "K2", "K3", "K4"}},
    YAxis:      &chart.Axis{},
    Series:     []chart.Series{...},
    Legend:     &chart.Legend{},
    PlotOptions: &chart.PlotOptions{...},
    Colors:     nil,                   // nil gebruikt DefaultColors
    Background: &pdf.Color{R: 250, G: 250, B: 250},
}
```

### chart.Title

```go
&chart.Title{
    Text:     "Grafiektitel",
    FontName: "bold",       // overschrijft Options.FontName
    FontSize: 11,           // overschrijft Options.FontSize als > 0
    Color:    &pdf.Color{R: 30, G: 30, B: 30},
}
```

### chart.Axis

```go
&chart.Axis{
    Title:         &chart.Title{Text: "Omzet (EUR)"},
    Categories:    []string{"K1", "K2", "K3", "K4"}, // discrete ticklabels
    Min:           chart.Float(0),    // minimale zichtbare waarde begrenzen
    Max:           chart.Float(500),  // maximale zichtbare waarde begrenzen
    TickInterval:  chart.Float(100),  // vaste rasterlijnafstand
    GridLineWidth: 0.5,               // 0 = standaard; negatief = verberg rasterlijnen
    GridLineColor: &pdf.Color{R: 220, G: 220, B: 220},
    Labels: &chart.AxisLabels{
        Enabled:  chart.Bool(true),
        Format:   "{value}%",         // "{value}" wordt vervangen door het ticklabel
        FontName: "regular",
        FontSize: 8,
    },
    Visible: chart.Bool(true),
}
```

### chart.Series

```go
chart.Series{
    Name:  "Product A",                       // weergegeven in de legenda
    Data:  []float64{43, 55, 57, 60},         // y-waarden voor lijn/gebied/kolom/balk/taart
    Color: &pdf.Color{R: 124, G: 181, B: 236}, // overschrijft paletttoewijzing
}

// Rijke gegevens voor scatter, bubbel, heatmap, bereikdiagrammen, boxplot, enz.
chart.Series{
    Name: "Metingen",
    Points: []chart.Point{
        {X: 1.5, Y: 23.4},
        {X: 2.3, Y: 17.8},
    },
}
```

### chart.Point

`Point` is een rijk gegevenspunt voor grafiektypen die meer dan een enkele
Y-waarde vereisen. Stel alleen de velden in die zinvol zijn voor jouw grafiektype.

```go
chart.Point{
    X:    1.5,     // horizontale waarde (scatter, bubbel, heatmap kolomindex)
    Y:    23.4,    // primaire waarde
    Z:    50,      // bellenstraal bron; heatmap celwaarde

    Low:    10.0,  // ondergrens (bereikdiagrammen, boxplot, foutbalk, dumbbell)
    Q1:     20.0,  // eerste kwartiel (alleen boxplot)
    Median: 30.0,  // mediaan (alleen boxplot)
    Q3:     40.0,  // derde kwartiel (alleen boxplot)
    High:   55.0,  // bovengrens (bereikdiagrammen, boxplot, foutbalk, dumbbell)

    Target: 220,   // referentie/doelwaarde (bulletdiagram)

    Name:  "Categorielabel", // watervalstappen, trechterfasen, treemapknooppunten
    Color: &pdf.Color{...},  // per-punt kleuroverride

    IsSum:             true, // waterval: toon cumulatief totaal (Y genegeerd)
    IsIntermediateSum: true, // waterval: toon lopend subtotaal
}
```

### chart.Legend

```go
&chart.Legend{
    Enabled:       chart.Bool(true),
    Layout:        "horizontal",   // "horizontal" (standaard) of "vertical"
    Align:         "center",       // "left", "center" (standaard), "right"
    VerticalAlign: "bottom",       // "top", "middle", "bottom" (standaard)
    FontName:      "regular",
    FontSize:      8,
}
```

### chart.PlotOptions

`PlotOptions` bevat per-grafiektype weergaveknoppen. Alleen het veld dat overeenkomt
met het grafiektype dat wordt weergegeven heeft effect.

```go
opts.PlotOptions = &chart.PlotOptions{
    Line:        &chart.LineOptions{...},
    Area:        &chart.AreaOptions{...},
    Column:      &chart.ColumnOptions{...},
    Bar:         &chart.BarOptions{...},  // alias voor ColumnOptions
    Pie:         &chart.PieOptions{...},
    Polar:       &chart.PolarOptions{...},
    Scatter:     &chart.ScatterOptions{...},
    Bubble:      &chart.BubbleOptions{...},
    Heatmap:     &chart.HeatmapOptions{...},
    Waterfall:   &chart.WaterfallOptions{...},
    Funnel:      &chart.FunnelOptions{...},
    Gauge:       &chart.GaugeOptions{...},
    Errorbar:    &chart.ErrorbarOptions{...},
    Boxplot:     &chart.BoxplotOptions{...},
    ColumnRange: &chart.ColumnRangeOptions{...},
    AreaRange:   &chart.AreaRangeOptions{...},
    Bullet:      &chart.BulletOptions{...},
    Dumbbell:    &chart.DumbbellOptions{...},
    Lollipop:    &chart.LollipopOptions{...},
    Treemap:     &chart.TreemapOptions{...},
}
```

Belangrijke velden voor veelgebruikte grafiektypen:

| Type | Belangrijke velden |
|------|-----------|
| `LineOptions` | `LineWidth` (standaard 2), `Marker`, `DataLabels` |
| `AreaOptions` | `LineWidth`, `FillAlpha` (0–1, standaard 0,3), `Marker`, `DataLabels` |
| `ColumnOptions` | `Stacking` (`""` gegroepeerd, `"normal"`, `"percent"`), `GroupPadding`, `PointPadding`, `BorderWidth`, `DataLabels` |
| `PieOptions` | `InnerSize` (`"50%"` voor donut), `StartAngle` (graden, standaard −90 = bovenkant), `DataLabels` |
| `PolarOptions` | `GridLineInterpolation` (`"polygon"` of `"circle"`), `FillAlpha`, `LineWidth`, `Marker`, `DataLabels` |
| `BubbleOptions` | `MinSize`, `MaxSize`, `ZMin`, `ZMax`, `DataLabels` |
| `HeatmapOptions` | `MinColor`, `MaxColor`, `BorderWidth`, `DataLabels` |
| `WaterfallOptions` | `UpColor`, `NegativeColor`, `LineWidth`, `DataLabels` |
| `FunnelOptions` | `NeckWidth`, `NeckHeight`, `Width`, `Reversed` (pyramide), `DataLabels` |
| `GaugeOptions` | `PaneStartAngle`, `PaneEndAngle`, `PlotBands`, `Solid` (solid-gauge), `DataLabels` |
| `BulletOptions` | `PlotBands`, `TargetWidth`, `TargetColor`, `DataLabels` |
| `TreemapOptions` | `ColorByPoint` (standaard true), `BorderWidth`, `BorderColor`, `DataLabels` |

### Hulpfuncties

```go
// Float geeft een *float64 terug — gebruik dit voor optionele float-velden.
chart.Float(0.5)

// Bool geeft een *bool terug — gebruik dit voor optionele bool-velden.
chart.Bool(true)

// SeriesColor geeft de kleur terug voor reeksindex i, roterend door het
// geconfigureerde palet (opts.Colors) of DefaultColors als opts.Colors nil is.
chart.SeriesColor(opts, i)

// DefaultColors is het ingebouwde 10-kleuren Highcharts-palet.
var chart.DefaultColors []pdf.Color
```

### DataLabels en Marker

```go
// DataLabels configureert waardelabels die naast gegevenspunten of balken worden weergegeven.
&chart.DataLabels{
    Enabled:  chart.Bool(true),
    Format:   "{y}",      // "{y}" wordt vervangen door de waarde; standaard "{y}"
    FontName: "regular",
    FontSize: 8,
    Color:    &pdf.Color{R: 50, G: 50, B: 50},
}

// Marker regelt het symbool dat wordt getekend op elk gegevenspunt op lijn/gebied/scatter-diagrammen.
&chart.Marker{
    Enabled: chart.Bool(true),
    Symbol:  "circle",   // "circle" (standaard), "square", "diamond"
    Radius:  3,          // straal in punten
}
```

### GaugePlotBand

Gekleurde boogbanden gebruikt door zowel `GaugeOptions` als `BulletOptions`:

```go
chart.GaugePlotBand{
    From:      0,
    To:        80,
    Color:     pdf.Color{R: 85, G: 191, B: 59},  // groene zone
    Thickness: 12,  // boogbreedte in punten; standaard 10
}
```

### Volledig voorbeeld — lijndiagram via layout-engine

```go
package main

import (
    "log"

    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/chart"
    "github.com/gvanbeck/nautilus/pdf/chart/line"
    "github.com/gvanbeck/nautilus/pdf/layout"
)

func main() {
    doc, _ := pdf.New(pdf.Config{
        PageSize: pdf.PageSizeA4,
        Margins:  pdf.UniformMargins(40),
    })
    doc.RegisterFont("regular", "/pad/naar/lettertype.ttf")
    doc.SetFont("regular", 11)

    opts := chart.Options{
        FontName: "regular",
        FontSize: 8,
        Title:    &chart.Title{Text: "Maandelijkse omzet"},
        XAxis:    &chart.Axis{Categories: []string{"Jan", "Feb", "Mrt", "Apr", "Mei", "Jun"}},
        YAxis:    &chart.Axis{},
        Legend:   &chart.Legend{},
        Series: []chart.Series{
            {Name: "2023", Data: []float64{120, 150, 130, 180, 160, 200}},
            {Name: "2024", Data: []float64{140, 165, 175, 195, 210, 240}},
        },
    }

    lc := &line.LineChart{Options: opts}

    story := []layout.Flowable{
        chart.NewFlowable(lc, 0, 220),
    }

    frame := &layout.LayoutFrame{
        X: doc.ContentX(), Y: doc.ContentY(),
        Width: doc.ContentWidth(), Height: doc.ContentHeight(),
    }
    tmpl := &layout.PageTemplate{ID: "main", Frames: []*layout.LayoutFrame{frame}}

    dt := layout.NewDocTemplate(doc)
    dt.AddPageTemplate(tmpl)
    if err := dt.Build(story); err != nil {
        log.Fatal(err)
    }
    doc.Save("grafiek.pdf")
}
```

### Veelgebruikte grafiekvoorbeelden

**Gestapeld kolomdiagram:**

```go
opts := chart.Options{
    FontName: "regular",
    FontSize: 8,
    XAxis:    &chart.Axis{Categories: []string{"K1", "K2", "K3", "K4"}},
    YAxis:    &chart.Axis{},
    Series: []chart.Series{
        {Name: "Noord", Data: []float64{43, 55, 57, 60}},
        {Name: "Zuid",  Data: []float64{23, 35, 41, 47}},
        {Name: "West",  Data: []float64{31, 28, 38, 44}},
    },
    PlotOptions: &chart.PlotOptions{
        Column: &chart.ColumnOptions{Stacking: "normal"},
    },
}
cc := &column.ColumnChart{Options: opts}
cc.Draw(doc, x, y, width, height)
```

**Donutdiagram:**

```go
opts := chart.Options{
    FontName: "regular",
    FontSize: 8,
    Series: []chart.Series{
        {Name: "Chrome",  Data: []float64{65}},
        {Name: "Firefox", Data: []float64{15}},
        {Name: "Safari",  Data: []float64{12}},
        {Name: "Overig",  Data: []float64{8}},
    },
    PlotOptions: &chart.PlotOptions{
        Pie: &chart.PieOptions{
            InnerSize:  "50%",
            DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
        },
    },
}
pc := &pie.PieChart{Options: opts}
```

**Meter met plotbanden:**

```go
opts := chart.Options{
    FontName: "regular",
    FontSize: 8,
    YAxis:    &chart.Axis{Min: chart.Float(0), Max: chart.Float(200)},
    Series:   []chart.Series{{Name: "Snelheid km/u", Data: []float64{120}}},
    PlotOptions: &chart.PlotOptions{
        Gauge: &chart.GaugeOptions{
            PaneStartAngle: -150,
            PaneEndAngle:   150,
            PlotBands: []chart.GaugePlotBand{
                {From: 0,   To: 80,  Color: pdf.Color{R: 85,  G: 191, B: 59},  Thickness: 12},
                {From: 80,  To: 140, Color: pdf.Color{R: 221, G: 223, B: 13},  Thickness: 12},
                {From: 140, To: 200, Color: pdf.Color{R: 223, G: 83,  B: 83},  Thickness: 12},
            },
            DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
        },
    },
}
gc := &gauge.GaugeChart{Options: opts}
```

**Watervaldiagram met somvlaggen:**

```go
opts := chart.Options{
    FontName: "regular",
    FontSize: 8,
    YAxis:    &chart.Axis{},
    Series: []chart.Series{{
        Points: []chart.Point{
            {Name: "Begin",      Y: 120000},
            {Name: "Omzet",      Y: 569000},
            {Name: "Kosten",     Y: -342000},
            {Name: "Subtotaal",  IsIntermediateSum: true},
            {Name: "Meer kosten", Y: -233000},
            {Name: "Saldo",      IsSum: true},
        },
    }},
}
wc := &waterfall.WaterfallChart{Options: opts}
```

## Voorbeelden

| Voorbeeld | Beschrijving |
|---------|-------------|
| [`examples/basic`](examples/basic/main.go) | Demo met meerdere pagina's, inclusief lettertypen, Unicode, emoji, randen, frames, tabellen, kop- en voetteksten en het tweefasige Build-mechanisme. |
| [`examples/html`](examples/html/main.go) | Demonstreert `pdf/html`: inline HTML-tags en klasseattributen parsen naar opgemaakte `Span`-waarden en weergeven met lettertypewissel. |
| [`examples/layout`](examples/layout/main.go) | Meerkolomsopmaak, framewissel, `KeepTogether`, `CondPageBreak`, `HRFlowable` en de `OnPage`-decorator. |
| [`examples/rtl`](examples/rtl/main.go) | Arabische en Hebreeuwse tekst van rechts naar links: contextuele vorming, lam-alef-ligaturen, BiDi-herordening, gemengd RTL/LTR en RTL binnen een Frame. |
| [`examples/xml`](examples/xml/main.go) | XML-gegevens → PDF-tabel: parseer een gestructureerd XML-bestand en geef het weer als een opgemaakte tabel. |
| [`examples/chart`](examples/chart/main.go) | Alle 20 grafiektypen weergegeven via de layout-engine over 11 pagina's. |

### Het basisvoorbeeld uitvoeren

```sh
go run ./examples/basic \
    -font  /Library/Fonts/Lato-Medium.ttf \
    -bold  /Library/Fonts/Lato-Black.ttf \
    -emoji path/to/noto-emoji/png/128 \
    -out   uitvoer.pdf
```

### Het HTML-opmaakvoorbeeld uitvoeren

```sh
go run ./examples/html \
    -font   /Library/Fonts/Lato-Regular.ttf \
    -bold   /Library/Fonts/Lato-Bold.ttf \
    -italic /Library/Fonts/Lato-Italic.ttf \
    -out    uitvoer.pdf
```

### Het layoutvoorbeeld uitvoeren

```sh
go run ./examples/layout \
    -font /Library/Fonts/Lato-Medium.ttf \
    -bold /Library/Fonts/Lato-Black.ttf \
    -out  uitvoer.pdf
```

### Het RTL-voorbeeld uitvoeren

```sh
go run ./examples/rtl \
    -arabic /System/Library/Fonts/Supplemental/DecoTypeNaskh.ttc \
    -hebrew /System/Library/Fonts/SFHebrew.ttf \
    -latin  /Library/Fonts/Lato-Regular.ttf \
    -out    uitvoer.pdf
```

### Het grafiekvoorbeeld uitvoeren

```sh
go run ./examples/chart \
    -font /Library/Fonts/Lato-Medium.ttf \
    -bold /Library/Fonts/Lato-Black.ttf \
    -out  grafiek_uitvoer.pdf
```

## Licentie

Zie [LICENSE](LICENSE) voor details.
