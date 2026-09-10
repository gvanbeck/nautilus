# Gevendorde gopdf

Kopie van [signintech/gopdf](https://github.com/signintech/gopdf), gevendord
zodat de tekstbreedte-berekening bit-exact gelijk gemaakt kan worden aan
ReportLab's `pdfmetrics.stringWidth`. Nautilus is bedoeld om documenten te
kunnen produceren die tot op de punt overeenkomen met een ReportLab-renderer,
en dat lukt niet zolang de breedtes afwijken.

| | |
|---|---|
| Upstream | https://github.com/signintech/gopdf |
| Versie | `v0.36.0` |
| Commit | `2dcf2ba99e1fe3be1d480eea69a45f34484a92b0` |
| Datum | 2026-02-08 |
| Licentie | MIT, zie `LICENSE` |

## Wat er is weggelaten

- `.github/` — CI van upstream, niet van toepassing.
- `go.mod`/`go.sum` — deze code hoort nu bij de nautilus-module.
- `examples/`, **op twee fixtures na**: `examples/outline_example/outline_demo.pdf`
  en `.../Ubuntu-L.ttf` worden door `TestImportPagesFromFile` gelezen. De map
  `examples/table/` bestaat leeg (met `.gitkeep`) omdat drie tabeltests hun
  output daarin schrijven — upstream schrijft testoutput in de boom; die
  artefacten staan in `.gitignore`.

Volledige testsuite groen na het vendoren: 28 packages, 0 failures.

## Wat er is gewijzigd t.o.v. upstream

Importpaden herschreven van `github.com/signintech/gopdf` naar
`github.com/gvanbeck/nautilus/internal/gopdf`. Functionele wijzigingen worden
hieronder bijgehouden, met de reden erbij — anders is een latere
upstream-vergelijking niet te doen.

Alle wijzigingen staan ook in de code zelf, gemarkeerd met
`AFWIJKING T.O.V. UPSTREAM` of `TOEGEVOEGD T.O.V. UPSTREAM`.

| Plek | Wijziging | Reden |
|---|---|---|
| `subset_font_obj.go` `GlyphIndexToPdfWidth` | `uint` → `float64`, factor `1000/upem` als float | upstream kapte af in uint: mediaan 0,49 fonteenheden per teken, 2,39pt op 200 tekens @12pt. ReportLab schaalt in float (`pdfbase/ttfonts.py:572-576`) |
| `subset_font_obj.go` `CharWidth` | `uint` → `float64` | volgt uit bovenstaande |
| `subset_font_obj.go` `DefaultWidth` | nieuw | breedte van glyph 0; ReportLab rekent die voor tekens buiten de cmap |
| `subset_font_obj.go` `AddChars` | geen rune-substitutie meer; onbekend teken → glyph 0 | ReportLab tekent `.notdef` en rekent zijn breedte. Empirisch: `splitString("A中B")` → codes `[65, 0, 66]` |
| `subset_font_obj.go` `CharCodeToGlyphIndex` | NBSP/spatie-glyphaliasing | `pdfbase/ttfonts.py:880-885`. Zonder alias liep een string van drie NBSP's in Lucida Sans 11,86pt uit de pas |
| `cache_content_text.go` `createContent` | int- → float-accumulatie; `defaultWidth` bij onbekend teken; `0.001*size*Σ` i.p.v. `Σ*(size/1000)` | `lib/rl_accel.py:106` |
| `cid_font_obj.go` `/W`-array | `%d` op afgekapte uint → reëel getal | anders accumuleert een viewer de afkapping binnen één `Tj` |
| `reportlab_real.go` | nieuw: `pdfReal` | `/W` formatteren volgens ReportLab's `fp_str`-regel (`lib/rl_accel.py:41-60`) |

`replaceGlyphThatNotFound` en de optie `OnGlyphNotFoundSubstitute` zijn hierdoor
dode code geworden. Ze blijven staan om de diff met upstream klein te houden.

## Verificatie

Een fixture van ReportLab-breedtes over 46 fonts × 63 strings × 4 groottes
(11.592 metingen) valt exact samen met `Document.MeasureText`: `Δ == 0.0`, geen
tolerantie. De fixture en de test horen bij de consument van deze bibliotheek,
niet bij nautilus zelf — de TTF's zijn niet vrij te verspreiden.
