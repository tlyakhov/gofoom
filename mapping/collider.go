package mapping

type ICollider interface {
	Collide() []*Segment
}
