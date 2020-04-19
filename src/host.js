import * as stats from "./stats"

const log = console.log.bind(console)

function assert(cond, msg) {
  if (DEBUG) { // for DCE
    if (!cond) {
      let e = new Error("assertion failed: " + (msg || cond))
      e.name = "AssertionError"
      throw e
    }
  }
}

function int16(n)  { return n << 16 >> 16 }
function uint16(n) { return n & 0xFFFF }
function int32(n)  { return n >> 0 }
function uint32(n) { return n >>> 0 }

// Host call message IDs
const HEventSubscribe   = uint32(2)
    , HEventUnsubscribe = uint32(3)
    , HWindowSize       = uint32(10)  // () -> i32 (i16x2)
    , HPixelRatio       = uint32(11)  // () -> f64
    , HMonotime         = uint32(12)  // () -> f64
    , HTime             = uint32(13)  // () -> f64
    , HReadRandom       = uint32(14)  // ([]byte) -> i32
    , HAnimationStatsUpdate = uint32(15) // () -> ()
      // GL (all these functions takes a JS object as the first parameter; context)
    , HGLdrawingBufferSize = uint32(1000) // () -> i32,i32
    , HGLcanvasSize = uint32(1001) // () -> i32,i32
    , HGLviewport = uint32(1002) // (i32,i32,i32,i32) -> ()
    , HGLclear = uint32(1003) // (u32) -> ()
    , HGLclearColor = uint32(1004) // (f32,f32,f32,f32) -> ()
    , HGLclearDepth = uint32(1005) // (f32) -> ()
    , HGLenable = uint32(1006) // (u32) -> ()
    , HGLdepthFunc = uint32(1007) // (u32) -> ()
    , HGLbindBuffer = uint32(1008) // (u32,u32) -> ()
    , HGLbufferData = uint32(1009) // (u32,u32,u32,[]uint8) -> ()
    , HGLvertexAttribPointer = uint32(1010)
    , HGLenableVertexAttribArray = uint32(1011)
    , HGLuseProgram = uint32(1012)
    , HGLuniformMatrix2fv = uint32(1013)
    , HGLuniformMatrix3fv = uint32(1014)
    , HGLuniformMatrix4fv = uint32(1015)
    , HGLuniformvf = uint32(1016)
    , HGLuniformvi = uint32(1017)
    , HGLdrawArrays = uint32(1018)
    , HGLdrawElements = uint32(1019)

// Event IDs
const EVNone           = 0
    , EVWindowResize   = 1
    , EVPointerMove    = 2
    , EVPointerDown    = 3
    , EVPointerUp      = 4
    , EVAnimationFrame = 5

function EVString(ev) {
  switch (ev) {
  case EVWindowResize:   return "EVWindowResize"
  case EVPointerMove:    return "EVPointerMove"
  case EVPointerDown:    return "EVPointerDown"
  case EVPointerUp:      return "EVPointerUp"
  case EVAnimationFrame: return "EVAnimationFrame"
  default:               return "(EV?)"
  }
}


// host calls
// Function name keywords:
//   i{N}  integer of {N} size
//   u{N}  unsigned integer of {N} size
//   f{N}  float of {N} size
//   v{T}  slice of type {T}
//   j     JS Object (js.Value in Go.) Load with getJSObject(addr)
//

const hostcallHandlers = {}

function regHCall(sig, msg, handler) {
  let v = hostcallHandlers[sig]
  if (!v) {
    hostcallHandlers[sig] = v = []
  }
  assert(!v[msg], `duplicate hostcall handler ${sig}[${msg}]`)
  v[msg] = handler
}

// -----------------------------------------------------------------------------------
// global

// regHCall("vu32_", HEventSubscribe, (mem, argc, argaddr) => {
//   let evs = new Array(argc)
//   let i = 0
//   for (let endaddr = argaddr + (argc * 4); argaddr < endaddr; argaddr += 4) {
//     evs[i++] = mem.getUint32(argaddr)
//   }
//   host.eventSubscribe(evs)
// })

regHCall("u32_", HEventSubscribe, (mem, events) => {
  host.eventSubscribe(events)
})

