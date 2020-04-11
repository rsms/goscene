package main

import (
  "syscall/js"
  "unsafe"
  "runtime"
)

// See https://www.khronos.org/registry/OpenGL/api/GLES/gl.h

type GLenum        = uint32
type GLboolean     = uint8
type GLbitfield    = uint32
type GLbyte        = int8
type GLshort       = int16
type GLint         = int32
type GLsizei       = int32
type GLubyte       = uint8
type GLushort      = uint16
type GLuint        = uint32
type GLfloat       = float32
type GLclampf      = float32
type GLintptrARB   = int32
type GLsizeiptrARB = int32
type GLfixed       = int32
type GLclampx      = int32

type GLBuffer = js.Value
type GLUniform = js.Value

type GLContext struct {
  jsv js.Value
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
  index, size, typ uint32, normalized bool, stride, offset uint32) {
  normalized_ := uint32(0)
  if normalized {
    normalized_ = 1
  }
  hostcall_jvu32_(HGLvertexAttribPointer, gl.jsv, index, size, typ, normalized_, stride, offset)
}

func (gl *GLContext) enableVertexAttribArray(index uint32) {
  hostcall_ju32_(HGLenableVertexAttribArray, gl.jsv, index)
}

func (gl *GLContext) useProgram(p *GLProgram) {
  hostcall_jx2_(HGLuseProgram, gl.jsv, p.jsv)
}

