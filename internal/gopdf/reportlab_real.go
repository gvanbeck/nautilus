package gopdf

import (
	"math"
	"strconv"
	"strings"
)

// pdfReal formats a width as a PDF real number following the same rule as
// reportlab's fp_str (reportlab/lib/rl_accel.py:41-60): six significant
// digits, trailing zeroes and a trailing period dropped, and anything below
// 1e-7 as "0".
//
// ADDED RELATIVE TO UPSTREAM. Upstream wrote the /W array with %d on a
// truncated uint; float widths need a formatting rule, and adopting
// reportlab's keeps the two documents comparable at the byte level as well.
// See VENDOR.md.
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
