# Vendored gopdf

Copy of [signintech/gopdf](https://github.com/signintech/gopdf), vendored so
that the text-width calculation can be made bit-exact with ReportLab's
`pdfmetrics.stringWidth`. Nautilus is meant to produce documents that match a
ReportLab renderer down to the point, and that is impossible as long as the
widths differ.

| | |
|---|---|
| Upstream | https://github.com/signintech/gopdf |
| Version | `v0.36.0` |
| Commit | `2dcf2ba99e1fe3be1d480eea69a45f34484a92b0` |
| Date | 2026-02-08 |
| License | MIT, see `LICENSE` |

## What was left out

- `.github/` — upstream CI, not applicable.
- `go.mod`/`go.sum` — this code is part of the nautilus module now.
- `examples/`, **except for two fixtures**: `examples/outline_example/outline_demo.pdf`
  and `.../Ubuntu-L.ttf` are read by `TestImportPagesFromFile`. The directory
  `examples/table/` exists but is empty (with a `.gitkeep`) because three table
  tests write their output there — upstream writes test output into the tree;
  those artifacts are listed in `.gitignore`.

Full test suite green after vendoring: 28 packages, 0 failures.

## What was changed relative to upstream

Import paths rewritten from `github.com/signintech/gopdf` to
`github.com/gvanbeck/nautilus/internal/gopdf`. Functional changes are tracked
below, together with the reason — without that, comparing against upstream
later is not feasible.

Every change is marked in the code itself as well, with
`DEVIATION FROM UPSTREAM` or `ADDED RELATIVE TO UPSTREAM`.

| Place | Change | Reason |
|---|---|---|
| `subset_font_obj.go` `GlyphIndexToPdfWidth` | `uint` → `float64`, factor `1000/upem` as a float | upstream truncated in uint: median 0.49 font units per character, 2.39pt over 200 characters @12pt. ReportLab scales in float (`pdfbase/ttfonts.py:572-576`) |
| `subset_font_obj.go` `CharWidth` | `uint` → `float64` | follows from the above |
| `subset_font_obj.go` `DefaultWidth` | new | width of glyph 0; ReportLab uses it for characters outside the cmap |
| `subset_font_obj.go` `AddChars` | no more rune substitution; unknown character → glyph 0 | ReportLab draws `.notdef` and counts its width. Empirically: `splitString("A中B")` → codes `[65, 0, 66]` |
| `subset_font_obj.go` `CharCodeToGlyphIndex` | NBSP/space glyph aliasing | `pdfbase/ttfonts.py:880-885`. Without the alias, a string of three NBSPs in Lucida Sans drifted out of step at 11.86pt |
| `cache_content_text.go` `createContent` | int → float accumulation; `defaultWidth` for an unknown character; `0.001*size*Σ` instead of `Σ*(size/1000)` | `lib/rl_accel.py:106` |
| `cid_font_obj.go` `/W` array | `%d` on a truncated uint → real number | otherwise a viewer accumulates the truncation within a single `Tj` |
| `reportlab_real.go` | new: `pdfReal` | format `/W` following ReportLab's `fp_str` rule (`lib/rl_accel.py:41-60`) |

`replaceGlyphThatNotFound` and the `OnGlyphNotFoundSubstitute` option have
become dead code as a result. They are kept to keep the diff with upstream
small.

## Verification

A fixture of ReportLab widths across 46 fonts × 63 strings × 4 sizes (11,592
measurements) matches `Document.MeasureText` exactly: `Δ == 0.0`, no tolerance.
The fixture and the test belong to the consumer of this library, not to
nautilus itself — the TTFs are not freely redistributable.
