package matrix

// Matrix3 is a 3 dimentional matrix
type Matrix3 [3][3]float64

// New3 returns a new 3x3 identity matrix
func New3() Matrix3 {
	m := Matrix3{}
	m[0][0] = 1
	m[1][1] = 1
	m[2][2] = 1

	return m
}

// IsIdentity checks whether the matrix is an identity matrix:
// [1, 0, 0
//  0, 1, 0
//  0, 0, 1]
func (m Matrix3) IsIdentity() bool {
	return m == New3()
}

// IsAffine checks whether the matrix represents an affine transformation
func (m Matrix3) IsAffine() bool {
	return m[2][0] == 0 && m[2][1] == 0 && m[2][2] == 1
}

// IsScale checks whether the matrix is a scaling matrix:
// [x, 0, x
//  0, x, x
//  0, 0, 1]
func (m Matrix3) IsScale() bool {
	return m[0][1] == 0 && m[1][0] == 0 && m.IsAffine()
}

// IsTranslate checks whether the matrix is a translate matrix
// [1, 0, x
//  0, 1, x
//  0, 0, 1]
func (m Matrix3) IsTranslate() bool {
	copy := m
	copy[0][2] = 0
	copy[1][2] = 0
	return copy.IsIdentity()
}

// Determinant returns the determinant of the matrix
func (m Matrix3) Determinant() float64 {
	return m[0][0]*(m[1][1]*m[2][2]-m[1][2]*m[2][1]) -
		m[0][1]*(m[1][0]*m[2][2]-m[1][2]*m[2][0]) +
		m[0][2]*(m[1][0]*m[2][1]-m[1][1]*m[2][0])
}

func (m *Matrix3) Invert() {
	copy := *m
	coeff := 1 / m.Determinant()

	m[0][0] = (copy[1][1]*copy[2][2] - copy[1][2]*copy[2][1]) * coeff
	m[1][0] = (copy[1][2]*copy[2][0] - copy[1][0]*copy[2][2]) * coeff
	m[2][0] = (copy[1][0]*copy[2][1] - copy[1][1]*copy[2][0]) * coeff

	m[0][1] = (copy[0][2]*copy[2][1] - copy[0][1]*copy[2][2]) * coeff
	m[1][1] = (copy[0][0]*copy[2][2] - copy[0][2]*copy[2][0]) * coeff
	m[2][1] = (copy[0][1]*copy[2][0] - copy[0][0]*copy[2][1]) * coeff

	m[0][2] = (copy[0][1]*copy[1][2] - copy[0][2]*copy[1][1]) * coeff
	m[1][2] = (copy[0][2]*copy[1][0] - copy[0][0]*copy[1][2]) * coeff
	m[2][2] = (copy[0][0]*copy[1][1] - copy[0][1]*copy[1][0]) * coeff
}
