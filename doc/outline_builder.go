package doc

// OutlineBuilder manages bookmarks, named destinations, and link annotations.
type OutlineBuilder interface {
	AddOutline(title string, pageIndex int) *DocumentBuilder
	AddOutlineURL(title string, url string) *DocumentBuilder
	AddOutlineToDest(title string, destName string) *DocumentBuilder
	AddNamedDest(name string, pageIndex int) *DocumentBuilder
	AddLinkToPage(pageIndex int, llx, lly, urx, ury float64, destPageIndex int) *DocumentBuilder
	AddLinkToDest(pageIndex int, llx, lly, urx, ury float64, destName string) *DocumentBuilder
	AddLinkToURL(pageIndex int, llx, lly, urx, ury float64, url string) *DocumentBuilder
}
