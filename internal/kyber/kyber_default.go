//go:build !cgo

package kyber

// newEncapsulator returns nil when CGO is disabled.
func newEncapsulator() encapsulator {
	return nil
}
