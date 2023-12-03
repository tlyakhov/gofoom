package concepts

type ControllerMethod uint32

const (
	ControllerAlways ControllerMethod = 1 << iota
	ControllerContainment
	ControllerLoaded
	ControllerRecalculate
)

type Controller interface {
	Parent(*ControllerSet)
	Priority() int
	Methods() ControllerMethod
	// Return false if controller shouldn't run for this entity
	Target(source *EntityRef) bool
	Always()
	Containment()
	Loaded()
	Recalculate()
}

type BaseController struct {
	*ControllerSet
	TargetEntity *EntityRef
}

func (c *BaseController) Priority() int {
	return 100
}

func (c *BaseController) Methods() ControllerMethod {
	return 0
}

func (c *BaseController) Parent(s *ControllerSet) {
	c.ControllerSet = s
}

func (c *BaseController) Target(target *EntityRef) bool {
	c.TargetEntity = target
	return true
}

func (c *BaseController) Always()      {}
func (c *BaseController) Containment() {}
func (c *BaseController) Loaded()      {}
func (c *BaseController) Recalculate() {}
