package main

import (
  "syscall/js"
  "unsafe"
)

// see https://godoc.org/syscall/js

// implemented in host_js.s + host.js
func hostcall___(msg uint32)
func hostcall__f64(msg uint32) float64
func hostcall__u32x2(msg uint32) (uint32, uint32)
func hostcall_j_i32(msg uint32, v1 js.Value) int32
func hostcall_j_u32x2(msg uint32, v1 js.Value) (uint32, uint32)
func hostcall_jf32_(msg uint32, v1 js.Value, v2 float32)
func hostcall_ju32_(msg uint32, v1 js.Value, v2 uint32)
func hostcall_ju32j_(msg uint32, v1 js.Value, v2 uint32, v3 js.Value)
func hostcall_ju32x3_(msg uint32, v1 js.Value, v2,v3,v4 uint32)
func hostcall_ju32x4_(msg uint32, v1 js.Value, v2,v3,v4,v5 uint32)
func hostcall_jvf32_(msg uint32, v1 js.Value, v... float32)
func hostcall_jvi32_(msg uint32, v1 js.Value, v... int32)
func hostcall_jvu32_(msg uint32, v1 js.Value, v... uint32)
func hostcall_jx2_(msg uint32, v1 js.Value, v2 js.Value)
func hostcall_jx2vf32_(msg uint32, v1,v2 js.Value, v... float32)
func hostcall_jx2vi32_(msg uint32, v1,v2 js.Value, v... int32)
func hostcall_jx2vu32_(msg uint32, v1,v2 js.Value, v... uint32)
func hostcall_u32_(msg uint32, v1 uint32)
func hostcall_vi32_i32(msg uint32, v... int32) int32
func hostcall_vu8_i32(msg uint32, v... uint8) int32

// hostcall message constants
const (
  // global
  HEventSubscribe       = uint32(2)
  HEventUnsubscribe     = uint32(3)
  HWindowSize           = uint32(10)  // () -> i32 (i16x2)
  HPixelRatio           = uint32(11)  // () -> f64
  HMonotime             = uint32(12)  // () -> f64
  HTime                 = uint32(13)  // () -> f64
  HReadRandom           = uint32(14)  // ([]byte) -> i32
  HAnimationStatsUpdate = uint32(15)  // () -> ()

  // GL (all these functions takes a JS object as the first parameter; context)
  HGLdrawingBufferSize = uint32(1000) // () -> i32,i32
  HGLcanvasSize = uint32(1001) // () -> i32,i32
  HGLviewport = uint32(1002) // (i32,i32,i32,i32) -> ()
  HGLclear = uint32(1003) // (u32) -> ()
  HGLclearColor = uint32(1004) // (f32,f32,f32,f32) -> ()
  HGLclearDepth = uint32(1005) // (f32) -> ()
  HGLenable = uint32(1006) // (u32) -> ()
  HGLdepthFunc = uint32(1007) // (u32) -> ()
  HGLbindBuffer = uint32(1008) // (u32,u32) -> ()
  HGLbufferData = uint32(1009) // (u32,u32,u32,[]uint8) -> ()
  HGLvertexAttribPointer = uint32(1010)
  HGLenableVertexAttribArray = uint32(1011)
  HGLuseProgram = uint32(1012)
  HGLuniformMatrix2fv = uint32(1013)
  HGLuniformMatrix3fv = uint32(1014)
  HGLuniformMatrix4fv = uint32(1015)
  HGLuniformvf = uint32(1016)
  HGLuniformvi = uint32(1017)
  HGLdrawArrays = uint32(1018)
  HGLdrawElements = uint32(1019)
)

// Events
type Event uint32
const (
  EVNone = Event(iota)
  EVWindowResize
  EVPointerMove
  EVPointerDown
  EVPointerUp
  EVAnimationFrame
)

