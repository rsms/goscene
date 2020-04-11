package main

import (
	"math"
  "github.com/go-gl/mathgl/mgl32"
)

type Vec2 = mgl32.Vec2
type Mat4 = mgl32.Mat4


func perspective(m *Mat4, fovy, aspect, near, far float32) {
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


func translate(m *Mat4, x, y, z float32) {
  m[12] = m[0] * x + m[4] * y + m[8]  * z + m[12]
  m[13] = m[1] * x + m[5] * y + m[9]  * z + m[13]
  m[14] = m[2] * x + m[6] * y + m[10] * z + m[14]
  m[15] = m[3] * x + m[7] * y + m[11] * z + m[15]
}

func translateZ(m *Mat4, z float32) {
  m[12] = m[8]  * z + m[12]
  m[13] = m[9]  * z + m[13]
  m[14] = m[10] * z + m[14]
  m[15] = m[11] * z + m[15]
}



// func (v Vec2) MulMut(c float32) {
// 	v[0] *= c
// 	v[1] *= c
// }

// type Vec struct {
//   x, y float32
// }

// func (v Vec) String() string {
//   return fmt.Sprintf("(%.1f, %.1f)", v.x, v.y)
// }
