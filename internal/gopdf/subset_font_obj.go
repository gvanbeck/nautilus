package gopdf

import (
	"errors"
	"fmt"
	"io"

	"github.com/gvanbeck/nautilus/internal/gopdf/fontmaker/core"
)

// ErrCharNotFound char not found
var ErrCharNotFound = errors.New("char not found")

// ErrGlyphNotFound font file not contain glyph
var ErrGlyphNotFound = errors.New("glyph not found")

// SubsetFontObj pdf subsetFont object
type SubsetFontObj struct {
	ttfp                  core.TTFParser
	Family                string
	CharacterToGlyphIndex *MapOfCharacterToGlyphIndex
	CountOfFont           int
	indexObjCIDFont       int
	indexObjUnicodeMap    int
	ttfFontOption         TtfOption
	funcKernOverride      FuncKernOverride
	funcGetRoot           func() *GoPdf
	addCharsBuff          []rune
}

func (s *SubsetFontObj) init(funcGetRoot func() *GoPdf) {
	s.CharacterToGlyphIndex = NewMapOfCharacterToGlyphIndex() //make(map[rune]uint)
	s.funcKernOverride = nil
	s.funcGetRoot = funcGetRoot

}

func (s *SubsetFontObj) write(w io.Writer, objID int) error {
	//me.AddChars("จ")
	io.WriteString(w, "<<\n")
	fmt.Fprintf(w, "/BaseFont /%s\n", CreateEmbeddedFontSubsetName(s.Family))
	fmt.Fprintf(w, "/DescendantFonts [%d 0 R]\n", s.indexObjCIDFont+1)
	io.WriteString(w, "/Encoding /Identity-H\n")
	io.WriteString(w, "/Subtype /Type0\n")
	fmt.Fprintf(w, "/ToUnicode %d 0 R\n", s.indexObjUnicodeMap+1)
	io.WriteString(w, "/Type /Font\n")
	io.WriteString(w, ">>\n")
	return nil
}

// SetIndexObjCIDFont set IndexObjCIDFont
func (s *SubsetFontObj) SetIndexObjCIDFont(index int) {
	s.indexObjCIDFont = index
}

// SetIndexObjUnicodeMap set IndexObjUnicodeMap
func (s *SubsetFontObj) SetIndexObjUnicodeMap(index int) {
	s.indexObjUnicodeMap = index
}

// SetFamily set font family name
func (s *SubsetFontObj) SetFamily(familyname string) {
	s.Family = familyname
}

// GetFamily get font family name
func (s *SubsetFontObj) GetFamily() string {
	return s.Family
}

// SetTtfFontOption set TtfOption must set before SetTTFByPath
func (s *SubsetFontObj) SetTtfFontOption(option TtfOption) {
	if option.OnGlyphNotFoundSubstitute == nil {
		option.OnGlyphNotFoundSubstitute = DefaultOnGlyphNotFoundSubstitute
	}
	s.ttfFontOption = option
}

// GetTtfFontOption get TtfOption must set before SetTTFByPath
func (s *SubsetFontObj) GetTtfFontOption() TtfOption {
	return s.ttfFontOption
}

// KernValueByLeft find kern value from kern table by left
func (s *SubsetFontObj) KernValueByLeft(left uint) (bool, *core.KernValue) {

	if !s.ttfFontOption.UseKerning {
		return false, nil
	}

	k := s.ttfp.Kern()
	if k == nil {
		return false, nil
	}

	if kval, ok := k.Kerning[left]; ok {
		return true, &kval
	}

	return false, nil
}

// SetTTFByPath set ttf
func (s *SubsetFontObj) SetTTFByPath(ttfpath string) error {
	useKerning := s.ttfFontOption.UseKerning
	s.ttfp.SetUseKerning(useKerning)
	err := s.ttfp.Parse(ttfpath)
	if err != nil {
		return err
	}
	return nil
}

