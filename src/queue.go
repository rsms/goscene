package main


type UintQueue struct {
  Items []uint
}

func (q *UintQueue) Len() int {
  return len(q.Items)
}

func (q *UintQueue) Front() uint {
  return q.Items[0]
}

func (q *UintQueue) Back() uint {
  return q.Items[len(q.Items)-1]
}

func (q *UintQueue) PushBack(v uint) {
  q.Items = append(q.Items, v)
}

func (q *UintQueue) PushFront(v uint) {
  q.Items = append(q.Items, v)
  copy(q.Items[1:], q.Items[:len(q.Items)-1])
  q.Items[0] = v
}

func (q *UintQueue) PopFront() uint {
  v := q.Items[0]
  q.Items = q.Items[1:]
  return v
}

func (q *UintQueue) PopBack() uint {
  i := len(q.Items)-1
  v := q.Items[i]
  q.Items = q.Items[:i]
  return v
}
