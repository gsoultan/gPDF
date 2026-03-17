package model

// Catalog is the root of the document (PDF Catalog dictionary).
// Keys include /Type (/Catalog), /Pages, /Metadata, /Outlines, etc.
type Catalog struct {
	Dict Dict
}

// MetadataRef returns the /Metadata stream reference (XMP) if present.
func (c *Catalog) MetadataRef() *Ref {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("Metadata")].(Ref); ok {
		return &v
	}
	return nil
}

// OutlinesRef returns the /Outlines (document outline/bookmarks) reference if present.
func (c *Catalog) OutlinesRef() *Ref {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("Outlines")].(Ref); ok {
		return &v
	}
	return nil
}

// DestsRef returns the /Dests (named destinations) dictionary reference if present.
func (c *Catalog) DestsRef() *Ref {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("Dests")].(Ref); ok {
		return &v
	}
	return nil
}

// AcroFormRef returns the /AcroForm (interactive form) dictionary reference if present.
func (c *Catalog) AcroFormRef() *Ref {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("AcroForm")].(Ref); ok {
		return &v
	}
	return nil
}

// OCPropertiesRef returns the /OCProperties (optional content) dictionary reference if present.
func (c *Catalog) OCPropertiesRef() *Ref {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("OCProperties")].(Ref); ok {
		return &v
	}
	return nil
}

// MarkInfoRef returns the /MarkInfo (tagged PDF) dictionary reference if present.
func (c *Catalog) MarkInfoRef() *Ref {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("MarkInfo")].(Ref); ok {
		return &v
	}
	return nil
}

// StructTreeRootRef returns the /StructTreeRoot (structure tree) reference if present.
func (c *Catalog) StructTreeRootRef() *Ref {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("StructTreeRoot")].(Ref); ok {
		return &v
	}
	return nil
}

// PermsRef returns the /Perms (document permissions, e.g. DocMDP for signatures) reference if present.
func (c *Catalog) PermsRef() *Ref {
	if c == nil || c.Dict == nil {
		return nil
	}
	if v, ok := c.Dict[Name("Perms")].(Ref); ok {
		return &v
	}
	return nil
}
