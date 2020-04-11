// +build js,wasm
// Single-threaded implementation

package main

import "reflect"


type EventListener func(ev Event, data ...uint32)


type EventSource struct {
  Enable   func(Event) // called when a first item is added
  Disable  func(Event) // called when the last item is removed
  m        map[Event][]eventSourceEntry
}

type eventSourceEntry struct {
  f EventListener
  p uintptr
}

func (s *EventSource) Len(ev Event) int {
  l := s.m[ev]
  if l == nil {
    return 0
  }
  return len(l)
}


func (s *EventSource) Listen(ev Event, f EventListener) EventListener {
  // logf("EventSource.Listen %s", ev)
  p := reflect.ValueOf(f).Pointer()
  if s.m == nil {
    s.m = make(map[Event][]eventSourceEntry)
  }
  l := s.m[ev]
  if l == nil {
    l = []eventSourceEntry{ eventSourceEntry{ f, p } }
    s.m[ev] = l
  } else {
    if s.findListener(l, p) != -1 {
      return f
    }
    l = append(l, eventSourceEntry{ f, p })
    s.m[ev] = l
  }
  if len(l) == 1 && s.Enable != nil {
    s.Enable(ev)
  }
  return f
}


func (s *EventSource) IsListening(ev Event, f EventListener) bool {
  if s.m == nil {
    return false
  }
  l := s.m[ev]
  if l == nil {
    return false
  }
  return s.findListener(l, reflect.ValueOf(f).Pointer()) != -1
}

func (s *EventSource) findListener(l []eventSourceEntry, p uintptr) int {
  for i, e := range l {
    if e.p == p {
      return i
    }
  }
  return -1
}


func (s *EventSource) StopListening(ev Event, f EventListener) bool {
  if s.m == nil {
    return false
  }
  l := s.m[ev]
  if l == nil {
    return false
  }
  p := reflect.ValueOf(f).Pointer()
  i := s.findListener(l, p)
  if i == -1 {
    return false
  }
  if len(l) == 1 {
    s.m[ev] = l[0:0]
    if s.Disable != nil {
      s.Disable(ev)
    }
  } else {
    s.m[ev] = append(l[:i], l[i+1:]...)
  }
  return true
}


func (s *EventSource) Trigger(ev Event, data ...uint32) {
  // logf("EventSource.Trigger %d %s", ev, ev)
  if s.m != nil {
    l := s.m[ev]
    if l != nil {
      for _, e := range l {
        e.f(ev, data...)
      }
    }
  }
}
