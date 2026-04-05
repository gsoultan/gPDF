package reader

// Table is a grid of TableCells detected from aligned text blocks on a page.
type Table struct {
	Page   int
	Rows   int
	Cols   int
	X      float64
	Y      float64
	Width  float64
	Height float64
	Cells  []TableCell
	// Border properties detected from surrounding vector shapes
	HasBorder   bool
	BorderColor ColorRGB
	BorderWidth float64
}

// Cell returns the text of the cell at row r, column c, or "" if absent.
func (t *Table) Cell(r, c int) string {
	for _, cell := range t.Cells {
		if cell.Row == r && cell.Col == c {
			return cell.Text
		}
	}
	return ""
}

// CellStyle returns the TextStyle of the cell at row r, column c.
// Returns a zero-value TextStyle if the cell is absent.
func (t *Table) CellStyle(r, c int) TextStyle {
	for _, cell := range t.Cells {
		if cell.Row == r && cell.Col == c {
			return cell.Style
		}
	}
	return TextStyle{}
}
