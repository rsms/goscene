package main

import (
  "syscall/js"
  "sync/atomic"
  "unsafe"
  "runtime"
)

// See https://www.khronos.org/registry/OpenGL/api/GLES/gl.h

type GLenum     = uint32
type GLboolean  = uint8
type GLbitfield = uint32
type GLbyte     = int8
type GLshort    = int16
type GLint      = int32
type GLsizei    = int32
type GLubyte    = uint8
type GLushort   = uint16
type GLuint     = uint32
type GLfloat    = float32
type GLclampf   = float32
type GLintptr   = int32
type GLsizeiptr = int32
type GLfixed    = int32
type GLclampx   = int32

type GLBuffer = js.Value
type GLUniform = js.Value

type GLContext struct {
  jsv          js.Value
  activeProgId uintptr  // tracks GLProgram.id to avoid redundant host calls
}

func NewGLContext(canvasHtmlElement js.Value) (*GLContext, error) {
  jsv := host.jsv.Call("getContext", canvasHtmlElement, "webgl")
  if jsv.Type() != js.TypeObject {
    return nil, errorf(`getContext("webgl") failed`)
  }
  gl := &GLContext{ jsv: jsv }
  return gl, nil
}

func (gl *GLContext) drawingBufferSize() (width, height uint32) {
  return hostcall_j_u32x2(HGLdrawingBufferSize, gl.jsv)
}

func (gl *GLContext) canvasSize() (width, height uint32) {
  return hostcall_j_u32x2(HGLcanvasSize, gl.jsv)
}

func (gl *GLContext) setCanvasSize(width, height uint32, pixelRatio float32) {
  canvas := gl.jsv.Get("canvas")
  // translate size from display points to pixels (intentionally floor() by truncation)
  width  = uint32(float32(width) * pixelRatio)
  height = uint32(float32(height) * pixelRatio)
  canvas.Get("style").Set("zoom", 1.0 / pixelRatio)
  canvas.Set("width",  width)
  canvas.Set("height", height)
}

func (gl *GLContext) viewport(x, y int32, width, height uint32) {
  hostcall_jvi32_(HGLviewport, gl.jsv, x, y, int32(width), int32(height))
}

func (gl *GLContext) clear(mask uint32) {
  hostcall_ju32_(HGLclear, gl.jsv, mask)
}

func (gl *GLContext) clearColor(r, g, b, a float32) {
  hostcall_jvf32_(HGLclearColor, gl.jsv, r, g, b, a)
}

func (gl *GLContext) clearDepth(d float32) {
  hostcall_jf32_(HGLclearDepth, gl.jsv, d)
}

func (gl *GLContext) enable(cap uint32) {
  hostcall_ju32_(HGLenable, gl.jsv, cap)
}

func (gl *GLContext) depthFunc(funcid uint32) {
  hostcall_ju32_(HGLdepthFunc, gl.jsv, funcid)
}

func (gl *GLContext) createBuffer() GLBuffer {
  return gl.jsv.Call("createBuffer")
}

func (gl *GLContext) deleteBuffer(b GLBuffer) {
  gl.jsv.Call("deleteBuffer", b)
}

func (gl *GLContext) bindBuffer(target uint32, buffer GLBuffer) {
  hostcall_ju32j_(HGLbindBuffer, gl.jsv, target, buffer)
}

func (gl *GLContext) bufferDataI8(target uint32, data []int8, usage uint32) {
  ptr := uint32(uintptr(unsafe.Pointer(&data[0]))) // take address+offset of underlying array
  hostcall_jvu32_(HGLbufferData, gl.jsv, target, usage, ptr, uint32(len(data)))
}
func (gl *GLContext) bufferDataI16(target uint32, data []int16, usage uint32) {
  ptr := uint32(uintptr(unsafe.Pointer(&data[0])))
  hostcall_jvu32_(HGLbufferData, gl.jsv, target, usage, ptr, uint32(len(data) * 2))
}
func (gl *GLContext) bufferDataU16(target uint32, data []uint16, usage uint32) {
  ptr := uint32(uintptr(unsafe.Pointer(&data[0])))
  hostcall_jvu32_(HGLbufferData, gl.jsv, target, usage, ptr, uint32(len(data) * 2))
}
func (gl *GLContext) bufferDataI32(target uint32, data []int32, usage uint32) {
  ptr := uint32(uintptr(unsafe.Pointer(&data[0])))
  hostcall_jvu32_(HGLbufferData, gl.jsv, target, usage, ptr, uint32(len(data) * 4))
}
func (gl *GLContext) bufferDataF32(target uint32, data []float32, usage uint32) {
  ptr := uint32(uintptr(unsafe.Pointer(&data[0])))
  hostcall_jvu32_(HGLbufferData, gl.jsv, target, usage, ptr, uint32(len(data) * 4))
}
func (gl *GLContext) bufferDataF64(target uint32, data []float64, usage uint32) {
  ptr := uint32(uintptr(unsafe.Pointer(&data[0])))
  hostcall_jvu32_(HGLbufferData, gl.jsv, target, usage, ptr, uint32(len(data) * 8))
}

