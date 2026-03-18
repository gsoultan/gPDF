package reader

import "math"

type matrix struct {
	a float64
	b float64
	c float64
	d float64
	e float64
	f float64
}

func identityMatrix() matrix {
	return matrix{a: 1, d: 1}
}

func matrixFromArgs(args []float64) matrix {
	if len(args) < 6 {
		return identityMatrix()
	}
	return matrix{a: args[0], b: args[1], c: args[2], d: args[3], e: args[4], f: args[5]}
}

func (m matrix) multiply(other matrix) matrix {
	return matrix{
		a: m.a*other.a + m.b*other.c,
		b: m.a*other.b + m.b*other.d,
		c: m.c*other.a + m.d*other.c,
		d: m.c*other.b + m.d*other.d,
		e: m.e*other.a + m.f*other.c + other.e,
		f: m.e*other.b + m.f*other.d + other.f,
	}
}

func (m matrix) apply(x, y float64) (float64, float64) {
	return x*m.a + y*m.c + m.e, x*m.b + y*m.d + m.f
}

func (m matrix) rotationDegrees() float64 {
	return math.Atan2(m.b, m.a) * 180 / math.Pi
}

func translationMatrix(tx, ty float64) matrix {
	return matrix{a: 1, d: 1, e: tx, f: ty}
}
