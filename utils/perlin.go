package utils

import (
	"math/rand"
)

const (
	B  = 0x100
	N  = 0x1000
	BM = 0xff
)

func NewPerlin(alpha, beta float64, n int, seed int64) *Perlin {
	return NewPerlinRandSource(alpha, beta, n, rand.NewSource(seed))
}

// Perlin is the noise generator
type Perlin struct {
	alpha float64
	beta  float64
	n     int

	p  [B + B + 2]int
	g3 [B + B + 2][3]float64
	g2 [B + B + 2][2]float64
	g1 [B + B + 2]float64
}

func NewPerlinRandSource(alpha, beta float64, n int, source rand.Source) *Perlin {
	var p Perlin
	var i int

	p.alpha = alpha
	p.beta = beta
	p.n = n

	r := rand.New(source) // #nosec G404

	for i = 0; i < B; i++ {
		p.p[i] = i
		p.g1[i] = float64((r.Int()%(B+B))-B) / B

		for j := 0; j < 2; j++ {
			p.g2[i][j] = float64((r.Int()%(B+B))-B) / B
		}
	}

	for i = B - 1; i >= 0; i-- {
		j := r.Int() % B
		p.p[i], p.p[j] = p.p[j], p.p[i]
	}

	for i = 0; i < B+2; i++ {
		p.p[B+i] = p.p[i]
		p.g1[B+i] = p.g1[i]
		for j := 0; j < 2; j++ {
			p.g2[B+i][j] = p.g2[i][j]
		}
		for j := 0; j < 3; j++ {
			p.g3[B+i][j] = p.g3[i][j]
		}
	}

	return &p
}

func (p *Perlin) Noise1D(x float64) float64 {
	var bx0, bx1 int
	var rx0, rx1, sx, t, u, v float64

	t = x + N
	bx0 = int(t) & BM
	bx1 = (bx0 + 1) & BM
	rx0 = t - float64(int(t))
	rx1 = rx0 - 1.0

	sx = sCurve(rx0)

	u = rx0 * p.g1[p.p[bx0]]
	v = rx1 * p.g1[p.p[bx1]]

	return lerp(sx, u, v)
}

func (p *Perlin) Noise2D(x, y float64) float64 {
	var bx0, bx1, by0, by1 int
	var b00, b10, b01, b11 int
	var rx0, rx1, ry0, ry1, sx, sy, a, b, t, u, v float64
	var q [2]float64
	var i, j int

	t = x + N
	bx0 = int(t) & BM
	bx1 = (bx0 + 1) & BM
	rx0 = t - float64(int(t))
	rx1 = rx0 - 1.0

	t = y + N
	by0 = int(t) & BM
	by1 = (by0 + 1) & BM
	ry0 = t - float64(int(t))
	ry1 = ry0 - 1.0

	i = p.p[bx0]
	j = p.p[bx1]

	b00 = p.p[i+by0]
	b10 = p.p[j+by0]
	b01 = p.p[i+by1]
	b11 = p.p[j+by1]

	sx = sCurve(rx0)
	sy = sCurve(ry0)

	q = p.g2[b00]
	u = at2(rx0, ry0, q)
	q = p.g2[b10]
	v = at2(rx1, ry0, q)
	a = lerp(sx, u, v)

	q = p.g2[b01]
	u = at2(rx0, ry1, q)
	q = p.g2[b11]
	v = at2(rx1, ry1, q)
	b = lerp(sx, u, v)

	return lerp(sy, a, b)
}

func sCurve(t float64) float64 {
	return t * t * (3.0 - 2.0*t)
}

func lerp(t, a, b float64) float64 {
	return a + t*(b-a)
}

func at2(rx, ry float64, q [2]float64) float64 {
	return rx*q[0] + ry*q[1]
}
