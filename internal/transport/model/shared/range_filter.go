package shared

type RangeFilter[T comparable] struct {
	Min T `json:"min"`
	Max T `json:"max"`
}
