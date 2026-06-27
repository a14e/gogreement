package unexpsrc

// Reader has an unexported method, so only types declared in this package can
// satisfy it.
type Reader interface {
	read() int
}
