package core

var (
	EmptyFilter = NewFilter()
)

// This filter describes the ways to search for a service
type Filter struct {
	DescContains *string `json:"desc_contains,omitempty"`
	NameContains *string `json:"name_contains,omitempty"`
}

// Builds a filter from a list of builder functions
func NewFilter(fns ...func(*Filter)) (ret Filter) {
	for _, fn := range fns {
		fn(&ret)
	}
	return
}

func (f Filter) Update(fns ...func(*Filter)) (ret Filter) {
	ret = f
	for _, fn := range fns {
		fn(&ret)
	}
	return
}

// Returns a filter function that matches services by
// name. Any services whose name contains the match
// string will be returned.
func FilterByName(match string) func(*Filter) {
	return func(f *Filter) {
		f.NameContains = &match
	}
}

// Returns a filter function that matches services by
// name. Any services whose description contains the match
// string will be returned.
func FilterByDesc(match string) func(*Filter) {
	return func(f *Filter) {
		f.DescContains = &match
	}
}
