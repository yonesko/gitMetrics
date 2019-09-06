package util

import (
	"fmt"
)

func PrettyBig(v uint64) string {
	if v >= 1e6 {
		return fmt.Sprintf("%v%v", v/1e6, "m")
	}
	if v >= 1e3 {
		return fmt.Sprintf("%v%v", v/1e3, "k")
	}

	return fmt.Sprint(v)
}
