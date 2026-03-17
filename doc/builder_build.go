package doc

import (
	"fmt"

	"gpdf/model"
)

// Build returns a Document that can be saved. The document is in-memory.
func (b *DocumentBuilder) Build() (Document, error) {
	objs := make(map[int]model.Object)
	nextNum := 1

	catalogNum := nextNum
	nextNum++
	pagesNum := nextNum
	nextNum++

	pageNums, pageRefs := b.allocatePageNums(nextNum)
	nextNum += len(b.pages)

	nextNum = b.buildMinimalStream(objs, nextNum)
	minimalStreamNum := nextNum - 1

	nextNum, err := b.buildPageObjects(objs, pageNums, pagesNum, minimalStreamNum, nextNum)
	if err != nil {
		return nil, err
	}

	pagesDict := model.Dict{
		model.Name("Type"):  model.Name("Pages"),
		model.Name("Kids"):  pageRefs,
		model.Name("Count"): model.Integer(len(b.pages)),
	}
	objs[pagesNum] = pagesDict

	catalogDict := model.Dict{
		model.Name("Type"):  model.Name("Catalog"),
		model.Name("Pages"): model.Ref{ObjectNumber: pagesNum, Generation: 0},
	}
	if b.metadata.Lang != "" {
		catalogDict[model.Name("Lang")] = model.String(b.metadata.Lang)
	}

	nextNum = b.buildCatalogExtras(objs, catalogDict, pageNums, nextNum)
	nextNum = b.buildEmbeddedFonts(objs, nextNum)

	objs[catalogNum] = catalogDict

	infoNum := nextNum
	nextNum++
	objs[infoNum] = b.metadata.BuildInfoDict()

	nextNum = b.metadata.BuildMetadataStream(objs, catalogDict, nextNum)

	trailer := model.Trailer{
		Dict: model.Dict{
			model.Name("Root"): model.Ref{ObjectNumber: catalogNum, Generation: 0},
			model.Name("Size"): model.Integer(nextNum),
			model.Name("Info"): model.Ref{ObjectNumber: infoNum, Generation: 0},
		},
	}
	return &builtDocument{
		trailer: trailer,
		objects: objs,
		size:    nextNum,
	}, nil
}

// allocatePageNums returns object numbers and refs for all pages.
func (b *DocumentBuilder) allocatePageNums(startNum int) (pageNums []int, pageRefs model.Array) {
	pageNums = make([]int, 0, len(b.pages))
	for i := range b.pages {
		num := startNum + i
		pageNums = append(pageNums, num)
		pageRefs = append(pageRefs, model.Ref{ObjectNumber: num, Generation: 0})
	}
	return
}

// buildMinimalStream creates a shared minimal content stream for blank pages.
func (b *DocumentBuilder) buildMinimalStream(objs map[int]model.Object, nextNum int) int {
	minimalContent := []byte("n\n")
	objs[nextNum] = &model.Stream{
		Dict:    model.Dict{model.Name("Length"): model.Integer(len(minimalContent))},
		Content: minimalContent,
	}
	return nextNum + 1
}

// buildPageObjects builds page dicts, content streams, image XObjects, and link annotations.
func (b *DocumentBuilder) buildPageObjects(objs map[int]model.Object, pageNums []int, pagesNum, minimalStreamNum, nextNum int) (int, error) {
	for idx, pageNum := range pageNums {
		spec := b.pages[idx]
		pageDict := copyPageDict(spec.dict)
		hasContent := len(spec.textRuns) > 0 || len(spec.imageRuns) > 0 || len(spec.graphicRuns) > 0

		if !hasContent {
			pageDict[model.Name("Contents")] = model.Ref{ObjectNumber: minimalStreamNum, Generation: 0}
			if _, ok := pageDict[model.Name("Resources")]; !ok {
				pageDict[model.Name("Resources")] = model.Dict{}
			}
		} else {
			contentStreamNum := nextNum
			nextNum++
			imageStreamNums := make([]int, len(spec.imageRuns))
			for i := range spec.imageRuns {
				imageStreamNums[i] = nextNum
				nextNum++
			}
			contentBytes, resources, buildErr := b.buildPageContent(spec.graphicRuns, spec.textRuns, spec.imageRuns, imageStreamNums)
			if buildErr != nil && buildErr != errFlateCompressed {
				return nextNum, fmt.Errorf("page %d: %w", idx+1, buildErr)
			}
			pageDict[model.Name("Contents")] = model.Ref{ObjectNumber: contentStreamNum, Generation: 0}
			if existing, ok := pageDict[model.Name("Resources")].(model.Dict); ok && existing != nil {
				for k, v := range resources {
					existing[k] = v
				}
				pageDict[model.Name("Resources")] = existing
			} else {
				pageDict[model.Name("Resources")] = resources
			}
			streamDict := model.Dict{model.Name("Length"): model.Integer(len(contentBytes))}
			if buildErr == errFlateCompressed {
				streamDict[model.Name("Filter")] = model.Name("FlateDecode")
			}
			objs[contentStreamNum] = &model.Stream{
				Dict:    streamDict,
				Content: contentBytes,
			}
			for i, im := range spec.imageRuns {
				if im.isJPEG {
					objs[imageStreamNums[i]] = b.jpegXObjectStream(im)
				} else {
					objs[imageStreamNums[i]] = b.imageXObjectStream(im)
				}
			}
		}

		annotRefs, updatedNext := b.outlines.BuildLinkAnnotations(objs, idx, pageNums, nextNum)
		nextNum = updatedNext
		if len(annotRefs) > 0 {
			pageDict[model.Name("Annots")] = annotRefs
		}

		objs[pageNum] = pageDict
		pageDict[model.Name("Parent")] = model.Ref{ObjectNumber: pagesNum, Generation: 0}
	}
	return nextNum, nil
}

