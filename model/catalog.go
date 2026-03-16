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
