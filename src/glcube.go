package main


type GLCube struct {
  program           *GLProgram
  vertexBuf         *GLBuf
  indexBuf          *GLBuf
  aVertexPosition   uint32
  uModelViewMatrix  GLUniform
  absoluteTransform Matrix4
}

var cubeVertices = GLVertexData(GL_STATIC_DRAW, []float32{
  // Front face
  -1.0, -1.0,  1.0,
   1.0, -1.0,  1.0,
   1.0,  1.0,  1.0,
  -1.0,  1.0,  1.0,

  // Back face
  -1.0, -1.0, -1.0,
  -1.0,  1.0, -1.0,
   1.0,  1.0, -1.0,
   1.0, -1.0, -1.0,

  // Top face
  -1.0,  1.0, -1.0,
  -1.0,  1.0,  1.0,
   1.0,  1.0,  1.0,
   1.0,  1.0, -1.0,

  // Bottom face
  -1.0, -1.0, -1.0,
   1.0, -1.0, -1.0,
   1.0, -1.0,  1.0,
  -1.0, -1.0,  1.0,

  // Right face
   1.0, -1.0, -1.0,
   1.0,  1.0, -1.0,
   1.0,  1.0,  1.0,
   1.0, -1.0,  1.0,

  // Left face
  -1.0, -1.0, -1.0,
  -1.0, -1.0,  1.0,
  -1.0,  1.0,  1.0,
  -1.0,  1.0, -1.0,
  },
  // This array defines each face as two triangles, using the indices into the
  // vertex array to specify each triangle's position.
  0,  1,  2,      0,  2,  3,    // front
  4,  5,  6,      4,  6,  7,    // back
  8,  9,  10,     8,  10, 11,   // top
  12, 13, 14,     12, 14, 15,   // bottom
  16, 17, 18,     16, 18, 19,   // right
  20, 21, 22,     20, 22, 23,   // left
)


func NewGLCube(program *GLProgram) (*GLCube, error) {
	gl := program.gl
	o := &GLCube{
		program: program,
		vertexBuf: gl.GetVertexBuffer(cubeVertices),
		indexBuf: gl.GetIndexBuffer(cubeVertices),
		absoluteTransform: Matrix4Identity,
	}
	// get shader positions
	var err error
	o.aVertexPosition, err = program.getAttribLocation("aVertexPosition")
	if err == nil {
		o.uModelViewMatrix, err = program.getUniformLocation("uModelViewMatrix")
	}
	return o, err
}


func (o *GLCube) Draw(r *Renderer) {
	gl := o.program.gl

  // Tell WebGL how to pull out the positions from the position
  // buffer into the vertexPosition attribute
	gl.bindBuffer(GL_ARRAY_BUFFER, o.vertexBuf.pos)
  gl.vertexAttribPointer(
    o.aVertexPosition,
    3,        // size: values per element
    GL_FLOAT, // type: the data in the buffer is 32bit floats
    false,    // normalize: no
    0,        // stride: 0 = use type and size value.
    o.vertexBuf.offset, // offset: offset in buffer to start at, in bytes.
  )
  gl.enableVertexAttribArray(o.aVertexPosition)

  // Tell WebGL which indices to use to index the vertices
  gl.bindBuffer(GL_ELEMENT_ARRAY_BUFFER, o.indexBuf.pos)

	// enable program
	gl.useProgram(o.program)


  // set model view matrix
	tm := o.absoluteTransform

	time := float32(host.scenetime)
  // tm = tm.WithRotation(time, 0.0, 0.0, 1.0)
  // tm = tm.WithRotation(time, 0, 1, 0.5)
  tm.Translate(sin32(time*0.5) * 1.0, cos32(time) * 1.5, 0.0)
  tm.Rotate(sin32(time * 0.8), cos32(time * 0.5), time * 2)

  // tm.RotateX((r.pointer[1] / r.resolution[1]) * PI)
  tm.RotateY((r.pointer[0] / r.resolution[0]) * PI)
  tm.RotateZ((r.pointer[1] / r.resolution[1]) * PI)

  tm.Scale(0.2 + abs32(sin32(time)), 0.2 + abs32(cos32(time)), 0.5)

  gl.uniformMatrix4fv(o.uModelViewMatrix, false, tm)


  gl.drawElements(GL_TRIANGLES, /*vertexCount*/ 36, GL_UNSIGNED_SHORT, /*offset*/ 0)
}
