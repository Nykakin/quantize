package quantize

type matrix interface {
	set(val float64)
	add(r, l matrix)
	sub(r, l matrix)
	rcount() int
	ccount() int
	at(i, j int) float64
}

type mat3x3 [][]float64

func newMat3x3() mat3x3 {
	v := make([][]float64, 3)
	for i := range v {
		v[i] = make([]float64, 3)
	}
	return mat3x3(v)
}

func (m mat3x3) at(r, c int) float64 {
	return m[r][c]
}

func (m mat3x3) rcount() int {
	return 3
}

func (m mat3x3) ccount() int {
	return 3
}

func (m mat3x3) set(val float64) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			m[i][j] = val
		}
	}
}

func (m mat3x3) mul(l, r matrix) {
	m.set(0.0)
	for i := 0; i < l.rcount(); i++ {
		for j := 0; j < r.ccount(); j++ {
			for k := 0; k < l.ccount(); k++ {
				m[i][j] += l.at(i, k) * r.at(k, j)
			}
		}
	}
}

func (m mat3x3) add(l, r matrix) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			m[i][j] = l.at(i, j) + r.at(i, j)
		}
	}
}

func (m mat3x3) sub(l, r matrix) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			m[i][j] = l.at(i, j) - r.at(i, j)
		}
	}
}

func (m mat3x3) apply(f func(i, j int, v float64) float64, r mat3x3) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			m[i][j] = f(i, j, r[i][j])
		}
	}
}

type vec3x1 [][]float64

func (m vec3x1) at(r, c int) float64 {
	return m[r][c]
}

func newVec3x1() vec3x1 {
	v := make([][]float64, 3)
	for i := range v {
		v[i] = make([]float64, 1)
	}
	return vec3x1(v)
}

func (m vec3x1) rcount() int {
	return 3
}

func (m vec3x1) ccount() int {
	return 1
}

func (m vec3x1) set(val float64) {
	for i := 0; i < 3; i++ {
		m[i][0] = val
	}
}

func (m vec3x1) setVec(vals []float64) {
	for i := 0; i < 3; i++ {
		m[i][0] = vals[i]
	}
}

func (m vec3x1) add(l, r matrix) {
	for i := 0; i < 3; i++ {
		m[i][0] = l.at(i, 0) + r.at(i, 0)
	}
}

func (m vec3x1) sub(l, r matrix) {
	for i := 0; i < 3; i++ {
		m[i][0] = l.at(i, 0) - r.at(i, 0)
	}
}

func (m vec3x1) atVec(i int) float64 {
	return m[i][0]
}

func (m vec3x1) t(v vec1x3) {
	m[0][0] = v[0][0]
	m[1][0] = v[0][1]
	m[2][0] = v[0][2]
}

type vec1x3 [][]float64

func (m vec1x3) at(r, c int) float64 {
	return m[r][c]
}

func newVec1x3() vec1x3 {
	v := make([][]float64, 1)
	for i := range v {
		v[i] = make([]float64, 3)
	}
	return vec1x3(v)
}

func (m vec1x3) rcount() int {
	return 1
}

func (m vec1x3) ccount() int {
	return 3
}

func (m vec1x3) set(val float64) {
	for i := 0; i < 3; i++ {
		m[0][i] = val
	}
}

func (m vec1x3) setVec(vals []float64) {
	for i := 0; i < 3; i++ {
		m[0][i] = vals[i]
	}
}

func (m vec1x3) add(l, r matrix) {
	for i := 0; i < 3; i++ {
		m[0][i] = l.at(0, i) + r.at(0, i)
	}
}

func (m vec1x3) sub(l, r matrix) {
	for i := 0; i < 3; i++ {
		m[0][i] = l.at(0, i) - r.at(0, i)
	}
}

func (m vec1x3) atVec(i int) float64 {
	return m[0][i]
}

func (m vec1x3) t(v vec3x1) {
	m[0][0] = v[0][0]
	m[0][1] = v[1][0]
	m[0][2] = v[2][0]
}
