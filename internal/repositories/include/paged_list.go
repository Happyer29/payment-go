package include

type PagedResultsList[T interface{}] struct {
	Total uint
	Items []*T
}
