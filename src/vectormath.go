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
  //  float32(f / aspect), 0, 0, 0,
  //  0, float32(f), 0, 0,
  //  0, 0, float32((near + far) / nmf), -1,
  //  0, 0, float32((2.0 * far * near) / nmf), 0,
  // }
}

// Matrix4Scale returns an identity matrix with scale
func Matrix4Scale(scaleX, scaleY, scaleZ float32) Matrix4 {
  return Matrix4{
    scaleX, 0, 0, 0,
    0, scaleY, 0, 0,
    0, 0, scaleZ, 0,
    0, 0,      0, 1,
  }
}

func Matrix4RotateX(radians float32) Matrix4 {
  sin := sin32(radians)
  cos := cos32(radians)
  return Matrix4{1, 0, 0, 0, 0, cos, sin, 0, 0, -sin, cos, 0, 0, 0, 0, 1}
}

func Matrix4RotateY(radians float32) Matrix4 {
  sin := sin32(radians)
  cos := cos32(radians)
  return Matrix4{cos, 0, -sin, 0, 0, 1, 0, 0, sin, 0, cos, 0, 0, 0, 0, 1}
}

func Matrix4RotateZ(radians float32) Matrix4 {
  sin := sin32(radians)
  cos := cos32(radians)
  return Matrix4{cos, sin, 0, 0, -sin, cos, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
}


func (m *Matrix4) Translate(x, y, z float32) *Matrix4 {
  m[12] = m[0]*x + m[4]*y + m[8] *z + m[12]
  m[13] = m[1]*x + m[5]*y + m[9] *z + m[13]
  m[14] = m[2]*x + m[6]*y + m[10]*z + m[14]
  m[15] = m[3]*x + m[7]*y + m[11]*z + m[15]
  return m
}

func (m *Matrix4) TranslateZ(z float32) {
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

func (m *Matrix4) Mul4Mut(m2 *Matrix4) {
  v0, v1,  v2,  v3,  v4,  v5,  v6,  v7 := m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7]
  v8, v9, v10, v11, v12, v13, v14, v15 := m[8], m[9], m[10], m[11], m[12], m[13], m[14], m[15]
  m[0]  = v0*m2[0]  + v4*m2[1]  + v8 *m2[2]  + v12*m2[3]
  m[1]  = v1*m2[0]  + v5*m2[1]  + v9 *m2[2]  + v13*m2[3]
  m[2]  = v2*m2[0]  + v6*m2[1]  + v10*m2[2]  + v14*m2[3]
  m[3]  = v3*m2[0]  + v7*m2[1]  + v11*m2[2]  + v15*m2[3]
  m[4]  = v0*m2[4]  + v4*m2[5]  + v8 *m2[6]  + v12*m2[7]
  m[5]  = v1*m2[4]  + v5*m2[5]  + v9 *m2[6]  + v13*m2[7]
  m[6]  = v2*m2[4]  + v6*m2[5]  + v10*m2[6]  + v14*m2[7]
  m[7]  = v3*m2[4]  + v7*m2[5]  + v11*m2[6]  + v15*m2[7]
  m[8]  = v0*m2[8]  + v4*m2[9]  + v8 *m2[10] + v12*m2[11]
  m[9]  = v1*m2[8]  + v5*m2[9]  + v9 *m2[10] + v13*m2[11]
  m[10] = v2*m2[8]  + v6*m2[9]  + v10*m2[10] + v14*m2[11]
  m[11] = v3*m2[8]  + v7*m2[9]  + v11*m2[10] + v15*m2[11]
  m[12] = v0*m2[12] + v4*m2[13] + v8 *m2[14] + v12*m2[15]
  m[13] = v1*m2[12] + v5*m2[13] + v9 *m2[14] + v13*m2[15]
  m[14] = v2*m2[12] + v6*m2[13] + v10*m2[14] + v14*m2[15]
  m[15] = v3*m2[12] + v7*m2[13] + v11*m2[14] + v15*m2[15]
}

func (m *Matrix4) WithRotation3D(radians, axisX, axisY, axisZ float32) Matrix4 {
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


func (m *Matrix4) Rotate(radiansX, radiansY, radiansZ float32) *Matrix4 {
  return m.RotateX(radiansX).RotateY(radiansY).RotateZ(radiansZ)
}


func (m *Matrix4) RotateX(radians float32) *Matrix4 {
  sin := sin32(radians)
  cos := cos32(radians)
  sinn := -sin

  v4, v5,  v6,  v7 := m[4], m[5], m[6], m[7]
  v8, v9, v10, v11 := m[8], m[9], m[10], m[11]

  m[4]  = v4*cos  + v8 *sin
  m[5]  = v5*cos  + v9 *sin
  m[6]  = v6*cos  + v10*sin
  m[7]  = v7*cos  + v11*sin
  m[8]  = v4*sinn + v8 *cos
  m[9]  = v5*sinn + v9 *cos
  m[10] = v6*sinn + v10*cos
  m[11] = v7*sinn + v11*cos

  return m
}

func (m *Matrix4) RotateY(radians float32) *Matrix4 {
  sin := sin32(radians)
  cos := cos32(radians)
  sinn := -sin

  v0, v1,  v2,  v3 := m[0], m[1], m[2], m[3]
  v8, v9, v10, v11 := m[8], m[9], m[10], m[11]

  m[0]  = v0*cos + v8 *sinn
  m[1]  = v1*cos + v9 *sinn
  m[2]  = v2*cos + v10*sinn
  m[3]  = v3*cos + v11*sinn

  m[8]  = v0*sin + v8 *cos
  m[9]  = v1*sin + v9 *cos
  m[10] = v2*sin + v10*cos
  m[11] = v3*sin + v11*cos

  return m
}

func (m *Matrix4) RotateZ(radians float32) *Matrix4 {
  // *m = m.RotateZ(radians)
  sin := sin32(radians)
  cos := cos32(radians)
  sinn := -sin

  v0, v1, v2, v3 := m[0], m[1], m[2], m[3]
  v4, v5, v6, v7 := m[4], m[5], m[6], m[7]

  m[0] = v0*cos  + v4*sin
  m[1] = v1*cos  + v5*sin
  m[2] = v2*cos  + v6*sin
  m[3] = v3*cos  + v7*sin

  m[4] = v0*sinn + v4*cos
  m[5] = v1*sinn + v5*cos
  m[6] = v2*sinn + v6*cos
  m[7] = v3*sinn + v7*cos

  return m
}


// // Scale returns a copy of m with scale x,y,z applied
// func (m *Matrix4) Scale(x, y, z float32) Matrix4 {
//   m2 := *m
//   m2.ScaleMut(x, y, z)
//   return m2
// }

func (m *Matrix4) Scale(x, y, z float32) *Matrix4 {
  m[0]  *= x
  m[1]  *= x
  m[2]  *= x
  m[3]  *= x
  m[4]  *= y
  m[5]  *= y
  m[6]  *= y
  m[7]  *= y
  m[8]  *= z
  m[9]  *= z
  m[10] *= z
  m[11] *= z
  return m
}


func sin32(v float32) float32 { return float32(math.Sin(float64(v))) }
func cos32(v float32) float32 { return float32(math.Cos(float64(v))) }
func abs32(v float32) float32 { return float32(math.Abs(float64(v))) }
