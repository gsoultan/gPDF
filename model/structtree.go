package model

// StructTreeRoot is the structure tree root (Catalog /StructTreeRoot) for tagged PDF.
// Keys: /Type (/StructTreeRoot), /K (kids), /ParentTree, /RoleMap.
type StructTreeRoot struct {
	Dict Dict
}

// K returns the immediate children (structure elements) of the tree root.
func (s *StructTreeRoot) K() Array {
	if s == nil || s.Dict == nil {
		return nil
	}
	if v, ok := s.Dict[Name("K")].(Array); ok {
		return v
	}
	return nil
}

// ParentTreeRef returns the reference to the parent tree (number tree) if present.
func (s *StructTreeRoot) ParentTreeRef() *Ref {
	if s == nil || s.Dict == nil {
		return nil
	}
	if v, ok := s.Dict[Name("ParentTree")].(Ref); ok {
		return &v
	}
	return nil
}

// RoleMap returns the /RoleMap dictionary if present.
func (s *StructTreeRoot) RoleMap() Dict {
	if s == nil || s.Dict == nil {
		return nil
	}
	if v, ok := s.Dict[Name("RoleMap")].(Dict); ok {
		return v
	}
	return nil
}
