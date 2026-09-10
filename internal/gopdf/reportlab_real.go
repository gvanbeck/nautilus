package gopdf

import (
	"math"
	"strconv"
	"strings"
)

// pdfReal formatteert een breedte als PDF-reëel getal volgens dezelfde regel
// als reportlab's fp_str (reportlab/lib/rl_accel.py:41-60): zes significante
// cijfers, trailing nullen en een trailing punt weggelaten, en alles onder
// 1e-7 als "0".
//
// TOEGEVOEGD T.O.V. UPSTREAM. Upstream schreef de /W-array met %d op een
// afgekapte uint; met float-breedtes is een formatteerregel nodig, en die van
// reportlab overnemen houdt de twee documenten ook op byteniveau vergelijkbaar.
// Zie VENDOR.md.
func pdfReal(v float64) string {
	a := math.Abs(v)
	if a <= 1e-7 {
		return "0"
	}
	decimals := 6
	if a > 1 {
		decimals = 6 - int(math.Log10(a))
		if decimals < 0 {
			decimals = 0
		} else if decimals > 6 {
			decimals = 6
		}
	}
	s := strconv.FormatFloat(v, 'f', decimals, 64)
	if decimals > 0 && strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	return s
}