// buildCatalogExtras adds named dests, outlines, form fields, tagged structure, OCGs, and ICC to the catalog.
func (b *DocumentBuilder) buildCatalogExtras(objs map[int]model.Object, catalogDict model.Dict, pageNums []int, nextNum int) int {
	nextNum = b.outlines.BuildNamedDests(objs, pageNums, catalogDict, nextNum)

	if b.forms.UseAcroForm || len(b.forms.Fields) > 0 {
		_, nextNum = b.forms.BuildFields(objs, pageNums, catalogDict, nextNum)
	}

	if b.useTagged {
		nextNum = b.tagging.BuildStructure(objs, catalogDict, pageNums, nextNum)
	}

	nextNum = b.layers.BuildOCProperties(objs, catalogDict, nextNum)

	nextNum = b.outlines.BuildOutlines(objs, pageNums, catalogDict, nextNum)

	nextNum = b.buildICCOutputIntent(objs, catalogDict, nextNum)

	return nextNum
}

// buildICCOutputIntent writes the ICC output intent objects.
func (b *DocumentBuilder) buildICCOutputIntent(objs map[int]model.Object, catalogDict model.Dict, nextNum int) int {
	if b.iccProfile == nil || len(b.iccProfile.Data) == 0 {
		return nextNum
	}
	iccStreamNum := nextNum
	nextNum++
	objs[iccStreamNum] = &model.Stream{
		Dict: model.Dict{
			model.Name("N"):         model.Integer(int64(b.iccProfile.N)),
			model.Name("Alternate"): model.Name(b.iccProfile.Alternate),
			model.Name("Length"):    model.Integer(int64(len(b.iccProfile.Data))),
			model.Name("Filter"):    model.Name("FlateDecode"),
		},
		Content: b.iccProfile.Data,
	}
	outputIntent := model.Dict{
		model.Name("Type"):                      model.Name("OutputIntent"),
		model.Name("S"):                         model.Name("GTS_PDFA1"),
		model.Name("OutputConditionIdentifier"): model.String("sRGB"),
		model.Name("DestOutputProfile"):         model.Ref{ObjectNumber: iccStreamNum},
	}
	intentNum := nextNum
	nextNum++
	objs[intentNum] = outputIntent
	catalogDict[model.Name("OutputIntents")] = model.Array{
		model.Ref{ObjectNumber: intentNum},
	}
	return nextNum
}

// buildEmbeddedFonts creates CID font objects and replaces placeholder refs in page resources.
func (b *DocumentBuilder) buildEmbeddedFonts(objs map[int]model.Object, nextNum int) int {
	embeddedFontObjNums := make(map[string]int)
	for psName, usage := range b.embeddedFonts {
		if len(usage.usedGIDs) == 0 {
			continue
		}
		type0Num := buildEmbeddedFontObjects(usage, objs, &nextNum)
		embeddedFontObjNums[psName] = type0Num
	}
	if len(embeddedFontObjNums) > 0 {
		b.replaceEmbeddedFontPlaceholders(objs, embeddedFontObjNums)
	}
	return nextNum
}