// SetTTFByReader set ttf
func (s *SubsetFontObj) SetTTFByReader(rd io.Reader) error {
	useKerning := s.ttfFontOption.UseKerning
	s.ttfp.SetUseKerning(useKerning)
	err := s.ttfp.ParseByReader(rd)
	if err != nil {
		return err
	}
	return nil
}

// SetTTFData set ttf
func (s *SubsetFontObj) SetTTFData(data []byte) error {
	useKerning := s.ttfFontOption.UseKerning
	s.ttfp.SetUseKerning(useKerning)
	err := s.ttfp.ParseFontData(data)
	if err != nil {
		return err
	}
	return nil
}

// AddChars add char to map CharacterToGlyphIndex
func (s *SubsetFontObj) AddChars(txt string) (string, error) {
	s.addCharsBuff = s.addCharsBuff[:0]
	for _, runeValue := range txt {
		if s.CharacterToGlyphIndex.KeyExists(runeValue) {
			s.addCharsBuff = append(s.addCharsBuff, runeValue)
			continue
		}
		glyphIndex, err := s.CharCodeToGlyphIndex(runeValue)
		if err == ErrGlyphNotFound {
			//never return error on this, just call function OnGlyphNotFound
			if s.ttfFontOption.OnGlyphNotFound != nil {
				s.ttfFontOption.OnGlyphNotFound(runeValue)
			}
			// AFWIJKING T.O.V. UPSTREAM: geen rune-substitutie. Upstream
			// verving het teken via OnGlyphNotFoundSubstitute (default een
			// spatie), waardoor zowel de getekende glyph als de breedte
			// afweken. Reportlab mapt een teken buiten de cmap op glyph 0
			// (.notdef) en rekent glyph 0's breedte. Empirisch bevestigd:
			// splitString("A中B") -> codes [65, 0, 66], en
			// stringWidth("A中B") - stringWidth("AB") == defaultWidth*0.001*size.
			// Zie VENDOR.md.
			s.CharacterToGlyphIndex.Set(runeValue, 0)
			s.addCharsBuff = append(s.addCharsBuff, runeValue)
			continue
		} else if err != nil {
			return "", err
		}
		s.CharacterToGlyphIndex.Set(runeValue, glyphIndex) // [runeValue] = glyphIndex
		s.addCharsBuff = append(s.addCharsBuff, runeValue)
	}
	return string(s.addCharsBuff), nil
}

/*
//AddChars add char to map CharacterToGlyphIndex
func (s *SubsetFontObj) AddChars(txt string) error {

	for _, runeValue := range txt {
		if s.CharacterToGlyphIndex.KeyExists(runeValue) {
			continue
		}
		glyphIndex, err := s.CharCodeToGlyphIndex(runeValue)
		if err == ErrGlyphNotFound {
			//never return error on this, just call function OnGlyphNotFound
			if s.ttfFontOption.OnGlyphNotFound != nil {
				s.ttfFontOption.OnGlyphNotFound(runeValue)
			}
			//start: try to find rune for replace
			runeValueReplace, glyphIndexReplace, ok := s.replaceGlyphThatNotFound(runeValue)
			if ok {
				s.CharacterToGlyphIndex.Set(runeValueReplace, glyphIndexReplace) // [runeValue] = glyphIndex
			}
			//end: try to find rune for replace
			continue
		} else if err != nil {
			return err
		}
		s.CharacterToGlyphIndex.Set(runeValue, glyphIndex) // [runeValue] = glyphIndex
	}
	return nil
}
*/

// replaceGlyphThatNotFound find glyph to replaced
// it returns
// - true if rune already add to CharacterToGlyphIndex
// - rune for replace
// - rune for replace is found or not
// - glyph index for replace
func (s *SubsetFontObj) replaceGlyphThatNotFound(runeNotFound rune) (bool, rune, uint) {
	if s.ttfFontOption.OnGlyphNotFoundSubstitute != nil {
		runeForReplace := s.ttfFontOption.OnGlyphNotFoundSubstitute(runeNotFound)
		if s.CharacterToGlyphIndex.KeyExists(runeForReplace) {
			return true, runeForReplace, 0
		}
		glyphIndexForReplace, err := s.CharCodeToGlyphIndex(runeForReplace)
		if err != nil {
			return false, runeForReplace, 0
		}
		return false, runeForReplace, glyphIndexForReplace
	}
	return false, runeNotFound, 0
}

