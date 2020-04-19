package main

import "fmt"


func init() {
  w := &World{}
  w.Init()

  player  := w.Ents.Alloc() ; logf("player  %v", player)
  gun     := w.Ents.Alloc() ; logf("gun     %v", gun)
  monster := w.Ents.Alloc() ; logf("monster %v", monster)
  bullet  := w.Ents.Alloc() ; logf("bullet  %v", bullet)

  movement := MovementSystem{}
  movement.Init()
  movement.Assoc(player,  MovementData{ mass: 1.0, position: Vec2{0.0, 0.0} })
  movement.Assoc(monster, MovementData{ mass: 1.2, position: Vec2{10.0, 10.0} })
  movement.Assoc(bullet,  MovementData{
    mass: 0.1,
    position: Vec2{0.0, 0.0},
    velocity: Vec2{0.001, 0.001},
  })

  playerNode := w.TransformSystem.CreateNode(player, Matrix4Identity)
  gunNode := w.TransformSystem.CreateNode(gun, Matrix4Identity)
  playerNode.AppendChild(gunNode)
  w.TransformSystem.Update(0.0)

  gunNode.SetLocal(&gunNode.local)

  w.TransformSystem.Update(0.0)

  // movement.Update(1.0 /*seconds*/ )
  // logf("bullet movement %v", movement.Get(bullet))

  // movement.Update(1.1)
  // logf("bullet movement %v", movement.Get(bullet))

  // host.events.Listen(EVAnimationFrame, func (_ Event, _ ...uint32) {
  //   movement.Update(Monotime())
    // logf("bullet movement %v", movement.Get(bullet))
  // })
}


type TransformSystem struct {
  world *World
  nodes []TransformNode
  dirty []*TransformNode
  m     map[Ent]*TransformNode
}

func (s *TransformSystem) Init(world *World) {
  s.world = world
  s.m = make(map[Ent]*TransformNode)
}

func (s *TransformSystem) CreateNode(ent Ent, local Matrix4) *TransformNode {
  s.nodes = append(s.nodes, TransformNode{
    system: s,
    ent: ent,
    local: local,
    dirty: true,
  })
  n := &s.nodes[len(s.nodes)-1]
  s.m[ent] = n
  s.markDirty(n)
  return n
}

func (s *TransformSystem) Get(ent Ent) *TransformNode {
  return s.m[ent]
}

func (s *TransformSystem) Update(time float64) {
  dirty := s.dirty // just in case some code we run here would call markDirty
  for _, n := range dirty {
    if n.dirty {
      n.computeAbsoluteTransform()
    }
  }
  assert(len(dirty) == len(s.dirty)) // or something called markDirty
  s.dirty = s.dirty[:0] // truncate
}

func (s *TransformSystem) markDirty(n *TransformNode) {
  s.dirty = append(s.dirty, n)
}

// -----------------------------------------------------------------------------

type TransformNode struct {
  system      *TransformSystem // Owning system pointer
  ent         Ent              // The entity owning this instance
  absolute    Matrix4          // Absolute "world" transform
  local       Matrix4          // Local transform relative to parent
  parent      *TransformNode   // The parent instance of this instance
  firstChild  *TransformNode   // The first child of this instance
  nextSibling *TransformNode   // The next sibling of this instance
  prevSibling *TransformNode   // The previous sibling of this instance
  dirty       bool             // true when local and absolute are not in sync
}

func (n *TransformNode) AppendChild(child *TransformNode) {
  child.nextSibling = nil
  child.parent = n
  if n.firstChild == nil {
    n.firstChild = child
    child.prevSibling = nil
  } else {
    // Note: We could add & maintain a lastChild pointer to TransformNode in order to
    // trade memory for speed if AppendChild turns out to be a frequent operation.
    // However, note that looping through children is really efficient as they are all
    // packed together in memory.
    lastChild := n.firstChild
    for {
      if lastChild.nextSibling == nil {
        lastChild.nextSibling = n
        n.prevSibling = lastChild
        break
      }
    }
  }
  n.markDirty()
}

func (n *TransformNode) markDirty() {
  if !n.dirty {
    n.dirty = true
    n.system.markDirty(n)
  }
}

func (n *TransformNode) SetLocal(m *Matrix4) {
  n.local = *m
  n.dirty = true
  n.system.markDirty(n)
}

func (n *TransformNode) Translate(x, y, z float32) {
  n.local.TranslateMut(x, y, z)
  n.markDirty()
}

func (n *TransformNode) computeAbsoluteTransform() {
  logf("computeAbsoluteTransform of %v", n.ent)
  // compute absolute "world" transform
  if n.parent != nil {
    n.absolute = n.local.Mul4(&n.parent.absolute)
  } else {
    n.absolute = n.local
  }

  // traverse all children to update their transforms
  child := n.firstChild
  for child != nil {
    child.computeAbsoluteTransform()
    child = child.nextSibling
  }

  n.dirty = false
}


// -------------------------------------------------------------------------------

type MovementSystem struct {
  data           []MovementData
  m              map[Ent]int
  lastUpdateTime float64
}

func (s *MovementSystem) Init() {
  s.m = make(map[Ent]int)
}

func (s *MovementSystem) Assoc(ent Ent, data MovementData) {
  index := len(s.data)
  s.m[ent] = index
  s.data = append(s.data, data)
}

func (s *MovementSystem) Get(ent Ent) *MovementData {
  index := s.m[ent]
  return &s.data[index]
}

func (s *MovementSystem) Update(time float64) {
  dt := float32(time - s.lastUpdateTime)
  s.lastUpdateTime = time

  for i := 0; i < len(s.data); i++ {
    data := &s.data[i]
    data.velocity = data.velocity.Add(data.acceleration.Mul(dt))
    data.position = data.position.Add(data.velocity.Mul(dt))
  }
}


type MovementData struct {
  mass         float32
  position     Vec2
  velocity     Vec2
  acceleration Vec2
}

func (d MovementData) String() string {
  return fmt.Sprintf("{mass %v, pos %v, vel %v, accel %v}",
    d.mass,
    d.position,
    d.velocity,
    d.acceleration,
  )
}



type World struct {
  Time float64
  Ents EntManager

  TransformSystem
}

func (w *World) Init() {
  w.Ents.Init()
  w.TransformSystem.Init(w)
}