regHCall("u32_", HEventUnsubscribe, (mem, events) => {
  host.eventUnsubscribe(events)
})

regHCall("__", HAnimationStatsUpdate, mem => {
  host.updateAnimationStats(performance.now())
})

// HPixelRatio: scale of display points to pixels for the current window.
// A display with 200% scaling factor yields the value 2.0 (i.e. 5dp = 10px.)
regHCall("_f64", HPixelRatio, () => window.devicePixelRatio || 1.0)

// HMonotime: high-precision monotonic clock (seconds)
regHCall("_f64", HMonotime, () => performance.now() * 1000)

// HTime: real time in seconds
const timeOrigin = Date.now() - performance.now()
regHCall("_f64", HTime, () => (timeOrigin + performance.now()) * 1000)

// HWindowSize: window size encoded as two uint16.
regHCall("_u32x2", HWindowSize, () => {
  return [window.innerWidth, window.innerHeight]
})
// // First 16 bits are width, last 16 bits are height.
// regHCall("_i32", HWindowSize, () => {
//   return (uint16(window.innerHeight) << 16) | uint16(window.innerWidth)
// })

// HReadRandom: fills argument array with random bytes. Returns number of bytes filled.
regHCall("vu8_i32", HReadRandom, (mem, argc, argaddr) => {
  let data = new Uint8Array(mem.buf, argaddr, argc)
  return crypto.getRandomValues(data).length
})

regHCall("vi32_i32", 123, (mem, argc, argaddr) => {
  log("hostcall", { argaddr, argc })
  for (let endaddr = argaddr + (argc * 4); argaddr < endaddr; argaddr += 4) {
    log(`  val#${argc - ((endaddr - argaddr) / 4)}`, mem.getInt32(argaddr))
  }
})

// -----------------------------------------------------------------------------------
// WebGL
regHCall("j_u32x2", HGLdrawingBufferSize, (mem, gl) =>
  [ gl.drawingBufferWidth, gl.drawingBufferHeight ]
)

regHCall("j_u32x2", HGLcanvasSize, (mem, gl) =>
  [ gl.canvas.clientWidth, gl.canvas.clientHeight ]
)

regHCall("jvi32_", HGLviewport, (mem, gl, argc, argaddr) => {
  assert(argc == 4)
  let x = mem.getInt32(argaddr)
  let y = mem.getInt32(argaddr + 4)
  let w = mem.getInt32(argaddr + 8)
  let h = mem.getInt32(argaddr + 12)
  gl.viewport(x, y, w, h)
})

regHCall("ju32j_", HGLbindBuffer, (mem, gl, target, bufferObj) => {
  gl.bindBuffer(target, bufferObj)
})

regHCall("ju32_", HGLclear, (mem, gl, mask) => {
  gl.clear(mask)
})

regHCall("ju32_", HGLenable, (mem, gl, cap) => {
  gl.enable(cap)
})

regHCall("ju32_", HGLdepthFunc, (mem, gl, funcid) => {
  gl.depthFunc(funcid)
})

regHCall("jvf32_", HGLclearColor, (mem, gl, argc, argaddr) => {
  assert(argc == 4)
  let r = mem.getFloat32(argaddr)
  let g = mem.getFloat32(argaddr + 4)
  let b = mem.getFloat32(argaddr + 8)
  let a = mem.getFloat32(argaddr + 12)
  gl.clearColor(r, g, b, a)
})

regHCall("jf32_", HGLclearDepth, (mem, gl, d) => {
  gl.clearDepth(d)
})

regHCall("jvu32_", HGLbufferData, (mem, gl, argc, argaddr) => {
  // Note: bufferData accepts a variety of data types. Rather than implementing
  // separate handlers for each possible data type, we instead accept a pointer to data
  // including the size in bytes of that data in memory. Then, we provide gl.bufferData
  // with a view into Go memory where the data is stored.
  //
  // Note: We can not avoid allocating a JS object here (the Uint8Array).
  // WebGL2 adds two additional arguments to bufferData—srcOffset, length—that allows
  // passing in an existing ArrayBuffer and have the API read a range. If this code is
  // ever converted or ported to WebGL2, keep that in mind.
  assert(argc == 4)
  const target   = mem.getUint32(argaddr)
  const usage    = mem.getUint32(argaddr + 4)
  const dataaddr = mem.getUint32(argaddr + 8)   // address of data in gomem
  const datasize = mem.getUint32(argaddr + 12)  // size of data in bytes
  gl.bufferData(target, new Uint8Array(mem.buf, dataaddr, datasize), usage)
})