// CharIndex index of char in glyph table
func (s *SubsetFontObj) CharIndex(r rune) (uint, error) {
	glyIndex, ok := s.CharacterToGlyphIndex.Val(r)
	if ok {
		return glyIndex, nil
	}
	return 0, ErrCharNotFound
}

// CharWidth with of char, in 1/1000 em.
//
// AFWIJKING T.O.V. UPSTREAM: retourtype uint -> float64, zie
// GlyphIndexToPdfWidth.
func (s *SubsetFontObj) CharWidth(r rune) (float64, error) {
	glyIndex, ok := s.CharacterToGlyphIndex.Val(r)
	if ok {
		return s.GlyphIndexToPdfWidth(glyIndex), nil
	}
	return 0, ErrCharNotFound
}

func (s *SubsetFontObj) getType() string {
	return "SubsetFont"
}

func (s *SubsetFontObj) charCodeToGlyphIndexFormat12(r rune) (uint, error) {

	value := uint(r)
	gTbs := s.ttfp.GroupingTables()
	for _, gTb := range gTbs {
		if value >= gTb.StartCharCode && value <= gTb.EndCharCode {
			gIndex := (value - gTb.StartCharCode) + gTb.GlyphID
			return gIndex, nil
		}
	}

	return uint(0), ErrGlyphNotFound
}

func (s *SubsetFontObj) charCodeToGlyphIndexFormat4(r rune) (uint, error) {
	value := uint(r)
	seg := uint(0)
	segCount := s.ttfp.SegCount
	for seg < segCount {
		if value <= s.ttfp.EndCount[seg] {
			break
		}
		seg++
	}
	//fmt.Printf("\ncccc--->%#v\n", me.ttfp.Chars())
	if value < s.ttfp.StartCount[seg] {
		return 0, ErrGlyphNotFound
	}

	if s.ttfp.IdRangeOffset[seg] == 0 {

		return (value + s.ttfp.IdDelta[seg]) & 0xFFFF, nil
	}
	//fmt.Printf("IdRangeOffset=%d\n", me.ttfp.IdRangeOffset[seg])
	idx := s.ttfp.IdRangeOffset[seg]/2 + (value - s.ttfp.StartCount[seg]) - (segCount - seg)

	if s.ttfp.GlyphIdArray[int(idx)] == uint(0) {
		return 0, nil
	}

	return (s.ttfp.GlyphIdArray[int(idx)] + s.ttfp.IdDelta[seg]) & 0xFFFF, nil
}

// CharCodeToGlyphIndex gets glyph index from char code.
//
// AFWIJKING T.O.V. UPSTREAM: NBSP/spatie-aliasing zoals reportlab die in
// TTFontFile.extractInfo doet (pdfbase/ttfonts.py:880-885). Heeft het font een
// spatie, dan krijgt U+00A0 diezelfde glyph — óók als het font een eigen
// NBSP-glyph heeft; ontbreekt de spatie, dan geldt het omgekeerde. Zonder deze
// alias liep een string van drie NBSP's in Lucida Sans tot 11,86pt uit de pas.
// Empirisch bevestigd: splitString("\u00a0") -> code 32. Zie VENDOR.md.
func (s *SubsetFontObj) CharCodeToGlyphIndex(r rune) (uint, error) {
	switch r {
	case '\u00a0':
		if g, err := s.charCodeToGlyphIndexRaw('\u0020'); err == nil {
			return g, nil
		}
	case '\u0020':
		if g, err := s.charCodeToGlyphIndexRaw('\u0020'); err == nil {
			return g, nil
		}
		if g, err := s.charCodeToGlyphIndexRaw('\u00a0'); err == nil {
			return g, nil
		}
	}
	return s.charCodeToGlyphIndexRaw(r)
}

