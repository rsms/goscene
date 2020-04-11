package main

import (
  "fmt"
  "math"
)

const PI = math.Pi

// func checkJSNotNull(v js.Value) {
// 	if v.IsNull() {
// 		panic("unexpected null js.Value")
// 	}
// 	if v.IsNull() || v.IsUndefined() {
// 		panic("unexpected undefined js.Value")
// 	}
// }

func logf(format string, v... interface{}) {
  fmt.Printf(format + "\n", v...)
}

func errorf(format string, v... interface{}) error {
  return fmt.Errorf(format, v...)
}

func panicf(format string, v... interface{}) {
  panic(fmt.Errorf(format, v...))
}

func max(x, y int) int {
  if x > y {
    return x
  }
  return y
}