regHCall("jvu32_", HGLvertexAttribPointer, (mem, gl, argc, argaddr) => {
  assert(argc == 6)
  const index      = mem.getUint32(argaddr)
  const size       = mem.getUint32(argaddr + 4)
  const type       = mem.getUint32(argaddr + 8)
  const normalized = mem.getUint32(argaddr + 12)
  const stride     = mem.getUint32(argaddr + 16)
  const offset     = mem.getUint32(argaddr + 20)
  gl.vertexAttribPointer(index, size, type, normalized, stride, offset)
})

regHCall("ju32x3_", HGLdrawArrays, (mem, gl, mode, first, count) => {
  gl.drawArrays(mode, first, count)
})

regHCall("ju32x4_", HGLdrawElements, (mem, gl, mode, count, typ, offset) => {
  gl.drawElements(mode, count, typ, offset)
})

for (let [size,msg] of [[2,HGLuniformMatrix2fv],[3,HGLuniformMatrix3fv],[4,HGLuniformMatrix4fv]]) {
  const f = WebGLRenderingContext.prototype[`uniformMatrix${size}fv`]
  const count = size * size
  regHCall("jx2vu32_", msg, (mem, gl, location, argc, argaddr) => {
    assert(argc == 2)
    const transpose  = mem.getUint32(argaddr)
    const ptr        = mem.getUint32(argaddr + 4)
    const value      = new Float32Array(mem.buf, ptr, count)
    f.call(gl, location, transpose, value)
  })
}

regHCall("jx2vf32_", HGLuniformvf, (mem, gl, location, argc, argaddr) => {
  if (argc == 1) {
    gl.uniform1f(location, mem.getFloat32(argaddr))
  } else {
    const values = new Float32Array(mem.buf, argaddr, argc)
    if (argc == 2) {
      gl.uniform2fv(location, values)
    } else if (argc == 3) {
      gl.uniform3fv(location, values)
    } else if (argc == 4) {
      gl.uniform4fv(location, values)
    } else {
      throw new Error(`invalid value count ${argc}`)
    }
  }
})

regHCall("jx2vi32_", HGLuniformvi, (mem, gl, location, argc, argaddr) => {
  if (argc == 1) {
    gl.uniform1i(location, mem.getInt32(argaddr))
  } else {
    const values = new Int32Array(mem.buf, argaddr, argc)
    if (argc == 2) {
      gl.uniform2iv(location, values)
    } else if (argc == 3) {
      gl.uniform3iv(location, values)
    } else if (argc == 4) {
      gl.uniform4iv(location, values)
    } else {
      throw new Error(`invalid value count ${argc}`)
    }
  }
})

regHCall("ju32_", HGLenableVertexAttribArray, (mem, gl, index) => {
  gl.enableVertexAttribArray(index)
})

regHCall("jx2_", HGLuseProgram, (mem, gl, program) => {
  gl.useProgram(program)
})


// -----------------------------------------------------------------------------------
// Go memory interface
class GoMemory {
  constructor(go) {
    this.go = go
    this.buf = null   // ArrayBuffer
    this.view = null  // DataView
    this.version = 0  // increments when this.buf is replaced
  }

  check() {
    let buf = this.go._inst.exports.mem.buffer
    if (this.buf !== buf) {
      // WASM memory changed
      this.buf = buf
      this.view = new DataView(buf)
      this.version++
    }
    return this
  }

  getInt8(addr) { return this.view.getInt8(addr, true) }
  getUint8(addr) { return this.view.getUint8(addr, true) }
  getInt16(addr) { return this.view.getInt16(addr, true) }
  getUint16(addr) { return this.view.getUint16(addr, true) }
  getInt32(addr) { return this.view.getInt32(addr, true) }
  getUint32(addr) { return this.view.getUint32(addr, true) }
  getFloat32(addr) { return this.view.getFloat32(addr, true) }
  getFloat64(addr) { return this.view.getFloat64(addr, true) }
  getBigInt64(addr) { return this.view.getBigInt64(addr, true) }
  getBigUint64(addr) { return this.view.getBigUint64(addr, true) }

