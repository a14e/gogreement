package ctorsource

// Widget must be constructed via NewWidget (which lives in this package).
// @constructor NewWidget
type Widget struct {
	Value int
}

func NewWidget(v int) *Widget {
	return &Widget{Value: v} // ✅ OK: in the declared constructor
}
