package impl

import (
	"fmt"

	"gpdf/content"
	"gpdf/model"
)

// StreamParser parses PDF content stream bytes into a sequence of content.Op.
type StreamParser struct{}

// NewStreamParser returns a content stream parser.
func NewStreamParser() *StreamParser {
	return &StreamParser{}
}

// Parse implements content.Parser.
func (p *StreamParser) Parse(stream []byte) ([]content.Op, error) {
	tok := newContentTokenizer(stream)
	var ops []content.Op
	var args []model.Object
	for {
		op, err := p.nextOp(tok, &args)
		if err != nil {
			return nil, err
		}
		if op == nil {
			break
		}
		ops = append(ops, *op)
	}
	return ops, nil
}

// nextOp reads tokens until an operator is seen; then returns Op and clears args.
// Returns nil Op at EOF. args is passed in and updated (operands pushed; cleared when op is emitted).
func (p *StreamParser) nextOp(tok *contentTokenizer, args *[]model.Object) (*content.Op, error) {
	for {
		t, err := tok.Next()
		if err != nil {
			return nil, err
		}
		switch t.kind {
		case ctEOF:
			return nil, nil
		case ctInteger:
			*args = append(*args, model.Integer(t.intVal))
		case ctReal:
			*args = append(*args, model.Real(t.fltVal))
		case ctName:
			*args = append(*args, model.Name(t.value))
		case ctString:
			*args = append(*args, model.String(t.value))
		case ctHex:
			*args = append(*args, model.String(t.value))
		case ctLArray:
			arr, err := p.parseArray(tok)
			if err != nil {
				return nil, err
			}
			*args = append(*args, arr)
		case ctRArray:
			return nil, fmt.Errorf("unexpected ]")
		case ctLDict:
			dict, err := p.parseDict(tok)
			if err != nil {
				return nil, err
			}
			*args = append(*args, dict)
		case ctRDict:
			return nil, fmt.Errorf("unexpected >>")
		case ctOp:
			name := t.value
			operands := make([]model.Object, len(*args))
			copy(operands, *args)
			*args = (*args)[:0]
			return &content.Op{Name: name, Args: operands}, nil
		}
	}
}

func (p *StreamParser) parseArray(tok *contentTokenizer) (model.Array, error) {
	var arr model.Array
	for {
		t, err := tok.Next()
		if err != nil {
			return nil, err
		}
		switch t.kind {
		case ctEOF:
			return nil, fmt.Errorf("array not closed")
		case ctInteger:
			arr = append(arr, model.Integer(t.intVal))
		case ctReal:
			arr = append(arr, model.Real(t.fltVal))
		case ctName:
			arr = append(arr, model.Name(t.value))
		case ctString:
			arr = append(arr, model.String(t.value))
		case ctHex:
			arr = append(arr, model.String(t.value))
		case ctLArray:
			nested, err := p.parseArray(tok)
			if err != nil {
				return nil, err
			}
			arr = append(arr, nested)
		case ctLDict:
			dict, err := p.parseDict(tok)
			if err != nil {
				return nil, err
			}
			arr = append(arr, dict)
		case ctRArray:
			return arr, nil
		case ctOp:
			return nil, fmt.Errorf("operator %s inside array", t.value)
		}
	}
}

func (p *StreamParser) parseDict(tok *contentTokenizer) (model.Dict, error) {
	dict := make(model.Dict)
	for {
		t, err := tok.Next()
		if err != nil {
			return nil, err
		}
		switch t.kind {
		case ctEOF:
			return nil, fmt.Errorf("dict not closed")
		case ctRDict:
			return dict, nil
		case ctName:
			key := model.Name(t.value)
			val, err := p.parseDictValue(tok)
			if err != nil {
				return nil, err
			}
			dict[key] = val
		default:
			return nil, fmt.Errorf("dict key: expected name, got %v", t.kind)
		}
	}
}

func (p *StreamParser) parseDictValue(tok *contentTokenizer) (model.Object, error) {
	t, err := tok.Next()
	if err != nil {
		return nil, err
	}
	switch t.kind {
	case ctInteger:
		return model.Integer(t.intVal), nil
	case ctReal:
		return model.Real(t.fltVal), nil
	case ctName:
		return model.Name(t.value), nil
	case ctString:
		return model.String(t.value), nil
	case ctHex:
		return model.String(t.value), nil
	case ctLArray:
		return p.parseArray(tok)
	case ctLDict:
		return p.parseDict(tok)
	default:
		return nil, fmt.Errorf("dict value: unexpected token %v %q", t.kind, t.value)
	}
}

// Ensure StreamParser implements content.Parser.
var _ content.Parser = (*StreamParser)(nil)
