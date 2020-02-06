package util

import (
	"fmt"
)

func AbrvInteger(v uint64) string {
	f := float64(v)
	if v >= 1e6 {
		return fmt.Sprintf("%.1fm", f/1e6)
	}
	if v >= 1e3 {
		return fmt.Sprintf("%.1fk", f/1e3)
	}

	return fmt.Sprint(v)
}
