package model

// Viewports returns the page's /VP array (viewports with measurement data, ISO
// 32000-2 §12.10). Entries are usually inline dictionaries.
func (p Page) Viewports() Array {
	if p.Dict == nil {
		return nil
	}
	if v, ok := p.Dict[Name("VP")].(Array); ok {
		return v
	}
	return nil
}

// Collection returns the catalog /Collection dictionary (portable collection) when
// it is inline, else nil; see CollectionRef.
func (c *Catalog) Collection() Dict {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("Collection")].(Dict); ok {
		return v
	}
	return nil
}

// CollectionRef returns the /Collection reference if present.
func (c *Catalog) CollectionRef() *Ref {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("Collection")].(Ref); ok {
		return &v
	}
	return nil
}
