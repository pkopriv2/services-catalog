package core

import uuid "github.com/satori/go.uuid"

var (
	EmptyFilter = NewFilter()
)

// This filter describes the ways to search for a service
type Filter struct {
	DescContains *string    `json:"desc_contains,omitempty"`
	NameContains *string    `json:"name_contains,omitempty"`
	ServiceId    *uuid.UUID `json:"service_id,omitempty"`
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

// Returns a filter function that matches services by id.
func FilterByServiceId(id uuid.UUID) func(*Filter) {
	return func(f *Filter) {
		f.ServiceId = &id
	}
}

// Returns a filter function that matches services by name.
func FilterByName(match string) func(*Filter) {
	return func(f *Filter) {
		f.NameContains = &match
	}
}

// Returns a filter function that matches services by description.
func FilterByDesc(match string) func(*Filter) {
	return func(f *Filter) {
		f.DescContains = &match
	}
}
