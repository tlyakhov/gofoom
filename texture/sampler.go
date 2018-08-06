package texture

type ISampler interface {
	Sample(x, y float64, scale float64) uint32
}
