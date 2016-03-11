package drawgl

import (
	"strconv"
	"strings"
)

// ParseLength parses a string containing a number an an optional unit of
// length. The supported units are '%' for a percentage, 'px' or none for a
// pixel. It returns both pixel and percentage, as well as any possible error
// that might have occured during parsing.
func ParseLength(dimen string) (px int, percent float64, err error) {
	if dimen != "" {
		if strings.HasSuffix(dimen, "%") {
			dimen = strings.TrimSpace(strings.TrimSuffix(dimen, "%"))
			var d float64
			if d, err = strconv.ParseFloat(dimen, 64); err == nil {
				percent = d / 100
			}
		} else {
			dimen = strings.TrimSpace(strings.TrimSuffix(dimen, "px"))
			var d int
			if d, err = strconv.Atoi(dimen); err == nil {
				px = d
			}
		}
	}

	return
}