func (gl *GLContext) uniformMatrix2fv(location GLUniform, transpose bool, value [4]float32) {
  transpose_ := uint32(0) ; if transpose { transpose_ = 1 }
  ptr := uint32(uintptr(unsafe.Pointer(&value[0])))
  hostcall_jx2vu32_(HGLuniformMatrix2fv, gl.jsv, location, transpose_, ptr)
}
func (gl *GLContext) uniformMatrix3fv(location GLUniform, transpose bool, value [9]float32) {
  transpose_ := uint32(0) ; if transpose { transpose_ = 1 }
  ptr := uint32(uintptr(unsafe.Pointer(&value[0])))
  hostcall_jx2vu32_(HGLuniformMatrix3fv, gl.jsv, location, transpose_, ptr)
}
func (gl *GLContext) uniformMatrix4fv(location GLUniform, transpose bool, value [16]float32) {
  transpose_ := uint32(0) ; if transpose { transpose_ = 1 }
  ptr := uint32(uintptr(unsafe.Pointer(&value[0])))
  hostcall_jx2vu32_(HGLuniformMatrix4fv, gl.jsv, location, transpose_, ptr)
}

func (gl *GLContext) uniformf(location GLUniform, value ...float32) {
  hostcall_jx2vf32_(HGLuniformvf, gl.jsv, location, value...)
}

func (gl *GLContext) uniformi(location GLUniform, value ...int32) {
  hostcall_jx2vi32_(HGLuniformvi, gl.jsv, location, value...)
}

func (gl *GLContext) vertexAttribPointer(
  index, size uint32, typ GLenum, normalized bool, stride, offset uint32) {
  normalized_ := uint32(0)
  if normalized {
    normalized_ = 1
  }
  hostcall_jvu32_(HGLvertexAttribPointer, gl.jsv, index, size, typ, normalized_, stride, offset)
}

func (gl *GLContext) enableVertexAttribArray(index uint32) {
  hostcall_ju32_(HGLenableVertexAttribArray, gl.jsv, index)
}

func (gl *GLContext) useProgram(p *GLProgram) bool {
  if gl.activeProgId == p.id {
    return false  // program already active
  }
  gl.activeProgId = p.id
  hostcall_jx2_(HGLuseProgram, gl.jsv, p.jsv)
  return true
}

func (gl *GLContext) drawArrays(mode, first, count uint32) {
  hostcall_ju32x3_(HGLdrawArrays, gl.jsv, mode, first, count)
}

// drawElements(mode: GLenum, count: GLsizei, type: GLenum, offset: GLintptr): void;
func (gl *GLContext) drawElements(mode, count, kind, offset uint32) {
  hostcall_ju32x4_(HGLdrawElements, gl.jsv, mode, count, kind, offset)
}


// --------------------------------

func (gl *GLContext) GetVertexBuffer(ref GLVertexDataRef) *GLBuf {
  gl.initGLVertexData()
  return &glVertexDataMap[ref].vertexBuf
}

func (gl *GLContext) GetIndexBuffer(ref GLVertexDataRef) *GLBuf {
  gl.initGLVertexData()
  return &glVertexDataMap[ref].indexBuf
}

type GLBuf struct {
  pos    GLBuffer // underlying buffer position
  offset uint32   // offset into underlying buffer
}

// --------------------------------

// GLVertexData is a registry and source of buffer data.
//
// Intended use:
// 1. One-time initializer calls GLVertexData() with its data and stores the
//    returned GLVertexDataRef.
// 2. Instance constructor calls gl.GetBuf(GLVertexDataRef) to retrieve a GLBuf
//    handle for the actual underlying buffer.
// 3. Draw function passes GLBuf.pos to gl.bindBuffer()
//    and GLBuf.offset to e.g. vertexAttribPointer().
//
type GLVertexDataRef uint32
type glVertexData struct {
  usage uint32          // e.g. GL_STATIC_DRAW
  vertexData []float32  // input vertex data
  indexData  []uint16   // input index data

  vertexBuf  GLBuf      // populated by initGLVertexData()
  indexBuf   GLBuf
}

