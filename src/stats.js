const log = console.log.bind(console)


export class Float64RingBuffer {
  constructor(size) {
    this.buf = new Float64Array(size)
    this.start = 0
    this.end = 0
    this.length = 0
  }
  reset(size) {
    this.buf = new Float64Array(size)
    this.start = 0
    this.end = 0
    this.length = 0
  }
  push(value) {
    let i = this.end
    if (this.length == this.buf.length) {
      this.start++
      i = i % this.length
    } else {
      this.length++
    }
    this.end++
    this.buf[i] = value
  }
  sum() { return this.buf.reduce((a, v) => a + v, 0.0) }
  avg() { return this.length > 0 ? this.sum() / this.length : 0 }
  toString() { return "[" + this.buf.subarray(0, this.length).join(" ") + "]" }
}


export class StatPanel {
  constructor(name, unit, initialValue) {
    this.name = name
    this.data = document.createTextNode(initialValue || "-")
    this.unit = document.createTextNode(unit)
    this.el = document.createElement("div")
    this.el.className = name
    this.el.appendChild(this.data)
    this.el.appendChild(this.unit)
    this.frameSamples = null
    this.lastRefreshTime = 0
    this.updateInterval = 100  // ms
    this.needsFrameSamples = true  // set to false to disable
  }
  set value(v) {
    this.data.nodeValue = v
  }
  resetFrameSamples(size) {
    if (this.needsFrameSamples) {
      this.frameSamples = new Float64RingBuffer(size)
    }
  }

  _update(time, dt) {
    if (this.needsFrameSamples) {
      this.onFrame(time, dt)
    }
    // called just before update() by Stats. Not overridden by subclasses.
    if (this.lastRefreshTime == 0) {
      if (!this.needsFrameSamples) {
        this.update(time)
      }
      this.lastRefreshTime = time
    } else if (time - this.lastRefreshTime >= this.updateInterval) {
      this.update(time)
      this.lastRefreshTime = time
    }
  }

  // for subclasses to implement
  update(time) {}      // called every updateInterval milliseconds; should update its data.
  onFrame(time, dt) {} // called for every frame if needsFrameSamples is true.
}


export class FPSPanel extends StatPanel {
  constructor() {
    super("fps", " FPS")
  }
  onFrame(time, dt) {
    this.frameSamples.push(dt)
  }
  update() {
    this.value = Math.round(1000 / this.frameSamples.avg())
  }
}


export class Stats {
  constructor() {
    this.rootElement = document.createElement("div")
    this.rootElement.className = "webglstats"
    this.panels = []
    this.lastTime = -1.0

    // First, estimate the nominal refresh rate, e.g. 60Hz.
    // This is used to configure sliding window of frameSamples to rougly one frame
    // worth of samples. This approach also has a nice side effect in that it delays
    // the actual FPS sampling by a few hundred milliseconds which _usually_ avoids
    // the one or two really long frames occuring during program initialization, which
    // would skew the first second of samples. It's not very important though and we
    // wouldn't write code specifically to avoid this, but is a free side effect :-)
    this.nominalRefreshRate = 0
    estimateNominalRefreshRate(20).then(hz => {
      this.nominalRefreshRate = Math.ceil(hz)
      for (let panel of this.panels) {
        panel.resetFrameSamples(this.nominalRefreshRate)
      }
    })

    // FPS
    this.addPanel(new FPSPanel())

    // JS memory
    if (self.performance && self.performance.memory) {
      const m = self.performance.memory
      // m.totalJSHeapSize: 23173485
      // m.usedJSHeapSize: 21401113
      // m.jsHeapSizeLimit: 4294705152
      this.addGenericPanel(() => `JS ${fmtByteSize(m.usedJSHeapSize)}`)
    }
  }

  mount(parentElement) {
    parentElement.appendChild(this.rootElement)
  }

  unmount() {
    this.rootElement.parent.removeChild(rootElement)
  }

  addPanel(panel) {
    this.rootElement.appendChild(panel.el)
    this.panels.push(panel)
    return panel
  }

  // props override properties of StatPanel
  //
  // addGenericPanel(props: {...}, templateFunc :(time:number)=>void)
  // addGenericPanel(templateFunc :(time:number)=>void)
  addGenericPanel(props, templateFunc) {
    if (!templateFunc) {
      templateFunc = props
      props = {}
    }
    this.addPanel(new class extends StatPanel {
      constructor() {
        super("generic", "")
        this.needsFrameSamples = false
        for (let k in props) {
          this[k] = props[k]
        }
      }
      update(time) {
        this.data.nodeValue = templateFunc(time)
      }
    })
  }

  update(time) {
    // if (this.needsFrameSamples)
    let isSamplingRefreshRate = this.nominalRefreshRate == 0
    if (this.lastTime >= 0.0) {
      let dt = time - this.lastTime
      for (let panel of this.panels) {
        if (!panel.needsFrameSamples || !isSamplingRefreshRate) {
          panel._update(time, dt)
        }
      }
    }
    this.lastTime = time
  }
}


function estimateNominalRefreshRate(samples) {
  return new Promise(resolve => {
    let min = 1e10, lastTime = -1.0
    var update = time => {
      if (lastTime >= 0.0) {
        min = Math.min(min, time - lastTime)
        if (samples-- < 1) {
          return resolve(1000/min)
        }
      }
      lastTime = time
      requestAnimationFrame(update)
    }
    requestAnimationFrame(update)
  })
}


export function fmtByteSize(z) {
  if (z < 1024)           { return `${z} B` }
  if (z < 1024*1024)      { return `${(z/1024).toFixed(1)} kB` }
  if (z < 1024*1024*1024) { return `${(z/(1024*1024)).toFixed(1)} MB` }
  return `${(z/1024).toFixed(2)}GB`
}
