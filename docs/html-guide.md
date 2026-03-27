# Nautilus HTML — Ontwikkelaarsgids

Het pakket `pdf/html` zet inline HTML-opmaak en HTML-tabellen om naar
structuren die de Nautilus PDF-bibliotheek kan renderen.  Er zijn twee
onafhankelijke functies:

1. **Inline HTML span-parsing** (`html.Parse`) — zet een string met HTML-tags
   om naar een slice van gestileerde `Span`-waarden voor het renderen van
   inline tekst.
2. **HTML-tabelparsing** (`html.ParseTable`) — zet een volledig `<table>`-element
   om naar een `HtmlTable` die het document via `doc.TableFromHTML` kan renderen.

---

## Inhoudsopgave

1. [Inline HTML — ondersteunde tags](#1-inline-html--ondersteunde-tags)
2. [Inline HTML parsen](#2-inline-html-parsen)
3. [CSS-klasse-koppeling](#3-css-klasse-koppeling)
4. [Inline spans renderen met WriteHTMLSpans](#4-inline-spans-renderen-met-writehtmlspans)
5. [HTML-tabel parsen](#5-html-tabel-parsen)
6. [HTML-tabellen renderen met TableFromHTML](#6-html-tabellen-renderen-met-tablefromhtml)
7. [Volledig voorbeeld — inline HTML](#7-volledig-voorbeeld--inline-html)
8. [Volledig voorbeeld — HTML-tabel](#8-volledig-voorbeeld--html-tabel)
9. [Bekende beperkingen](#9-bekende-beperkingen)

---

## 1. Inline HTML — ondersteunde tags

De inline parser herkent de volgende HTML-tags.  Tags mogen vrij genest worden.
Alle andere tags worden genegeerd (hun inhoud blijft wel als platte tekst
beschikbaar).

| Tag(s)                              | Effect          |
|-------------------------------------|-----------------|
| `<b>`, `<strong>`                   | Vet             |
| `<i>`, `<em>`, `<cite>`, `<var>`, `<dfn>` | Cursief  |
| `<u>`, `<ins>`                      | Onderstrepen    |
| `<s>`, `<strike>`, `<del>`          | Doorhalen       |
| `<code>`, `<tt>`, `<kbd>`, `<samp>` | Monospace       |
| elke tag met `class="…"`            | CSS-klasse      |

Zelfsluitende tags (bijv. `<br/>`) worden geaccepteerd maar leveren geen uitvoer;
gebruik `\n` of converteer `<br>` naar `\n` voor regelafbrekingen.

---

## 2. Inline HTML parsen

```go
import "github.com/gvanbeck/nautilus/pdf/html"

spans, err := html.Parse(
    `Hallo <b>wereld</b> en <i>cursief <b>vet-cursief</b></i> tekst.`,
    nil, // geen CSS-klasse-koppeling
)
```

`Parse` geeft een `[]html.Span` terug waarbij elke span bevat:

```go
type Span struct {
    Text  string     // platte tekstinhoud van deze span
    Style Style      // gecumuleerde opmaak
    Class string     // binnenste CSS-klassenaam, indien aanwezig
}

type Style struct {
    Bold          bool
    Italic        bool
    Underline     bool
    Strikethrough bool
    Monospace     bool
}
```

Opeenvolgende tekens met dezelfde stijl worden gegroepeerd in één span.
De parser gaat soepel om met niet-gesloten of verkeerd geordende tags.

---

## 3. CSS-klasse-koppeling

Geef een `ClassStyle`-map mee aan `Parse` om CSS-klassenamen te vertalen naar
opmaakmarkeringen.  De klassenaam wordt altijd bewaard in `Span.Class`,
ongeacht of er een koppeling bestaat.

```go
klassen := html.ClassStyle{
    "markering": html.Style{Bold: true},
    "waarschuwing": html.Style{Italic: true, Underline: true},
    "code":     html.Style{Monospace: true},
}

spans, err := html.Parse(
    `Normale <span class="markering">belangrijke</span> tekst.`,
    klassen,
)
// spans[1].Style.Bold == true
// spans[1].Class      == "markering"
```

Als er geen koppeling opgegeven is (of een klassenaam niet in de map staat),
wordt de klassenaam toch opgeslagen in `Span.Class` zodat de aanroeper
aangepaste renderlogica kan toepassen.

---

## 4. Inline spans renderen met WriteHTMLSpans

`doc.WriteHTMLSpans` rendert een `[]html.Span` inline op de pagina.  U geeft
een callback (`fontFor`) mee die een `Style` omzet naar een geregistreerde
lettertypenaam, zodat u kunt wisselen tussen uw reguliere, vette, cursieve en
monospace lettertypen.

```go
import (
    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/html"
)

// Registreer vier lettervarianten.
doc.RegisterFont("regular", "/pad/naar/Roboto-Regular.ttf")
doc.RegisterFont("bold",    "/pad/naar/Roboto-Bold.ttf")
doc.RegisterFont("italic",  "/pad/naar/Roboto-Italic.ttf")
doc.RegisterFont("mono",    "/pad/naar/RobotoMono-Regular.ttf")

fontFor := func(s html.Style) string {
    switch {
    case s.Bold && s.Italic: return "bold"
    case s.Bold:             return "bold"
    case s.Italic:           return "italic"
    case s.Monospace:        return "mono"
    default:                 return "regular"
    }
}

spans, _ := html.Parse(`Prijs: <b>€ 99</b> <i>(excl. btw)</i>`, nil)

// endX is de X-positie na het laatste teken — handig om inhoud op
// dezelfde regel voort te zetten.
endX, err := doc.WriteHTMLSpans(spans, fontFor, 11, x, y)
```

Onderstrepen en doorhalen worden automatisch getekend als dunne horizontale
lijnen direct na de tekst.

---

## 5. HTML-tabel parsen

`html.ParseTable` parst een string met een `<table>`-element en geeft een
`HtmlTable` terug.

```go
htmlSrc := `
<table>
  <thead>
    <tr><th>Product</th><th>Prijs</th><th>Voorraad</th></tr>
  </thead>
  <tbody>
    <tr><td>Widget A</td><td align="right">€ 12,50</td><td>145</td></tr>
    <tr bgcolor="#f0f0f0"><td>Widget B</td><td align="right">€ 8,00</td><td>32</td></tr>
  </tbody>
</table>`

tabel, err := html.ParseTable(htmlSrc)
```

### Ondersteunde tabelstructuurelementen

| Element             | Beschrijving                                                     |
|---------------------|------------------------------------------------------------------|
| `<table>`           | Rootelement                                                      |
| `<caption>`         | Optioneel bijschrift (beschikbaar als `HtmlTable.Caption`)       |
| `<thead>`           | Koptekstsectie — rijen worden gemarkeerd als `IsHeader: true`    |
| `<tbody>`           | Gedeelte met gegevens                                            |
| `<tfoot>`           | Voettekstsectie — rijen worden gemarkeerd als `IsFooter: true`   |
| `<tr>`              | Tabelrij                                                         |
| `<th>`              | Kopcel — altijd `Bold: true`, markeert rij als koptekst          |
| `<td>`              | Gegevenscel                                                      |

### Ondersteunde cel- en rij-attributen

| Attribuut / eigenschap        | Waar       | Effect                                                    |
|-------------------------------|------------|-----------------------------------------------------------|
| `colspan="N"`                 | `<td>/<th>`| Cel beslaat N kolommen (`HtmlCell.ColSpan`)               |
| `rowspan="N"`                 | `<td>/<th>`| Cel beslaat N rijen (`HtmlCell.RowSpan`)                  |
| `align="left|center|right"`   | `<tr>/<td>`| Horizontale tekstuitlijning                               |
| `valign="top|middle|bottom"`  | `<tr>/<td>`| Verticale tekstuitlijning                                 |
| `bgcolor="#RRGGBB"`           | `<tr>/<td>`| Achtergrondkleur (ook via `style="background-color:…"`)   |
| `style="color:…"`             | `<td>/<th>`| Tekstkleur                                                |
| `style="font-weight:bold"`    | `<td>/<th>`| Vette tekst                                               |
| `style="text-align:…"`        | `<td>/<th>`| Horizontale uitlijning (overschrijft `align`)             |
| `style="vertical-align:…"`    | `<td>/<th>`| Verticale uitlijning (overschrijft `valign`)              |
| `nowrap`                      | `<td>/<th>`| Stelt `HtmlCell.NoWrap: true` in                         |
| `width="…"`                   | `<td>/<th>`| Breedtehint — opgeslagen in `HtmlCell.Width` als string   |

Celinhoud mag inline HTML-tags (`<b>`, `<i>`, `<u>`, enz.) bevatten en
`<br>` (omgezet naar `\n`).

---

## 6. HTML-tabellen renderen met TableFromHTML

Na het parsen zet u de `HtmlTable` om naar een `pdf.Table` met
`doc.TableFromHTML` en tekent u hem als elke andere tabel.

```go
import (
    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/html"
)

htmlTabel, err := html.ParseTable(htmlSrc)
if err != nil {
    log.Fatal(err)
}

cfg := pdf.TableConfig{
    ColWidths: []float64{200, 100, 80},  // verplicht — geen automatische breedte
    DefaultCellStyle: pdf.CellStyle{
        FontName: "regular",
        FontSize: 10,
        Padding:  pdf.Padding{Top: 4, Right: 6, Bottom: 4, Left: 6},
    },
}

htmlOpts := pdf.HtmlTableOptions{
    SpanFontFor: func(s html.Style) string {
        switch {
        case s.Bold:      return "bold"
        case s.Italic:    return "italic"
        case s.Monospace: return "mono"
        default:          return "regular"
        }
    },
    HeaderStyle: pdf.CellStyle{
        FontName:   "bold",
        FontSize:   10,
        Background: &pdf.Color{R: 0.2, G: 0.4, B: 0.7},
        TextColor:  &pdf.Color{R: 1, G: 1, B: 1},
    },
    FooterStyle: pdf.CellStyle{
        FontName: "italic",
        FontSize: 9,
    },
}

pdfTabel, err := doc.TableFromHTML(htmlTabel, cfg, htmlOpts)
if err != nil {
    log.Fatal(err)
}

// Teken de tabel op (x=50, y=100). Geeft de Y-positie onder de tabel terug.
eindY, err := pdfTabel.Draw(doc, 50, 100)
```

### HtmlTableOptions-velden

| Veld          | Type                             | Beschrijving                                             |
|---------------|----------------------------------|----------------------------------------------------------|
| `SpanFontFor` | `func(html.Style) string`        | Koppelt een Style aan een geregistreerde lettertypenaam  |
| `HeaderStyle` | `pdf.CellStyle`                  | Toegepast op `<thead>`-rijen en alle-`<th>`-rijen        |
| `FooterStyle` | `pdf.CellStyle`                  | Toegepast op `<tfoot>`-rijen                             |

Kleurwaarden uit `bgcolor` en inline `style="color:…"` worden automatisch
verwerkt en ondersteunen:
- Benoemde kleuren: `red`, `blue`, `green`, `gray`, `white`, `black`, …
- Hex: `#RRGGBB` en `#RGB`
- RGB-functie: `rgb(255, 128, 0)`

---

## 7. Volledig voorbeeld — inline HTML

```go
package main

import (
    "log"

    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/html"
)

func main() {
    doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
    doc.RegisterFont("regular", "/pad/naar/Roboto-Regular.ttf")
    doc.RegisterFont("bold",    "/pad/naar/Roboto-Bold.ttf")
    doc.RegisterFont("italic",  "/pad/naar/Roboto-Italic.ttf")
    doc.AddPage()

    fontFor := func(s html.Style) string {
        switch {
        case s.Bold:   return "bold"
        case s.Italic: return "italic"
        default:       return "regular"
        }
    }

    src := `Status: <b>Actief</b> — eigenaar: <i>Alice</i> — ref: <u>FAC-2024-0042</u>`
    spans, err := html.Parse(src, nil)
    if err != nil {
        log.Fatal(err)
    }

    doc.SetFont("regular", 11)
    if _, err := doc.WriteHTMLSpans(spans, fontFor, 11, 50, 80); err != nil {
        log.Fatal(err)
    }

    if err := doc.Save("inline.pdf"); err != nil {
        log.Fatal(err)
    }
}
```

---

## 8. Volledig voorbeeld — HTML-tabel

```go
package main

import (
    "log"

    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/html"
)

func main() {
    doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
    doc.RegisterFont("regular", "/pad/naar/Roboto-Regular.ttf")
    doc.RegisterFont("bold",    "/pad/naar/Roboto-Bold.ttf")
    doc.AddPage()

    src := `
    <table>
      <thead>
        <tr>
          <th>Artikel</th>
          <th>Aantal</th>
          <th align="right">Prijs</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>Widget <b>Pro</b></td>
          <td>3</td>
          <td align="right">€ 45,00</td>
        </tr>
        <tr bgcolor="#f5f5f5">
          <td>Gadget</td>
          <td>1</td>
          <td align="right">€ 12,50</td>
        </tr>
      </tbody>
    </table>`

    htmlTabel, err := html.ParseTable(src)
    if err != nil {
        log.Fatal(err)
    }

    cfg := pdf.TableConfig{
        ColWidths: []float64{250, 60, 100},
        DefaultCellStyle: pdf.CellStyle{
            FontName: "regular",
            FontSize: 10,
            Padding:  pdf.Padding{Top: 4, Right: 6, Bottom: 4, Left: 6},
        },
    }

    opts := pdf.HtmlTableOptions{
        SpanFontFor: func(s html.Style) string {
            if s.Bold {
                return "bold"
            }
            return "regular"
        },
        HeaderStyle: pdf.CellStyle{
            FontName:   "bold",
            Background: &pdf.Color{R: 0.15, G: 0.35, B: 0.65},
            TextColor:  &pdf.Color{R: 1, G: 1, B: 1},
        },
    }

    pdfTabel, err := doc.TableFromHTML(htmlTabel, cfg, opts)
    if err != nil {
        log.Fatal(err)
    }

    if _, err := pdfTabel.Draw(doc, 50, 80); err != nil {
        log.Fatal(err)
    }

    if err := doc.Save("tabel.pdf"); err != nil {
        log.Fatal(err)
    }
}
```

---

## 9. Bekende beperkingen

- **Blokelementen** — alleen `<table>` wordt herkend als blokelement. Divs,
  paragrafen, koppen en andere blok-tags binnen inline HTML worden genegeerd;
  hun tekstinhoud blijft wel beschikbaar.
- **Geneste tabellen** — geneste `<table>`-elementen binnen een cel worden
  overgeslagen.
- **CSS-stijlbladen** — slechts een kleine subset van inline `style="…"`-
  eigenschappen wordt verwerkt (zie de attributentabel in sectie 5). Externe
  of ingesloten CSS wordt niet ondersteund.
- **`<br>` in cellen** — `<br>`-tags worden omgezet naar `\n` in celinhoud.
- **`width`-hints** — het `width`-attribuut op `<td>` / `<th>` wordt opgeslagen
  in `HtmlCell.Width` als ruwe string maar wordt **niet** automatisch toegepast.
  Vertaal het zelf naar expliciete `ColWidths` in `TableConfig` indien nodig.
- **HTML-entiteiten** — alleen de meest voorkomende entiteiten (`&amp;`, `&lt;`,
  `&gt;`, `&nbsp;`, `&quot;`) worden gedecodeerd. Numerieke tekenverwijzingen
  (`&#160;`) worden niet ondersteund.
