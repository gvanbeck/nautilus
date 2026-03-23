# Nautilus Layout Systeem — Gebruikershandleiding

Het layout-pakket (`pdf/layout`) is een hoog-niveau layout-motor bovenop de Nautilus
`pdf.Document`-tekenlaag, geïnspireerd door het Platypus-systeem van de Python
ReportLab-bibliotheek.

In plaats van coördinaten handmatig te berekenen, beschrijft u uw document als een
**Story**: een platte lijst van `Flowable`-objecten.  De motor — `DocTemplate` —
verdeelt die inhoud automatisch over frames en pagina's.

---

## Inhoudsopgave

1. [Kernconcepten](#1-kernconcepten)
2. [Meeteenheden en coördinaten](#2-meeteenheden-en-coördinaten)
3. [Minimaal voorbeeld](#3-minimaal-voorbeeld)
4. [Document en lettertypes](#4-document-en-lettertypes)
5. [Paragraph](#5-paragraph)
6. [Spacer](#6-spacer)
7. [HRFlowable](#7-hrflowable)
8. [KeepTogether](#8-keeptogether)
9. [Action Flowables](#9-action-flowables)
10. [LayoutFrame](#10-layoutframe)
11. [PageTemplate](#11-pagetemplate)
12. [DocTemplate en Build](#12-doctemplate-en-build)
13. [Pagina-decorators (headers en footers)](#13-pagina-decorators-headers-en-footers)
   - [OnPage versus OnPageEnd](#onpage-versus-onpageend)
14. [Template wisselen](#14-template-wisselen)
15. [Meerkoloms-layout](#15-meerkoloms-layout)
16. [Foutopsporing met ShowBoundary](#16-foutopsporing-met-showboundary)
17. [Eigen Flowable implementeren](#17-eigen-flowable-implementeren)
18. [Veelgestelde vragen](#18-veelgestelde-vragen)

---

## 1. Kernconcepten

Het systeem bestaat uit vier lagen die op elkaar stapelen:

```
Story  ([]Flowable)
   ↓  consumed by
DocTemplate    —  beheert frames en pagina's
   ↓  manages
PageTemplate   —  pagina-geometrie: geordende LayoutFrames + decorators
   ↓  contains
LayoutFrame    —  rechthoekig gebied met een dalende Y-cursor
   ↓  draws into
pdf.Document   —  de onderliggende PDF-tekenlaag
```

### Story

Een `Story` is gewoon een `[]layout.Flowable`.  U bouwt de volledige
documentinhoud op als een platte lijst en geeft die door aan `DocTemplate.Build`.

```go
story := []layout.Flowable{
    &layout.Paragraph{Text: "Titel", Style: titleStyle},
    &layout.Spacer{Height: 12},
    &layout.Paragraph{Text: "Eerste alinea...", Style: bodyStyle},
    &layout.PageBreak{},
    &layout.Paragraph{Text: "Begint op een nieuwe pagina.", Style: bodyStyle},
}
```

### Flowable

Een `Flowable` is alles wat gemeten en getekend kan worden.  Het interface:

```go
type Flowable interface {
    Wrap(doc *pdf.Document, availWidth, availHeight float64) (width, height float64)
    Draw(doc *pdf.Document, x, y float64) error
    Split(doc *pdf.Document, availWidth, availHeight float64) []Flowable
    SpaceBefore() float64
    SpaceAfter()  float64
    KeepWithNext() bool
    MinWidth() float64
}
```

De motor roept altijd eerst `Wrap` aan (voor meting), daarna `Draw` (voor rendering).
`Split` wordt alleen aangeroepen als een flowable niet past in de resterende ruimte.

`MinWidth` geeft de absolute minimale breedte aan die de flowable nodig heeft.
Momenteel gebruikt de motor dit als controlegetal; als een frame smaller is dan
`MinWidth`, wordt de flowable als niet-plaatsbaar beschouwd.  Voor de meeste
elementen is `0` de juiste waarde.

### ActionFlowable

`ActionFlowable` is een sub-interface van `Flowable` voor onzichtbare,
nul-hoge elementen die de motor sturen in plaats van inhoud te renderen.
Ingebouwde voorbeelden: `PageBreak`, `FrameBreak`, `CondPageBreak`,
`NextPageTemplate`.

```go
type ActionFlowable interface {
    Flowable
    apply(doc *DocTemplate)
}
```

> **Let op:** `apply` is ongeëxporteerd.  U kunt geen eigen `ActionFlowable`
> implementeren buiten het `layout`-pakket.  Voor aangepaste paginabesturing
> kunt u de ingebouwde acties combineren of `OnPage`/`OnPageEnd` gebruiken.

### LayoutFrame

Een `LayoutFrame` is een rechthoekig gebied op de pagina met een interne
Y-cursor die naar beneden beweegt naarmate inhoud wordt toegevoegd.  Als de
cursor de bodem bereikt, is het frame vol en schakelt de motor naar het volgende
frame of een nieuwe pagina.

### PageTemplate

Een `PageTemplate` koppelt een naam aan een geordende lijst van `LayoutFrame`-objecten
plus optionele callbacks voor kop- en voetteksten.  Meerdere templates maken
wisselende paginalay-outs mogelijk (titelpagina, één kolom, twee kolommen, enz.).

### DocTemplate

`DocTemplate` is de motor.  Het verwerkt de story stap voor stap, plaatst
flowables in frames, wisselt frames, start nieuwe pagina's en voert
template-wisselingen uit.

---

## 2. Meeteenheden en coördinaten

Alle afmetingen zijn in **punten** (pt).  1 punt = 1/72 inch ≈ 0,353 mm.

Veelgebruikte omrekeningen:

| Eenheid | Punten |
|---------|--------|
| 1 cm    | 28,35 pt |
| 1 inch  | 72 pt  |
| 10 mm   | 28,35 pt |

Het coördinatensysteem heeft de oorsprong **linksboven** op de pagina.
De Y-as loopt **naar beneden** (overeenkomstig de rest van de Nautilus-bibliotheek).

Standaard paginaformaten:

```go
pdf.PageSizeA3     // 841,89 × 1190,55 pt
pdf.PageSizeA4     // 595,28 × 841,89 pt
pdf.PageSizeA5     // 419,53 × 595,28 pt
pdf.PageSizeLetter // 612 × 792 pt
pdf.PageSizeLegal  // 612 × 1008 pt
```

---

## 3. Minimaal voorbeeld

```go
package main

import (
    "log"
    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/layout"
)

func main() {
    // 1. Maak een document aan.
    doc, err := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
    if err != nil {
        log.Fatal(err)
    }

    // 2. Registreer en activeer een lettertype.
    if err := doc.RegisterFont("regular", "/pad/naar/NotoSans-Regular.ttf"); err != nil {
        log.Fatal(err)
    }
    doc.SetFont("regular", 12) // stel standaardlettertype in vóór Build

    // 3. Definieer een stijl en bouw de story op.
    body := layout.ParagraphStyle{FontName: "regular", FontSize: 12, SpaceAfter: 8}
    story := []layout.Flowable{
        &layout.Paragraph{Text: "Hallo, Nautilus!", Style: body},
        &layout.Spacer{Height: 12},
        &layout.Paragraph{Text: "Tweede alinea.", Style: body},
    }

    // 4. Definieer het frame: het complete inhoudsgebied van de pagina.
    const margin = 50.0
    frame := &layout.LayoutFrame{
        X:      margin,
        Y:      margin,
        Width:  doc.PageWidth() - 2*margin,
        Height: doc.PageHeight() - 2*margin,
    }

    // 5. Koppel het frame aan een PageTemplate.
    tmpl := &layout.PageTemplate{
        ID:     "main",
        Frames: []*layout.LayoutFrame{frame},
    }

    // 6. Maak de DocTemplate, registreer het template en bouw.
    dt := layout.NewDocTemplate(doc)
    dt.AddPageTemplate(tmpl)
    if err := dt.Build(story); err != nil {
        log.Fatal(err)
    }

    // 7. Sla het PDF-bestand op.
    if err := doc.Save("output.pdf"); err != nil {
        log.Fatal(err)
    }
}
```

> **Let op:** `doc.SetFont` moet worden aangeroepen vóór `Build`, zodat tekstmeting
> al werkt vanaf de allereerste alinea.

---

## 4. Document en lettertypes

```go
doc, err := pdf.New(pdf.Config{
    PageSize:         pdf.PageSizeA4,
    DefaultFontSize:  12,
    LineHeightFactor: 1.4,
})
```

Lettertypes registreren en wisselen:

```go
doc.RegisterFont("regular", "NotoSans-Regular.ttf")
doc.RegisterFont("bold",    "NotoSans-Bold.ttf")
doc.RegisterFont("italic",  "NotoSans-Italic.ttf")

doc.SetFont("regular", 12) // activeer vóór Build
```

### Pagina- en inhoudsgebied-helpers

`pdf.Document` biedt hulpmethoden om pagina- en margegegevens op te vragen
zonder handmatige rekensom:

```go
doc.PageWidth()       // paginabreedte in pt
doc.PageHeight()      // paginahoogte in pt
doc.PageCount()       // aantal pagina's na Build (bruikbaar in OnPageEnd)

doc.ContentX()        // linkerrand van het inhoudsgebied (= leftMargin)
doc.ContentY()        // bovenrand van het inhoudsgebied (= topMargin)
doc.ContentWidth()    // breedte van het inhoudsgebied
doc.ContentHeight()   // hoogte van het inhoudsgebied
doc.ContentRightX()   // rechterrand van het inhoudsgebied
doc.ContentBottomY()  // onderrand van het inhoudsgebied
```

Gebruik deze methoden bij het definiëren van frames:

```go
frame := &layout.LayoutFrame{
    X:      doc.ContentX(),
    Y:      doc.ContentY(),
    Width:  doc.ContentWidth(),
    Height: doc.ContentHeight(),
}
```

In een `ParagraphStyle` schakelt `FontName` automatisch naar het juiste lettertype
bij rendering.  Als `FontName` leeg is, gebruikt de alinea het lettertype dat op
dat moment actief is op het document.

---

## 5. Paragraph

`Paragraph` is de primaire tekstelement.  Ondersteunt woordafbreking,
lettertypewisseling, kleur, inspringing en horizontale uitlijning.

### Struct-definitie

```go
type Paragraph struct {
    Text  string         // de te renderen tekst; gebruik \n voor expliciete regeleinden
    Style ParagraphStyle // visuele opmaak
}
```

### ParagraphStyle

```go
type ParagraphStyle struct {
    FontName         string     // geregistreerde lettertypenaam; leeg = huidig lettertype
    FontSize         float64    // grootte in punten; 0 = 12 pt
    Leading          float64    // regelafstand in punten; 0 = FontSize × 1,2
    Alignment        HAlign     // AlignLeft (standaard), AlignCenter, AlignRight
    SpaceBefore      float64    // extra witruimte boven de alinea
    SpaceAfter       float64    // extra witruimte onder de alinea
    KeepWithNextPara bool       // geen frame-/paginasprong na deze alinea
    LeftIndent       float64    // inspringing links in punten
    RightIndent      float64    // inspringing rechts in punten
    TextColor        *pdf.Color // tekstkleur; nil = huidig documentkleur
}
```

### Uitlijning

```go
const (
    AlignLeft   HAlign = iota // standaard
    AlignCenter
    AlignRight
)
```

### Voorbeelden

**Basisalinea:**
```go
body := layout.ParagraphStyle{
    FontName:   "regular",
    FontSize:   11,
    SpaceAfter: 6,
}
p := &layout.Paragraph{Text: "De snelle bruine vos springt over de luie hond.", Style: body}
```

**Gecentreerde ondertitel:**
```go
subtitle := layout.ParagraphStyle{
    FontName:   "regular",
    FontSize:   14,
    Alignment:  layout.AlignCenter,
    SpaceAfter: 16,
}
```

**Gekleurd en ingesprongen:**
```go
navy := pdf.ColorNavy
callout := layout.ParagraphStyle{
    FontName:    "bold",
    FontSize:    10,
    LeftIndent:  20,
    RightIndent: 20,
    SpaceBefore: 8,
    SpaceAfter:  8,
    TextColor:   &navy,
}
```

**Regelafstand aanpassen:**
```go
ruim := layout.ParagraphStyle{
    FontName: "regular",
    FontSize: 11,
    Leading:  18, // vaste regelafstand van 18 pt in plaats van 11 × 1,2 = 13,2 pt
}
```

**Expliciete regeleinden met \n:**
```go
&layout.Paragraph{
    Text: "Regel één\nRegel twee\nRegel drie",
    Style: body,
}
```

### Splitsing over frames

Lange alinea's worden automatisch gesplitst als ze niet passen in de resterende
frameruimte.  Het eerste deel wordt in het huidige frame getekend; de rest wordt
naar het volgende frame doorgeschoven.  U hoeft hier niets voor te doen.

---

## 6. Spacer

`Spacer` reserveert een vaste hoeveelheid verticale ruimte zonder iets te tekenen.

```go
type Spacer struct {
    Width  float64 // 0 of negatief = volledige beschikbare breedte
    Height float64 // te reserveren hoogte in punten
}
```

**Voorbeelden:**

```go
// 12 punten witruimte
&layout.Spacer{Height: 12}

// Half inch (36 pt) witruimte
&layout.Spacer{Height: 36}
```

> **Tip:** Gebruik `SpaceBefore` en `SpaceAfter` in `ParagraphStyle` voor
> automatische tussenruimte rondom alinea's.  Gebruik `Spacer` voor eenmalige,
> expliciete witruimte.

---

## 7. HRFlowable

`HRFlowable` tekent een horizontale lijn als een gevulde balk.

```go
type HRFlowable struct {
    Width     float64   // breedte: > 1.0 = absolute punten;
                        //          0..1.0 = fractie van beschikbare breedte (bv. 0.8 = 80 %)
                        //          0 = volledige beschikbare breedte
    Thickness float64   // hoogte van de balk in punten; standaard 1
    Color     pdf.Color // vulkleur
    Align     HAlign    // uitlijning als Width < beschikbare breedte
    Before    float64   // witruimte boven de lijn
    After     float64   // witruimte onder de lijn
}
```

**Voorbeelden:**

```go
// Dunne grijze lijn over de volledige breedte
&layout.HRFlowable{
    Thickness: 0.75,
    Color:     pdf.ColorLightGray,
    Before:    8,
    After:     8,
}

// Dikke marineblauwe lijn, 60 % breedte, gecentreerd
&layout.HRFlowable{
    Width:     0.6,
    Thickness: 2,
    Color:     pdf.ColorNavy,
    Align:     layout.AlignCenter,
    Before:    12,
    After:     12,
}

// Rode lijn met absolute breedte van 100 punten, rechts uitgelijnd
&layout.HRFlowable{
    Width:     100,
    Thickness: 1,
    Color:     pdf.ColorRed,
    Align:     layout.AlignRight,
}
```

---

## 8. KeepTogether

`KeepTogether` voorkomt dat een groep flowables over frames of pagina's wordt
gesplitst.

```go
type KeepTogether struct {
    Flowables []Flowable // de samen te houden elementen
}
```

**Gedrag:**

1. Past de groep in de resterende frameruimte → direct tekenen.
2. Past niet → motor voegt een `FrameBreak` in en probeert het opnieuw in het
   volgende frame.
3. Past zelfs in een leeg frame niet → motor splitst de individuele flowables
   afzonderlijk (om oneindige lussen te vermijden).

**Typisch gebruik — koptekst altijd samen met eerste alinea:**

```go
h1 := layout.ParagraphStyle{FontName: "bold", FontSize: 14, SpaceAfter: 4}
body := layout.ParagraphStyle{FontName: "regular", FontSize: 11, SpaceAfter: 6}

story = append(story, &layout.KeepTogether{
    Flowables: []layout.Flowable{
        &layout.Paragraph{Text: "Hoofdstuk 3 — Resultaten", Style: h1},
        &layout.Paragraph{Text: "In dit hoofdstuk bespreken wij...", Style: body},
    },
})
```

**Alternatief via KeepWithNextPara:**

Als een alineastijl `KeepWithNextPara: true` heeft, groepeert de motor die alinea
automatisch samen met de volgende.  Dit is handig voor koptekststijlen:

```go
h1Style := layout.ParagraphStyle{
    FontName:         "bold",
    FontSize:         14,
    KeepWithNextPara: true, // nooit van de eerste alinea eronder gescheiden
}
```

> **Let op:** `KeepWithNextPara` koppelt de alinea alleen aan de *eerstvolgende*
> flowable.  Voor langere groepen gebruikt u `KeepTogether` expliciet.

---

## 9. Action Flowables

Action Flowables zijn onzichtbare, nul-hoge elementen die de motor sturen in
plaats van zichtbare inhoud te tekenen.  Ze worden in de story-lijst ingevoegd
als besturingssignalen.

### PageBreak

Forceert een directe paginasprong.

```go
type PageBreak struct {
    NextTemplate string // optionele template-ID voor de nieuwe pagina
}
```

```go
// Eenvoudige paginasprong
&layout.PageBreak{}

// Paginasprong én meteen wisselen naar het "twee-kolommen" template
&layout.PageBreak{NextTemplate: "twee-kolommen"}
```

### FrameBreak

Schakelt naar het volgende frame in het huidige template (of naar een nieuwe
pagina als het huidige frame het laatste is).

```go
&layout.FrameBreak{}
```

Gebruik dit om inhoud expliciet naar de tweede kolom te sturen:

```go
story = append(story,
    linkerKolomInhoud...,
    &layout.FrameBreak{}, // spring naar rechterkolom
    rechterKolomInhoud...,
)
```

### CondPageBreak

Voegt een paginasprong in *alleen* als er minder dan `MinHeight` punten over zijn
in het huidige frame.  Handig om te voorkomen dat een sectie op een kleine
resterende ruimte begint.

```go
type CondPageBreak struct {
    MinHeight float64
}
```

```go
// Paginasprong als er minder dan 72 pt (1 inch) over is
&layout.CondPageBreak{MinHeight: 72}

// Paginasprong als er minder dan 3 regels à 14 pt over zijn
&layout.CondPageBreak{MinHeight: 3 * 14 * 1.2}
```

**Goede positie:** vlak vóór een sectiekop of een groep die op een
"schone" nieuwe pagina moet beginnen.

### NextPageTemplate

Plant een template-wissel die van kracht wordt bij de *volgende* paginasprong.
De huidige pagina blijft het huidige template gebruiken.

```go
type NextPageTemplate struct {
    TemplateID string
}
```

```go
// Schakel naar twee kolommen op de volgende pagina
&layout.NextPageTemplate{TemplateID: "twee-kolommen"},
&layout.PageBreak{},
```

Het verschil met `PageBreak{NextTemplate: "..."}`:

| | `NextPageTemplate` + `PageBreak{}` | `PageBreak{NextTemplate: "..."}` |
|---|---|---|
| Wanneer wisselt het? | Op de volgende `PageBreak` | Direct bij de `PageBreak` |
| Huidige pagina | blijft ongewijzigd | wordt al beëindigd |
| Gebruik | Altijd veilig | Als u meteen wilt wisselen |

In de praktijk zijn beide gelijkwaardig als ze vlak na elkaar staan.

---

## 10. LayoutFrame

`LayoutFrame` definieert een rechthoekig inhoudsgebied op de pagina.

```go
type LayoutFrame struct {
    X, Y         float64     // linksboven-coördinaat op de pagina (punten)
    Width, Height float64    // buitenafmetingen (punten)
    Padding      pdf.Padding // binnenste witruimte
    ID           string      // optionele naam (voor foutopsporing)
    ShowBoundary bool        // teken omlijning om het frame (foutopsporing)
}
```

### Padding

```go
frame := &layout.LayoutFrame{
    X: 50, Y: 80,
    Width: 495, Height: 700,
    Padding: pdf.Padding{Top: 8, Right: 12, Bottom: 8, Left: 12},
}

// Snelkoppelingen:
frame.Padding = pdf.UniformPadding(10)       // 10 pt op alle zijden
frame.Padding = pdf.HorizontalPadding(12, 8) // 12 pt links/rechts, 8 pt boven/onder
```

### Beschikbare breedte

De *binnenste* breedte waarop flowables kunnen renderen:

```
binnenste breedte = Width − Padding.Left − Padding.Right
```

Flowables ontvangen deze waarde als `availWidth` in hun `Wrap`-aanroep.

### Handige berekeningsmethode

Gebruik de afmetingen van het document als startpunt:

```go
const (
    margin  = 50.0
    headerH = 40.0
    footerH = 36.0
)
frame := &layout.LayoutFrame{
    X:      margin,
    Y:      margin + headerH,
    Width:  doc.PageWidth() - 2*margin,
    Height: doc.PageHeight() - margin - headerH - footerH,
}
```

### ShowBoundary

Stel `ShowBoundary: true` in tijdens het ontwikkelen om de framecontouren als
dunne grijze rechthoeken zichtbaar te maken.  Verwijder dit in productie.

```go
frame := &layout.LayoutFrame{
    X: 50, Y: 80, Width: 495, Height: 700,
    ShowBoundary: true, // alleen voor debuggen
}
```

---

## 11. PageTemplate

`PageTemplate` koppelt een naam aan een geordende lijst van frames plus
optionele page-decorators.

```go
type PageTemplate struct {
    ID               string          // unieke naam, gebruikt door NextPageTemplate
    Frames           []*LayoutFrame  // frames op volgorde van vulling
    OnPage           PageDecorator   // callback na AddPage (headers, watermerken)
    OnPageEnd        PageDecorator   // callback voor paginaafsluiting (footers)
    AutoNextTemplate string          // template-ID na deze pagina; leeg = zelfde template
}
```

### AutoNextTemplate

Met `AutoNextTemplate` hoeft u geen `NextPageTemplate`-acties in uw story te zetten
als u wilt dat alle pagina's van hetzelfde type zijn:

```go
singleTemplate := &layout.PageTemplate{
    ID:               "single",
    Frames:           []*layout.LayoutFrame{singleFrame},
    OnPage:           headerFooter,
    AutoNextTemplate: "single", // elke volgende pagina ook "single"
}
```

Zonder `AutoNextTemplate` gebruikt de motor steeds het huidige template totdat
u het expliciet wisselt.

### Meerdere templates registreren

Het eerste geregistreerde template wordt gebruikt voor pagina 1.

```go
dt := layout.NewDocTemplate(doc)
dt.AddPageTemplate(titleTemplate)   // pagina 1
dt.AddPageTemplate(bodyTemplate)    // wordt actief na een wissel
dt.AddPageTemplate(twoColTemplate)
```

---

## 12. DocTemplate en Build

```go
dt := layout.NewDocTemplate(doc)
dt.AddPageTemplate(tmpl)            // registreer minimaal één template
err := dt.Build(story)              // verwerk de story
```

### Wat Build doet

```
startPage()
  → pageNum++
  → doc.AddPage()
  → reset alle frames in het huidige template
  → roep OnPage-decorator aan

Hoofdlus (voor elk element in de story):
  1. ActionFlowable? → direct uitvoeren, ga door
  2. KeepWithNext=true? → verzamel keten, wikkel in KeepTogether
  3. frame.add(flowable) → past het?
     JA  → teken, cursor omlaag, ga door
     NEE → probeer frame.split(flowable)
           Splitsen gelukt? → zet onderdelen terug in de wachtrij
           Splitsen mislukt? → advanceFrame() (naar volgend frame of pagina)

endPage()
  → roep OnPageEnd-decorator aan
  → verwerk pending template-wissel
```

### Foutafhandeling

`Build` geeft een fout terug als:
- er geen templates zijn geregistreerd.
- een flowable meer dan 10 keer achter elkaar niet in een frame past (kringloop
  of te grote content).

```go
if err := dt.Build(story); err != nil {
    log.Fatalf("build mislukt: %v", err)
}
```

### Na Build

Het `pdf.Document` bevat alle pagina's.  Sla op als gewoonlijk:

```go
doc.Save("output.pdf")
// of naar een io.Writer:
doc.Output(os.Stdout)
```

---

## 13. Pagina-decorators (headers en footers)

`PageDecorator` is een functie die de motor aanroept bij het begin of einde
van elke pagina:

```go
type PageDecorator func(doc *pdf.Document, pageNum int)
```

`pageNum` is 1-gebaseerd.  U kunt alles tekenen wat `pdf.Document` ondersteunt:
tekst, kaders, lijnen, afbeeldingen.

### Voorbeeld

```go
const (
    margin  = 50.0
    headerH = 40.0
    footerH = 36.0
)
pageW := doc.PageWidth()
pageH := doc.PageHeight()
contentW := pageW - 2*margin

pageDecorator := func(d *pdf.Document, pageNum int) {
    // ── Koptekst ──────────────────────────────────────────────────────────
    spec := &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}
    d.DrawBorder(margin, margin, contentW, headerH-4, pdf.Border{Bottom: spec})

    d.SetFont("regular", 8)
    d.SetTextColor(100, 100, 100)
    d.WriteLine("Mijn Document", margin, margin+10)

    num := fmt.Sprintf("Pagina %d", pageNum)
    w, _ := d.MeasureText(num)
    d.WriteLine(num, pageW-margin-w, margin+10)

    // ── Voettekst ─────────────────────────────────────────────────────────
    footerY := pageH - footerH
    d.DrawBorder(margin, footerY, contentW, 0, pdf.Border{
        Top: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray},
    })
    d.SetFont("regular", 8)
    d.SetTextColor(150, 150, 150)
    d.WriteLine("Vertrouwelijk", margin, footerY+8)
}

tmpl := &layout.PageTemplate{
    ID:     "main",
    Frames: []*layout.LayoutFrame{frame},
    OnPage: pageDecorator, // wordt aangeroepen aan het begin van elke pagina
}
```

### OnPage versus OnPageEnd

| Decorator  | Wordt aangeroepen | Typisch gebruik |
|------------|-------------------|-----------------|
| `OnPage`   | Direct na `AddPage` (vóór inhoud) | Koptekst, watermerk, framecontouren |
| `OnPageEnd`| Net vóór paginaafsluiting (ná inhoud) | Voettekst, paginanummer, dynamische informatie |

```go
tmpl := &layout.PageTemplate{
    ID:     "main",
    Frames: []*layout.LayoutFrame{frame},
    OnPage: func(d *pdf.Document, pageNum int) {
        // Koptekst — getekend vóór de inhoud van de pagina
        d.SetFont("regular", 8)
        d.WriteLine("Mijn Document", margin, margin+10)
    },
    OnPageEnd: func(d *pdf.Document, pageNum int) {
        // Voettekst — getekend ná de inhoud van de pagina
        footerY := d.PageHeight() - margin + 8
        d.SetFont("regular", 8)
        num := fmt.Sprintf("Pagina %d", pageNum)
        w, _ := d.MeasureText(num)
        d.WriteLine(num, d.PageWidth()-margin-w, footerY)
    },
}
```

### Eerste pagina overslaan

```go
pageDecorator := func(d *pdf.Document, pageNum int) {
    if pageNum == 1 {
        return // geen koptekst op de titelpagina
    }
    // ... rest van de koptekst
}
```

### Frames reserveren voor koptekst en voettekst

De decorator tekent *buiten* de `LayoutFrame`-gebieden.  Zorg dat uw frames
voldoende marge laten voor de koptekst en voettekst:

```go
frame := &layout.LayoutFrame{
    X:      margin,
    Y:      margin + headerH, // onder de koptekst
    Width:  pageW - 2*margin,
    Height: pageH - margin - headerH - footerH, // boven de voettekst
}
```

---

## 14. Template wisselen

Gebruik template-wisseling wanneer verschillende pagina's een andere lay-out
nodig hebben: een titelpagina, een inhoudsopgave, eenkoloms inhoud, en twee
kolommen voor dense tabellen.

### Methode 1 — `NextPageTemplate` + `PageBreak`

De meest gebruikelijke aanpak.  `NextPageTemplate` plant de wissel; `PageBreak`
voert de sprong uit.

```go
story = append(story,
    // ... eenkoloms inhoud ...
    &layout.NextPageTemplate{TemplateID: "twee-kolommen"},
    &layout.PageBreak{},
    // ... tweetkoloms inhoud ...
    &layout.NextPageTemplate{TemplateID: "single"},
    &layout.PageBreak{},
    // ... eenkoloms inhoud ...
)
```

### Methode 2 — `PageBreak` met `NextTemplate`

Combineert paginasprong en template-wissel in één stap:

```go
&layout.PageBreak{NextTemplate: "twee-kolommen"},
```

### Methode 3 — `AutoNextTemplate`

Automatische wissel na elke pagina, zonder story-acties.  Handig voor
"titelpagina → rest":

```go
titleTemplate := &layout.PageTemplate{
    ID:               "title",
    Frames:           []*layout.LayoutFrame{titleFrame},
    OnPage:           titleDecorator,
    AutoNextTemplate: "body", // na pagina 1 automatisch naar "body"
}
bodyTemplate := &layout.PageTemplate{
    ID:               "body",
    Frames:           []*layout.LayoutFrame{bodyFrame},
    OnPage:           bodyDecorator,
    AutoNextTemplate: "body", // blijf op "body"
}
```

### Registratiepriorteit

Het **eerste** geregistreerde template wordt gebruikt voor pagina 1:

```go
dt.AddPageTemplate(titleTemplate) // pagina 1
dt.AddPageTemplate(bodyTemplate)  // actief na wissel
```

---

## 15. Meerkoloms-layout

Meerdere kolommen worden gerealiseerd door meerdere `LayoutFrame`-objecten aan
een `PageTemplate` toe te voegen.  De motor vult ze van links naar rechts
(in de volgorde van de `Frames`-lijst).

```go
const (
    margin    = 50.0
    colGutter = 12.0  // ruimte tussen de kolommen
)
contentW := doc.PageWidth() - 2*margin
colW     := (contentW - colGutter) / 2
contentY := 90.0
contentH := doc.PageHeight() - contentY - 50.0

leftFrame := &layout.LayoutFrame{
    X: margin,          Y: contentY,
    Width: colW,        Height: contentH,
}
rightFrame := &layout.LayoutFrame{
    X: margin + colW + colGutter, Y: contentY,
    Width: colW,                  Height: contentH,
}

twoColTemplate := &layout.PageTemplate{
    ID:               "two-column",
    Frames:           []*layout.LayoutFrame{leftFrame, rightFrame},
    OnPage:           pageDecorator,
    AutoNextTemplate: "two-column",
}
```

> **Belangrijk:** De content stroomt *sequentieel* van kolom naar kolom.
> De linkerkolom wordt volledig gevuld voordat de rechterkolom inhoud ontvangt.
> Dit is hetzelfde gedrag als Platypus.  Er is geen automatische balancering
> (zoals bij krantenkolomopmaak).

### Expliciet naar de rechterkolom springen

```go
story = append(story,
    linkerKolomInhoud...,
    &layout.FrameBreak{}, // forceer overgang naar rechterkolom
    rechterKolomInhoud...,
)
```

### Driekoloms-layout

```go
colW := (contentW - 2*colGutter) / 3

col1 := &layout.LayoutFrame{X: margin,              Y: contentY, Width: colW, Height: contentH}
col2 := &layout.LayoutFrame{X: margin+colW+colGutter,     Y: contentY, Width: colW, Height: contentH}
col3 := &layout.LayoutFrame{X: margin+2*(colW+colGutter), Y: contentY, Width: colW, Height: contentH}

threeColTemplate := &layout.PageTemplate{
    ID:     "three-column",
    Frames: []*layout.LayoutFrame{col1, col2, col3},
    OnPage: pageDecorator,
    AutoNextTemplate: "three-column",
}
```

---

## 16. Foutopsporing met ShowBoundary

Schakel `ShowBoundary: true` in op frames om hun contouren als dunne grijze
rechthoeken te zien.  Zo kunt u bevestigen dat:

- Frames op de juiste positie staan.
- Het juiste template actief is.
- Beide kolommen worden geactiveerd (ook als één leeg is).

```go
leftFrame := &layout.LayoutFrame{
    X: margin, Y: contentY,
    Width: colW, Height: contentH,
    ShowBoundary: true, // ← zichtbaar in het PDF
}
```

Verwijder `ShowBoundary: true` wanneer het document klaar is voor productie.

---

## 17. Eigen Flowable implementeren

Implementeer de `Flowable`-interface om aangepaste elementen te maken.  Alle
zeven methoden zijn vereist.

### Minimale implementatie

```go
type MijnFlowable struct {
    Width, Height float64
    // ... eigen velden ...
}

func (m *MijnFlowable) Wrap(_ *pdf.Document, availWidth, _ float64) (float64, float64) {
    w := m.Width
    if w <= 0 || w > availWidth {
        w = availWidth
    }
    return w, m.Height
}

func (m *MijnFlowable) Draw(doc *pdf.Document, x, y float64) error {
    // Teken hier met doc.WriteLine, doc.DrawBorder, doc.FillRect, enz.
    return nil
}

func (m *MijnFlowable) Split(_ *pdf.Document, _, _ float64) []Flowable { return nil }
func (m *MijnFlowable) SpaceBefore() float64                           { return 0 }
func (m *MijnFlowable) SpaceAfter() float64                            { return 0 }
func (m *MijnFlowable) KeepWithNext() bool                             { return false }
func (m *MijnFlowable) MinWidth() float64                              { return 0 }
```

### Richtlijnen

| Methode | Richtlijn |
|---------|-----------|
| `Wrap` | Sla de beschikbare breedte op voor gebruik in `Draw`. Negeer `availHeight` als de hoogte vast is. |
| `Draw` | Teken alleen binnen het gebied `(x, y)` tot `(x+width, y+height)`. |
| `Split` | Geef `nil` terug als splitsing niet mogelijk is (motor verschuift de flowable naar het volgende frame). Geef twee flowables terug bij splitsing: eerste deel past in `availHeight`, tweede deel bevat de rest. |
| `SpaceBefore/After` | Gebruik voor extra witruimte; de motor past dit toe en onderdrukt `SpaceBefore` bij het begin van een frame. |
| `KeepWithNext` | Geef `true` terug om te voorkomen dat er een framesprong na dit element komt. |

### Voorbeeld — gekleurde kader-flowable

```go
type ColorBox struct {
    BoxWidth, BoxHeight float64
    FillColor           pdf.Color
    BorderColor         pdf.Color
    SpaceBeforeVal      float64
    SpaceAfterVal       float64
}

func (cb *ColorBox) Wrap(_ *pdf.Document, availWidth, _ float64) (float64, float64) {
    w := cb.BoxWidth
    if w <= 0 || w > availWidth {
        w = availWidth
    }
    return w, cb.BoxHeight
}

func (cb *ColorBox) Draw(doc *pdf.Document, x, y float64) error {
    doc.FillRect(x, y, cb.BoxWidth, cb.BoxHeight, cb.FillColor)
    spec := pdf.BorderSpec{Thickness: 1, Color: cb.BorderColor}
    return doc.DrawBorder(x, y, cb.BoxWidth, cb.BoxHeight, pdf.NewUniformBorder(spec))
}

func (cb *ColorBox) Split(_ *pdf.Document, _, _ float64) []layout.Flowable { return nil }
func (cb *ColorBox) SpaceBefore() float64                                   { return cb.SpaceBeforeVal }
func (cb *ColorBox) SpaceAfter() float64                                    { return cb.SpaceAfterVal }
func (cb *ColorBox) KeepWithNext() bool                                     { return false }
func (cb *ColorBox) MinWidth() float64                                      { return cb.BoxWidth }
```

Gebruik:

```go
story = append(story,
    &ColorBox{
        BoxWidth: 0, BoxHeight: 40, // 0 breedte = volledige beschikbare breedte
        FillColor:   pdf.Color{R: 235, G: 245, B: 255},
        BorderColor: pdf.ColorNavy,
        SpaceBeforeVal: 8,
        SpaceAfterVal:  8,
    },
)
```

### Splitsbare flowable

Als een flowable over frames gesplitst kan worden, implementeer dan `Split`:

```go
func (mf *MijnFlowable) Split(doc *pdf.Document, availWidth, availHeight float64) []layout.Flowable {
    if availHeight < mf.MinimumHeight() {
        return nil // past zelfs in een klein stuk niet; verschuif naar volgend frame
    }
    deel1 := &MijnFlowable{/* inhoud die past in availHeight */}
    deel2 := &MijnFlowable{/* resterende inhoud */}
    return []layout.Flowable{deel1, deel2}
}
```

> **Contractregel:** de som van de hoogten van alle teruggegeven onderdelen moet
> gelijk zijn aan de originele hoogte.  Er mag geen inhoud verloren gaan.

---

## 18. Veelgestelde vragen

### Waarom verschijnt alle inhoud in één kolom bij een tweetkolomslayout?

De motor vult kolommen sequentieel: de linkerkolom wordt eerst volledig gevuld
voordat de rechterkolom inhoud ontvangt.  Als uw inhoud past in de linkerkolom,
blijft de rechterkolom leeg.

**Oplossing:** Voeg meer inhoud toe, of gebruik `FrameBreak` om expliciet naar
de rechterkolom te springen.

Schakel `ShowBoundary: true` in om de framecontouren te controleren.

---

### Waarom loopt tekst over de onderrand van het frame?

Flowables worden pas gesplitst als ze *niet passen* in de resterende ruimte.  Als
een flowable geen `Split`-implementatie heeft (geeft `nil` terug), verschuift de
motor hem naar het volgende frame.  Controleer of uw frames groot genoeg zijn en
of `Split` correct werkt voor aangepaste flowables.

---

### Hoe voeg ik "Pagina X van Y" toe aan de voettekst?

`DocTemplate` biedt geen ingebouwde twee-pass rendering.  Er zijn twee
praktische oplossingen:

**Optie 1 — Totaal vooraf bekend of handmatig berekend:**

```go
var totalPages int  // bepaal dit vooraf als het document een vaste lengte heeft

tmpl.OnPageEnd = func(d *pdf.Document, pageNum int) {
    label := fmt.Sprintf("Pagina %d van %d", pageNum, totalPages)
    // ... teken label in voettekst
}
```

**Optie 2 — Totaal na `Build` beschikbaar:**

```go
dt.Build(story)                       // genereer het document
total := doc.PageCount()              // lees het totale aantal pagina's

// Teken paginanummers handmatig in een tweede stap, of
// gebruik totaal als geheugensteuntje in een volgende versie van het document.
```

Als het totale paginatelling niet nodig is, volstaat `Pagina %d`
met alleen `pageNum` in `OnPageEnd`.

---

### Kan ik `pdf.Frame` en `DocTemplate` door elkaar gebruiken?

Ja.  `pdf.Frame` (de lager-niveau frame-API) en `DocTemplate` zijn volledig
onafhankelijk.  Gebruik `pdf.Frame` voor vrijgeplaatste kaders (bijv.
tekstblokken naast afbeeldingen in een koptekst); gebruik `DocTemplate` voor de
hoofdinhoudsstroom.

---

### Mijn flowable past niet in enig frame en Build geeft een fout.

`Build` retourneert een fout als een flowable meer dan 10 keer niet geplaatst kan
worden.  Controleer of:

- De `Height` van uw frame groot genoeg is.
- `Wrap` een realistische hoogte teruggeeft (niet groter dan de framehoogte).
- Er geen kringloop is in een aangepaste `Split`-implementatie.

---

### Hoe stel ik de lettergrootte opnieuw in na een paginasprong?

De `PageDecorator` (`OnPage`) wordt na elke `AddPage` aangeroepen en stelt
typisch het lettertype in voor de koptekst.  Uw eerste alinea-stijl op de nieuwe
pagina roept `applyFont` aan bij `Wrap`, zodat het lettertype direct wordt
hersteld.  Zorg dat elke `ParagraphStyle` een `FontName` en `FontSize` heeft.

---

## Referentie-overzicht

### Types

| Type | Beschrijving |
|------|-------------|
| `Flowable` | Interface voor alle plaatsbare elementen |
| `Paragraph` | Tekstelement met woordafbreking |
| `ParagraphStyle` | Visuele opmaak voor `Paragraph` |
| `HAlign` | Uitlijningsconstante: `AlignLeft`, `AlignCenter`, `AlignRight` |
| `Spacer` | Reserveert verticale ruimte |
| `HRFlowable` | Horizontale lijn als gevulde balk |
| `KeepTogether` | Houdt een groep elementen samen |
| `PageBreak` | Forceert een paginasprong |
| `FrameBreak` | Schakelt naar het volgende frame |
| `CondPageBreak` | Voorwaardelijke paginasprong |
| `NextPageTemplate` | Plant een template-wissel |
| `LayoutFrame` | Rechthoekig inhoudsgebied op een pagina |
| `PageTemplate` | Pagina-lay-out: frames + decorators |
| `PageDecorator` | `func(doc *pdf.Document, pageNum int)` |
| `DocTemplate` | De layout-motor |

### Functies

| Functie | Beschrijving |
|---------|-------------|
| `NewDocTemplate(doc)` | Maak een nieuwe motor aan |
| `(dt) AddPageTemplate(pt)` | Registreer een template |
| `(dt) Build(story)` | Verwerk de story en genereer het PDF |

### pdf.Document — pagina-helpers

| Methode | Beschrijving |
|---------|-------------|
| `PageWidth()` | Paginabreedte in pt |
| `PageHeight()` | Paginahoogte in pt |
| `PageCount()` | Aantal pagina's (beschikbaar na `Build`) |
| `ContentX()` | Linkerrand van het inhoudsgebied (= leftMargin) |
| `ContentY()` | Bovenrand van het inhoudsgebied (= topMargin) |
| `ContentWidth()` | Breedte van het inhoudsgebied |
| `ContentHeight()` | Hoogte van het inhoudsgebied |
| `ContentRightX()` | Rechterrand van het inhoudsgebied |
| `ContentBottomY()` | Onderrand van het inhoudsgebied |

### ParagraphStyle — snel overzicht

| Veld | Type | Standaard | Beschrijving |
|------|------|-----------|-------------|
| `FontName` | `string` | `""` | Geregistreerde lettertypenaam |
| `FontSize` | `float64` | `0` (→ 12) | Grootte in punten |
| `Leading` | `float64` | `0` (→ FontSize×1,2) | Regelafstand |
| `Alignment` | `HAlign` | `AlignLeft` | Uitlijning |
| `SpaceBefore` | `float64` | `0` | Witruimte boven |
| `SpaceAfter` | `float64` | `0` | Witruimte onder |
| `KeepWithNextPara` | `bool` | `false` | Koppelen aan volgende flowable |
| `LeftIndent` | `float64` | `0` | Inspringing links |
| `RightIndent` | `float64` | `0` | Inspringing rechts |
| `TextColor` | `*pdf.Color` | `nil` | Tekstkleur |

### LayoutFrame — snel overzicht

| Veld | Type | Beschrijving |
|------|------|-------------|
| `X, Y` | `float64` | Linksboven-positie op de pagina |
| `Width, Height` | `float64` | Buitenafmetingen |
| `Padding` | `pdf.Padding` | Binnenste witruimte |
| `ID` | `string` | Optionele naam |
| `ShowBoundary` | `bool` | Teken debug-omlijning |