  setInt8(addr, v) { return this.view.setInt8(addr, v, true) }
  setUint8(addr, v) { return this.view.setUint8(addr, v, true) }
  setInt16(addr, v) { return this.view.setInt16(addr, v, true) }
  setUint16(addr, v) { return this.view.setUint16(addr, v, true) }
  setInt32(addr, v) { return this.view.setInt32(addr, v, true) }
  setUint32(addr, v) { return this.view.setUint32(addr, v, true) }
  setFloat32(addr, v) { return this.view.setFloat32(addr, v, true) }
  setFloat64(addr, v) { return this.view.setFloat64(addr, v, true) }
  setBigInt64(addr, v) { return this.view.setBigInt64(addr, v, true) }
  setBigUint64(addr, v) { return this.view.setBigUint64(addr, v, true) }

  getInt64(addr) {
    const low = this.view.getUint32(addr + 0, true)
    const high = this.view.getInt32(addr + 4, true)
    return low + high * 4294967296
  }

  setInt64(addr, v) {
    this.view.setUint32(addr + 0, v, true)
    this.view.setUint32(addr + 4, Math.floor(v / 4294967296), true)
  }

  getUint8Slice(addr) {
    const arrayAddr = this.getInt64(addr)
    const len = this.getInt64(addr + 8)
    return new Uint8Array(this.buf, arrayAddr, len)
  }

  getInt32Slice(addr) {
    const arrayAddr = this.getInt64(addr)
    const len = this.getInt64(addr + 8)
    return new Int32Array(this.buf, arrayAddr, len)
  }

  getJSObject(addr) {
    const id = this.getUint32(addr)
    return this.go._values[id]
  }
}

// -----------------------------------------------------------------------------------