// charCodeToGlyphIndexRaw is de onveranderde upstream-lookup, zonder de
// NBSP-alias hierboven.
func (s *SubsetFontObj) charCodeToGlyphIndexRaw(r rune) (uint, error) {
	value := uint64(r)
	if value <= 0xFFFF {
		gIndex, err := s.charCodeToGlyphIndexFormat4(r)
		if err != nil {
			return 0, err
		}
		return gIndex, nil
	}
	gIndex, err := s.charCodeToGlyphIndexFormat12(r)
	if err != nil {
		return 0, err
	}
	return gIndex, nil
}

// GlyphIndexToPdfWidth gets width from glyphIndex, in 1/1000 em.
//
// AFWIJKING T.O.V. UPSTREAM: reportlab-getrouwe schaling in float64. Upstream
// rekende `width * 1000 / unitsPerEm` in uint en kapte dus af — mediaan 0,49
// fonteenheden per teken, wat op een cel van 30 tekens @8pt al 0,12pt scheelt
// en in de /W-array zichtbaar accumuleert. Reportlab
// (pdfbase/ttfonts.py:572-576) berekent de factor 1000/upem als float en
// vermenigvuldigt per glyph; bij upem==1000 schaalt het niet. Zie VENDOR.md.
func (s *SubsetFontObj) GlyphIndexToPdfWidth(glyphIndex uint) float64 {

	numberOfHMetrics := s.ttfp.NumberOfHMetrics()
	unitsPerEm := s.ttfp.UnitsPerEm()
	if glyphIndex >= numberOfHMetrics {
		glyphIndex = numberOfHMetrics - 1
	}

	width := s.ttfp.Widths()[glyphIndex]
	if unitsPerEm == 1000 {
		return float64(width)
	}
	return float64(width) * (1000.0 / float64(unitsPerEm))
}

// DefaultWidth is de breedte van glyph 0 (.notdef), in 1/1000 em.
//
// TOEGEVOEGD T.O.V. UPSTREAM. Reportlab rekent deze breedte voor elk teken dat
// niet in de cmap van het font zit (rl_accel.py:106 valt terug op
// face.defaultWidth, dat in extractInfo op glyph 0 wordt gezet).
func (s *SubsetFontObj) DefaultWidth() float64 {
	return s.GlyphIndexToPdfWidth(0)
}

// GetTTFParser gets TTFParser.
func (s *SubsetFontObj) GetTTFParser() *core.TTFParser {
	return &s.ttfp
}

// GetUnderlineThickness underlineThickness.
func (s *SubsetFontObj) GetUnderlineThickness() int {
	return s.ttfp.UnderlineThickness()
}

func (s *SubsetFontObj) GetUnderlineThicknessPx(fontSize float64) float64 {
	return (float64(s.ttfp.UnderlineThickness()) / float64(s.ttfp.UnitsPerEm())) * fontSize
}

// GetUnderlinePosition underline position.
func (s *SubsetFontObj) GetUnderlinePosition() int {
	return s.ttfp.UnderlinePosition()
}

func (s *SubsetFontObj) GetUnderlinePositionPx(fontSize float64) float64 {
	return (float64(s.ttfp.UnderlinePosition()) / float64(s.ttfp.UnitsPerEm())) * fontSize
}

func (s *SubsetFontObj) GetAscender() int {
	return s.ttfp.Ascender()
}

func (s *SubsetFontObj) GetAscenderPx(fontSize float64) float64 {
	return (float64(s.ttfp.Ascender()) / float64(s.ttfp.UnitsPerEm())) * fontSize
}

func (s *SubsetFontObj) GetDescender() int {
	return s.ttfp.Descender()
}

func (s *SubsetFontObj) GetDescenderPx(fontSize float64) float64 {
	return (float64(s.ttfp.Descender()) / float64(s.ttfp.UnitsPerEm())) * fontSize
}
