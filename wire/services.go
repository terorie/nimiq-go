package wire

// Service type IDs.
const (
	ServicesNone = uint32(0)
	ServicesNano = uint32(1 << iota)
	ServicesLight
	ServicesFull
)