const host = new class Host {
  constructor() {
    this.eventMsgBuf  = new ArrayBuffer(4096)
    this.eventMsgView = new DataView(this.eventMsgBuf)

    this.wasmInstance = null

    this.eventSubscriptions = new Set()  // subscribed events
    this.transientEvents = new Map() // queued event data keyed by event code
    this.persistentEvents = new Map()  // active persistent events keyed by event code
    this.persistentEventsVersion = 0  // increments when persistentEvents changes

    this.runloopTickFunc = null


    let go = this.go = new Go()
    let gomem = new GoMemory(go)
    this.gomem = gomem

    function bindHostcall(sigs, handler) {
      if (!Array.isArray(sigs)) {
        sigs = [ sigs ]
      }
      for (let sig of sigs) {
        const impl = hostcallHandlers[sig]
        assert(impl, `no hostcallHandlers[sig=${sig}]`)
        delete hostcallHandlers[sig]
        go.importObject.go["main.hostcall_" + sig] = sp => {
          const mem = gomem.check()
          const msg = mem.getInt32(sp + 8)
          const f = impl[msg]
          if (!f) { throw new Error(`invalid hostcall_${sig} #${msg}`) }
          sp += 16
          handler(mem, sp, f)
        }
      }
    }

    bindHostcall("__", (mem, sp, f) => {
      f(mem)
    })

    bindHostcall("_f64", (mem, sp, f) => {
      mem.setFloat64(sp, f(mem))
    })

    bindHostcall("_u32x2", (mem, sp, f) => {
      const result = f(mem)
      mem.setUint32(sp, result[0])
      mem.setUint32(sp + 4, result[1])
    })

    bindHostcall(["vu8_i32","vi32_i32"], (mem, sp, f) => {
      const argaddr = mem.getInt64(sp)
      const argcount = mem.getInt64(sp + 8)
      const result = f(mem, argcount, argaddr)
      mem.setInt32(sp + 24, result)
    })

    // bindHostcall(["vu32_"], (mem, sp, f) => {
    //   const argaddr = mem.getInt64(sp)
    //   const argcount = mem.getInt64(sp + 8)
    //   f(mem, argcount, argaddr)
    // })

    bindHostcall("j_u32x2", (mem, sp, f) => {
      const obj1 = mem.getJSObject(sp)
      const result = f(mem, obj1)
      mem.setUint32(sp + 8, result[0])
      mem.setUint32(sp + 12, result[1])
    })

    bindHostcall(["jvi32_", "jvu32_", "jvf32_"], (mem, sp, f) => {
      const obj1 = mem.getJSObject(sp)
      const argaddr = mem.getInt64(sp + 8)
      const argcount = mem.getInt64(sp + 16)
      f(mem, obj1, argcount, argaddr)
    })

    bindHostcall("ju32x3_", (mem, sp, f) => {
      const obj1 = mem.getJSObject(sp)
      const v1 = mem.getUint32(sp + 8)
      const v2 = mem.getUint32(sp + 12)
      const v3 = mem.getUint32(sp + 16)
      f(mem, obj1, v1, v2, v3)
    })

    bindHostcall("ju32x4_", (mem, sp, f) => {
      const obj1 = mem.getJSObject(sp)
      const v1 = mem.getUint32(sp + 8)
      const v2 = mem.getUint32(sp + 12)
      const v3 = mem.getUint32(sp + 16)
      const v4 = mem.getUint32(sp + 20)
      f(mem, obj1, v1, v2, v3, v4)
    })

    bindHostcall(["jx2vi32_", "jx2vu32_", "jx2vf32_"], (mem, sp, f) => {
      const obj1 = mem.getJSObject(sp)
      const obj2 = mem.getJSObject(sp + 8)
      const argaddr = mem.getInt64(sp + 16)
      const argcount = mem.getInt64(sp + 24)
      f(mem, obj1, obj2, argcount, argaddr)
    })

    bindHostcall("ju32_", (mem, sp, f) => {
      f(mem, mem.getJSObject(sp), mem.getUint32(sp + 8))
    })

    bindHostcall("jf32_", (mem, sp, f) => {
      f(mem, mem.getJSObject(sp), mem.getFloat32(sp + 8))
    })

    bindHostcall("ju32j_", (mem, sp, f) => {
      f(mem, mem.getJSObject(sp), mem.getUint32(sp + 8), mem.getJSObject(sp + 16))
    })

    bindHostcall("jx2_", (mem, sp, f) => {
      f(mem, mem.getJSObject(sp), mem.getJSObject(sp + 8))
    })

    bindHostcall("u32_", (mem, sp, f) => {
      f(mem, mem.getUint32(sp - 4)) // -4 wtf
    })

    // check to make sure all hostcall handlers were bound
    if (DEBUG) {
      for (let _ in hostcallHandlers) {
        throw new Error(
          `not all hostcallHandlers were bound with bindHostcall.\n` +
          `Unhandled signatures:\n  ` +
          Object.keys(hostcallHandlers).join('\n  ') + "\n"
        )
      }
    }

    this.initEvents() // XXX
  }

  log() { console.log.apply(console, arguments) }
  error(err) { console.error(typeof err == "object" && err.stack ? err.stack : String(err)) }


  enableAnimationStats() {
    if (!this._animationStats) {
      const s = this._animationStats = new stats.Stats()
      // add Go WASM memory size stats
      s.addGenericPanel(
        { updateInterval: 1000 },
        () => `Go ${stats.fmtByteSize(this.gomem.check().buf.byteLength)}`
      )
      s.mount(document.body)
    }
  }

  disableAnimationStats() {
    if (this._animationStats) {
      this._animationStats.unmount()
      this._animationStats = null
    }
  }

  updateAnimationStats(time) {
    if (!this._animationStats) {
      this.enableAnimationStats()
    }
    this._animationStats.update(time)
  }


  getContext(canvas, contextType) {
    return canvas.getContext(contextType)
    // let g = canvas.getContext(contextType)
    // // log(`canvas.getContext(${contextType}) =>`, g)
    // // must wrap and bind all members to make go's js/Invoke() work
    // let g2 = { __proto__:g }
    // for (let k in g) {
    //   let v = g[k]
    //   if (typeof v == "function") {
    //     g2[k] = v.bind(g)
    //   } else {
    //     Object.defineProperty(g2, k, {
    //       get() { return g[k] },
    //       set(v) { g[k] = v },
    //       enumerable: true,
    //     })
    //   }
    // }
    // return g2
  }

  // main() :Promise<void>
  main(wasmInstance) {
    // log("Host.main() wasm exports:")
    // for (let k in wasmInstance.exports) {
    //   log(`  ${k}`, wasmInstance.exports[k])
    // }
    this.wasmInstance = wasmInstance
    return this.go.run(wasmInstance) // Promise<void> resolved when main() exits
  }

  initEvents() {
    const h = this

    const jsEvent = (obj, name, handler) => ({
      ev: 0,
      _jsevent: true,
      handler,
      enable() { obj.addEventListener(name, this.handler) },
      disable() { obj.removeEventListener(name, this.handler) },
    })

    const persistentEvent = (ev, data) => ({
      ev,
      enable() { h.enablePersistentEvent(ev, data) },
      disable() { h.disablePersistentEvent(ev) },
    })

    // handlePointerEvent translates a PointerEvent into data that is sent to WASM.
    // Note: Changing this requires changes to host.go
    const handlePointerEvent = ev => [ ev.pointerId, ev.x*10, ev.y*10, ev.buttons ]

    // all supported events
    h.events = {
      [EVPointerMove]:  jsEvent(window, "pointermove", handlePointerEvent),
      [EVPointerDown]:  jsEvent(window, "pointerdown", handlePointerEvent),
      [EVPointerUp]:    jsEvent(window, "pointerup",   handlePointerEvent),
      [EVWindowResize]: jsEvent(window, "resize", ev => [window.innerWidth, window.innerHeight] ),

      [EVAnimationFrame]: persistentEvent(EVAnimationFrame),
    }
    for (let k in h.events) {
      let e = h.events[k]
      let ev = e.ev = parseInt(k)
      if (e._jsevent) {
        let handler = e.handler
        e.handler = (jsevent) => {
          h.eventEnqueue(ev, handler(jsevent))
        }
      }
    }
  }

  eventSubscribe(events) {
    log("host.eventSubscribe", events, EVString(events))
    for (let k in this.events) {
      let e = this.events[k]
      if (events & e.ev && !this.eventSubscriptions.has(e.ev)) {
        this.eventSubscriptions.add(e.ev)
        e.enable()
      }
    }
  }

  eventUnsubscribe(events) {
    log("host.eventUnsubscribe", events, EVString(events))
    for (let e of this.events) {
      if (events & e.ev && this.eventSubMask.has(e.ev)) {
        e.disable()
        this.eventSubscriptions.delete(e.ev)
        this.transientEvents.delete(e.ev)
        this.persistentEvents.delete(e.ev)
      }
    }
  }

  eventEnqueue(ev, data) {
    // log(`eventEnqueue ${EVString(ev)}`, data)
    assert(!data || data instanceof Array)
    this.transientEvents.set(ev, data)
    if (this.runloopWake) {
      this.runloopWakeA()
    }
  }

  enablePersistentEvent(ev, data) {
    this.persistentEvents.set(ev, data)
    this.persistentEventsVersion++
    if (this.runloopWake) {
      this.runloopWake()
    }
  }

  disablePersistentEvent(ev) {
    this.persistentEvents.delete(ev)
    this.persistentEventsVersion++
  }

  stopRunLoop() {
    log("Host.stopRunLoop")
    this.runloopTickFunc = null
  }

  startRunLoop(gocb, msgbufaddr, msgbufsize) {
    // Go caller provides a memory segment at msgbufaddr of msgbufsize size in bytes
    // where we write and read input and output data.
    //
    // Example
    // this.eventEnqueue(EVPointerMove, [123, 456])
    // this.eventEnqueue(EVPointerDown, [700, 800])
    //
    // uint32 data written to gomem:
    //  0  3     EVENT_MASK
    //  1  2     EVENT_COUNT
    //  2  2     event# EVPointerMove
    //  3  2     len(event_data) = 2
    //  4  123   event_data[0]
    //  5  456   event_data[1]
    //  6  3     event# EVPointerDown
    //  7  2     len(event_data) = 2
    //  8  700   event_data[0]
    //  9  800   event_data[1]
    log("Host.startRunLoop")
    if (this.runloopTickFunc) {
      throw new Error("runloop already running")
    }

    const timeOrigin = performance.now()

    // address/offset into gomem
    let a = msgbufaddr
    const TIME        = a
    const EVENT_COUNT = TIME + 8
    const EVENT_DATA  = EVENT_COUNT + 4 // until msgbufaddr + msgbufsize

    // number of p. events written in gomem. This needs to be tracked separately from
    // this.persistentEvents.size as this.persistentEvents may change while we write
    // event data to gomem
    let persistentEventInMemCount = 0

    // address in gomem where transient event data start (after persistent data)
    let transientEventDataAddr = 0

    // writes all persistent events to memory
    const writePersistentEvents = (mem) => {
      let addr = EVENT_DATA
      for (let [ev, data] of this.persistentEvents) {
        addr = writeEventData(mem, addr, ev, data)
      }
      // transient events follows persistent events in memory
      persistentEventInMemCount = this.persistentEvents.size
      transientEventDataAddr = addr
    }

    // Writes event and its optional data to mem at addr. Returns adjusted addr.
    const writeEventData = (mem, addr, ev, data) => {
      mem.setUint32(addr, ev) ; addr += 4
      mem.setUint32(addr, data ? data.length : 0) ; addr += 4
      if (data) for (let v of data) {
        mem.setUint32(addr, v) ; addr += 4
      }
      return addr
    }

    // memVersion is compared to gomem.version to discover when gomem was either
    // found initially or relocated.
    let memVersion = 0

    // tracks dirty state of this.persistentEvents by comparing to this.persistentEventsVersion
    let persistentEventsVersion = 0

    var tick = () => {
      if (this.runloopTickFunc !== tick) {
        // runloop stopped or restarted with different tick function
        return
      }

      let mem = this.gomem.check()

      if (mem.version != memVersion) {
        // gomem initialized or relocated
        writePersistentEvents(mem)
        memVersion = mem.version
        persistentEventsVersion = this.persistentEventsVersion
      } else if (this.persistentEventsVersion != persistentEventsVersion) {
        // persistent events changed
        persistentEventsVersion = this.persistentEventsVersion
        writePersistentEvents(mem)
      }

      // write input
      mem.setFloat64(TIME, (performance.now() - timeOrigin) / 1000)
      mem.setUint32(EVENT_COUNT, persistentEventInMemCount + this.transientEvents.size)

      // write transient events
      if (this.transientEvents.size > 0) {
        let addr = transientEventDataAddr
        for (let [ev, data] of this.transientEvents) {
          addr = writeEventData(mem, addr, ev, data)
        }
        // reset (in case gocb sets new events)
        this.transientEvents.clear()
      }

      // reset wake signal
      this.runloopWake = null

      // call into go
      gocb()

      if (this.persistentEvents.size > 0) {
        // there are still events queued
        requestAnimationFrame(tick)
      } else {
        this.runloopWake = tick
      }
    }

    this.runloopWakeA = () => {
      this.runloopWake = null
      requestAnimationFrame(tick)
    }

    this.runloopTickFunc = tick
    tick()
  }
}


// tear off initialization data from the global object
const appinit = window["_appinit"]
let wasmStartTime = 0
appinit["host"] = host
appinit["initCallback"] = () => {
  // called when the wasm program deems itself initialized
  const wasmTime = (performance.now() - wasmStartTime).toFixed(1)
  console.debug(`wasm initialized in ${wasmTime}ms`)
  delete window["_appinit"]
  if (appinit.onLoaded) { appinit.onLoaded() }
}
console.debug(`host boot ${Date.now()-appinit.time}ms`)

// load, instantiate and run wasm module
;(WebAssembly.instantiateStreaming ?
  WebAssembly.instantiateStreaming(appinit.wasmFetch, host.go.importObject) :
  appinit.wasmFetch.then(r => r.arrayBuffer()).then(buf =>
    WebAssembly.instantiate(buf, host.go.importObject))
).then(m => {
  wasmStartTime = performance.now()
  host.main(m.instance)
})
