package rtl

// joinType returns the joining behaviour of a character for Arabic shaping.
//
//	'D' – dual-joining: connects on both sides (most letters).
//	'R' – right-joining: only accepts a connection from its predecessor.
//	'T' – transparent: diacritics; skipped when scanning for neighbours.
//	'U' – non-joining: no connection in either direction.
func joinType(r rune) byte {
	switch {
	// Arabic diacritics (harakat, shadda, sukun, tatweel, …) are transparent.
	case r >= 0x064B && r <= 0x0652,
		r == 0x0653, r == 0x0654, r == 0x0655,
		r == 0x0656, r == 0x0657, r == 0x0658,
		r == 0x0640, // tatweel (kashida) – transparent
		r == 0x0670: // superscript alef
		return 'T'

	// Right-joining: alef variants, ta marbuta, dal, dhal, ra, zain, waw,
	// alef maqsura – these only connect on their right side.
	case r == 0x0622, r == 0x0623, r == 0x0624, r == 0x0625,
		r == 0x0627, r == 0x0629,
		r == 0x062F, r == 0x0630, r == 0x0631, r == 0x0632,
		r == 0x0648, r == 0x0649:
		return 'R'

	// Lam-alef ligatures behave as right-joining after they are formed.
	case r >= 0xFEF5 && r <= 0xFEFC:
		return 'R'
	}

	if _, ok := arabicForms[r]; ok {
		return 'D'
	}
	return 'U'
}

// arabicForms maps Arabic base letters to their four contextual presentation
// forms: [isolated, final, initial, medial].
// A zero value means the form does not exist; isolated is used as fallback.
var arabicForms = map[rune][4]rune{
	// ── Dual-joining ──────────────────────────────────────────────────────
	0x0626: {0xFE89, 0xFE8A, 0xFE8B, 0xFE8C}, // ئ ya with hamza above
	0x0628: {0xFE8F, 0xFE90, 0xFE91, 0xFE92}, // ب ba
	0x062A: {0xFE95, 0xFE96, 0xFE97, 0xFE98}, // ت ta
	0x062B: {0xFE99, 0xFE9A, 0xFE9B, 0xFE9C}, // ث tha
	0x062C: {0xFE9D, 0xFE9E, 0xFE9F, 0xFEA0}, // ج jeem
	0x062D: {0xFEA1, 0xFEA2, 0xFEA3, 0xFEA4}, // ح ha
	0x062E: {0xFEA5, 0xFEA6, 0xFEA7, 0xFEA8}, // خ kha
	0x0633: {0xFEB1, 0xFEB2, 0xFEB3, 0xFEB4}, // س sin
	0x0634: {0xFEB5, 0xFEB6, 0xFEB7, 0xFEB8}, // ش shin
	0x0635: {0xFEB9, 0xFEBA, 0xFEBB, 0xFEBC}, // ص sad
	0x0636: {0xFEBD, 0xFEBE, 0xFEBF, 0xFEC0}, // ض dad
	0x0637: {0xFEC1, 0xFEC2, 0xFEC3, 0xFEC4}, // ط tah
	0x0638: {0xFEC5, 0xFEC6, 0xFEC7, 0xFEC8}, // ظ zah
	0x0639: {0xFEC9, 0xFECA, 0xFECB, 0xFECC}, // ع ain
	0x063A: {0xFECD, 0xFECE, 0xFECF, 0xFED0}, // غ ghain
	0x0641: {0xFED1, 0xFED2, 0xFED3, 0xFED4}, // ف fa
	0x0642: {0xFED5, 0xFED6, 0xFED7, 0xFED8}, // ق qaf
	0x0643: {0xFED9, 0xFEDA, 0xFEDB, 0xFEDC}, // ك kaf
	0x0644: {0xFEDD, 0xFEDE, 0xFEDF, 0xFEE0}, // ل lam
	0x0645: {0xFEE1, 0xFEE2, 0xFEE3, 0xFEE4}, // م meem
	0x0646: {0xFEE5, 0xFEE6, 0xFEE7, 0xFEE8}, // ن nun
	0x0647: {0xFEE9, 0xFEEA, 0xFEEB, 0xFEEC}, // ه ha
	0x064A: {0xFEF1, 0xFEF2, 0xFEF3, 0xFEF4}, // ي ya

	// ── Right-joining (only isolated and final forms) ─────────────────────
	0x0622: {0xFE81, 0xFE82, 0, 0}, // آ alef with madda above
	0x0623: {0xFE83, 0xFE84, 0, 0}, // أ alef with hamza above
	0x0624: {0xFE85, 0xFE86, 0, 0}, // ؤ waw with hamza above
	0x0625: {0xFE87, 0xFE88, 0, 0}, // إ alef with hamza below
	0x0627: {0xFE8D, 0xFE8E, 0, 0}, // ا alef
	0x0629: {0xFE93, 0xFE94, 0, 0}, // ة ta marbuta
	0x062F: {0xFEA9, 0xFEAA, 0, 0}, // د dal
	0x0630: {0xFEAB, 0xFEAC, 0, 0}, // ذ dhal
	0x0631: {0xFEAD, 0xFEAE, 0, 0}, // ر ra
	0x0632: {0xFEAF, 0xFEB0, 0, 0}, // ز zain
	0x0648: {0xFEED, 0xFEEE, 0, 0}, // و waw
	0x0649: {0xFEEF, 0xFEF0, 0, 0}, // ى alef maqsura
}

