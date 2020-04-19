package main


type GLPlane struct {
	program           *GLProgram
	buf               *GLBuf
	aVertexPosition   uint32
	uModelViewMatrix  GLUniform
	absoluteTransform Matrix4     // TODO: replace with TransformNode in ECS
}

var planeVertices = GLVertexData(GL_STATIC_DRAW, []float32{
  -1.0,  1.0,
   1.0,  1.0,
  -1.0, -1.0,
   1.0, -1.0,
})


func NewGLPlane(program *GLProgram) (*GLPlane, error) {
	gl := program.gl
	buf := gl.GetVertexBuffer(planeVertices)

	// posbuf := gl.createBuffer()
	// gl.bindBuffer(GL_ARRAY_BUFFER, posbuf)
	// gl.bufferDataF32(GL_ARRAY_BUFFER, planeVertices, GL_STATIC_DRAW)

	o := &GLPlane{
		program: program,
		buf: buf,
		absoluteTransform: Matrix4Identity,  // identity; the center of the scene
	}

	// move back a little away from the camera
	o.absoluteTransform.TranslateZ(-5)

	// get shader positions
	var err error
	o.aVertexPosition, err = program.getAttribLocation("aVertexPosition")
	if err == nil {
		o.uModelViewMatrix, err = program.getUniformLocation("uModelViewMatrix")
	}
	return o, err
}


func (o *GLPlane) Draw(r *Renderer) {
	gl := o.program.gl

	// activate vertex buffer
	gl.bindBuffer(GL_ARRAY_BUFFER, o.buf.pos)
  gl.vertexAttribPointer(
    o.aVertexPosition,
    2,        // size: values per element
    GL_FLOAT, // type: the data in the buffer is 32bit floats
    false,    // normalize: no
    0,        // stride: 0 = use type and size value.
    o.buf.offset, // offset: how many bytes inside the buffer to start from.
  )
  gl.enableVertexAttribArray(o.aVertexPosition)

	// enable program
	gl.useProgram(o.program)

  // set model view matrix
	tm := o.absoluteTransform
  tm.RotateY((r.pointer[0] / r.resolution[0]) * PI)
  tm.RotateZ((r.pointer[1] / r.resolution[1]) * PI)
  gl.uniformMatrix4fv(o.uModelViewMatrix, false, tm)

  // draw
  gl.drawArrays(GL_TRIANGLE_STRIP, /*offset*/ 0, /*vertexCount*/ 4)
}
