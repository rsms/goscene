package main

import (
  "syscall/js"
)

type Renderer struct {
  gl         *GLContext
  pixelRatio float32
  width      uint32    // width of canvas in display points
  height     uint32    // height of canvas in display points
  resolution Vec2      // rendering size in pixels (may be smaller than size*pixelRatio)
  needResize bool      // true when size or pixelRatio has changed and a resize() call is needed
  pointer    Vec3      // position of pointer in canvas space. xy: pos, z: click
  canvasel   js.Value
  projectionMatrix Matrix4
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

  // Update projection matrix.
  // We use a perspective matrix; a special matrix that is used to simulate the
  // distortion of perspective in a camera.
  // Our field of view is 55 degrees, with a width/height ratio that matches the
  // display size of the canvas and we only want to see objects between 0.1 units
  // and 100 units away from the camera.
  const fov float32         = 45.0 * (PI / 180.0)   // in radians
  const zNear, zFar float32 = 0.1, 100.0
  aspect := r.resolution[0] / r.resolution[1]
  r.projectionMatrix = Matrix4Perspective(fov, aspect, zNear, zFar)

  // mark the renderer as needing resize, like calling gl.viewport()
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
`

const fragmentShaderSrc = `
#ifdef GL_ES
precision mediump float;
#endif

uniform vec2        uResolution;  // Canvas size, viewport resolution (in pixels)
uniform vec3        uPointer;     // mouse pointer pixel coords. xy: current, z: click
uniform highp float uTime;        // Time in seconds since load

void main() {
  vec2 st = gl_FragCoord.xy / uResolution;
  vec2 pointer = 1.0 - vec2(uPointer) / uResolution;
  gl_FragColor = vec4(st.x * pointer.x, st.y * pointer.y + uPointer.z, abs(sin(uTime)), 1.0);
}
`

const cubeVertexShaderSrc = `
attribute vec4 aVertexPosition;
//attribute vec4 aVertexColor;

uniform mat4 uModelViewMatrix;
uniform mat4 uProjectionMatrix;

varying lowp vec4 vColor;

void main(void) {
  gl_Position = uProjectionMatrix * uModelViewMatrix * aVertexPosition;
  vColor = vec4(aVertexPosition.xyz, 1.0);
}
`

const cubeFragmentShaderSrc = `
#ifdef GL_ES
precision mediump float;
#endif

uniform vec2 uResolution;
varying lowp vec4 vColor;

void main() {
  //vec2 xy = gl_FragCoord.xy / uResolution;
  //gl_FragColor = vec4(xy, 0.5, 1.0);
  gl_FragColor = vColor;
}
`


// program constants, for now (TODO move into GLProgram)
var (
  // vertex shader
  aVertexPosition   uint32
  uModelViewMatrix  GLUniform
  uProjectionMatrix GLUniform

  // fragment shader
  uResolution      GLUniform // vec2  // Canvas size, viewport resolution (in pixels)
  uPointer         GLUniform // vec3  // Pointer pixel coords
  uTime            GLUniform // float // Time in seconds since load

  // cube shader
  cube_uProjectionMatrix GLUniform
  cube_uResolution GLUniform
)

var (
  program *GLProgram
  planeobj *GLPlane
  planeobj2 *GLPlane

  cubeProgram *GLProgram
  cubeobj1 *GLCube
  cubeobj2 *GLCube
)


func (r *Renderer) init() {
  logf("Renderer.init resolution %v, pixelRatio %.1f", r.resolution, r.pixelRatio)

  r.initShaders()
  // r.initBuffers()
  // r.initTextures()

  // make planes
  var err error
  planeobj, err = NewGLPlane(program)
  planeobj2, err = NewGLPlane(program)
  planeobj2.absoluteTransform.Translate(0.4, 0.4, -1.0).RotateZ(1.0)

  // make cubes
  cubeobj1, err = NewGLCube(cubeProgram)
  cubeobj1.absoluteTransform.Translate(0, 0, -5.5)

  cubeobj2, err = NewGLCube(cubeProgram)
  // cubeobj2.absoluteTransform.ScaleMut(1.5, 0.5, 0.5)
  cubeobj2.absoluteTransform.Translate(0.5, 0.5, -5.5)
  if err != nil {
    panic(err)
  }
}

func (r *Renderer) start() {
  // update pointer
  onPointerEvent := func (_ Event, data ...uint32) {
    r.pointer[0] = host.pointer.x * r.pixelRatio
    r.pointer[1] = host.pointer.y * r.pixelRatio
    if host.pointer.buttons == 0 {
      r.pointer[2] = 0.0
    } else {
      r.pointer[2] = 1.0
    }
  }
  host.events.Listen(EVPointerMove, onPointerEvent)
  host.events.Listen(EVPointerDown, onPointerEvent)
  host.events.Listen(EVPointerUp, onPointerEvent)

  // render on each frame
  host.events.Listen(EVAnimationFrame, func (_ Event, _ ...uint32) {
    host.UpdateAnimationStats()
    r.render(float32(host.scenetime))
  })
  r.render(0.0)
}


func (r *Renderer) initShaders() {
  // Create shader program
  var err error
  program, err = NewGLProgramSource(r.gl, vertexShaderSrc, fragmentShaderSrc)
  if err != nil {
    panic(err)
  }

  cubeProgram, err = NewGLProgramSource(r.gl, cubeVertexShaderSrc, cubeFragmentShaderSrc)
  if err != nil {
    panic(err)
  }

  cube_uProjectionMatrix, err = cubeProgram.getUniformLocation("uProjectionMatrix")
  cube_uResolution, err = cubeProgram.getUniformLocation("uResolution")

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
}


func (r *Renderer) render(time float32) {
  // logf("Renderer.render")
  gl := r.gl
  width, height := r.resolution[0], r.resolution[1]

  if r.needResize {
    r.needResize = false
    r.gl.viewport(0, 0, uint32(width), uint32(height))
  }

  gl.clearColor(0.2, 0.25, 0.3, 1.0) // Clear to color, fully opaque
  gl.clearDepth(1.0)                 // Clear everything
  gl.enable(GL_DEPTH_TEST)              // Enable depth testing
  gl.depthFunc(GL_LEQUAL)               // Near things obscure far things

  // Clear the canvas before we start drawing on it.
  gl.clear(GL_COLOR_BUFFER_BIT | GL_DEPTH_BUFFER_BIT)

  // TODO: figure out how to setup the projection matrix and uniforms for whatever programs
  // are being used by drawables automatically and efficiently.
  // I.e. plane and cube uses different shader programs and need to have their.
  //
  // For example, we could do something like this:
  //   r.useProgram(p):
  //     gl.useProgram(p):  // returns true when program was not already active
  //       if currframe != progInitFrame[p]:
  //         progInitFrame[p] = currframe
  //         gl.uniformMatrix4fv(uProjectionMatrix, false, r.projectionMatrix)
  //         gl.uniformf(uResolution, r.resolution[0], r.resolution[1])
  //
  //   drawable.Draw(r):
  //     r.useProgram(drawable.program)
  //
  // For now, just setup all programs manually, hard coded.

  // Tell WebGL to use our program for planes when drawing
  gl.useProgram(program)
  gl.uniformMatrix4fv(uProjectionMatrix, false, r.projectionMatrix)
  gl.uniformf(uResolution, r.resolution[0], r.resolution[1])
  gl.uniformf(uTime, time)
  gl.uniformf(uPointer, r.pointer[:]...)

  // planeobj.Draw(r)
  // planeobj2.Draw(r)

  // Tell WebGL to use our program for meshes when drawing
  gl.useProgram(cubeProgram)
  gl.uniformMatrix4fv(cube_uProjectionMatrix, false, r.projectionMatrix)
  gl.uniformf(cube_uResolution, r.resolution[0], r.resolution[1])

  cubeobj1.Draw(r)
  // cubeobj2.Draw(r)

  gl.useProgram(program)
  // planeobj2.Draw(r)
}