var glVertexDataDirty bool
var glVertexDataMap []glVertexData

func GLVertexData(usage uint32, vertexData []float32, index ...uint16) GLVertexDataRef {
  ref := GLVertexDataRef(len(glVertexDataMap))
  glVertexDataMap = append(glVertexDataMap, glVertexData{
    usage: usage,
    vertexData: vertexData,
    indexData: index,
  })
  glVertexDataDirty = true
  return ref
}

func (gl *GLContext) initGLVertexData() {
  if !glVertexDataDirty {
    return
  }

  // sort vertexData into categories of usage
  groups := make(map[uint32][]*glVertexData, len(glVertexDataMap)/2)
  for i := 0; i < len(glVertexDataMap); i++ {
    d := &glVertexDataMap[i]
    groups[d.usage] = append(groups[d.usage], d)
  }

  for _, v := range groups {
    // calculate size of buffer
    vertexDataSize := 0
    indexDataSize := 0
    for _, d := range v {
      vertexDataSize += len(d.vertexData)
      indexDataSize += len(d.indexData)
    }

    // allocate GL buffer to get position
    vertexBuf := gl.createBuffer()
    indexBuf := gl.createBuffer()

    // build one contiguous array of all data and update glVertexData.GLBuf
    vertexData := make([]float32, vertexDataSize)
    vertexOffset := uint32(0)
    indexData := make([]uint16, indexDataSize)
    indexOffset := uint32(0)

    for _, d := range v {
      d.vertexBuf.offset = vertexOffset * 4  // offset is in bytes; sizeof(float32)=4
      d.vertexBuf.pos = vertexBuf
      copy(vertexData[vertexOffset:], d.vertexData)
      vertexOffset += uint32(len(d.vertexData))

      if len(d.indexData) > 0 {
        d.indexBuf.offset = indexOffset * 4  // offset is in bytes; sizeof(float32)=4
        d.indexBuf.pos = indexBuf
        copy(indexData[indexOffset:], d.indexData)
        indexOffset += uint32(len(d.indexData))
      }
    }

    // TODO: compress vertexData slab:
    //
    // 1  for each glVertexData d:
    // 2    let offset be the first match of d.vertexData in vertexData
    // 3    if offset != sparse_offset then:
    // 4      splice out sparse_offset+len from vertexData
    //
    // Note: #2 is essentially substring search
    //

    // copy vertexData to GL buffer
    gl.bindBuffer(GL_ARRAY_BUFFER, vertexBuf)
    gl.bufferDataF32(GL_ARRAY_BUFFER, vertexData, v[0].usage)

    // copy indexData to GL buffer
    if indexDataSize > 0 {
      gl.bindBuffer(GL_ELEMENT_ARRAY_BUFFER, indexBuf)
      gl.bufferDataU16(GL_ELEMENT_ARRAY_BUFFER, indexData, v[0].usage)
    }
  }
}


// --------------------------------------------------------------------------------------


type GLShader struct {
  gl   *GLContext
  jsv  js.Value
  kind GLenum  // VERTEX_SHADER or FRAGMENT_SHADER
}

func NewGLShader(gl *GLContext, kind GLenum, source string) (*GLShader, error) {
  jsv := gl.jsv.Call("createShader", kind)
  gl.jsv.Call("shaderSource", jsv, source)  // Send the source to the shader object
  gl.jsv.Call("compileShader", jsv)  // Compile the shader program
  if (!gl.jsv.Call("getShaderParameter", jsv, GL_COMPILE_STATUS).Bool()) {
    info := gl.jsv.Call("getShaderInfoLog", jsv)
    gl.jsv.Call("deleteShader", jsv)
    return nil, errorf("shader failed to compile: %+v", info)
  }
  return &GLShader{ gl: gl, jsv: jsv, kind: kind  }, nil
}

func (s *GLShader) Free() {
  s.gl.jsv.Call("deleteShader", s.jsv)
  s.jsv = js.Null()
}


// --------------------------------------------------------------------------------------

var glNextID uintptr = 0

func glGenID() uintptr {
  return atomic.AddUintptr(&glNextID, 1)
}

// --------------------------------------------------------------------------------------

type GLProgram struct {
  id      uintptr
  gl      *GLContext
  jsv     js.Value
  shaders []*GLShader
}