func (gl *GLContext) drawArrays(mode, first, count uint32) {
  hostcall_ju32x3_(HGLdrawArrays, gl.jsv, mode, first, count)
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
  if (!gl.jsv.Call("getShaderParameter", jsv, COMPILE_STATUS).Bool()) {
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


type GLProgram struct {
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
  if (!gl.jsv.Call("getProgramParameter", jsv, LINK_STATUS).Bool()) {
    info := gl.jsv.Call("getProgramInfoLog", jsv)
    gl.jsv.Call("deleteProgram", jsv)
    return nil, errorf("shader failed to compile: %+v", info)
  }
  p := &GLProgram{ gl: gl, jsv: jsv, shaders: shaders }
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
  vertextShader, err := NewGLShader(gl, VERTEX_SHADER, vSource)
  if err != nil {
    return nil, err
  }
  fragmentShader, err := NewGLShader(gl, FRAGMENT_SHADER, hSource)
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
  ACTIVE_ATTRIBUTES = GLenum(35721)
  ACTIVE_TEXTURE = GLenum(34016)
  ACTIVE_UNIFORMS = GLenum(35718)
  ALIASED_LINE_WIDTH_RANGE = GLenum(33902)
  ALIASED_POINT_SIZE_RANGE = GLenum(33901)
  ALPHA = GLenum(6406)
  ALPHA_BITS = GLenum(3413)
  ALWAYS = GLenum(519)
  ARRAY_BUFFER = GLenum(34962)
  ARRAY_BUFFER_BINDING = GLenum(34964)
  ATTACHED_SHADERS = GLenum(35717)
  BACK = GLenum(1029)
  BLEND = GLenum(3042)
  BLEND_COLOR = GLenum(32773)
  BLEND_DST_ALPHA = GLenum(32970)
  BLEND_DST_RGB = GLenum(32968)
  BLEND_EQUATION = GLenum(32777)
  BLEND_EQUATION_ALPHA = GLenum(34877)
  BLEND_EQUATION_RGB = GLenum(32777)
  BLEND_SRC_ALPHA = GLenum(32971)
  BLEND_SRC_RGB = GLenum(32969)
  BLUE_BITS = GLenum(3412)
  BOOL = GLenum(35670)
  BOOL_VEC2 = GLenum(35671)
  BOOL_VEC3 = GLenum(35672)
  BOOL_VEC4 = GLenum(35673)
  BROWSER_DEFAULT_WEBGL = GLenum(37444)
  BUFFER_SIZE = GLenum(34660)
  BUFFER_USAGE = GLenum(34661)
  BYTE = GLenum(5120)
  CCW = GLenum(2305)
  CLAMP_TO_EDGE = GLenum(33071)
  COLOR_ATTACHMENT0 = GLenum(36064)
  COLOR_BUFFER_BIT = GLenum(16384)
  COLOR_CLEAR_VALUE = GLenum(3106)
  COLOR_WRITEMASK = GLenum(3107)
  COMPILE_STATUS = GLenum(35713)
  COMPRESSED_TEXTURE_FORMATS = GLenum(34467)
  CONSTANT_ALPHA = GLenum(32771)
  CONSTANT_COLOR = GLenum(32769)
  CONTEXT_LOST_WEBGL = GLenum(37442)
  CULL_FACE = GLenum(2884)
  CULL_FACE_MODE = GLenum(2885)
  CURRENT_PROGRAM = GLenum(35725)
  CURRENT_VERTEX_ATTRIB = GLenum(34342)
  CW = GLenum(2304)
  DECR = GLenum(7683)
  DECR_WRAP = GLenum(34056)
  DELETE_STATUS = GLenum(35712)
  DEPTH_ATTACHMENT = GLenum(36096)
  DEPTH_BITS = GLenum(3414)
  DEPTH_BUFFER_BIT = GLenum(256)
  DEPTH_CLEAR_VALUE = GLenum(2931)
  DEPTH_COMPONENT = GLenum(6402)
  DEPTH_COMPONENT16 = GLenum(33189)
  DEPTH_FUNC = GLenum(2932)
  DEPTH_RANGE = GLenum(2928)
  DEPTH_STENCIL = GLenum(34041)
  DEPTH_STENCIL_ATTACHMENT = GLenum(33306)
  DEPTH_TEST = GLenum(2929)
  DEPTH_WRITEMASK = GLenum(2930)
  DITHER = GLenum(3024)
  DONT_CARE = GLenum(4352)
  DST_ALPHA = GLenum(772)
  DST_COLOR = GLenum(774)
  DYNAMIC_DRAW = GLenum(35048)
  ELEMENT_ARRAY_BUFFER = GLenum(34963)
  ELEMENT_ARRAY_BUFFER_BINDING = GLenum(34965)
  EQUAL = GLenum(514)
  FASTEST = GLenum(4353)
  FLOAT = GLenum(5126)
  FLOAT_MAT2 = GLenum(35674)
  FLOAT_MAT3 = GLenum(35675)
  FLOAT_MAT4 = GLenum(35676)
  FLOAT_VEC2 = GLenum(35664)
  FLOAT_VEC3 = GLenum(35665)
  FLOAT_VEC4 = GLenum(35666)
  FRAGMENT_SHADER = GLenum(35632)
  FRAMEBUFFER = GLenum(36160)
  FRAMEBUFFER_ATTACHMENT_OBJECT_NAME = GLenum(36049)
  FRAMEBUFFER_ATTACHMENT_OBJECT_TYPE = GLenum(36048)
  FRAMEBUFFER_ATTACHMENT_TEXTURE_CUBE_MAP_FACE = GLenum(36051)
  FRAMEBUFFER_ATTACHMENT_TEXTURE_LEVEL = GLenum(36050)
  FRAMEBUFFER_BINDING = GLenum(36006)
  FRAMEBUFFER_COMPLETE = GLenum(36053)
  FRAMEBUFFER_INCOMPLETE_ATTACHMENT = GLenum(36054)
  FRAMEBUFFER_INCOMPLETE_DIMENSIONS = GLenum(36057)
  FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT = GLenum(36055)
  FRAMEBUFFER_UNSUPPORTED = GLenum(36061)
  FRONT = GLenum(1028)
  FRONT_AND_BACK = GLenum(1032)
  FRONT_FACE = GLenum(2886)
  FUNC_ADD = GLenum(32774)
  FUNC_REVERSE_SUBTRACT = GLenum(32779)
  FUNC_SUBTRACT = GLenum(32778)
  GENERATE_MIPMAP_HINT = GLenum(33170)
  GEQUAL = GLenum(518)
  GREATER = GLenum(516)
  GREEN_BITS = GLenum(3411)
  HIGH_FLOAT = GLenum(36338)
  HIGH_INT = GLenum(36341)
  IMPLEMENTATION_COLOR_READ_FORMAT = GLenum(35739)
  IMPLEMENTATION_COLOR_READ_TYPE = GLenum(35738)
  INCR = GLenum(7682)
  INCR_WRAP = GLenum(34055)
  INT = GLenum(5124)
  INT_VEC2 = GLenum(35667)
  INT_VEC3 = GLenum(35668)
  INT_VEC4 = GLenum(35669)
  INVALID_ENUM = GLenum(1280)
  INVALID_FRAMEBUFFER_OPERATION = GLenum(1286)
  INVALID_OPERATION = GLenum(1282)
  INVALID_VALUE = GLenum(1281)
  INVERT = GLenum(5386)
  KEEP = GLenum(7680)
  LEQUAL = GLenum(515)
  LESS = GLenum(513)
  LINEAR = GLenum(9729)
  LINEAR_MIPMAP_LINEAR = GLenum(9987)
  LINEAR_MIPMAP_NEAREST = GLenum(9985)
  LINES = GLenum(1)
  LINE_LOOP = GLenum(2)
  LINE_STRIP = GLenum(3)
  LINE_WIDTH = GLenum(2849)
  LINK_STATUS = GLenum(35714)
  LOW_FLOAT = GLenum(36336)
  LOW_INT = GLenum(36339)
  LUMINANCE = GLenum(6409)
  LUMINANCE_ALPHA = GLenum(6410)
  MAX_COMBINED_TEXTURE_IMAGE_UNITS = GLenum(35661)
  MAX_CUBE_MAP_TEXTURE_SIZE = GLenum(34076)
  MAX_FRAGMENT_UNIFORM_VECTORS = GLenum(36349)
  MAX_RENDERBUFFER_SIZE = GLenum(34024)
  MAX_TEXTURE_IMAGE_UNITS = GLenum(34930)
  MAX_TEXTURE_SIZE = GLenum(3379)
  MAX_VARYING_VECTORS = GLenum(36348)
  MAX_VERTEX_ATTRIBS = GLenum(34921)
  MAX_VERTEX_TEXTURE_IMAGE_UNITS = GLenum(35660)
  MAX_VERTEX_UNIFORM_VECTORS = GLenum(36347)
  MAX_VIEWPORT_DIMS = GLenum(3386)
  MEDIUM_FLOAT = GLenum(36337)
  MEDIUM_INT = GLenum(36340)
  MIRRORED_REPEAT = GLenum(33648)
  NEAREST = GLenum(9728)
  NEAREST_MIPMAP_LINEAR = GLenum(9986)
  NEAREST_MIPMAP_NEAREST = GLenum(9984)
  NEVER = GLenum(512)
  NICEST = GLenum(4354)
  NONE = GLenum(0)
  NOTEQUAL = GLenum(517)
  NO_ERROR = GLenum(0)
  ONE = GLenum(1)
  ONE_MINUS_CONSTANT_ALPHA = GLenum(32772)
  ONE_MINUS_CONSTANT_COLOR = GLenum(32770)
  ONE_MINUS_DST_ALPHA = GLenum(773)
  ONE_MINUS_DST_COLOR = GLenum(775)
  ONE_MINUS_SRC_ALPHA = GLenum(771)
  ONE_MINUS_SRC_COLOR = GLenum(769)
  OUT_OF_MEMORY = GLenum(1285)
  PACK_ALIGNMENT = GLenum(3333)
  POINTS = GLenum(0)
  POLYGON_OFFSET_FACTOR = GLenum(32824)
  POLYGON_OFFSET_FILL = GLenum(32823)
  POLYGON_OFFSET_UNITS = GLenum(10752)
  RED_BITS = GLenum(3410)
  RENDERBUFFER = GLenum(36161)
  RENDERBUFFER_ALPHA_SIZE = GLenum(36179)
  RENDERBUFFER_BINDING = GLenum(36007)
  RENDERBUFFER_BLUE_SIZE = GLenum(36178)
  RENDERBUFFER_DEPTH_SIZE = GLenum(36180)
  RENDERBUFFER_GREEN_SIZE = GLenum(36177)
  RENDERBUFFER_HEIGHT = GLenum(36163)
  RENDERBUFFER_INTERNAL_FORMAT = GLenum(36164)
  RENDERBUFFER_RED_SIZE = GLenum(36176)
  RENDERBUFFER_STENCIL_SIZE = GLenum(36181)
  RENDERBUFFER_WIDTH = GLenum(36162)
  RENDERER = GLenum(7937)
  REPEAT = GLenum(10497)
  REPLACE = GLenum(7681)
  RGB = GLenum(6407)
  RGB565 = GLenum(36194)
  RGB5_A1 = GLenum(32855)
  RGBA = GLenum(6408)
  RGBA4 = GLenum(32854)
  SAMPLER_2D = GLenum(35678)
  SAMPLER_CUBE = GLenum(35680)
  SAMPLES = GLenum(32937)
  SAMPLE_ALPHA_TO_COVERAGE = GLenum(32926)
  SAMPLE_BUFFERS = GLenum(32936)
  SAMPLE_COVERAGE = GLenum(32928)
  SAMPLE_COVERAGE_INVERT = GLenum(32939)
  SAMPLE_COVERAGE_VALUE = GLenum(32938)
  SCISSOR_BOX = GLenum(3088)
  SCISSOR_TEST = GLenum(3089)
  SHADER_TYPE = GLenum(35663)
  SHADING_LANGUAGE_VERSION = GLenum(35724)
  SHORT = GLenum(5122)
  SRC_ALPHA = GLenum(770)
  SRC_ALPHA_SATURATE = GLenum(776)
  SRC_COLOR = GLenum(768)
  STATIC_DRAW = GLenum(35044)
  STENCIL_ATTACHMENT = GLenum(36128)
  STENCIL_BACK_FAIL = GLenum(34817)
  STENCIL_BACK_FUNC = GLenum(34816)
  STENCIL_BACK_PASS_DEPTH_FAIL = GLenum(34818)
  STENCIL_BACK_PASS_DEPTH_PASS = GLenum(34819)
  STENCIL_BACK_REF = GLenum(36003)
  STENCIL_BACK_VALUE_MASK = GLenum(36004)
  STENCIL_BACK_WRITEMASK = GLenum(36005)
  STENCIL_BITS = GLenum(3415)
  STENCIL_BUFFER_BIT = GLenum(1024)
  STENCIL_CLEAR_VALUE = GLenum(2961)
  STENCIL_FAIL = GLenum(2964)
  STENCIL_FUNC = GLenum(2962)
  STENCIL_INDEX8 = GLenum(36168)
  STENCIL_PASS_DEPTH_FAIL = GLenum(2965)
  STENCIL_PASS_DEPTH_PASS = GLenum(2966)
  STENCIL_REF = GLenum(2967)
  STENCIL_TEST = GLenum(2960)
  STENCIL_VALUE_MASK = GLenum(2963)
  STENCIL_WRITEMASK = GLenum(2968)
  STREAM_DRAW = GLenum(35040)
  SUBPIXEL_BITS = GLenum(3408)
  TEXTURE = GLenum(5890)
  TEXTURE0 = GLenum(33984)
  TEXTURE1 = GLenum(33985)
  TEXTURE2 = GLenum(33986)
  TEXTURE3 = GLenum(33987)
  TEXTURE4 = GLenum(33988)
  TEXTURE5 = GLenum(33989)
  TEXTURE6 = GLenum(33990)
  TEXTURE7 = GLenum(33991)
  TEXTURE8 = GLenum(33992)
  TEXTURE9 = GLenum(33993)
  TEXTURE10 = GLenum(33994)
  TEXTURE11 = GLenum(33995)
  TEXTURE12 = GLenum(33996)
  TEXTURE13 = GLenum(33997)
  TEXTURE14 = GLenum(33998)
  TEXTURE15 = GLenum(33999)
  TEXTURE16 = GLenum(34000)
  TEXTURE17 = GLenum(34001)
  TEXTURE18 = GLenum(34002)
  TEXTURE19 = GLenum(34003)
  TEXTURE20 = GLenum(34004)
  TEXTURE21 = GLenum(34005)
  TEXTURE22 = GLenum(34006)
  TEXTURE23 = GLenum(34007)
  TEXTURE24 = GLenum(34008)
  TEXTURE25 = GLenum(34009)
  TEXTURE26 = GLenum(34010)
  TEXTURE27 = GLenum(34011)
  TEXTURE28 = GLenum(34012)
  TEXTURE29 = GLenum(34013)
  TEXTURE30 = GLenum(34014)
  TEXTURE31 = GLenum(34015)
  TEXTURE_2D = GLenum(3553)
  TEXTURE_BINDING_2D = GLenum(32873)
  TEXTURE_BINDING_CUBE_MAP = GLenum(34068)
  TEXTURE_CUBE_MAP = GLenum(34067)
  TEXTURE_CUBE_MAP_NEGATIVE_X = GLenum(34070)
  TEXTURE_CUBE_MAP_NEGATIVE_Y = GLenum(34072)
  TEXTURE_CUBE_MAP_NEGATIVE_Z = GLenum(34074)
  TEXTURE_CUBE_MAP_POSITIVE_X = GLenum(34069)
  TEXTURE_CUBE_MAP_POSITIVE_Y = GLenum(34071)
  TEXTURE_CUBE_MAP_POSITIVE_Z = GLenum(34073)
  TEXTURE_MAG_FILTER = GLenum(10240)
  TEXTURE_MIN_FILTER = GLenum(10241)
  TEXTURE_WRAP_S = GLenum(10242)
  TEXTURE_WRAP_T = GLenum(10243)
  TRIANGLES = GLenum(4)
  TRIANGLE_FAN = GLenum(6)
  TRIANGLE_STRIP = GLenum(5)
  UNPACK_ALIGNMENT = GLenum(3317)
  UNPACK_COLORSPACE_CONVERSION_WEBGL = GLenum(37443)
  UNPACK_FLIP_Y_WEBGL = GLenum(37440)
  UNPACK_PREMULTIPLY_ALPHA_WEBGL = GLenum(37441)
  UNSIGNED_BYTE = GLenum(5121)
  UNSIGNED_INT = GLenum(5125)
  UNSIGNED_SHORT = GLenum(5123)
  UNSIGNED_SHORT_4_4_4_4 = GLenum(32819)
  UNSIGNED_SHORT_5_5_5_1 = GLenum(32820)
  UNSIGNED_SHORT_5_6_5 = GLenum(33635)
  VALIDATE_STATUS = GLenum(35715)
  VENDOR = GLenum(7936)
  VERSION = GLenum(7938)
  VERTEX_ATTRIB_ARRAY_BUFFER_BINDING = GLenum(34975)
  VERTEX_ATTRIB_ARRAY_ENABLED = GLenum(34338)
  VERTEX_ATTRIB_ARRAY_NORMALIZED = GLenum(34922)
  VERTEX_ATTRIB_ARRAY_POINTER = GLenum(34373)
  VERTEX_ATTRIB_ARRAY_SIZE = GLenum(34339)
  VERTEX_ATTRIB_ARRAY_STRIDE = GLenum(34340)
  VERTEX_ATTRIB_ARRAY_TYPE = GLenum(34341)
  VERTEX_SHADER = GLenum(35633)
  VIEWPORT = GLenum(2978)
  ZERO = GLenum(0)
)
