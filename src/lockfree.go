package main

import (
  "sync/atomic"
  "unsafe"
)

// From http://scottlobdell.me/2016/09/thread-safe-non-blocking-linked-lists-golang/

type ThreadSafeList struct {
  head *listNode
}

type listNode struct {
  markableNext *markablePointer
  value        unsafe.Pointer
}

type markablePointer struct {
  marked bool
  next   *listNode
}

func rrange(n *listNode, f func(unsafe.Pointer)bool) {
  if n != nil {
    rrange(n.markableNext.next, f)
    f(n.value)
  }
}

func (t *ThreadSafeList) Range(f func(unsafe.Pointer)bool) {
  rrange(t.head, f)
}

func (t *ThreadSafeList) RangeReverse(f func(unsafe.Pointer)bool) {
  n := t.head
  for n != nil {
    if !f(n.value) {
      break
    }
    n = n.markableNext.next
  }
}

func (t *ThreadSafeList) Insert(
  value unsafe.Pointer,
  lessThanFn func(a, b unsafe.Pointer) bool,
) {
  currentHeadAddress := &t.head
  currentHead := t.head

  if currentHead == nil || lessThanFn(value, currentHead.value) {
    newNode := listNode{
      value: value,
      markableNext: &markablePointer{
        next: currentHead,
      },
    }

    operationSucceeded := atomic.CompareAndSwapPointer(
      (*unsafe.Pointer)(unsafe.Pointer(currentHeadAddress)),
      unsafe.Pointer(currentHead),
      unsafe.Pointer(&newNode),
    )
    if !operationSucceeded {
      t.Insert(value, lessThanFn)
      return
    }
    return
  }

  cursor := t.head
  for {
    if cursor.markableNext.next == nil || lessThanFn(value, cursor.markableNext.next.value) {
      currentNext := cursor.markableNext
      if currentNext.marked {
        continue
      }
      newNode := listNode{
        value: value,
        markableNext: &markablePointer{
          next: currentNext.next,
        },
      }
      newNext := markablePointer{
        next: &newNode,
      }
      operationSucceeded := atomic.CompareAndSwapPointer(
        (*unsafe.Pointer)(unsafe.Pointer(&(cursor.markableNext))),
        unsafe.Pointer(currentNext),
        unsafe.Pointer(&newNext),
      )
      if !operationSucceeded {
        t.Insert(value, lessThanFn)
        return
      }
      break
    }
    cursor = cursor.markableNext.next
  }
}


func (t *ThreadSafeList) Delete(value unsafe.Pointer) (success bool) {
  var previous *listNode
  currentHeadAddress := &t.head
  currentHead := t.head
  cursor := currentHead
  for {
    if cursor == nil {
      break
    }
    if cursor.value == value {
      nextNode := cursor.markableNext.next
      newNext := markablePointer{
        marked: true,
        next:   nextNode,
      }
      operationSucceeded := atomic.CompareAndSwapPointer(
        (*unsafe.Pointer)(unsafe.Pointer(&(cursor.markableNext))),
        unsafe.Pointer(cursor.markableNext),
        unsafe.Pointer(&newNext),
      )
      if !operationSucceeded {
        return t.Delete(value)
      }
      newNext = markablePointer{
        next: nextNode,
      }
      if previous != nil {
        operationSucceeded = atomic.CompareAndSwapPointer(
          (*unsafe.Pointer)(unsafe.Pointer(&(previous.markableNext))),
          unsafe.Pointer(previous.markableNext),
          unsafe.Pointer(&newNext),
        )
      } else {
        // we just deleted head
        operationSucceeded = atomic.CompareAndSwapPointer(
          (*unsafe.Pointer)(unsafe.Pointer(currentHeadAddress)),
          unsafe.Pointer(currentHead),
          unsafe.Pointer(nextNode),
        )
      }
      if !operationSucceeded {
        success = t.Delete(value)
      } else {
        success = true
      }
      break
    }

    previous = cursor
    cursor = cursor.markableNext.next
  }
  return success
}
