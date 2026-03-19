package doc

import "gpdf/doc/pagesize"

// Standard page sizes shortcuts for DocumentBuilder.

// A4 sets the page size to A4 (210 x 297 mm).
func (b *DocumentBuilder) A4() *DocumentBuilder { return b.ApplyPageSize(pagesize.A4) }

// A3 sets the page size to A3 (297 x 420 mm).
func (b *DocumentBuilder) A3() *DocumentBuilder { return b.ApplyPageSize(pagesize.A3) }

// A5 sets the page size to A5 (148 x 210 mm).
func (b *DocumentBuilder) A5() *DocumentBuilder { return b.ApplyPageSize(pagesize.A5) }

// Letter sets the page size to Letter (8.5 x 11 in).
func (b *DocumentBuilder) Letter() *DocumentBuilder { return b.ApplyPageSize(pagesize.Letter) }

// Legal sets the page size to Legal (8.5 x 14 in).
func (b *DocumentBuilder) Legal() *DocumentBuilder { return b.ApplyPageSize(pagesize.Legal) }

// Tabloid sets the page size to Tabloid (11 x 17 in).
func (b *DocumentBuilder) Tabloid() *DocumentBuilder { return b.ApplyPageSize(pagesize.Tabloid) }

// Executive sets the page size to Executive (7.25 x 10.5 in).
func (b *DocumentBuilder) Executive() *DocumentBuilder { return b.ApplyPageSize(pagesize.Executive) }

// Statement sets the page size to Statement (5.5 x 8.5 in).
func (b *DocumentBuilder) Statement() *DocumentBuilder { return b.ApplyPageSize(pagesize.Statement) }

// B4JIS sets the page size to B4 (JIS) (257 x 364 mm).
func (b *DocumentBuilder) B4JIS() *DocumentBuilder { return b.ApplyPageSize(pagesize.B4JIS) }

// B5JIS sets the page size to B5 (JIS) (182 x 257 mm).
func (b *DocumentBuilder) B5JIS() *DocumentBuilder { return b.ApplyPageSize(pagesize.B5JIS) }

// Folio sets the page size to Folio (8.5 x 13 in).
func (b *DocumentBuilder) Folio() *DocumentBuilder { return b.ApplyPageSize(pagesize.Folio) }

// Quarto sets the page size to Quarto (215 x 275 mm).
func (b *DocumentBuilder) Quarto() *DocumentBuilder { return b.ApplyPageSize(pagesize.Quarto) }

// Note sets the page size to Note (8.5 x 11 in).
func (b *DocumentBuilder) Note() *DocumentBuilder { return b.ApplyPageSize(pagesize.Note) }

// Ledger sets the page size to Ledger (17 x 11 in).
func (b *DocumentBuilder) Ledger() *DocumentBuilder { return b.ApplyPageSize(pagesize.Ledger) }

// Landscape sets the current page size to landscape orientation.
func (b *DocumentBuilder) Landscape() *DocumentBuilder {
	if b.pc.pageSize[0] < b.pc.pageSize[1] {
		b.pc.pageSize[0], b.pc.pageSize[1] = b.pc.pageSize[1], b.pc.pageSize[0]
	}
	return b
}

// Portrait sets the current page size to portrait orientation.
func (b *DocumentBuilder) Portrait() *DocumentBuilder {
	if b.pc.pageSize[0] > b.pc.pageSize[1] {
		b.pc.pageSize[0], b.pc.pageSize[1] = b.pc.pageSize[1], b.pc.pageSize[0]
	}
	return b
}
