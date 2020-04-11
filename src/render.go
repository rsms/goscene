package main

import (
  "syscall/js"
  "github.com/go-gl/mathgl/mgl32"
)

type Renderer struct {
  gl         *GLContext
  pixelRatio float32
  width      uint32    // width of canvas in display points
  height     uint32    // height of canvas in display points
  resolution Vec2      // rendering size in pixels (may be smaller than size*pixelRatio)
  needResize bool      // true when size or pixelRatio has changed and a resize() call is needed
  pointerPos Vec2      // position of pointer in canvas space
  canvasel   js.Value
}


func NewRenderer(canvas js.Value, width, height uint32, pixelRatio float32) (*Renderer, error) {
  gl, err := NewGLContext(canvas)
  if err != nil {
    return nil, err
  }
  r := &Renderer{ gl: gl }
  r.setSize(width, height, pixelRatio)
  return r, nil
}


// setSize sets the size in display points
func (r *Renderer) setSize(width, height uint32, pixelRatio float32) {
  r.width = width
  r.height = height
  r.pixelRatio = pixelRatio

  // set canvas size and scale
  r.gl.setCanvasSize(width, height, pixelRatio)

  // read effective rendering resolution
  resx, resy := r.gl.drawingBufferSize()
  r.resolution[0] = float32(resx)
  r.resolution[1] = float32(resy)

  // mark r for needing resize updates like gl.viewport
  r.needResize = true
}


// Each iteration of the vertex shader receives the next value from the buffer assigned
// to that attribute.
// Uniforms are similar to JavaScript global variables. They stay the same value for all
// iterations of a shader.
const vertexShaderSrc = `
attribute vec4 aVertexPosition;

uniform mat4 uModelViewMatrix;
uniform mat4 uProjectionMatrix;

void main() {
  gl_Position = uProjectionMatrix * uModelViewMatrix * aVertexPosition;
}
`;

const fragmentShaderSrc = `
#ifdef GL_ES
precision mediump float;
#endif

uniform vec2  uResolution;  // Canvas size, viewport resolution (in pixels)
uniform vec2  uPointer;     // mouse pointer pixel coords. xy: current, zw: click
uniform float uTime;        // Time in seconds since load

void main() {
  vec2 st = gl_FragCoord.xy / uResolution;
  vec2 pointer = 1.0 - uPointer / uResolution;
  gl_FragColor = vec4(st.x * pointer.x, st.y * pointer.y, abs(sin(uTime)), 1.0);
}
`;


// program constants, for now (TODO move into GLProgram)
var (
  // vertex shader
  aVertexPosition   uint32
  uModelViewMatrix  GLUniform
  uProjectionMatrix GLUniform

  // fragment shader
  uResolution      GLUniform // vec2  // Canvas size, viewport resolution (in pixels)
  uPointer         GLUniform // vec2  // Pointer pixel coords
  uTime            GLUniform // float // Time in seconds since load

  positionBuffer   GLBuffer

  program *GLProgram
)


func (r *Renderer) init() {
  logf("Renderer.init resolution %v, pixelRatio %.1f", r.resolution, r.pixelRatio)

  // Create shader program
  var err error
  program, err = NewGLProgramSource(r.gl, vertexShaderSrc, fragmentShaderSrc)
  if err != nil {
    panic(err)
  }

  // attribLocations
  aVertexPosition, err = program.getAttribLocation("aVertexPosition")

  // uniformLocations
  uProjectionMatrix, err = program.getUniformLocation("uProjectionMatrix")
  uModelViewMatrix, err  = program.getUniformLocation("uModelViewMatrix")
  uResolution, err       = program.getUniformLocation("uResolution")
  uPointer, err          = program.getUniformLocation("uPointer")
  uTime, err             = program.getUniformLocation("uTime")
  if err != nil {
    panic(err)
  }

  positionBuffer = r.initBuffers()

  // track mouse pointer
  host.events.Listen(EVPointerMove, func (_ Event, xy ...uint32) {
    x := float32(xy[0]) * r.pixelRatio
    y := float32(xy[1]) * r.pixelRatio
    r.onPointerMove(x, y)
  })

  host.events.Listen(EVAnimationFrame, func (_ Event, _ ...uint32) {
    r.render(host.scenetime)
  })

  r.render(0.0)
}