func (e Event) String() string {
  switch e {
  case EVWindowResize:   return "EVWindowResize"
  case EVPointerMove:    return "EVPointerMove"
  case EVPointerDown:    return "EVPointerDown"
  case EVPointerUp:      return "EVPointerUp"
  case EVAnimationFrame: return "EVAnimationFrame"
  default:               return "(EV?)"
  }
}


var host = CreateHostEnv(js.Global().Get("_appinit").Get("host"))

func Monotime() float64 { return host.Monotime() }
func Time() float64 { return host.Time() }


type Pointer struct {
  id      int      // system-specific unique identifier for the virtual pointer device
  x, y    float32  // position in display points
  buttons uint32   // bitmask of buttons' depressed state. 0 = none, 1 = left, 2 = right, 4 ...
}


type HostEnv struct {
  jsv           js.Value  // reference to the Host JS object in the host runtime
  windowWidth   uint32    // width of host window in display points
  windowHeight  uint32    // height of host window in display points
  pixelRatio    float32   // display point to pixel ratio
  pointer       Pointer   // input pointer
  scenetime     float64   // monotonically incrementing time, in seconds
  runloopStopCh chan bool // used by RunLoop() to signal stop
  events        EventSource
}

func CreateHostEnv(jsv js.Value) *HostEnv {
  h := &HostEnv{
    jsv: jsv,
    pixelRatio: 1.0,
  }

  h.events.Enable = func (ev Event) {
    // called when the first handler for ev was added
    h.eventSubscribe(ev)
  }
  h.events.Disable = func (ev Event) {
    // called when the last handler for ev was removed
    h.eventUnsubscribe(ev)
  }

  // Track window size and pixel ratio
  h.pixelRatio = float32(h.getPixelRatio())
  h.windowWidth, h.windowHeight = h.getWindowSize()
  h.events.Listen(EVWindowResize, func (ev Event, xy ...uint32) {
    h.pixelRatio = float32(h.getPixelRatio())
    h.windowWidth, h.windowHeight = xy[0], xy[1]
  })

  // Track mouse pointer
  h.pointer.id = -1
  onPointerEvent := func (_ Event, data ...uint32) {
    // data = [ pointerId, x, y, buttons ]
    // Note: /10 is a constant that is matched in host.js's handlePointerEvent
    id := int(data[0])
    if h.pointer.id == -1 {
      h.pointer.id = id
    } else if h.pointer.id != id {
      // ignore non-primary pointer
      return
    }
    h.pointer.x = float32(data[1]) / 10
    h.pointer.y = float32(data[2]) / 10
    h.pointer.buttons = data[3]
  }
  h.events.Listen(EVPointerMove, onPointerEvent)
  h.events.Listen(EVPointerDown, onPointerEvent)
  h.events.Listen(EVPointerUp, onPointerEvent)

  /*// XXX test hostcall
  r := hostcall_vi32_i32(123, 111, 222, 333, 444)
  logf("hostcall_vi32_i32(123) => %v", r)

  logf("getPixelRatio() => %v", h.getPixelRatio())
  // logf("WindowSize() => %v, %v", h.WindowSize())
  logf("Monotime() => %v", h.Monotime())
  logf("Time() => %v", h.Time())

  rand := make([]byte, 8)
  h.ReadRandom(rand)
  logf("ReadRandom() => %v", rand)*/

  return h
}


func (h *HostEnv) eventSubscribe(ev Event) {
  hostcall_u32_(HEventSubscribe, uint32(ev))
}

func (h *HostEnv) eventUnsubscribe(ev Event) {
  hostcall_u32_(HEventUnsubscribe, uint32(ev))
}


// Monotime returns a timestamp in seconds with high resolution that is guaranteed to
// increment for every call, unlike Time() which could go backwards if the computer clock changes.
func (h *HostEnv) Monotime() float64 {
  // Note: float32 as seconds gives us 97 days before we completely run out of precision.
  // A float32 for seconds will run out of millisecond precision after about 24 days (2095000s)
  // since IEEE 754 float has 23 bits of mantissa, (2^23)/60/60/24 = 97 days.
  // A flaot64 for seconds means we can keep running "forever" without precision loss, or
  // 142 808 207 years.
  return hostcall__f64(HMonotime)
}

