package testutil

import "fmt"

// panicf wraps errors with context and panics.
func panicf(format string, args ...interface{}) {
	panic("testutil: " + fmt.Sprintf(format, args...))
}
