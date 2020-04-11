// +build !js,!wasm
// Multi-threaded implementation

package main

import (
  "reflect"
  "sync/atomic"
  "unsafe"
)


type EventSource struct {
  Enable func()  // called when a first item is added
  Disable func() // called when the last item is removed
  l ThreadSafeList
  size int32
}

func (s *EventSource) Len() int {
  return int(atomic.LoadInt32(&s.size))
}

func (s *EventSource) Listen(f func()) func() {
  s.l.Insert(unsafe.Pointer(&f), func(a, b unsafe.Pointer) bool {
    // Returning false makes is so that adding items adds them to the end of the list.
    // This increases time complexity of inserts but speeds up iteration.
    return false
  })
  if atomic.AddInt32(&s.size, 1) == 1 && s.Enable != nil {
    s.Enable()
  }
  return f
}

func (s *EventSource) IsListening(f func()) bool {
  p := reflect.ValueOf(f).Pointer()
  n := s.l.head
  for n != nil {
    f := *(*func())(n.value)
    if reflect.ValueOf(f).Pointer() == p {
      return true
    }
    n = n.markableNext.next
  }
  return false
}

func (s *EventSource) StopListening(f func()) bool {
  if !s.l.Delete(unsafe.Pointer(&f)) {
    return false
  }
  if atomic.AddInt32(&s.size, -1) == 0 && s.Disable != nil {
    s.Disable()
  }
  return true
}

func (s *EventSource) Trigger() {
  n := s.l.head
  for n != nil {
    f := *(*func())(n.value)
    f()
    n = n.markableNext.next
  }
}