func (h *HostEnv) Time() float64 {
  return hostcall__f64(HTime)
}

func (h *HostEnv) getPixelRatio() float64 {
  return hostcall__f64(HPixelRatio)
}

func (h *HostEnv) getWindowSize() (width, height uint32) {
  return hostcall__u32x2(HWindowSize)
}

func (h *HostEnv) ReadRandom(buf []byte) int {
  return int(hostcall_vu8_i32(HReadRandom, buf...))
}

func (h *HostEnv) dispatchEvent(ev Event, data []uint32) {
  h.events.Trigger(ev, data...)
}

func (h *HostEnv) UpdateAnimationStats() {
  hostcall___(HAnimationStatsUpdate)
}


// StopRunLoop causes an existing call to RunLoop() to return.
// Note: This may caues the program to exit in case main() called RunLoop().
func (h *HostEnv) StopRunLoop() bool {
  if h.runloopStopCh != nil {
    host.jsv.Call("stopRunLoop")
    select {
      case h.runloopStopCh <- true:
        h.runloopStopCh = nil
        return true
      default:
    }
  }
  return false
}


func (h *HostEnv) RunLoop() {
  h.StopRunLoop()
  h.runloopStopCh = make(chan bool)

  // msgbuf is memory shared between the program and the host, used to communicate.
  //
  // The host will write information about time and events here before it calls our
  // callback function cb. Next, cb reads msgbuf updating state like scenetime and
  // dispatches any events incoming from the host.
  // Currently cb does not write to msgbuf but it could if we needed to "talk back"
  // to the host.
  //
  // Note: msgbus if heap-allocated so that it doesn't move in case stack splits.
  // This because the host will use its address.
  msgbuf := make([]byte, 1024)

  // byte offsets into msgbuf (must match constants in host.js)
  const TIME         = 0
  const EVENT_COUNT  = TIME + 8
  const EVENT_DATA   = EVENT_COUNT + 4
  const MAX_VALCOUNT = 32

  cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    h.scenetime = *(*float64)(unsafe.Pointer(&msgbuf[TIME]))
    evcount    := *(*uint32)(unsafe.Pointer(&msgbuf[EVENT_COUNT]))
    dataptr    := uintptr(unsafe.Pointer(&msgbuf[EVENT_DATA]))
    // logf("runloop callback. time: %g, events: %s, evcount: %v",
    //   time, Event(*events), evcount)
    for i := uint32(0); i < evcount; i++ {
      ev := *(*Event)(unsafe.Pointer(dataptr))
      dataptr += 4

      valcount := *(*uint32)(unsafe.Pointer(dataptr))
      dataptr += 4
      // logf("got event %v valcount=%v", ev, valcount)

      if valcount == 0 {
        h.dispatchEvent(ev, nil)
      } else {
        if valcount > MAX_VALCOUNT {
          logf("[RunLoop] warning: received event data larger than MAX_VALCOUNT")
          valcount = MAX_VALCOUNT
        }
        a := *(*[MAX_VALCOUNT]uint32)(unsafe.Pointer(dataptr))
        h.dispatchEvent(ev, a[:valcount])
        dataptr += uintptr(valcount) * 4
      }
    }

    return nil
  })

  // start the runloop host driver, passing the address and length of msgbuf
  host.jsv.Call("startRunLoop", cb, uintptr(unsafe.Pointer(&msgbuf[0])), len(msgbuf))

  // wait for stop signal
  <-h.runloopStopCh

  // release function handle in the JS environment
  cb.Release()
}


func (h *HostEnv) error(e interface{}) {
  h.jsv.Call("error", e)
}


func (h *HostEnv) log(e... interface{}) {
  h.jsv.Call("log", e...)
}