func (r *Renderer) onPointerMove(x, y float32) {
  // logf("Renderer.onPointerMove %v, %v", x, y)
  // r.gl.useProgram(program)
  r.gl.uniformf(uPointer, x, y)
}


func (r *Renderer) initBuffers() (positionBuffer GLBuffer) {
  // Create a buffer for the square's positions
  gl := r.gl
  positionBuffer = gl.createBuffer()

  // Select the positionBuffer as the one to apply buffer
  // operations to from here out.
  gl.bindBuffer(ARRAY_BUFFER, positionBuffer)

  // Now create an array of positions for the square.
  positions := []float32{
    -1.0,  1.0,
     1.0,  1.0,
    -1.0, -1.0,
     1.0, -1.0,
  }

  // Now pass the list of positions into WebGL to build the
  // shape. We do this by creating a Float32Array from the
  // JavaScript array, then use it to fill the current buffer.
  gl.bufferDataF32(ARRAY_BUFFER, positions, STATIC_DRAW)

  return positionBuffer
}


func (r *Renderer) render(time float32) {
  // logf("Renderer.render")
  gl := r.gl
  width, height := r.resolution[0], r.resolution[1]

  if r.needResize {
    r.needResize = false
    r.gl.viewport(0, 0, uint32(width), uint32(height))
  }

  gl.clearColor(0.2, 0.25, 0.3, 1.0) // Clear to black, fully opaque
  gl.clearDepth(1.0)                 // Clear everything
  gl.enable(DEPTH_TEST)              // Enable depth testing
  gl.depthFunc(LEQUAL)               // Near things obscure far things

  // Clear the canvas before we start drawing on it.
  gl.clear(COLOR_BUFFER_BIT | DEPTH_BUFFER_BIT)

  // Create a perspective matrix, a special matrix that is
  // used to simulate the distortion of perspective in a camera.
  // Our field of view is 45 degrees, with a width/height
  // ratio that matches the display size of the canvas
  // and we only want to see objects between 0.1 units
  // and 100 units away from the camera.
  const fov   float32 = 45.0 * (PI / 180.0)   // in radians
  const zNear float32 = 0.1
  const zFar  float32 = 100.0
  aspect := width / height
  projectionMatrix := mgl32.Perspective(fov, aspect, zNear, zFar)

  // Set the drawing position to the "identity" point, which is
  // the center of the scene.
  modelViewMatrix := mgl32.Ident4()

  // Now move the drawing position a bit to where we want to
  // start drawing the square.
  // translate(&modelViewMatrix, 0.0, 0.0, -6)
  translateZ(&modelViewMatrix, -6)

  // Tell WebGL how to pull out the positions from the position buffer into the
  // aVertexPosition attribute.
  gl.bindBuffer(ARRAY_BUFFER, positionBuffer)
  gl.vertexAttribPointer(
    aVertexPosition,
    2,     // size; pull out 2 values per iteration
    FLOAT, // the data in the buffer is 32bit floats
    false, // don't normalize
    0,     // stride; use type and size value.
    0,     // offset; how many bytes inside the buffer to start from.
  )
  gl.enableVertexAttribArray(aVertexPosition)

  // Tell WebGL to use our program when drawing
  gl.useProgram(program)

  // Set the shader uniforms
  gl.uniformMatrix4fv(uProjectionMatrix, false, projectionMatrix)
  gl.uniformMatrix4fv(uModelViewMatrix, false, modelViewMatrix)

  gl.uniformf(uTime, time)
  gl.uniformf(uResolution, r.resolution[0], r.resolution[1])

  // draw
  offset := uint32(0)
  vertexCount := uint32(4)
  // gl.jsv.Call("drawArrays", TRIANGLE_STRIP, offset, vertexCount)
  gl.drawArrays(TRIANGLE_STRIP, offset, vertexCount)
}



// strconv.FormatFloat(f, 'E', -1, 64)
