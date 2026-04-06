package tagged

import "github.com/gsoultan/gpdf/model"

// inSection returns whether the given content index belongs to any section.
func inSection(sections []Section, blockIdx, figureIdx, tableIdx, listIdx int, kind byte) bool {
	for _, s := range sections {
		switch kind {
		case 'b':
			for _, i := range s.BlockIndices {
				if i == blockIdx {
					return true
				}
			}
		case 'f':
			for _, i := range s.FigureIndices {
				if i == figureIdx {
					return true
				}
			}
		case 't':
			for _, i := range s.TableIndices {
				if i == tableIdx {
					return true
				}
			}
		case 'l':
			for _, i := range s.ListIndices {
				if i == listIdx {
					return true
				}
			}
		}
	}
	return false
}

// BuildStructure builds MarkInfo, StructTreeRoot, and ParentTree for tagged PDF.
// It writes objects into objs, updates catalogDict, and returns the updated next object number.
func (s *Support) BuildStructure(objs map[int]model.Object, catalogDict model.Dict, pageNums []int, nextNum int) int {
	markInfoNum := nextNum
	nextNum++
	objs[markInfoNum] = model.Dict{
		model.Name("Marked"): model.Boolean(true),
	}
	catalogDict[model.Name("MarkInfo")] = model.Ref{ObjectNumber: markInfoNum, Generation: 0}
	structTreeNum := nextNum
	nextNum++

	type cellElemRef struct {
		pageIndex int
		mcid      int
		objNum    int
	}
	tableObjNums := make([]int, len(s.Tables))
	rowObjNums := make([][]int, len(s.Tables))
	var cellElems []cellElemRef

	for ti, tbl := range s.Tables {
		if len(tbl.Rows) == 0 {
			continue
		}
		tableObjNums[ti] = nextNum
		nextNum++
		rowObjNums[ti] = make([]int, len(tbl.Rows))
		for ri, row := range tbl.Rows {
			if len(row.Cells) == 0 {
				continue
			}
			rowObjNums[ti][ri] = nextNum
			nextNum++
			for _, c := range row.Cells {
				for _, mcid := range c.MCIDs {
					cellNum := nextNum
					nextNum++
					cellElems = append(cellElems, cellElemRef{
						pageIndex: c.PageIndex,
						mcid:      mcid,
						objNum:    cellNum,
					})
				}
			}
		}
	}

	cellIdx := 0
	for ti, tbl := range s.Tables {
		for ri, row := range tbl.Rows {
			for _, c := range row.Cells {
				if len(c.MCIDs) == 0 {
					continue
				}
				pageRef := model.Ref{ObjectNumber: pageNums[c.PageIndex], Generation: 0}
				rowRef := model.Ref{ObjectNumber: rowObjNums[ti][ri], Generation: 0}

				var kids model.Array
				var cellRefNums []int
				for _, mcid := range c.MCIDs {
					if cellIdx >= len(cellElems) {
						break
					}
					cellRef := cellElems[cellIdx]
					cellIdx++
					if cellRef.pageIndex != c.PageIndex || cellRef.mcid != mcid {
						continue
					}
					kids = append(kids, model.Dict{
						model.Name("Type"): model.Name("MCR"),
						model.Name("Pg"):   pageRef,
						model.Name("MCID"): model.Integer(int64(mcid)),
					})
					cellRefNums = append(cellRefNums, cellRef.objNum)
				}
				if len(kids) == 0 || len(cellRefNums) == 0 {
					continue
				}

				elemDict := model.Dict{
					model.Name("Type"): model.Name("StructElem"),
					model.Name("S"):    c.Role,
					model.Name("P"):    rowRef,
					model.Name("Pg"):   pageRef,
				}
				if len(kids) == 1 {
					elemDict[model.Name("K")] = kids[0]
				} else {
					elemDict[model.Name("K")] = kids
				}
				attr := model.Dict{}
				if c.Role == model.Name("TH") && c.Scope != "" {
					attr[model.Name("Scope")] = model.Name(c.Scope)
				}
				if c.Alt != "" {
					attr[model.Name("Alt")] = model.String(c.Alt)
				}
				if c.Lang != "" {
					elemDict[model.Name("Lang")] = model.String(c.Lang)
				}
				if len(attr) > 0 {
					elemDict[model.Name("A")] = attr
				}
				for _, num := range cellRefNums {
					objs[num] = elemDict
				}
			}
		}
	}

	for ti, tbl := range s.Tables {
		if len(tbl.Rows) == 0 || tableObjNums[ti] == 0 {
			continue
		}
		tableNum := tableObjNums[ti]
		var rowRefs model.Array
		for ri, row := range tbl.Rows {
			if len(row.Cells) == 0 || rowObjNums[ti][ri] == 0 {
				continue
			}
			rowNum := rowObjNums[ti][ri]
			var cellRefs model.Array
			for _, ce := range cellElems {
				for _, c := range row.Cells {
					for _, mcid := range c.MCIDs {
						if ce.pageIndex == c.PageIndex && ce.mcid == mcid {
							cellRefs = append(cellRefs, model.Ref{ObjectNumber: ce.objNum, Generation: 0})
							break
						}
					}
				}
			}
			rowDict := model.Dict{
				model.Name("Type"): model.Name("StructElem"),
				model.Name("S"):    model.Name("TR"),
				model.Name("P"):    model.Ref{ObjectNumber: tableNum, Generation: 0},
				model.Name("K"):    cellRefs,
			}
			objs[rowNum] = rowDict
			rowRefs = append(rowRefs, model.Ref{ObjectNumber: rowNum, Generation: 0})
		}
		tableDict := model.Dict{
			model.Name("Type"): model.Name("StructElem"),
			model.Name("S"):    model.Name("Table"),
			model.Name("K"):    rowRefs,
		}
		objs[tableNum] = tableDict
	}

	blockObjNums := make([]int, len(s.Blocks))
	for i := range s.Blocks {
		blockObjNums[i] = nextNum
		nextNum++
	}
	figureObjNums := make([]int, len(s.Figures))
	for i, fig := range s.Figures {
		if fig.PageIndex >= 0 && fig.PageIndex < len(pageNums) {
			figureObjNums[i] = nextNum
			nextNum++
		}
	}
	for i, blk := range s.Blocks {
		if blk.PageIndex < 0 || blk.PageIndex >= len(pageNums) {
			continue
		}
		pageRef := model.Ref{ObjectNumber: pageNums[blk.PageIndex], Generation: 0}
		k := model.Dict{
			model.Name("Type"): model.Name("MCR"),
			model.Name("Pg"):   pageRef,
			model.Name("MCID"): model.Integer(int64(blk.MCID)),
		}
		elemDict := model.Dict{
			model.Name("Type"): model.Name("StructElem"),
			model.Name("S"):    blk.Role,
			model.Name("Pg"):   pageRef,
			model.Name("K"):    k,
		}
		if blk.Lang != "" {
			elemDict[model.Name("Lang")] = model.String(blk.Lang)
		}
		if blk.Alt != "" {
			elemDict[model.Name("Alt")] = model.String(blk.Alt)
		}
		objs[blockObjNums[i]] = elemDict
	}

	for i, fig := range s.Figures {
		if figureObjNums[i] == 0 {
			continue
		}
		pageRef := model.Ref{ObjectNumber: pageNums[fig.PageIndex], Generation: 0}
		mcr := model.Dict{
			model.Name("Type"): model.Name("MCR"),
			model.Name("Pg"):   pageRef,
			model.Name("MCID"): model.Integer(int64(fig.MCID)),
		}
		elemDict := model.Dict{
			model.Name("Type"): model.Name("StructElem"),
			model.Name("S"):    model.Name("Figure"),
			model.Name("Pg"):   pageRef,
			model.Name("K"):    mcr,
		}
		if fig.Alt != "" {
			elemDict[model.Name("A")] = model.Dict{
				model.Name("Alt"): model.String(fig.Alt),
			}
		}
		objs[figureObjNums[i]] = elemDict
	}

	listObjNums := make([]int, len(s.Lists))
	itemObjNums := make([][]int, len(s.Lists))
	for i, lst := range s.Lists {
		if lst.PageIndex < 0 || lst.PageIndex >= len(pageNums) || len(lst.Items) == 0 {
			continue
		}
		listObjNums[i] = nextNum
		nextNum++
		itemObjNums[i] = make([]int, len(lst.Items))
		for j := range lst.Items {
			itemObjNums[i][j] = nextNum
			nextNum++
		}
	}
	for i, lst := range s.Lists {
		if listObjNums[i] == 0 || len(lst.Items) == 0 || lst.PageIndex < 0 || lst.PageIndex >= len(pageNums) {
			continue
		}
		pageRef := model.Ref{ObjectNumber: pageNums[lst.PageIndex], Generation: 0}
		var liRefs model.Array
		for j, it := range lst.Items {
			mcr := model.Dict{
				model.Name("Type"): model.Name("MCR"),
				model.Name("Pg"):   pageRef,
				model.Name("MCID"): model.Integer(int64(it.MCID)),
			}
			liDict := model.Dict{
				model.Name("Type"): model.Name("StructElem"),
				model.Name("S"):    model.Name("LI"),
				model.Name("P"):    model.Ref{ObjectNumber: listObjNums[i], Generation: 0},
				model.Name("Pg"):   pageRef,
				model.Name("K"):    mcr,
			}
			objs[itemObjNums[i][j]] = liDict
			liRefs = append(liRefs, model.Ref{ObjectNumber: itemObjNums[i][j], Generation: 0})
		}
		listDict := model.Dict{
			model.Name("Type"): model.Name("StructElem"),
			model.Name("S"):    model.Name("L"),
			model.Name("K"):    liRefs,
		}
		objs[listObjNums[i]] = listDict
	}

	sectObjNums := make([]int, len(s.Sections))
	for i := range s.Sections {
		sectObjNums[i] = nextNum
		nextNum++
	}
	for i, sec := range s.Sections {
		var kids model.Array
		for _, bi := range sec.BlockIndices {
			if bi >= 0 && bi < len(blockObjNums) && blockObjNums[bi] != 0 {
				kids = append(kids, model.Ref{ObjectNumber: blockObjNums[bi], Generation: 0})
			}
		}
		for _, fi := range sec.FigureIndices {
			if fi >= 0 && fi < len(figureObjNums) && figureObjNums[fi] != 0 {
				kids = append(kids, model.Ref{ObjectNumber: figureObjNums[fi], Generation: 0})
			}
		}
		for _, ti := range sec.TableIndices {
			if ti >= 0 && ti < len(tableObjNums) && tableObjNums[ti] != 0 {
				kids = append(kids, model.Ref{ObjectNumber: tableObjNums[ti], Generation: 0})
			}
		}
		for _, li := range sec.ListIndices {
			if li >= 0 && li < len(listObjNums) && listObjNums[li] != 0 {
				kids = append(kids, model.Ref{ObjectNumber: listObjNums[li], Generation: 0})
			}
		}
		objs[sectObjNums[i]] = model.Dict{
			model.Name("Type"): model.Name("StructElem"),
			model.Name("S"):    model.Name("Sect"),
			model.Name("K"):    kids,
		}
	}

	var docKids model.Array
	if len(s.Sections) > 0 {
		for _, snum := range sectObjNums {
			docKids = append(docKids, model.Ref{ObjectNumber: snum, Generation: 0})
		}
		for ti, tnum := range tableObjNums {
			if tnum != 0 && !inSection(s.Sections, 0, 0, ti, 0, 't') {
				docKids = append(docKids, model.Ref{ObjectNumber: tnum, Generation: 0})
			}
		}
		for li, lnum := range listObjNums {
			if lnum != 0 && !inSection(s.Sections, 0, 0, 0, li, 'l') {
				docKids = append(docKids, model.Ref{ObjectNumber: lnum, Generation: 0})
			}
		}
		for bi, bnum := range blockObjNums {
			if bnum != 0 && !inSection(s.Sections, bi, 0, 0, 0, 'b') {
				docKids = append(docKids, model.Ref{ObjectNumber: bnum, Generation: 0})
			}
		}
		for fi, fnum := range figureObjNums {
			if fnum != 0 && !inSection(s.Sections, 0, fi, 0, 0, 'f') {
				docKids = append(docKids, model.Ref{ObjectNumber: fnum, Generation: 0})
			}
		}
	} else {
		for _, tnum := range tableObjNums {
			if tnum != 0 {
				docKids = append(docKids, model.Ref{ObjectNumber: tnum, Generation: 0})
			}
		}
		for _, lnum := range listObjNums {
			if lnum != 0 {
				docKids = append(docKids, model.Ref{ObjectNumber: lnum, Generation: 0})
			}
		}
		for _, bnum := range blockObjNums {
			if bnum != 0 {
				docKids = append(docKids, model.Ref{ObjectNumber: bnum, Generation: 0})
			}
		}
		for _, fnum := range figureObjNums {
			if fnum != 0 {
				docKids = append(docKids, model.Ref{ObjectNumber: fnum, Generation: 0})
			}
		}
	}
	documentNum := nextNum
	nextNum++
	documentDict := model.Dict{
		model.Name("Type"): model.Name("StructElem"),
		model.Name("S"):    model.Name("Document"),
		model.Name("K"):    docKids,
	}
	objs[documentNum] = documentDict

	structRoot := model.Dict{
		model.Name("Type"): model.Name("StructTreeRoot"),
		model.Name("K"):    model.Array{model.Ref{ObjectNumber: documentNum, Generation: 0}},
	}

	roleMap := model.Dict{
		model.Name("Document"): model.Name("Document"),
		model.Name("Sect"):     model.Name("Sect"),
		model.Name("Table"):    model.Name("Table"),
		model.Name("TR"):       model.Name("TR"),
		model.Name("TH"):       model.Name("TH"),
		model.Name("TD"):       model.Name("TD"),
		model.Name("P"):        model.Name("P"),
		model.Name("H1"):       model.Name("H1"),
		model.Name("H2"):       model.Name("H2"),
		model.Name("H3"):       model.Name("H3"),
		model.Name("H4"):       model.Name("H4"),
		model.Name("H5"):       model.Name("H5"),
		model.Name("H6"):       model.Name("H6"),
		model.Name("Caption"):  model.Name("Caption"),
		model.Name("Quote"):    model.Name("Quote"),
		model.Name("Code"):     model.Name("Code"),
		model.Name("Figure"):   model.Name("Figure"),
		model.Name("L"):        model.Name("L"),
		model.Name("LI"):       model.Name("LI"),
	}
	structRoot[model.Name("RoleMap")] = roleMap

	if len(cellElems) > 0 || len(s.Blocks) > 0 || len(s.Lists) > 0 || len(s.Figures) > 0 {
		parentTreeNum := nextNum
		nextNum++
		pageCells := make(map[int]model.Array)
		for _, ce := range cellElems {
			pageCells[ce.pageIndex] = append(pageCells[ce.pageIndex],
				model.Ref{ObjectNumber: ce.objNum, Generation: 0},
			)
		}
		for i, blk := range s.Blocks {
			if blk.PageIndex < 0 || blk.PageIndex >= len(pageNums) {
				continue
			}
			pageCells[blk.PageIndex] = append(pageCells[blk.PageIndex],
				model.Ref{ObjectNumber: blockObjNums[i], Generation: 0},
			)
		}
		for i, lst := range s.Lists {
			if lst.PageIndex < 0 || lst.PageIndex >= len(pageNums) {
				continue
			}
			for j := range lst.Items {
				pageCells[lst.PageIndex] = append(pageCells[lst.PageIndex],
					model.Ref{ObjectNumber: itemObjNums[i][j], Generation: 0},
				)
			}
		}
		for i, fig := range s.Figures {
			if figureObjNums[i] == 0 || fig.PageIndex < 0 || fig.PageIndex >= len(pageNums) {
				continue
			}
			pageCells[fig.PageIndex] = append(pageCells[fig.PageIndex],
				model.Ref{ObjectNumber: figureObjNums[i], Generation: 0},
			)
		}
		var nums model.Array
		parentIdx := 0
		for pageIndex, cellsArr := range pageCells {
			nums = append(nums,
				model.Integer(int64(parentIdx)),
				cellsArr,
			)
			if pageIndex >= 0 && pageIndex < len(pageNums) {
				if pd, ok := objs[pageNums[pageIndex]].(model.Dict); ok {
					pd[model.Name("StructParents")] = model.Integer(int64(parentIdx))
					objs[pageNums[pageIndex]] = pd
				}
			}
			parentIdx++
		}
		parentTreeDict := model.Dict{
			model.Name("Nums"): nums,
		}
		objs[parentTreeNum] = parentTreeDict
		structRoot[model.Name("ParentTree")] = model.Ref{ObjectNumber: parentTreeNum, Generation: 0}
	}

	objs[structTreeNum] = structRoot
	catalogDict[model.Name("StructTreeRoot")] = model.Ref{ObjectNumber: structTreeNum, Generation: 0}

	return nextNum
}
