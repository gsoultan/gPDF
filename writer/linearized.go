package writer

import (
	"fmt"
	"io"
	"sort"

	"gpdf/model"
)

// WriteLinearized writes a linearized (fast web view) PDF to ws.
// The writer must be seekable so that the linearization parameter dictionary (object 0)
// and the first-page trailer can be updated with correct offsets after the rest is written.
// Linearization places first-page objects and a first xref/trailer at the start so that
// a viewer can display page 1 before downloading the full file.
func (pw *PDFWriter) WriteLinearized(ws WriteSeeker, doc Document) error {
	allNums := doc.ObjectNumbers()
	if len(allNums) == 0 {
		return fmt.Errorf("document has no objects")
	}
	sort.Ints(allNums)
	firstSet, err := pw.firstPageObjectSet(doc)
	if err != nil {
		return err
	}
	var firstPageNums []int
	for _, n := range allNums {
		if firstSet[n] {
			firstPageNums = append(firstPageNums, n)
		}
	}
	sort.Ints(firstPageNums)
	var restNums []int
	for _, n := range allNums {
		if !firstSet[n] {
			restNums = append(restNums, n)
		}
	}
	// First page object number (for /O): the Page dict that is the first kid of Pages
	firstPageObjNum := 0
	if len(firstPageNums) > 0 {
		firstPageObjNum = firstPageNums[len(firstPageNums)-1] // heuristic: often the last in dependency order
		root := doc.Trailer().Root()
		if root != nil {
			cat, _ := doc.Resolve(model.Ref{ObjectNumber: root.ObjectNumber, Generation: 0})
			if catDict, ok := cat.(model.Dict); ok {
				if pagesRef, ok := catDict[model.Name("Pages")].(model.Ref); ok {
					pagesObj, _ := doc.Resolve(pagesRef)
					if pagesDict, ok := pagesObj.(model.Dict); ok {
						if kids, ok := pagesDict[model.Name("Kids")].(model.Array); ok && len(kids) > 0 {
							if firstRef, ok := kids[0].(model.Ref); ok {
								firstPageObjNum = firstRef.ObjectNumber
							}
						}
					}
				}
			}
		}
	}
	numPages := 1
	if root := doc.Trailer().Root(); root != nil {
		if pagesObj, err := doc.Resolve(model.Ref{ObjectNumber: root.ObjectNumber, Generation: 0}); err == nil {
			if catDict, ok := pagesObj.(model.Dict); ok {
				if pagesRef, ok := catDict[model.Name("Pages")].(model.Ref); ok {
					if pObj, err := doc.Resolve(pagesRef); err == nil {
						if pDict, ok := pObj.(model.Dict); ok {
							if c, ok := pDict[model.Name("Count")].(model.Integer); ok {
								numPages = int(c)
							}
						}
					}
				}
			}
		}
	}
	maxNum := allNums[len(allNums)-1]
	totalObjs := maxNum + 1 // include object 0

	const header = "%PDF-1.4\n"
	if _, err := ws.Write([]byte(header)); err != nil {
		return err
	}
	pos0, _ := ws.Seek(0, io.SeekCurrent)
	// Write object 0: linearization parameter dict (placeholders for /L and /H)
	linDict := fmt.Sprintf("%d 0 obj\n<< /Linearized 1 /L 0000000000 /H [ 0000000000 0000000000 ] /O %d /N %d /T %d >>\nendobj\n",
		0, firstPageObjNum, numPages, totalObjs)
	if _, err := io.WriteString(ws, linDict); err != nil {
		return err
	}
	firstPageBodyEnd, _ := ws.Seek(0, io.SeekCurrent)
	// Write first-page objects (object 0 already written; firstPageNums are 1..N)
	offsetsFirst := make(map[int]int64)
	offsetsFirst[0] = pos0
	for _, num := range firstPageNums {
		ref := model.Ref{ObjectNumber: num, Generation: 0}
		obj, err := doc.Resolve(ref)
		if err != nil {
			return err
		}
		pos, _ := ws.Seek(0, io.SeekCurrent)
		offsetsFirst[num] = pos
		if err := pw.writeIndirectObject(ws, num, 0, obj); err != nil {
			return err
		}
	}
	firstXrefStart, _ := ws.Seek(0, io.SeekCurrent)
	// First xref: 0 through max of first section
	maxFirst := 0
	for n := range offsetsFirst {
		if n > maxFirst {
			maxFirst = n
		}
	}
	if err := pw.writeXRefTable(ws, doc, firstPageNums, offsetsFirst, maxFirst); err != nil {
		return err
	}
	// First trailer with 10-digit /Prev placeholder so we can seek back and fix after we know mainXrefStart
	rootNum := 1
	if r := doc.Trailer().Root(); r != nil {
		rootNum = r.ObjectNumber
	}
	prevPlaceholder := "0000000000"
	if _, err := fmt.Fprintf(ws, "trailer\n<< /Root %d 0 R /Size %d /Prev %s >>\nstartxref\n%d\n%%%%EOF\n",
		rootNum, maxNum+1, prevPlaceholder, firstXrefStart); err != nil {
		return err
	}
	firstSectionEnd, _ := ws.Seek(0, io.SeekCurrent)
	// Offset of the 10-digit /Prev value (we'll overwrite with mainXrefStart at the end)
	xrefLineLen := 20
	prevValueOffset := firstXrefStart + int64(5+2+(maxFirst+1)*xrefLineLen) + int64(len("trailer\n<< /Root ")+len(fmt.Sprintf("%d", rootNum))+len(" 0 R /Size ")+len(fmt.Sprintf("%d", maxNum+1))+len(" /Prev "))
	// Rest: remaining objects, then main xref, main trailer
	offsetsRest := make(map[int]int64)
	for _, num := range restNums {
		ref := model.Ref{ObjectNumber: num, Generation: 0}
		obj, err := doc.Resolve(ref)
		if err != nil {
			return err
		}
		pos, _ := ws.Seek(0, io.SeekCurrent)
		offsetsRest[num] = pos
		if err := pw.writeIndirectObject(ws, num, 0, obj); err != nil {
			return err
		}
	}
	mainXrefStart, _ := ws.Seek(0, io.SeekCurrent)
	allOffsets := make(map[int]int64)
	for k, v := range offsetsFirst {
		allOffsets[k] = v
	}
	for k, v := range offsetsRest {
		allOffsets[k] = v
	}
	allNumsWith0 := append([]int{0}, allNums...)
	sort.Ints(allNumsWith0)
	if err := pw.writeXRefTable(ws, doc, allNumsWith0, allOffsets, maxNum); err != nil {
		return err
	}
	trailerMain := copyTrailerWithSize(doc.Trailer().Dict, maxNum+1)
	if _, err := io.WriteString(ws, "trailer\n"); err != nil {
		return err
	}
	if err := pw.writeDict(ws, trailerMain); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(ws, "\nstartxref\n%d\n%%%%EOF\n", mainXrefStart); err != nil {
		return err
	}
	totalLen, _ := ws.Seek(0, io.SeekCurrent)
	// Seek back and fix object 0: /L and /H
	// Lin dict format: /L 0000000000 /H [ 0000000000 0000000000 ]
	// We need to overwrite with: /L totalLen /H [ firstPageBodyEnd firstSectionEnd ]
	seekBack := pos0 + int64(len("0 0 obj\n<< /Linearized 1 "))
	if _, err := ws.Seek(seekBack, io.SeekStart); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(ws, "/L %d /H [ %d %d ]", totalLen, firstPageBodyEnd, firstSectionEnd); err != nil {
		return err
	}
	// Fix first trailer /Prev to point to main xref
	if _, err := ws.Seek(prevValueOffset, io.SeekStart); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(ws, "%010d", mainXrefStart); err != nil {
		return err
	}
	return nil
}

