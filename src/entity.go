package main

import "fmt"

type entint = uint32
const entintSize uint = 32 << (^entint(0) >> 63)
const entintMax entint = ^entint(0)

// The idea here is that the index part directly gives us the index of the entity
// in a lookup array. The generation part is used to distinguish entities created
// at the same index slot. As we create and destroy entities we will at some point
// have to reuse an index in the array. By changing the generation value when that
// happens we ensure that we still get a unique ID.
const (
  EntityGenerationBits = entint(8)
  EntityGenerationMask = entint((1 << EntityGenerationBits) - 1)
  EntityIndexBits = entint(entintSize) - EntityGenerationBits
  EntityIndexMask = entint((1 << EntityIndexBits) - 1)
)

// We've split up our 32 bits into 24 bits for the index and 8 bits for the
// generation. This means that we support a maximum of 16 million simultaneous
// entities (2^24). It also means that we can only distinguish between 256 different
// entities created at the same index slot. If more than 256 entities are created
// at the same index slot, the generation value will wrap around and our new entity
// will get the same ID as an old entity.
//
// To prevent that from happening too often we need to make sure that we don't
// reuse the same index slot too often. There are various possible ways of doing
// that. Our solution is to put recycled indices in a queue and only reuse values
// from that queue when it contains at least MinimumFreeIndices = 1024 items.
// Since we have 256 generations, an ID will never reappear until its index has run
// 256 laps through the queue. So this means that you must create and destroy at
// least 256 * 1024 entities until an ID can reappear. This seems reasonably safe,
// but if you want you can play with the numbers to get different margins. For
// example, if you don't need 16 M entities, you can steal some bits from index and
// give to generation.
//
const MinimumFreeIndices = 1024



type Ent entint
const NilEnt = Ent(0)

func (e Ent) Index() entint {
  return entint(e) & EntityIndexMask
}

func (e Ent) Generation() uint8 {
  return uint8((entint(e) >> EntityIndexBits) & EntityGenerationMask)
}

func (e Ent) String() string {
  return fmt.Sprintf("Ent#%X", entint(e))
}

func (e Ent) GoString() string {
  return fmt.Sprintf("Ent{%d,%d}", e.Index(), e.Generation())
}

func createEnt(index entint, generation uint8) Ent {
  return Ent(entint(generation) << EntityIndexBits | index)
}



type EntManager struct {
  generation  []byte
  freeIndices EntQueue  // 1-based as Ent#0 == NilEnt
}

func (em *EntManager) Init() {
  // pre-allocate freeIndices since we will use at least that many slots
  em.generation = make([]byte, 1, MinimumFreeIndices + 1)  // 1-based
  em.freeIndices.Items = make([]Ent, 0, MinimumFreeIndices + 1)[:]
}

func (em *EntManager) Alloc() Ent {
  var index entint
  if em.freeIndices.Len() >= MinimumFreeIndices {
   index = entint(em.freeIndices.PopFront())
  } else {
   em.generation = append(em.generation, 0)
   index = entint(len(em.generation)) - 1
   // assert(index < (1 << EntityIndexBits))
  }
  return createEnt(index, em.generation[index])
}

func (em *EntManager) Free(e Ent) {
  index := e.Index()
  em.generation[index]++
  em.freeIndices.PushBack(Ent(index))
}

func (em *EntManager) IsAlive(e Ent) bool {
  return entint(e) != 0 && em.generation[e.Index()] == e.Generation()
}



// double-ended queue
// TODO: more efficient implementation, esp. for PopFront()
type EntQueue struct {
  Items []Ent
}

func (q *EntQueue) Len() int {
  return len(q.Items)
}

func (q *EntQueue) Front() Ent {
  return q.Items[0]
}

func (q *EntQueue) Back() Ent {
  return q.Items[len(q.Items)-1]
}

func (q *EntQueue) PushBack(v Ent) {
  q.Items = append(q.Items, v)
}

func (q *EntQueue) PushFront(v Ent) {
  q.Items = append(q.Items, v)
  copy(q.Items[1:], q.Items[:len(q.Items)-1])
  q.Items[0] = v
}

func (q *EntQueue) PopFront() Ent {
  v := q.Items[0]
  q.Items = q.Items[1:]
  return v
}

func (q *EntQueue) PopBack() Ent {
  i := len(q.Items)-1
  v := q.Items[i]
  q.Items = q.Items[:i]
  return v
}
