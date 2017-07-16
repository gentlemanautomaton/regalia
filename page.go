package regalia

// TODO: Consider naming this a "Block" instead.

// Page is a page of data. Its size may vary.
//
// TODO: Store hash at beginning or end of page?
type Page []byte

// Header returns the header for the page, if it has one.
func (p Page) Header() PageHeader {
	return PageHeader{}
}

// Data returns the data contained in the page.
func (p Page) Data() []byte {
	return nil
}

// PageHeader is header data for a page.
type PageHeader struct {
	// FIXME: Store offset in variable length encoding.
	DataOffset uint8 // Byte offset to the start of the data.
	// TODO: Add type information
}

// PageSet represents a set of data pages that store a particular data series.
//type PageSet
