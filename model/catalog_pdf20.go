package model

// DPartRef returns the /DPart (document part hierarchy) reference if present (PDF 2.0).
func (c *Catalog) DPartRef() *Ref {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("DPart")].(Ref); ok {
		return &v
	}
	return nil
}

// DPartRootRef returns the /DPartRoot (document part root) reference if present (PDF 2.0).
func (c *Catalog) DPartRootRef() *Ref {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("DPartRoot")].(Ref); ok {
		return &v
	}
	return nil
}

// Extensions returns the /Extensions dictionary (developer extensions, ISO 32000-2
// §7.12), inline or nil. Each entry maps a prefix to a dictionary with /BaseVersion,
// /ExtensionLevel and, in PDF 2.0, an optional /URL.
func (c *Catalog) Extensions() Dict {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("Extensions")].(Dict); ok {
		return v
	}
	return nil
}

// OutputIntents returns the catalog /OutputIntents array inline, if present.
func (c *Catalog) OutputIntents() Array {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("OutputIntents")].(Array); ok {
		return v
	}
	return nil
}

// AFRef returns the /AF (associated files) array reference if present (PDF 2.0).
func (c *Catalog) AFRef() *Ref {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("AF")].(Ref); ok {
		return &v
	}
	return nil
}

// AssociatedFilesArray returns the /AF array inline (not as a reference) if present (PDF 2.0).
func (c *Catalog) AssociatedFilesArray() Array {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("AF")].(Array); ok {
		return v
	}
	return nil
}

// NamespacesRef returns the /Namespaces array reference if present (PDF 2.0).
func (c *Catalog) NamespacesRef() *Ref {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("Namespaces")].(Ref); ok {
		return &v
	}
	return nil
}

// PieceInfoRef returns the /PieceInfo (application private data) dictionary reference if present.
func (c *Catalog) PieceInfoRef() *Ref {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("PieceInfo")].(Ref); ok {
		return &v
	}
	return nil
}

// Version returns the /Version name (e.g. /1.4) overriding the header version, if present.
func (c *Catalog) Version() Name {
	if c == nil || c.Dict == nil {
		return ""
	}
	if v, ok := c.Dict[Name("Version")].(Name); ok {
		return v
	}
	return ""
}

// Lang returns the /Lang (natural language) string from the catalog if present (PDF 1.4+).
func (c *Catalog) Lang() String {
	if c == nil || c.Dict == nil {
		return ""
	}
	if v, ok := c.Dict[Name("Lang")].(String); ok {
		return v
	}
	return ""
}
