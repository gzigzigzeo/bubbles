package validate

import (
	"fmt"
	"strings"
)

// InRange returns a validator that checks lo ≤ v ≤ hi.
func InRange(lo, hi int) func(int) string {
	return func(v int) string {
		if v < lo || v > hi {
			return fmt.Sprintf("must be in range: %d–%d", lo, hi)
		}

		return ""
	}
}

// NotEmptyString returns a validator that rejects empty or whitespace-only strings.
func NotEmptyString() func(string) string {
	return func(v string) string {
		if strings.TrimSpace(v) == "" {
			return "must be set"
		}

		return ""
	}
}
