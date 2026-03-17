package impl

// ctokenKind is a content-stream token kind.
type ctokenKind int

const (
	ctEOF ctokenKind = iota
	ctInteger
	ctReal
	ctName   // /Name operand
	ctString // ( ... )
	ctHex    // < ... >
	ctLArray // [
	ctRArray // ]
	ctLDict  // <<
	ctRDict  // >>
	ctOp     // operator name: q, Q, cm, BT, ET, Tf, Tj, etc.
)

// ctoken is a single content stream token.
type ctoken struct {
	kind   ctokenKind
	value  string
	intVal int64
	fltVal float64
}
