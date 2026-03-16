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
		case ctRArray:
			return arr, nil
		case ctOp:
			return nil, fmt.Errorf("operator %s inside array", t.value)
		}
	}
}

// Ensure StreamParser implements content.Parser.
var _ content.Parser = (*StreamParser)(nil)