func NewGLProgram(gl *GLContext, shaders... *GLShader) (*GLProgram, error) {
  jsv := gl.jsv.Call("createProgram")
  for _, shader := range shaders {
    gl.jsv.Call("attachShader", jsv, shader.jsv)
  }
  gl.jsv.Call("linkProgram", jsv)
  if (!gl.jsv.Call("getProgramParameter", jsv, GL_LINK_STATUS).Bool()) {
    info := gl.jsv.Call("getProgramInfoLog", jsv)
    gl.jsv.Call("deleteProgram", jsv)
    return nil, errorf("shader failed to compile: %+v", info)
  }
  p := &GLProgram{ id: glGenID(), gl: gl, jsv: jsv, shaders: shaders }
  runtime.SetFinalizer(p, finalizeGLProgram)
  return p, nil
}

func finalizeGLProgram(p *GLProgram) {
  // Called when the object is marked for garbage collection.
  // See https://golang.org/pkg/runtime/#SetFinalizer for details on semantics.
  p.gl.jsv.Call("deleteProgram", p.jsv)
  p.jsv = js.Null()
}

func NewGLProgramSource(gl *GLContext, vSource, hSource string) (*GLProgram, error) {
  vertextShader, err := NewGLShader(gl, GL_VERTEX_SHADER, vSource)
  if err != nil {
    return nil, err
  }
  fragmentShader, err := NewGLShader(gl, GL_FRAGMENT_SHADER, hSource)
  if err != nil {
    return nil, err
  }
  return NewGLProgram(gl, vertextShader, fragmentShader)
}

func (p *GLProgram) getUniformLocation(name string) (u GLUniform, err error) {
  u = p.gl.jsv.Call("getUniformLocation", p.jsv, name)
  if (u.Type() != js.TypeObject) {
    err = errorf("uniform %#v not found", name)
  }
  return
}

func (p *GLProgram) getAttribLocation(name string) (location uint32, err error) {
  v := p.gl.jsv.Call("getAttribLocation", p.jsv, name)
  if (v.Type() == js.TypeNumber) {
    location = uint32(v.Int())
  } else {
    err = errorf("attribute %#v not found", name)
  }
  return
}




