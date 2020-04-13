package main

import (
	"math"
	"github.com/go-gl/mathgl/mgl32"
)

type Vec2    mgl32.Vec2
type Vec3    mgl32.Vec3
type Matrix4 mgl32.Mat4


func (v Vec2) Add(b Vec2) Vec2 { return Vec2(mgl32.Vec2(v).Add(mgl32.Vec2(b))) }
func (v Vec2) Sub(b Vec2) Vec2 { return Vec2(mgl32.Vec2(v).Sub(mgl32.Vec2(b))) }
func (v Vec2) Mul(b float32) Vec2 { return Vec2(mgl32.Vec2(v).Mul(b)) }

func (v Vec3) Add(b Vec3) Vec3 { return Vec3(mgl32.Vec3(v).Add(mgl32.Vec3(b))) }
func (v Vec3) Sub(b Vec3) Vec3 { return Vec3(mgl32.Vec3(v).Sub(mgl32.Vec3(b))) }
func (v Vec3) Mul(b float32) Vec3 { return Vec3(mgl32.Vec3(v).Mul(b)) }


var Matrix4Identity = Matrix4{
	1, 0, 0, 0,
	0, 1, 0, 0,
	0, 0, 1, 0,
	0, 0, 0, 1,
}


func Matrix4Perspective(fovy, aspect, near, far float32) Matrix4 {
	return Matrix4(mgl32.Perspective(fovy, aspect, near, far))
	// // fovy = (fovy * math.Pi) / 180.0 // convert from degrees to radians
	// nmf, f := near-far, float32(1.0/math.Tan(float64(fovy)/2.0))
	// return Matrix4{
	// 	float32(f / aspect), 0, 0, 0,
	// 	0, float32(f), 0, 0,
	// 	0, 0, float32((near + far) / nmf), -1,
	// 	0, 0, float32((2.0 * far * near) / nmf), 0,
	// }
}

func (m *Matrix4) PerspectiveMut(fovy, aspect, near, far float32) {
	f := 1.0 / float32(math.Tan(float64(fovy) / 2.0))
	nf := 1.0 / (near - far)
	m[0] = f / aspect
	m[1] = 0
	m[2] = 0
	m[3] = 0
	m[4] = 0
	m[5] = f
	m[6] = 0
	m[7] = 0
	m[8] = 0
	m[9] = 0
	m[10] = (far + near) * nf
	m[11] = -1
	m[12] = 0
	m[13] = 0
	m[14] = 2 * far * near * nf
	m[15] = 0
}

func (m *Matrix4) TranslateMut(x, y, z float32) {
	m[12] = m[0]*x + m[4]*y + m[8] *z + m[12]
	m[13] = m[1]*x + m[5]*y + m[9] *z + m[13]
	m[14] = m[2]*x + m[6]*y + m[10]*z + m[14]
	m[15] = m[3]*x + m[7]*y + m[11]*z + m[15]
}

func (m *Matrix4) TranslateZMut(z float32) {
	m[12] = m[8]  * z + m[12]
	m[13] = m[9]  * z + m[13]
	m[14] = m[10] * z + m[14]
	m[15] = m[11] * z + m[15]
}

func (m1 *Matrix4) Mul(v float32) Matrix4 { return Matrix4(mgl32.Mat4(*m1).Mul(v)) }

func (m1 *Matrix4) MulMut(c float32) {
	m1[0]  *= c; m1[1]  *= c; m1[2]  *= c; m1[3]  *= c
	m1[4]  *= c; m1[5]  *= c; m1[6]  *= c; m1[7]  *= c
	m1[8]  *= c; m1[9]  *= c; m1[10] *= c; m1[11] *= c
	m1[12] *= c; m1[13] *= c; m1[14] *= c; m1[15] *= c
}