// lamAlefForms maps an alef-variant to its lam-alef ligature forms:
// [isolated, final].  "Final" is used when lam has a right-side connection.
var lamAlefForms = map[rune][2]rune{
	0x0622: {0xFEF5, 0xFEF6}, // ل + آ
	0x0623: {0xFEF7, 0xFEF8}, // ل + أ
	0x0625: {0xFEF9, 0xFEFA}, // ل + إ
	0x0627: {0xFEFB, 0xFEFC}, // ل + ا
}

// shapeArabic replaces Arabic base letters with their contextual presentation
// forms based on their position within a word, and applies mandatory lam-alef
// ligatures.  The input must be in logical (storage) order; the output
// remains in logical order with shaped code points.
func shapeArabic(text string) string {
	src := []rune(text)

	// ── Pass 1: mandatory lam-alef ligatures ──────────────────────────────
	// Replace ل + [alef-variant] with a single ligature glyph.
	merged := make([]rune, 0, len(src))
	for i := 0; i < len(src); i++ {
		if src[i] == 0x0644 && i+1 < len(src) {
			if ligs, ok := lamAlefForms[src[i+1]]; ok {
				// Lam has a right-side connection if the nearest non-transparent
				// predecessor (already placed in merged) is dual-joining.
				prevDual := false
				for j := len(merged) - 1; j >= 0; j-- {
					if jt := joinType(merged[j]); jt != 'T' {
						prevDual = jt == 'D'
						break
					}
				}
				if prevDual {
					merged = append(merged, ligs[1]) // final form
				} else {
					merged = append(merged, ligs[0]) // isolated form
				}
				i++ // consume the alef
				continue
			}
		}
		merged = append(merged, src[i])
	}

	// ── Pass 2: contextual forms for remaining Arabic letters ─────────────
	result := make([]rune, 0, len(merged))
	for i, r := range merged {
		forms, ok := arabicForms[r]
		if !ok {
			result = append(result, r)
			continue
		}

		prev := prevNonTransparent(merged, i)
		next := nextNonTransparent(merged, i)

		// Right-side connection: nearest predecessor is dual-joining.
		connectsRight := prev >= 0 && joinType(merged[prev]) == 'D'
		// Left-side connection: current char is dual AND nearest successor is
		// dual or right-joining (it "accepts" our left arm).
		selfDual := joinType(r) == 'D'
		connectsLeft := selfDual && next >= 0 &&
			(joinType(merged[next]) == 'D' || joinType(merged[next]) == 'R')

		form := 0 // isolated
		switch {
		case connectsRight && connectsLeft:
			form = 3 // medial
		case connectsRight:
			form = 1 // final
		case connectsLeft:
			form = 2 // initial
		}

		shaped := forms[form]
		if shaped == 0 {
			shaped = forms[0] // no such form → fall back to isolated
		}
		result = append(result, shaped)
	}

	return string(result)
}

// prevNonTransparent returns the index of the nearest non-transparent
// character before position i, or -1 if none.
func prevNonTransparent(rs []rune, i int) int {
	for j := i - 1; j >= 0; j-- {
		if joinType(rs[j]) != 'T' {
			return j
		}
	}
	return -1
}

// nextNonTransparent returns the index of the nearest non-transparent
// character after position i, or -1 if none.
func nextNonTransparent(rs []rune, i int) int {
	for j := i + 1; j < len(rs); j++ {
		if joinType(rs[j]) != 'T' {
			return j
		}
	}
	return -1
}
