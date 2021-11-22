package core

// A Page describes a range of a query result.  This is typically
// used for pagination where required.
//
// NOTE: Normally, this would likely be common across many different
// libraries needing transport/storage apis but leaving it here for
// simplicity
type Page struct {
	Offset  uint64 `json:"offset"`
	Limit   uint64 `json:"limit"`
	OrderBy string `json:"order_by"`
}

// Constructs a new page from a list of options
func NewPage(fns ...func(*Page)) (ret Page) {
	ret = Page{Limit: 1024, OrderBy: "name"}
	for _, fn := range fns {
		fn(&ret)
	}
	return
}

func (p Page) Update(fns ...func(*Page)) (ret Page) {
	ret = p
	for _, fn := range fns {
		fn(&ret)
	}
	return
}

// Returns a page option that sets the limit of results
func Limit(num uint64) func(*Page) {
	return func(o *Page) {
		o.Limit = num
	}
}

// Returns a page option that sets the offset of the result batch
func Offset(num uint64) func(*Page) {
	return func(o *Page) {
		o.Offset = num
	}
}

// Returns a page option that sets the order by field
func OrderBy(field string) func(*Page) {
	return func(o *Page) {
		o.OrderBy = field
	}
}