/*  ------------------------------------------------------------------------------
Reminder of this source file are GL constants, generated from the following
javascript snippet, run in a browser:

console.log(((o)=>
  Object.keys(o)
    .filter(k => { let c = k.charCodeAt(); return c >= 0x41 && c <= 0x5A })
    .sort((a, b) => {
      const s = "TEXTURE"  // e.g. TEXTURE2, TEXTURE5, TEXTURE21 etc.
      if (a.length > s.length && a.length <= s.length + 2 && a.startsWith(s)) {
        let an = parseInt(a.substr(s.length))
        let bn = parseInt(b.substr(s.length))
        if (!isNaN(an) && !isNaN(bn)) {
          return an < bn ? -1 : bn < an ? 1 : 0
        }
      }
      return a < b ? -1 : b < a ? 1 : 0
    })
    .map(k => `  ${k} = GLenum(${o[k]})`)
    .join("\n")
)(WebGLRenderingContext.prototype))

*/
var (
  GL_ACTIVE_ATTRIBUTES = GLenum(35721)
  GL_ACTIVE_TEXTURE = GLenum(34016)
  GL_ACTIVE_UNIFORMS = GLenum(35718)
  GL_ALIASED_LINE_WIDTH_RANGE = GLenum(33902)
  GL_ALIASED_POINT_SIZE_RANGE = GLenum(33901)
  GL_ALPHA = GLenum(6406)
  GL_ALPHA_BITS = GLenum(3413)
  GL_ALWAYS = GLenum(519)
  GL_ARRAY_BUFFER = GLenum(34962)
  GL_ARRAY_BUFFER_BINDING = GLenum(34964)
  GL_ATTACHED_SHADERS = GLenum(35717)
  GL_BACK = GLenum(1029)
  GL_BLEND = GLenum(3042)
  GL_BLEND_COLOR = GLenum(32773)
  GL_BLEND_DST_ALPHA = GLenum(32970)
  GL_BLEND_DST_RGB = GLenum(32968)
  GL_BLEND_EQUATION = GLenum(32777)
  GL_BLEND_EQUATION_ALPHA = GLenum(34877)
  GL_BLEND_EQUATION_RGB = GLenum(32777)
  GL_BLEND_SRC_ALPHA = GLenum(32971)
  GL_BLEND_SRC_RGB = GLenum(32969)
  GL_BLUE_BITS = GLenum(3412)
  GL_BOOL = GLenum(35670)
  GL_BOOL_VEC2 = GLenum(35671)
  GL_BOOL_VEC3 = GLenum(35672)
  GL_BOOL_VEC4 = GLenum(35673)
  GL_BROWSER_DEFAULT_WEBGL = GLenum(37444)
  GL_BUFFER_SIZE = GLenum(34660)
  GL_BUFFER_USAGE = GLenum(34661)
  GL_BYTE = GLenum(5120)
  GL_CCW = GLenum(2305)
  GL_CLAMP_TO_EDGE = GLenum(33071)
  GL_COLOR_ATTACHMENT0 = GLenum(36064)
  GL_COLOR_BUFFER_BIT = GLenum(16384)
  GL_COLOR_CLEAR_VALUE = GLenum(3106)
  GL_COLOR_WRITEMASK = GLenum(3107)
  GL_COMPILE_STATUS = GLenum(35713)
  GL_COMPRESSED_TEXTURE_FORMATS = GLenum(34467)
  GL_CONSTANT_ALPHA = GLenum(32771)
  GL_CONSTANT_COLOR = GLenum(32769)
  GL_CONTEXT_LOST_WEBGL = GLenum(37442)
  GL_CULL_FACE = GLenum(2884)
  GL_CULL_FACE_MODE = GLenum(2885)
  GL_CURRENT_PROGRAM = GLenum(35725)
  GL_CURRENT_VERTEX_ATTRIB = GLenum(34342)
  GL_CW = GLenum(2304)
  GL_DECR = GLenum(7683)
  GL_DECR_WRAP = GLenum(34056)
  GL_DELETE_STATUS = GLenum(35712)
  GL_DEPTH_ATTACHMENT = GLenum(36096)
  GL_DEPTH_BITS = GLenum(3414)
  GL_DEPTH_BUFFER_BIT = GLenum(256)
  GL_DEPTH_CLEAR_VALUE = GLenum(2931)
  GL_DEPTH_COMPONENT = GLenum(6402)
  GL_DEPTH_COMPONENT16 = GLenum(33189)
  GL_DEPTH_FUNC = GLenum(2932)
  GL_DEPTH_RANGE = GLenum(2928)
  GL_DEPTH_STENCIL = GLenum(34041)
  GL_DEPTH_STENCIL_ATTACHMENT = GLenum(33306)
  GL_DEPTH_TEST = GLenum(2929)
  GL_DEPTH_WRITEMASK = GLenum(2930)
  GL_DITHER = GLenum(3024)
  GL_DONT_CARE = GLenum(4352)
  GL_DST_ALPHA = GLenum(772)
  GL_DST_COLOR = GLenum(774)
  GL_DYNAMIC_DRAW = GLenum(35048)
  GL_ELEMENT_ARRAY_BUFFER = GLenum(34963)
  GL_ELEMENT_ARRAY_BUFFER_BINDING = GLenum(34965)
  GL_EQUAL = GLenum(514)
  GL_FASTEST = GLenum(4353)
  GL_FLOAT = GLenum(5126)
  GL_FLOAT_MAT2 = GLenum(35674)
  GL_FLOAT_MAT3 = GLenum(35675)
  GL_FLOAT_MAT4 = GLenum(35676)
  GL_FLOAT_VEC2 = GLenum(35664)
  GL_FLOAT_VEC3 = GLenum(35665)
  GL_FLOAT_VEC4 = GLenum(35666)
  GL_FRAGMENT_SHADER = GLenum(35632)
  GL_FRAMEBUFFER = GLenum(36160)
  GL_FRAMEBUFFER_ATTACHMENT_OBJECT_NAME = GLenum(36049)
  GL_FRAMEBUFFER_ATTACHMENT_OBJECT_TYPE = GLenum(36048)
  GL_FRAMEBUFFER_ATTACHMENT_TEXTURE_CUBE_MAP_FACE = GLenum(36051)
  GL_FRAMEBUFFER_ATTACHMENT_TEXTURE_LEVEL = GLenum(36050)
  GL_FRAMEBUFFER_BINDING = GLenum(36006)
  GL_FRAMEBUFFER_COMPLETE = GLenum(36053)
  GL_FRAMEBUFFER_INCOMPLETE_ATTACHMENT = GLenum(36054)
  GL_FRAMEBUFFER_INCOMPLETE_DIMENSIONS = GLenum(36057)
  GL_FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT = GLenum(36055)
  GL_FRAMEBUFFER_UNSUPPORTED = GLenum(36061)
  GL_FRONT = GLenum(1028)
  GL_FRONT_AND_BACK = GLenum(1032)
  GL_FRONT_FACE = GLenum(2886)
  GL_FUNC_ADD = GLenum(32774)
  GL_FUNC_REVERSE_SUBTRACT = GLenum(32779)
  GL_FUNC_SUBTRACT = GLenum(32778)
  GL_GENERATE_MIPMAP_HINT = GLenum(33170)
  GL_GEQUAL = GLenum(518)
  GL_GREATER = GLenum(516)
  GL_GREEN_BITS = GLenum(3411)
  GL_HIGH_FLOAT = GLenum(36338)
  GL_HIGH_INT = GLenum(36341)
  GL_IMPLEMENTATION_COLOR_READ_FORMAT = GLenum(35739)
  GL_IMPLEMENTATION_COLOR_READ_TYPE = GLenum(35738)
  GL_INCR = GLenum(7682)
  GL_INCR_WRAP = GLenum(34055)
  GL_INT = GLenum(5124)
  GL_INT_VEC2 = GLenum(35667)
  GL_INT_VEC3 = GLenum(35668)
  GL_INT_VEC4 = GLenum(35669)
  GL_INVALID_ENUM = GLenum(1280)
  GL_INVALID_FRAMEBUFFER_OPERATION = GLenum(1286)
  GL_INVALID_OPERATION = GLenum(1282)
  GL_INVALID_VALUE = GLenum(1281)
  GL_INVERT = GLenum(5386)
  GL_KEEP = GLenum(7680)
  GL_LEQUAL = GLenum(515)
  GL_LESS = GLenum(513)
  GL_LINEAR = GLenum(9729)
  GL_LINEAR_MIPMAP_LINEAR = GLenum(9987)
  GL_LINEAR_MIPMAP_NEAREST = GLenum(9985)
  GL_LINES = GLenum(1)
  GL_LINE_LOOP = GLenum(2)
  GL_LINE_STRIP = GLenum(3)
  GL_LINE_WIDTH = GLenum(2849)
  GL_LINK_STATUS = GLenum(35714)
  GL_LOW_FLOAT = GLenum(36336)
  GL_LOW_INT = GLenum(36339)
  GL_LUMINANCE = GLenum(6409)
  GL_LUMINANCE_ALPHA = GLenum(6410)
  GL_MAX_COMBINED_TEXTURE_IMAGE_UNITS = GLenum(35661)
  GL_MAX_CUBE_MAP_TEXTURE_SIZE = GLenum(34076)
  GL_MAX_FRAGMENT_UNIFORM_VECTORS = GLenum(36349)
  GL_MAX_RENDERBUFFER_SIZE = GLenum(34024)
  GL_MAX_TEXTURE_IMAGE_UNITS = GLenum(34930)
  GL_MAX_TEXTURE_SIZE = GLenum(3379)
  GL_MAX_VARYING_VECTORS = GLenum(36348)
  GL_MAX_VERTEX_ATTRIBS = GLenum(34921)
  GL_MAX_VERTEX_TEXTURE_IMAGE_UNITS = GLenum(35660)
  GL_MAX_VERTEX_UNIFORM_VECTORS = GLenum(36347)
  GL_MAX_VIEWPORT_DIMS = GLenum(3386)
  GL_MEDIUM_FLOAT = GLenum(36337)
  GL_MEDIUM_INT = GLenum(36340)
  GL_MIRRORED_REPEAT = GLenum(33648)
  GL_NEAREST = GLenum(9728)
  GL_NEAREST_MIPMAP_LINEAR = GLenum(9986)
  GL_NEAREST_MIPMAP_NEAREST = GLenum(9984)
  GL_NEVER = GLenum(512)
  GL_NICEST = GLenum(4354)
  GL_NONE = GLenum(0)
  GL_NOTEQUAL = GLenum(517)
  GL_NO_ERROR = GLenum(0)
  GL_ONE = GLenum(1)
  GL_ONE_MINUS_CONSTANT_ALPHA = GLenum(32772)
  GL_ONE_MINUS_CONSTANT_COLOR = GLenum(32770)
  GL_ONE_MINUS_DST_ALPHA = GLenum(773)
  GL_ONE_MINUS_DST_COLOR = GLenum(775)
  GL_ONE_MINUS_SRC_ALPHA = GLenum(771)
  GL_ONE_MINUS_SRC_COLOR = GLenum(769)
  GL_OUT_OF_MEMORY = GLenum(1285)
  GL_PACK_ALIGNMENT = GLenum(3333)
  GL_POINTS = GLenum(0)
  GL_POLYGON_OFFSET_FACTOR = GLenum(32824)
  GL_POLYGON_OFFSET_FILL = GLenum(32823)
  GL_POLYGON_OFFSET_UNITS = GLenum(10752)
  GL_RED_BITS = GLenum(3410)
  GL_RENDERBUFFER = GLenum(36161)
  GL_RENDERBUFFER_ALPHA_SIZE = GLenum(36179)
  GL_RENDERBUFFER_BINDING = GLenum(36007)
  GL_RENDERBUFFER_BLUE_SIZE = GLenum(36178)
  GL_RENDERBUFFER_DEPTH_SIZE = GLenum(36180)
  GL_RENDERBUFFER_GREEN_SIZE = GLenum(36177)
  GL_RENDERBUFFER_HEIGHT = GLenum(36163)
  GL_RENDERBUFFER_INTERNAL_FORMAT = GLenum(36164)
  GL_RENDERBUFFER_RED_SIZE = GLenum(36176)
  GL_RENDERBUFFER_STENCIL_SIZE = GLenum(36181)
  GL_RENDERBUFFER_WIDTH = GLenum(36162)
  GL_RENDERER = GLenum(7937)
  GL_REPEAT = GLenum(10497)
  GL_REPLACE = GLenum(7681)
  GL_RGB = GLenum(6407)
  GL_RGB565 = GLenum(36194)
  GL_RGB5_A1 = GLenum(32855)
  GL_RGBA = GLenum(6408)
  GL_RGBA4 = GLenum(32854)
  GL_SAMPLER_2D = GLenum(35678)
  GL_SAMPLER_CUBE = GLenum(35680)
  GL_SAMPLES = GLenum(32937)
  GL_SAMPLE_ALPHA_TO_COVERAGE = GLenum(32926)
  GL_SAMPLE_BUFFERS = GLenum(32936)
  GL_SAMPLE_COVERAGE = GLenum(32928)
  GL_SAMPLE_COVERAGE_INVERT = GLenum(32939)
  GL_SAMPLE_COVERAGE_VALUE = GLenum(32938)
  GL_SCISSOR_BOX = GLenum(3088)
  GL_SCISSOR_TEST = GLenum(3089)
  GL_SHADER_TYPE = GLenum(35663)
  GL_SHADING_LANGUAGE_VERSION = GLenum(35724)
  GL_SHORT = GLenum(5122)
  GL_SRC_ALPHA = GLenum(770)
  GL_SRC_ALPHA_SATURATE = GLenum(776)
  GL_SRC_COLOR = GLenum(768)
  GL_STATIC_DRAW = GLenum(35044)
  GL_STENCIL_ATTACHMENT = GLenum(36128)
  GL_STENCIL_BACK_FAIL = GLenum(34817)
  GL_STENCIL_BACK_FUNC = GLenum(34816)
  GL_STENCIL_BACK_PASS_DEPTH_FAIL = GLenum(34818)
  GL_STENCIL_BACK_PASS_DEPTH_PASS = GLenum(34819)
  GL_STENCIL_BACK_REF = GLenum(36003)
  GL_STENCIL_BACK_VALUE_MASK = GLenum(36004)
  GL_STENCIL_BACK_WRITEMASK = GLenum(36005)
  GL_STENCIL_BITS = GLenum(3415)
  GL_STENCIL_BUFFER_BIT = GLenum(1024)
  GL_STENCIL_CLEAR_VALUE = GLenum(2961)
  GL_STENCIL_FAIL = GLenum(2964)
  GL_STENCIL_FUNC = GLenum(2962)
  GL_STENCIL_INDEX8 = GLenum(36168)
  GL_STENCIL_PASS_DEPTH_FAIL = GLenum(2965)
  GL_STENCIL_PASS_DEPTH_PASS = GLenum(2966)
  GL_STENCIL_REF = GLenum(2967)
  GL_STENCIL_TEST = GLenum(2960)
  GL_STENCIL_VALUE_MASK = GLenum(2963)
  GL_STENCIL_WRITEMASK = GLenum(2968)
  GL_STREAM_DRAW = GLenum(35040)
  GL_SUBPIXEL_BITS = GLenum(3408)
  GL_TEXTURE = GLenum(5890)
  GL_TEXTURE0 = GLenum(33984)
  GL_TEXTURE1 = GLenum(33985)
  GL_TEXTURE2 = GLenum(33986)
  GL_TEXTURE3 = GLenum(33987)
  GL_TEXTURE4 = GLenum(33988)
  GL_TEXTURE5 = GLenum(33989)
  GL_TEXTURE6 = GLenum(33990)
  GL_TEXTURE7 = GLenum(33991)
  GL_TEXTURE8 = GLenum(33992)
  GL_TEXTURE9 = GLenum(33993)
  GL_TEXTURE10 = GLenum(33994)
  GL_TEXTURE11 = GLenum(33995)
  GL_TEXTURE12 = GLenum(33996)
  GL_TEXTURE13 = GLenum(33997)
  GL_TEXTURE14 = GLenum(33998)
  GL_TEXTURE15 = GLenum(33999)
  GL_TEXTURE16 = GLenum(34000)
  GL_TEXTURE17 = GLenum(34001)
  GL_TEXTURE18 = GLenum(34002)
  GL_TEXTURE19 = GLenum(34003)
  GL_TEXTURE20 = GLenum(34004)
  GL_TEXTURE21 = GLenum(34005)
  GL_TEXTURE22 = GLenum(34006)
  GL_TEXTURE23 = GLenum(34007)
  GL_TEXTURE24 = GLenum(34008)
  GL_TEXTURE25 = GLenum(34009)
  GL_TEXTURE26 = GLenum(34010)
  GL_TEXTURE27 = GLenum(34011)
  GL_TEXTURE28 = GLenum(34012)
  GL_TEXTURE29 = GLenum(34013)
  GL_TEXTURE30 = GLenum(34014)
  GL_TEXTURE31 = GLenum(34015)
  GL_TEXTURE_2D = GLenum(3553)
  GL_TEXTURE_BINDING_2D = GLenum(32873)
  GL_TEXTURE_BINDING_CUBE_MAP = GLenum(34068)
  GL_TEXTURE_CUBE_MAP = GLenum(34067)
  GL_TEXTURE_CUBE_MAP_NEGATIVE_X = GLenum(34070)
  GL_TEXTURE_CUBE_MAP_NEGATIVE_Y = GLenum(34072)
  GL_TEXTURE_CUBE_MAP_NEGATIVE_Z = GLenum(34074)
  GL_TEXTURE_CUBE_MAP_POSITIVE_X = GLenum(34069)
  GL_TEXTURE_CUBE_MAP_POSITIVE_Y = GLenum(34071)
  GL_TEXTURE_CUBE_MAP_POSITIVE_Z = GLenum(34073)
  GL_TEXTURE_MAG_FILTER = GLenum(10240)
  GL_TEXTURE_MIN_FILTER = GLenum(10241)
  GL_TEXTURE_WRAP_S = GLenum(10242)
  GL_TEXTURE_WRAP_T = GLenum(10243)
  GL_TRIANGLES = GLenum(4)
  GL_TRIANGLE_FAN = GLenum(6)
  GL_TRIANGLE_STRIP = GLenum(5)
  GL_UNPACK_ALIGNMENT = GLenum(3317)
  GL_UNPACK_COLORSPACE_CONVERSION_WEBGL = GLenum(37443)
  GL_UNPACK_FLIP_Y_WEBGL = GLenum(37440)
  GL_UNPACK_PREMULTIPLY_ALPHA_WEBGL = GLenum(37441)
  GL_UNSIGNED_BYTE = GLenum(5121)
  GL_UNSIGNED_INT = GLenum(5125)
  GL_UNSIGNED_SHORT = GLenum(5123)
  GL_UNSIGNED_SHORT_4_4_4_4 = GLenum(32819)
  GL_UNSIGNED_SHORT_5_5_5_1 = GLenum(32820)
  GL_UNSIGNED_SHORT_5_6_5 = GLenum(33635)
  GL_VALIDATE_STATUS = GLenum(35715)
  GL_VENDOR = GLenum(7936)
  GL_VERSION = GLenum(7938)
  GL_VERTEX_ATTRIB_ARRAY_BUFFER_BINDING = GLenum(34975)
  GL_VERTEX_ATTRIB_ARRAY_ENABLED = GLenum(34338)
  GL_VERTEX_ATTRIB_ARRAY_NORMALIZED = GLenum(34922)
  GL_VERTEX_ATTRIB_ARRAY_POINTER = GLenum(34373)
  GL_VERTEX_ATTRIB_ARRAY_SIZE = GLenum(34339)
  GL_VERTEX_ATTRIB_ARRAY_STRIDE = GLenum(34340)
  GL_VERTEX_ATTRIB_ARRAY_TYPE = GLenum(34341)
  GL_VERTEX_SHADER = GLenum(35633)
  GL_VIEWPORT = GLenum(2978)
  GL_ZERO = GLenum(0)
)
