package doc

import "gpdf/doc/style"

const (
	A4Width       = style.A4Width
	A4Height      = style.A4Height
	A3Width       = style.A3Width
	A3Height      = style.A3Height
	A5Width       = style.A5Width
	A5Height      = style.A5Height
	LetterWidth   = style.LetterWidth
	LetterHeight  = style.LetterHeight
	LegalWidth    = style.LegalWidth
	LegalHeight   = style.LegalHeight
	TabloidWidth  = style.TabloidWidth
	TabloidHeight = style.TabloidHeight
)

func (b *DocumentBuilder) A4() *DocumentBuilder      { return b.PageSize(A4Width, A4Height) }
func (b *DocumentBuilder) A3() *DocumentBuilder      { return b.PageSize(A3Width, A3Height) }
func (b *DocumentBuilder) A5() *DocumentBuilder      { return b.PageSize(A5Width, A5Height) }
func (b *DocumentBuilder) Letter() *DocumentBuilder  { return b.PageSize(LetterWidth, LetterHeight) }
func (b *DocumentBuilder) Legal() *DocumentBuilder   { return b.PageSize(LegalWidth, LegalHeight) }
func (b *DocumentBuilder) Tabloid() *DocumentBuilder { return b.PageSize(TabloidWidth, TabloidHeight) }
func (b *DocumentBuilder) Landscape() *DocumentBuilder {
	b.pc.pageSize[0], b.pc.pageSize[1] = b.pc.pageSize[1], b.pc.pageSize[0]
	return b
}