// Mul4 performs a "matrix product" between this matrix and another of the given dimension
func (m1 *Matrix4) Mul4(m2 *Matrix4) Matrix4 {
	return Matrix4{
		m1[0]*m2[0]  + m1[4]*m2[1]  + m1[8] *m2[2]  + m1[12]*m2[3],
		m1[1]*m2[0]  + m1[5]*m2[1]  + m1[9] *m2[2]  + m1[13]*m2[3],
		m1[2]*m2[0]  + m1[6]*m2[1]  + m1[10]*m2[2]  + m1[14]*m2[3],
		m1[3]*m2[0]  + m1[7]*m2[1]  + m1[11]*m2[2]  + m1[15]*m2[3],
		m1[0]*m2[4]  + m1[4]*m2[5]  + m1[8] *m2[6]  + m1[12]*m2[7],
		m1[1]*m2[4]  + m1[5]*m2[5]  + m1[9] *m2[6]  + m1[13]*m2[7],
		m1[2]*m2[4]  + m1[6]*m2[5]  + m1[10]*m2[6]  + m1[14]*m2[7],
		m1[3]*m2[4]  + m1[7]*m2[5]  + m1[11]*m2[6]  + m1[15]*m2[7],
		m1[0]*m2[8]  + m1[4]*m2[9]  + m1[8] *m2[10] + m1[12]*m2[11],
		m1[1]*m2[8]  + m1[5]*m2[9]  + m1[9] *m2[10] + m1[13]*m2[11],
		m1[2]*m2[8]  + m1[6]*m2[9]  + m1[10]*m2[10] + m1[14]*m2[11],
		m1[3]*m2[8]  + m1[7]*m2[9]  + m1[11]*m2[10] + m1[15]*m2[11],
		m1[0]*m2[12] + m1[4]*m2[13] + m1[8] *m2[14] + m1[12]*m2[15],
		m1[1]*m2[12] + m1[5]*m2[13] + m1[9] *m2[14] + m1[13]*m2[15],
		m1[2]*m2[12] + m1[6]*m2[13] + m1[10]*m2[14] + m1[14]*m2[15],
		m1[3]*m2[12] + m1[7]*m2[13] + m1[11]*m2[14] + m1[15]*m2[15],
	}
}


func (m *Matrix4) Rotate(radians, axisX, axisY, axisZ float32) Matrix4 {
	s, c := sin32(radians), cos32(radians)
	x, y, z := axisX, axisY, axisZ
	k := 1 - c
	m2 := Matrix4{
		x*x*k + c,    x*y*k + z*s,  x*z*k - y*s,  0,
		x*y*k - z*s,  y*y*k + c,    y*z*k + x*s,  0,
		x*z*k + y*s,  y*z*k - x*s,  z*z*k + c,    0,
		0, 0, 0, 1,
	}
	return m.Mul4(&m2)
}


func (m *Matrix4) RotateX(radians float32) Matrix4 {
	sin := sin32(radians)
	cos := cos32(radians)
  m2 := Matrix4{1, 0, 0, 0, 0, cos, sin, 0, 0, -sin, cos, 0, 0, 0, 0, 1}
  return m.Mul4(&m2)
}


func (m *Matrix4) RotateY(radians float32) Matrix4 {
	sin := sin32(radians)
	cos := cos32(radians)
  m2 := Matrix4{cos, 0, -sin, 0, 0, 1, 0, 0, sin, 0, cos, 0, 0, 0, 0, 1}
  return m.Mul4(&m2)
}

func (m *Matrix4) RotateZ(radians float32) Matrix4 {
	sin := sin32(radians)
	cos := cos32(radians)
  m2 := Matrix4{cos, sin, 0, 0, -sin, cos, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
  return m.Mul4(&m2)
}

func (m *Matrix4) RotateXMut(radians float32) { *m = m.RotateX(radians) }
func (m *Matrix4) RotateYMut(radians float32) { *m = m.RotateY(radians) }
func (m *Matrix4) RotateZMut(radians float32) { *m = m.RotateZ(radians) }

func sin32(v float32) float32 { return float32(math.Sin(float64(v))) }
func cos32(v float32) float32 { return float32(math.Cos(float64(v))) }

