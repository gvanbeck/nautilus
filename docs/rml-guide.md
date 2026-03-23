# Nautilus RML — Gebruikershandleiding

Het `pdf/rml`-pakket maakt het mogelijk om PDF-documenten te beschrijven via een
XML-dialect dat gebaseerd is op **RML (Report Markup Language)** van ReportLab.
In plaats van Go-code te schrijven, beschrijft u de volledige lay-out in een
XML-bestand dat door de bibliotheek wordt geparseerd en omgezet naar een PDF.

---

## Inhoudsopgave

1. [Snelstart](#1-snelstart)
2. [Documentstructuur](#2-documentstructuur)
3. [Docinit — lettertypes registreren](#3-docinit--lettertypes-registreren)
4. [Template — pagina-indeling](#4-template--pagina-indeling)
5. [PageGraphics — kop- en voetteksten](#5-pagegraphics--kop--en-voetteksten)
6. [Stylesheet](#6-stylesheet)
   - [paraStyle](#61-parastyle)
   - [blockTableStyle](#62-blocktablestyle)
7. [Story — inhoud](#7-story--inhoud)
   - [para](#71-para)
   - [spacer](#72-spacer)
   - [blockTable](#73-blocktable)
   - [image](#74-image)
   - [ul / ol](#75-ul--ol)
   - [indent](#76-indent)
   - [keepTogether](#77-keeptogether)
   - [condPageBreak](#78-condpagebreak)
   - [pageBreak / frameBreak](#79-pagebreak--framebreak)
   - [nextPageTemplate](#710-nextpagetemplate)
   - [hr / hRule](#711-hr--hrule)
8. [Meeteenheden en kleuren](#8-meeteenheden-en-kleuren)
9. [Coördinatenstelsel](#9-coördinatenstelsel)
10. [Go API](#10-go-api)
11. [Volledig voorbeeld](#11-volledig-voorbeeld)

---

## 1. Snelstart

```bash
go run ./examples/rml \
  -rml  examples/rml/invoice.rml \
  -fontdir /Library/Fonts \
  -out  invoice.pdf
```

Of vanuit Go-code:

```go
import "github.com/gvanbeck/nautilus/pdf/rml"

doc, err := rml.ParseFile("invoice.rml", rml.Options{
    FontDir: "/Library/Fonts",
})
if err != nil {
    log.Fatal(err)
}
if err := doc.Save("invoice.pdf"); err != nil {
    log.Fatal(err)
}
```

---

## 2. Documentstructuur

Een RML-document heeft altijd deze volgorde van vier secties:

```xml
<?xml version="1.0" encoding="utf-8"?>
<document>
  <docinit>   <!-- lettertypes -->  </docinit>
  <template>  <!-- pagina-indeling --> </template>
  <stylesheet><!-- stijlen -->       </stylesheet>
  <story>     <!-- inhoud -->        </story>
</document>
```

---

## 3. Docinit — lettertypes registreren

### `<registerTTFont>`

Registreert een TrueType-lettertype onder een interne naam.

```xml
<docinit>
  <registerTTFont fontName="regular"     fontFile="Lato-Regular.ttf"/>
  <registerTTFont fontName="regularBold" fontFile="Lato-Bold.ttf"/>
</docinit>
```

| Attribuut  | Beschrijving                                         |
|------------|------------------------------------------------------|
| `fontName` | Interne naam voor gebruik in stijlen                 |
| `fontFile` | Bestandsnaam (relatief t.o.v. `FontDir` of absoluut) |

### `<registerFontFamily>`

Groepeert vier varianten van een lettertype zodat inline-markup (`<b>`, `<i>`)
automatisch de juiste variant kiest.

```xml
<docinit>
  <registerTTFont fontName="sans"           fontFile="Roboto-Regular.ttf"/>
  <registerTTFont fontName="sans-bold"      fontFile="Roboto-Bold.ttf"/>
  <registerTTFont fontName="sans-italic"    fontFile="Roboto-Italic.ttf"/>
  <registerTTFont fontName="sans-boldital"  fontFile="Roboto-BoldItalic.ttf"/>

  <registerFontFamily name="sans"
                      fontName="sans"
                      bold="sans-bold"
                      italic="sans-italic"
                      boldItalic="sans-boldital"/>
</docinit>
```

| Attribuut    | Beschrijving                          |
|--------------|---------------------------------------|
| `name`       | Familienaam (gebruik in `fontName=`)  |
| `fontName`   | Reguliere variant                     |
| `bold`       | Vet variant                           |
| `italic`     | Cursief variant                       |
| `boldItalic` | Vet-cursief variant                   |

---

## 4. Template — pagina-indeling

```xml
<template pageSize="A4"
          leftMargin="55"  rightMargin="55"
          topMargin="65"   bottomMargin="62"
          title="Factuur 2026-0042"
          author="Nautilus Systems BV"
          subject="Factuur"
          creator="MijnApp 1.0">

  <pageTemplate id="main">
    <pageGraphics>…</pageGraphics>   <!-- optioneel -->
    <frame id="body" x1="55" y1="62" width="485" height="720"/>
  </pageTemplate>

</template>
```

### `<template>`-attributen

| Attribuut           | Standaard | Beschrijving                                              |
|---------------------|-----------|-----------------------------------------------------------|
| `pageSize`          | `A4`      | `A3`, `A4`, `A5`, `letter`, `legal`, of `(breedte,hoogte)` in pt |
| `leftMargin`        | `55`      | Linkermarge in pt                                         |
| `rightMargin`       | `55`      | Rechtermarge in pt                                        |
| `topMargin`         | `55`      | Bovenmarge in pt                                          |
| `bottomMargin`      | `55`      | Ondermarge in pt                                          |
| `title`             | —         | PDF-metadatatitel                                         |
| `author`            | —         | PDF-metadataauteur                                        |
| `subject`           | —         | PDF-metadataonderwerp                                     |
| `creator`           | —         | PDF-metadatacreator                                       |
| `firstPageTemplate` | eerste    | ID van het template te gebruiken voor pagina 1            |

### `<pageTemplate>`

| Attribuut | Beschrijving                      |
|-----------|-----------------------------------|
| `id`      | Unieke naam voor dit paginamodel  |

### `<frame>`

Definieert een tekstvlak waarin de story-inhoud stroomt. Coördinaten zijn
**PDF-coördinaten** (nulpunt linksonder).

| Attribuut | Beschrijving                            |
|-----------|-----------------------------------------|
| `id`      | Naam van het frame                      |
| `x1`      | Linkerrand (van linkerkant pagina)      |
| `y1`      | Onderrand (van onderkant pagina)        |
| `width`   | Breedte van het frame                   |
| `height`  | Hoogte van het frame                    |

---

## 5. PageGraphics — kop- en voetteksten

`<pageGraphics>` bevat tekenopdrachten die op **elke pagina** worden uitgevoerd
vóór de story-inhoud. Ideaal voor vaste kop- en voetteksten.

```xml
<pageGraphics>
  <saveState/>
  <setFont name="regular" size="8"/>
  <fill color="gray"/>

  <!-- Koptekstlijn -->
  <lines>55 790 540 790</lines>
  <drawString x="55" y="793">Bedrijfsnaam BV</drawString>
  <drawRightString x="540" y="793">Factuur #2026-0042</drawRightString>

  <!-- Voettekst met paginanummer -->
  <lines>55 52 540 52</lines>
  <drawCentredString x="297" y="41">Pagina %p</drawCentredString>

  <restoreState/>
</pageGraphics>
```

### Ondersteunde tekenopdrachten

| Element             | Attributen                            | Beschrijving                              |
|---------------------|---------------------------------------|-------------------------------------------|
| `<saveState/>`      | —                                     | Grafische toestand opslaan                |
| `<restoreState/>`   | —                                     | Grafische toestand herstellen             |
| `<setFont/>`        | `name`, `size`                        | Lettertype en -grootte instellen          |
| `<fill/>`           | `color`                               | Opvulkleur / tekstkleur instellen         |
| `<stroke/>`         | `color`, `width`                      | Lijnkleur instellen                       |
| `<drawString/>`     | `x`, `y`                              | Tekst links-uitgelijnd tekenen            |
| `<drawRightString/>`| `x`, `y`                              | Tekst rechts-uitgelijnd (x = rechterrand) |
| `<drawCentredString/>`| `x`, `y`                            | Tekst gecentreerd op x                    |
| `<lines/>`          | inhoud: `x1 y1 x2 y2 …`              | Eén of meer lijnstukken tekenen           |
| `<line/>`           | `x1`, `y1`, `x2`, `y2`, `width`, `color` | Enkel lijnstuk                        |
| `<rect/>`           | `x`, `y`, `width`, `height`, `fill`, `stroke`, `round` | Rechthoek          |
| `<circle/>`         | `x`, `y`, `radius`, `fill`, `stroke` | Cirkel                                    |

### Paginanummer-variabelen

Binnen tekst-elementen van `<pageGraphics>`:

| Variabele | Inhoud                    |
|-----------|---------------------------|
| `%p`      | Huidig paginanummer       |
| `%P`      | Totaal aantal pagina's *(nog niet ondersteund — zie §11)* |

---

## 6. Stylesheet

### 6.1 `<paraStyle>`

Definieert een herbruikbare alinea-stijl.

```xml
<stylesheet>
  <paraStyle name="titel"
             fontName="regularBold" fontSize="22"
             spaceAfter="6"/>

  <paraStyle name="brood"
             fontName="regular" fontSize="11"
             leading="16" spaceAfter="5"/>

  <paraStyle name="klein"
             fontName="regular" fontSize="9"
             textColor="gray" spaceAfter="3"/>
</stylesheet>
```

| Attribuut        | Beschrijving                                          |
|------------------|-------------------------------------------------------|
| `name`           | Unieke stijlnaam                                      |
| `parent`         | Overerft instellingen van een andere stijl            |
| `fontName`       | Lettertype                                            |
| `fontSize`       | Lettergrootte in pt                                   |
| `leading`        | Regelafstand in pt                                    |
| `alignment`      | `left`, `center`, `right`                             |
| `spaceBefore`    | Ruimte vóór de alinea in pt                           |
| `spaceAfter`     | Ruimte na de alinea in pt                             |
| `textColor`      | Tekstkleur (naam of `#rrggbb`)                        |
| `backColor`      | Achtergrondkleur van de alinea                        |
| `leftIndent`     | Inspringing links in pt                               |
| `rightIndent`    | Inspringing rechts in pt                              |
| `firstLineIndent`| Extra inspringing van de eerste regel in pt           |
| `underline`      | `1` of `true` voor onderstreping                      |
| `strike`         | `1` of `true` voor doorstreping                       |
| `keepWithNext`   | `1` of `true` om samen met volgende alinea te houden  |

### 6.2 `<blockTableStyle>`

Definieert de opmaak van een tabel.

```xml
<blockTableStyle id="factuur">
  <!-- Volledige grid in lichtgrijs, dikke buitenrand in donkerblauw -->
  <lineStyle kind="GRID"    colorName="lightgray" thickness="0.4"
             start="0,0" stop="-1,-1"/>
  <lineStyle kind="OUTLINE" colorName="navy"      thickness="1.2"
             start="0,0" stop="-1,-1"/>

  <!-- Koprijachtergrond: donkerblauw, witte vette tekst -->
  <blockBackground colorName="navy"  start="0,0" stop="-1,0"/>
  <blockTextColor  colorName="white" start="0,0" stop="-1,0"/>
  <blockFont       name="regularBold" start="0,0" stop="-1,0"/>

  <!-- Laatste rij: lichtblauwe achtergrond, vet -->
  <blockBackground colorName="#e8f0fe" start="0,-1" stop="-1,-1"/>
  <blockFont       name="regularBold"  start="0,-1" stop="-1,-1"/>

  <!-- Numerieke kolommen rechts uitlijnen (kol 3 t/m 4) -->
  <blockAlignment  value="right" start="3,0" stop="4,-1"/>

  <!-- Verticale centrering in alle cellen -->
  <blockValign     value="middle" start="0,0" stop="-1,-1"/>

  <!-- Padding -->
  <blockPadding      value="4"   start="0,0" stop="-1,-1"/>
  <blockLeftPadding  value="6"   start="0,0" stop="-1,-1"/>
</blockTableStyle>
```

#### `<lineStyle>`

| Attribuut   | Waarden                          | Beschrijving                          |
|-------------|----------------------------------|---------------------------------------|
| `kind`      | `GRID`, `OUTLINE`, `INNERGRID`, `BOX` | Type lijn                        |
| `colorName` | kleurwaarde                      | Lijnkleur                             |
| `thickness` | getal in pt                      | Lijndikte (standaard `0.5`)           |
| `start`     | `col,rij`                        | Begincel (inclusief)                  |
| `stop`      | `col,rij`                        | Eindcel (inclusief, negatief = vanuit einde) |

#### Cel-stijlopdrachten

| Element             | Relevante attributen         | Beschrijving                   |
|---------------------|------------------------------|--------------------------------|
| `<blockBackground>` | `colorName`, `start`, `stop` | Achtergrondkleur van cellen    |
| `<blockTextColor>`  | `colorName`, `start`, `stop` | Tekstkleur van cellen          |
| `<blockFont>`       | `name`, `size`, `start`, `stop` | Lettertype van cellen       |
| `<blockAlignment>`  | `value`, `start`, `stop`     | Horizontale uitlijning         |
| `<blockValign>`     | `value`, `start`, `stop`     | Verticale uitlijning           |
| `<blockPadding>`    | `value`, `start`, `stop`     | Gelijkmatige opvulling         |
| `<blockTopPadding>` | `value`, `start`, `stop`     | Opvulling boven                |
| `<blockRightPadding>`| `value`, `start`, `stop`    | Opvulling rechts               |
| `<blockBottomPadding>`| `value`, `start`, `stop`   | Opvulling onder                |
| `<blockLeftPadding>`| `value`, `start`, `stop`     | Opvulling links                |

#### Celbereik met negatieve indexering

`start="0,0" stop="-1,-1"` geldt voor **alle** cellen.
`start="0,0" stop="-1,0"` geldt voor **de koprijbalk** (rij 0).
`start="0,-1" stop="-1,-1"` geldt voor **de laatste rij**.

---

## 7. Story — inhoud

### 7.1 `<para>`

Een alinea met optionele inline-opmaak.

```xml
<para style="brood">Gewone tekst zonder opmaak.</para>
<para style="brood">Tekst met <b>vetgedrukt</b> en <i>cursief</i>.</para>
<para style="brood">Tekst met <u>onderstreping</u>.</para>
```

Ondersteunde inline-tags: `<b>`, `<i>`, `<u>`.
Wanneer een alinea inline-markup bevat én een `<registerFontFamily>` is gedeclareerd,
worden de vet/cursief-varianten automatisch gekozen.

Stijlverkorting `h1`–`h6` is beschikbaar als alternatief voor `<para style="…">`:

```xml
<h1>Hoofdstuk 1</h1>
<h2>Paragraaf</h2>
```

### 7.2 `<spacer>`

Voegt verticale (of horizontale) witruimte toe.

```xml
<spacer length="18"/>        <!-- 18 pt verticaal -->
<spacer length="10" width="0"/>
```

### 7.3 `<blockTable>`

Een tabel die automatisch over pagina's kan doorlopen.

```xml
<blockTable colWidths="55,220,40,80,90"
            rowHeights="24,22,22,22"
            style="factuur"
            repeatRows="1"
            align="left"
            spaceAfter="10">

  <!-- Koprij -->
  <tr height="24">
    <td>Ref.</td>
    <td>Omschrijving</td>
    <td>Aantal</td>
    <td>Prijs</td>
    <td>Subtotaal</td>
  </tr>

  <!-- Datarijen met rijachtergrond -->
  <tr>
    <td>A-001</td>
    <td>Mechanisch toetsenbord</td>
    <td>2</td>
    <td>€ 89,95</td>
    <td>€ 179,90</td>
  </tr>
  <tr bg="#eef4ff">
    <td>A-002</td>
    <td>27" 4K IPS-monitor</td>
    <td>1</td>
    <td>€ 549,00</td>
    <td>€ 549,00</td>
  </tr>

</blockTable>
```

#### `<blockTable>`-attributen

| Attribuut    | Beschrijving                                                  |
|--------------|---------------------------------------------------------------|
| `colWidths`  | Kommagescheiden kolombreedten in pt (verplicht)               |
| `rowHeights` | Kommagescheiden rijhoogten in pt (optioneel)                  |
| `style`      | Verwijzing naar een `blockTableStyle`-id                      |
| `repeatRows` | Aantal koprijbaren dat bij paginaoverloop wordt herhaald      |
| `align`      | Horizontale uitlijning van de tabel: `left`, `center`, `right`|
| `spaceBefore`| Ruimte vóór de tabel in pt                                    |
| `spaceAfter` | Ruimte na de tabel in pt                                      |

#### `<tr>`-attributen

| Attribuut | Beschrijving                                   |
|-----------|------------------------------------------------|
| `height`  | Vaste rijhoogte (overschrijft `rowHeights`)    |
| `bg`      | Achtergrondkleur van de hele rij               |

#### `<td>`-attributen

| Attribuut      | Beschrijving                                        |
|----------------|-----------------------------------------------------|
| `colspan`      | Aantal kolommen samenvoegen (standaard 1)           |
| `rowspan`      | Aantal rijen samenvoegen (standaard 1)              |
| `style`        | Paragraafstijl voor de celinhoud                    |
| `fontName`     | Lettertype overschrijven                            |
| `fontSize`     | Lettergrootte overschrijven                         |
| `bold`         | `1` of `true` voor vetgedrukt                       |
| `bg`           | Achtergrondkleur van de cel                         |
| `textColor`    | Tekstkleur van de cel                               |
| `halign`       | `left`, `center`, `right`                           |
| `valign`       | `top`, `middle`, `bottom`                           |
| `topPadding`   | Opvulling boven in pt                               |
| `rightPadding` | Opvulling rechts in pt                              |
| `bottomPadding`| Opvulling onder in pt                               |
| `leftPadding`  | Opvulling links in pt                               |

### 7.4 `<image>`

Voegt een afbeelding in.

```xml
<image file="logo.png" width="120" height="60"
       align="right" spaceAfter="10"/>
```

| Attribuut    | Beschrijving                                      |
|--------------|---------------------------------------------------|
| `file`       | Pad naar de afbeelding (JPEG of PNG)              |
| `width`      | Breedte in pt (optioneel, schaalt proportioneel)  |
| `height`     | Hoogte in pt (optioneel)                          |
| `align`      | `left`, `center`, `right`                         |
| `spaceBefore`| Ruimte vóór in pt                                 |
| `spaceAfter` | Ruimte na in pt                                   |

### 7.5 `<ul>` / `<ol>`

Ongeordende of geordende lijst.

```xml
<ul bulletIndent="12">
  <li>Eerste punt</li>
  <li style="brood">Tweede punt</li>
</ul>

<ol start="3">
  <li>Derde item</li>
  <li>Vierde item</li>
</ol>
```

#### `<ul>` / `<ol>`-attributen

| Attribuut      | Beschrijving                                           |
|----------------|--------------------------------------------------------|
| `bulletIndent` | Inspringing van het opsommingsteken in pt              |
| `style`        | Standaard paragraafstijl voor alle lijstitems          |
| `start`        | Beginwaarde voor genummerde lijst (standaard `1`)      |

#### `<li>`-attributen

| Attribuut | Beschrijving                             |
|-----------|------------------------------------------|
| `style`   | Paragraafstijl (overschrijft lijststijl) |

### 7.6 `<indent>`

Wikkelt onderliggende inhoud in een inspringing.

```xml
<indent left="30" right="15">
  <para style="brood">Ingesprongen tekst.</para>
  <blockTable …>…</blockTable>
</indent>
```

| Attribuut | Beschrijving                  |
|-----------|-------------------------------|
| `left`    | Inspringing links in pt       |
| `right`   | Inspringing rechts in pt      |

### 7.7 `<keepTogether>`

Houdt alle onderliggende flowables op dezelfde pagina.

```xml
<keepTogether maxHeight="200">
  <para style="h2">Sectietitel</para>
  <para style="brood">Eerste alinea van de sectie.</para>
</keepTogether>
```

| Attribuut   | Beschrijving                                                      |
|-------------|-------------------------------------------------------------------|
| `maxHeight` | Maximale hoogte in pt (standaard onbeperkt). Indien de groep groter is, wordt ze toch gesplitst. |

### 7.8 `<condPageBreak>`

Voegt een pagina-einde in **als** de resterende ruimte kleiner is dan de opgegeven hoogte.

```xml
<condPageBreak height="150"/>
```

| Attribuut | Beschrijving                          |
|-----------|---------------------------------------|
| `height`  | Minimale vereiste hoogte in pt        |

### 7.9 `<pageBreak>` / `<frameBreak>`

```xml
<pageBreak/>    <!-- Harde pagina-overgang -->
<frameBreak/>   <!-- Overgang naar het volgende frame -->
```

### 7.10 `<nextPageTemplate>`

Schakelt over naar een ander paginasjabloon vanaf de volgende pagina.

```xml
<nextPageTemplate id="tweekolommen"/>
```

| Attribuut | Beschrijving                    |
|-----------|---------------------------------|
| `id`      | ID van het doelpaginasjabloon   |

### 7.11 `<hr>` / `<hRule>`

Horizontale scheidingslijn.

```xml
<hr width="100%" thickness="0.5" colorName="lightgray"/>
```

| Attribuut   | Beschrijving                              |
|-------------|-------------------------------------------|
| `width`     | Breedte in pt of `%` van beschikbare breedte |
| `thickness` | Lijndikte in pt (standaard `0.5`)         |
| `colorName` | Kleur (standaard zwart)                   |

---

## 8. Meeteenheden en kleuren

### Meeteenheden

Alle numerieke waarden worden standaard geïnterpreteerd als **punten (pt)**.
De volgende eenheden worden ook herkend:

| Invoer     | Voorbeeld  | Omrekening        |
|------------|------------|-------------------|
| `pt`       | `12pt`     | 1 pt = 1/72 inch  |
| `cm`       | `2.1cm`    | 1 cm = 28.35 pt   |
| `mm`       | `21mm`     | 1 mm = 2.835 pt   |
| `in`       | `0.83in`   | 1 in = 72 pt      |
| *(getal)*  | `595`      | Geïnterpreteerd als pt |

### Kleuren

Kleuren kunnen worden opgegeven als:

| Formaat       | Voorbeeld           | Beschrijving                              |
|---------------|---------------------|-------------------------------------------|
| Naam          | `navy`, `lightgray` | Bekende CSS-kleurnamen                    |
| Hex           | `#e8f0fe`           | RGB hexadecimaal                          |
| Komma         | `220,220,220`       | Drie decimale waarden (0–255)             |

Ondersteunde kleurnamen (selectie): `black`, `white`, `red`, `green`, `blue`,
`navy`, `gray` / `grey`, `lightgray`, `darkgray`, `yellow`, `orange`, `purple`,
`cyan`, `magenta`, `pink`, `brown`, `lime`, `teal`, `silver`, `gold`,
`transparent` / `none`.

---

## 9. Coördinatenstelsel

RML gebruikt **PDF-coördinaten**: het nulpunt ligt **linksonder** op de pagina.
De Y-as loopt omhoog.

```
(0, 842)  ──────────────── (595, 842)   ← bovenkant A4
          │                │
          │   pagina-inhoud│
          │                │
(0,  0)   ──────────────── (595,   0)   ← onderkant A4
```

Dit geldt voor `<frame>`, `<drawString>`, `<lines>`, `<rect>` enz.

De bibliotheek converteert intern naar het Nautilus-coördinatenstelsel
(nulpunt linksboven).

**Richtlijn voor A4 (595 × 842 pt):**

- Koptekst op y ≈ 790–800 (dicht bij de bovenkant)
- Voettekst op y ≈ 30–55 (dicht bij de onderkant)
- Tekstframe: `y1=62, height=720` laat ruimte voor kop en voet

---

## 10. Go API

```go
package rml

// Options bevat configuratieparameters voor het parseren.
type Options struct {
    FontDir string // map met lettertypebestanden
}

// ParseFile leest een RML-bestand en geeft een pdf.Document terug.
func ParseFile(path string, opts Options) (*pdf.Document, error)

// Parse leest een RML-stream en geeft een pdf.Document terug.
func Parse(r io.Reader, opts Options) (*pdf.Document, error)
```

### Voorbeeld CLI-tool

```go
package main

import (
    "flag"
    "log"

    "github.com/gvanbeck/nautilus/pdf/rml"
)

func main() {
    rmlFile := flag.String("rml",     "document.rml", "RML-invoerbestand")
    fontDir := flag.String("fontdir", ".",             "Map met lettertypes")
    outFile := flag.String("out",     "document.pdf",  "PDF-uitvoerbestand")
    flag.Parse()

    doc, err := rml.ParseFile(*rmlFile, rml.Options{FontDir: *fontDir})
    if err != nil {
        log.Fatalf("parse: %v", err)
    }
    if err := doc.Save(*outFile); err != nil {
        log.Fatalf("save: %v", err)
    }
}
```

---

## 11. Volledig voorbeeld

Zie `examples/rml/invoice.rml` voor een volledig factuurvoorbeeld met:

- `<docinit>` met twee lettertyperegistraties
- `<template>` met metadata, marges en een frameomschrijving
- `<pageGraphics>` met koptekst (bedrijfsnaam + factuurnummer), scheidingslijnen
  en voettekst met paginanummer (`%p`)
- `<stylesheet>` met meerdere paragraafstijlen en twee tabelstijlen
  (`invoice` met navy kopbalk, en `meta` voor metagegevenstabel)
- `<story>` met briefhoofd, metagegevenstabel, factuurregeltabel
  (met rijachtergrondkleuren) en een samenvattingstabel

```bash
go run ./examples/rml \
  -rml     examples/rml/invoice.rml \
  -fontdir /Library/Fonts \
  -out     invoice.pdf
```

---

## Bekende beperkingen

De volgende RML-functies worden momenteel **niet** ondersteund:

| Functie                | Beschrijving                                          |
|------------------------|-------------------------------------------------------|
| `%P` in pageGraphics   | Totaal aantal pagina's (vereist twee-pass rendering)  |
| `<storyPlace>`         | Inhoud op vaste coördinaten plaatsen                  |
| `<balancedColumns>`    | Tekst automatisch over meerdere kolommen verdelen     |
| Barcodes               | `<barCode>`, QR-codes enz.                            |
| Formuliervelden        | Interactieve PDF-formulieren                          |
| Kruisverwijzingen      | `<seq>`, `<bookmark>`, `<getName>` enz.               |
| `<illustration>`       | Inline SVG/vector-tekeningen                          |
| `<plugInFlowable>`     | Externe flowable-plug-ins                             |