// firstPageObjectSet returns the set of object numbers required to display the first page
// (catalog, pages tree, first page, and all refs from that page).
func (pw *PDFWriter) firstPageObjectSet(doc Document) (map[int]bool, error) {
	root := doc.Trailer().Root()
	if root == nil {
		return nil, fmt.Errorf("no catalog")
	}
	seen := make(map[int]bool)
	pw.collectRefs(doc, root.ObjectNumber, seen)
	cat, err := doc.Resolve(model.Ref{ObjectNumber: root.ObjectNumber, Generation: 0})
	if err != nil {
		return nil, err
	}
	catDict, ok := cat.(model.Dict)
	if !ok {
		return seen, nil
	}
	pagesRef, ok := catDict[model.Name("Pages")].(model.Ref)
	if !ok {
		return seen, nil
	}
	pw.collectRefs(doc, pagesRef.ObjectNumber, seen)
	pagesObj, err := doc.Resolve(pagesRef)
	if err != nil {
		return seen, nil
	}
	pagesDict, ok := pagesObj.(model.Dict)
	if !ok {
		return seen, nil
	}
	kids, ok := pagesDict[model.Name("Kids")].(model.Array)
	if !ok || len(kids) == 0 {
		return seen, nil
	}
	firstRef, ok := kids[0].(model.Ref)
	if !ok {
		return seen, nil
	}
	pw.collectRefs(doc, firstRef.ObjectNumber, seen)
	return seen, nil
}

func (pw *PDFWriter) collectRefs(doc Document, objNum int, seen map[int]bool) {
	if seen[objNum] {
		return
	}
	seen[objNum] = true
	obj, err := doc.Resolve(model.Ref{ObjectNumber: objNum, Generation: 0})
	if err != nil || obj == nil {
		return
	}
	switch v := obj.(type) {
	case model.Dict:
		for _, val := range v {
			if ref, ok := val.(model.Ref); ok {
				pw.collectRefs(doc, ref.ObjectNumber, seen)
			}
		}
	case model.Array:
		for _, val := range v {
			if ref, ok := val.(model.Ref); ok {
				pw.collectRefs(doc, ref.ObjectNumber, seen)
			}
		}
	case *model.Stream:
		for _, val := range v.Dict {
			if ref, ok := val.(model.Ref); ok {
				pw.collectRefs(doc, ref.ObjectNumber, seen)
			}
		}
	}
}

func copyTrailerWithSize(d model.Dict, size int) model.Dict {
	out := make(model.Dict, len(d)+1)
	for k, v := range d {
		out[k] = v
	}
	out[model.Name("Size")] = model.Integer(int64(size))
	return out
}
