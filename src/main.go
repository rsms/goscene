package main

import (
  "syscall/js"
  "fmt"
)

func main() {
  defer func() {
    if r := recover(); r != nil {
      host.error(fmt.Sprintf("%v", r))
    } else {
      // stop go program from exiting by waiting on a channel forever
      <- make(chan bool)
    }
  }()

  // create renderer
  canvasHtmlElement := js.Global().Get("document").Call("querySelector", "canvas")
  r, err := NewRenderer(canvasHtmlElement, host.windowWidth, host.windowHeight, host.pixelRatio)
  if err != nil {
    panic(err)
  }
  r.init()

  // resize canvas when window size changes
  host.events.Listen(EVWindowResize, func (ev Event, xy ...uint32) {
    logf("window resized %d, %d, %f", host.windowWidth, host.windowHeight, host.pixelRatio)
    r.setSize(host.windowWidth, host.windowHeight, host.pixelRatio)
  })

  // signal to host that initialization is complete
  js.Global().Get("_appinit").Call("initCallback")

  host.RunLoop()
}


func hostRunloop(r *Renderer) {
  host.jsv.Call("runloop", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    // time := float32(args[0].Float() / 1000.0)
    // r.render(time)
    return nil  // return js.False() to stop animation
  }))
}

