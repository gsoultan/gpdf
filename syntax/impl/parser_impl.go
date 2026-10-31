package impl

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/syntax"
)

// ParserImpl parses PDF syntax from a reader supporting random access.
type ParserImpl struct {
	r       io.ReaderAt
	size    int64
	pos     int64
	tok     syntax.Tokenizer
	lastTok syntax.Token
	hasPeek bool
	// pushback holds tokens that were consumed but must be replayed (e.g. "0 0 595" where we thought "0 0 R").
	pushback []syntax.Token
	// lengthResolver resolves an indirect stream /Length (e.g. "/Length 5 0 R").
	lengthResolver func(model.Ref) (int64, bool)
	depth          int
}

// maxNestingDepth bounds array/dictionary nesting so hostile input cannot exhaust the stack.
const maxNestingDepth = 512

// SetLengthResolver installs a callback used to resolve indirect stream /Length values.
func (p *ParserImpl) SetLengthResolver(fn func(model.Ref) (int64, bool)) {
	p.lengthResolver = fn
}

// NewParserWithLengthResolver creates a parser that resolves indirect stream lengths with fn.
func NewParserWithLengthResolver(r io.ReaderAt, size int64, fn func(model.Ref) (int64, bool)) syntax.Parser {
	p := NewParser(r, size).(*ParserImpl)
	p.lengthResolver = fn
	return p
}

// NewParser creates a parser that reads from r (size bytes). Initial position is 0.
func NewParser(r io.ReaderAt, size int64) syntax.Parser {
	p := &ParserImpl{r: r, size: size, pos: 0}
	p.resetTokenizer(0)
	return p
}

func (p *ParserImpl) resetTokenizer(offset int64) {
	p.pos = offset
	end := p.size - offset
	if end < 0 {
		end = 0
	}
	p.tok = NewTokenizer(p.r, offset, end)
	p.hasPeek = false
	p.pushback = p.pushback[:0]
}

// SetPosition sets the read position for subsequent parsing.
func (p *ParserImpl) SetPosition(offset int64) error {
	p.resetTokenizer(offset)
	return nil
}

func (p *ParserImpl) nextToken() (syntax.Token, error) {
	if len(p.pushback) > 0 {
		tok := p.pushback[len(p.pushback)-1]
		p.pushback = p.pushback[:len(p.pushback)-1]
		return tok, nil
	}
	if p.hasPeek {
		p.hasPeek = false
		return p.lastTok, nil
	}
	tok, err := p.tok.Next()
	if err != nil {
		return syntax.Token{}, err
	}
	p.lastTok = tok
	return tok, nil
}

func (p *ParserImpl) peekToken() (syntax.Token, error) {
	if len(p.pushback) > 0 {
		return p.pushback[len(p.pushback)-1], nil
	}
	if p.hasPeek {
		return p.lastTok, nil
	}
	tok, err := p.tok.Next()
	if err != nil {
		return syntax.Token{}, err
	}
	p.hasPeek = true
	p.lastTok = tok
	return tok, nil
}

// ParseObject parses the next object. If it is an indirect object (N M obj ... endobj),
// returns indirect and nil inline. Otherwise returns nil and the inline object.
func (p *ParserImpl) ParseObject() (indirect *syntax.IndirectObject, inline model.Object, err error) {
	tok, err := p.nextToken()
	if err != nil {
		return nil, nil, err
	}
	if tok.Kind == syntax.TokenEOF {
		return nil, nil, io.EOF
	}
	// Check for indirect object: Integer Integer "obj"
	if tok.Kind == syntax.TokenInteger {
		n := tok.Int
		tok2, err := p.nextToken()
		if err != nil {
			return nil, nil, err
		}
		if tok2.Kind == syntax.TokenInteger {
			gen := tok2.Int
			tok3, err := p.nextToken()
			if err != nil {
				return nil, nil, err
			}
			if tok3.Kind == syntax.TokenKeyword && tok3.Value == "obj" {
				val, err := p.parseValue(true)
				if err != nil {
					return nil, nil, err
				}
				endTok, err := p.nextToken()
				if err != nil {
					return nil, nil, err
				}
				if endTok.Kind != syntax.TokenKeyword || endTok.Value != "endobj" {
					// Missing endobj is a common producer bug; viewers accept it. Replay the token.
					if endTok.Kind == syntax.TokenKeyword && endTok.Value == "obj" {
						return nil, nil, fmt.Errorf("expected endobj, got %v", endTok)
					}
					p.pushback = append(p.pushback, endTok)
				}
				return &syntax.IndirectObject{
					ObjectNumber: int(n),
					Generation:   int(gen),
					Value:        val,
				}, nil, nil
			}
			if tok3.Kind == syntax.TokenKeyword && tok3.Value == "R" {
				return nil, model.Ref{ObjectNumber: int(n), Generation: int(gen)}, nil
			}
			// Two integers followed by something else: return the first integer and replay
			// the second integer and tok3 on the next calls (pushback is LIFO).
			p.pushback = append(p.pushback, tok3, tok2)
			return nil, model.Integer(n), nil
		}
		// One integer, put back tok2
		p.hasPeek = true
		p.lastTok = tok2
		inline, err = p.parseIntOrReal(syntax.Token{Kind: syntax.TokenInteger, Int: tok.Int})
		return nil, inline, err
	}
	// Not indirect object start; parse as single value
	inline, err = p.parseValueFromToken(tok, false)
	return nil, inline, err
}

func (p *ParserImpl) parseValue(allowStream bool) (model.Object, error) {
	tok, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	return p.parseValueFromToken(tok, allowStream)
}

func (p *ParserImpl) parseValueFromToken(tok syntax.Token, allowStream bool) (model.Object, error) {
	switch tok.Kind {
	case syntax.TokenEOF:
		return nil, io.EOF
	case syntax.TokenInteger:
		return p.parseIntOrReal(tok)
	case syntax.TokenReal:
		return model.Real(tok.Float), nil
	case syntax.TokenKeyword:
		switch tok.Value {
		case "true":
			return model.Boolean(true), nil
		case "false":
			return model.Boolean(false), nil
		case "null":
			return model.Null{}, nil
		case "endobj", "endstream":
			return nil, fmt.Errorf("unexpected keyword %s (possible wrong object offset)", tok.Value)
		default:
			return nil, fmt.Errorf("unexpected keyword %s", tok.Value)
		}
	case syntax.TokenName:
		return model.Name(tok.Value), nil
	case syntax.TokenLiteral:
		return model.String(tok.Value), nil
	case syntax.TokenHex:
		return model.String(tok.Value), nil
	case syntax.TokenLBracket:
		return p.parseArray()
	case syntax.TokenLDict:
		dict, err := p.parseDict()
		if err != nil {
			return nil, err
		}
		if allowStream {
			next, err := p.peekToken()
			if err != nil {
				return nil, err
			}
			if next.Kind == syntax.TokenKeyword && next.Value == "stream" {
				p.nextToken()
				stream, err := p.parseStream(dict)
				if err != nil {
					return nil, err
				}
				return stream, nil
			}
		}
		return dict, nil
	case syntax.TokenRDict:
		// Stray ">>" when a value was expected (malformed dict); treat as null so parsing can continue.
		return model.Null{}, nil
	default:
		return nil, fmt.Errorf("unexpected token kind %v (value %q)", tok.Kind, tok.Value)
	}
}

func (p *ParserImpl) parseIntOrReal(tok syntax.Token) (model.Object, error) {
	// Check for N M R (indirect reference)
	if tok.Kind == syntax.TokenInteger {
		next, err := p.peekToken()
		if err != nil {
			return model.Integer(tok.Int), nil
		}
		if next.Kind == syntax.TokenInteger {
			n2 := next.Int
			p.nextToken()
			rTok, err := p.nextToken()
			if err != nil {
				return nil, err
			}
			if rTok.Kind == syntax.TokenKeyword && rTok.Value == "R" {
				return model.Ref{ObjectNumber: int(tok.Int), Generation: int(n2)}, nil
			}
			// Two integers but not N M R: return first, push back second and the non-R token (e.g. "0 0 595" in array).
			p.pushback = append(p.pushback, rTok, syntax.Token{Kind: syntax.TokenInteger, Int: n2})
		}
		return model.Integer(tok.Int), nil
	}
	return model.Integer(tok.Int), nil
}

func (p *ParserImpl) parseArray() (model.Array, error) {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > maxNestingDepth {
		return nil, fmt.Errorf("array nesting deeper than %d", maxNestingDepth)
	}
	var arr model.Array
	for {
		tok, err := p.peekToken()
		if err != nil {
			return nil, err
		}
		if tok.Kind == syntax.TokenRBracket {
			p.nextToken()
			break
		}
		obj, err := p.parseValue(false)
		if err != nil {
			return nil, err
		}
		arr = append(arr, obj)
	}
	return arr, nil
}

func (p *ParserImpl) parseDict() (model.Dict, error) {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > maxNestingDepth {
		return nil, fmt.Errorf("dictionary nesting deeper than %d", maxNestingDepth)
	}
	dict := make(model.Dict)
	for {
		tok, err := p.nextToken()
		if err != nil {
			return nil, err
		}
		if tok.Kind == syntax.TokenRDict {
			break
		}
		if tok.Kind != syntax.TokenName {
			return nil, fmt.Errorf("expected name in dict, got %v", tok.Kind)
		}
		key := model.Name(tok.Value)
		val, err := p.parseValue(false)
		if err != nil {
			return nil, err
		}
		dict[key] = val
	}
	return dict, nil
}

func (p *ParserImpl) parseStream(dict model.Dict) (*model.Stream, error) {
	// PDF: after "stream" comes EOL (CR, LF, or CRLF), then stream bytes, then "endstream".
	offset := p.tok.CurrentOffset()
	skip := int64(0)
	var eol [2]byte
	if n, _ := p.r.ReadAt(eol[:], offset); n > 0 {
		switch {
		case eol[0] == '\r' && n > 1 && eol[1] == '\n':
			skip = 2
		case eol[0] == '\r' || eol[0] == '\n':
			skip = 1
		}
	}
	streamStart := offset + skip
	if streamStart > p.size {
		return nil, fmt.Errorf("stream starts beyond end of file")
	}

	// /Length is untrusted: it may be missing, indirect, negative, too short or far past EOF.
	length := int64(-1)
	switch l := dict[model.Name("Length")].(type) {
	case model.Integer:
		length = int64(l)
	case model.Real:
		length = int64(l)
	case model.Ref:
		if p.lengthResolver != nil {
			if n, ok := p.lengthResolver(l); ok {
				length = n
			}
		}
	}
	var end int64 // offset just past the stream data
	if length >= 0 && length <= p.size-streamStart && p.endstreamAt(streamStart+length) {
		end = streamStart + length
	} else {
		// Recover by scanning for the endstream keyword (what viewers do for damaged files).
		kw, err := p.findEndstream(streamStart)
		if err != nil {
			return nil, err
		}
		end = kw
		// The EOL before endstream is not part of the data.
		if end > streamStart {
			var b [1]byte
			if _, err := p.r.ReadAt(b[:], end-1); err == nil && b[0] == '\n' {
				end--
				if end > streamStart {
					if _, err := p.r.ReadAt(b[:], end-1); err == nil && b[0] == '\r' {
						end--
					}
				}
			} else if err == nil && b[0] == '\r' {
				end--
			}
		}
		dict[model.Name("Length")] = model.Integer(end - streamStart)
	}
	content := make([]byte, end-streamStart)
	if _, err := p.r.ReadAt(content, streamStart); err != nil && err != io.EOF {
		return nil, err
	}
	// Advance parser past stream body and consume "endstream".
	p.SetPosition(end)
	tok, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if tok.Kind != syntax.TokenKeyword || tok.Value != "endstream" {
		return nil, fmt.Errorf("expected endstream at offset %d, got %q", end, tok.Value)
	}
	return &model.Stream{Dict: dict, Content: content}, nil
}

var endstreamKeyword = []byte("endstream")

// endstreamAt reports whether "endstream" follows pos after optional whitespace.
func (p *ParserImpl) endstreamAt(pos int64) bool {
	buf := make([]byte, 32)
	n, _ := p.r.ReadAt(buf, pos)
	buf = buf[:n]
	i := 0
	for i < len(buf) && isWhitespace(buf[i]) {
		i++
	}
	return bytes.HasPrefix(buf[i:], endstreamKeyword)
}

// findEndstream returns the offset of the first "endstream" at or after start.
func (p *ParserImpl) findEndstream(start int64) (int64, error) {
	const chunk = 64 << 10
	overlap := int64(len(endstreamKeyword) - 1)
	buf := make([]byte, chunk)
	for pos := start; pos < p.size; pos += chunk - overlap {
		n, err := p.r.ReadAt(buf, pos)
		if n <= 0 {
			if err != nil && err != io.EOF {
				return 0, err
			}
			break
		}
		if i := bytes.Index(buf[:n], endstreamKeyword); i >= 0 {
			return pos + int64(i), nil
		}
		if int64(n) < chunk {
			break
		}
	}
	return 0, fmt.Errorf("stream at offset %d: endstream not found", start)
}

// ParseXRefTable parses the xref table. Parser must be positioned at the "xref" keyword.
func (p *ParserImpl) ParseXRefTable() (entries map[int]syntax.XRefEntry, err error) {
	entries = make(map[int]syntax.XRefEntry)
	// Consume leading "xref" keyword if present
	if tok, err := p.nextToken(); err == nil {
		if tok.Kind != syntax.TokenKeyword || tok.Value != "xref" {
			p.hasPeek, p.lastTok = true, tok
		}
	}
	for {
		tok, err := p.nextToken()
		if err != nil {
			return nil, err
		}
		if tok.Kind == syntax.TokenKeyword && tok.Value == "trailer" {
			// Put "trailer" back so ParseTrailer can be called
			p.hasPeek = true
			p.lastTok = tok
			break
		}
		if tok.Kind != syntax.TokenInteger {
			return nil, fmt.Errorf("xref: expected integer start, got %v", tok.Kind)
		}
		start := int(tok.Int)
		tok, err = p.nextToken()
		if err != nil {
			return nil, err
		}
		if tok.Kind != syntax.TokenInteger {
			return nil, fmt.Errorf("xref: expected integer count, got %v", tok.Kind)
		}
		count := int(tok.Int)
		// Tokenizer unread the delimiter after "count", so CurrentOffset is at last digit. Skip to first digit of first 20-byte line.
		lineStart := p.tok.CurrentOffset()
		b := make([]byte, 1)
		afterNewline := false
		for {
			if _, err := p.r.ReadAt(b, lineStart); err != nil {
				break
			}
			if b[0] == '\n' {
				lineStart++
				afterNewline = true
				continue
			}
			if b[0] == ' ' || b[0] == '\t' {
				lineStart++
				continue
			}
			if b[0] >= '0' && b[0] <= '9' {
				if afterNewline {
					break
				}
				lineStart++
				continue
			}
			if b[0] == '\r' {
				lineStart++
				if n, _ := p.r.ReadAt(b, lineStart); n > 0 && b[0] == '\n' {
					lineStart++
				}
				break
			}
			break
		}
		for i := range count {
			entryOffset := lineStart + int64(i*20)
			buf := make([]byte, 20)
			_, err := p.r.ReadAt(buf, entryOffset)
			if err != nil {
				return nil, err
			}
			// Format: 10-digit offset, space, 5-digit gen, space, n/f, EOL
			offsetStr := bytes.TrimSpace(buf[0:10])
			genStr := bytes.TrimSpace(buf[11:16])
			flag := buf[17]
			objNum := start + i
			off, _ := strconv.ParseInt(string(offsetStr), 10, 64)
			gen, _ := strconv.Atoi(string(genStr))
			entries[objNum] = syntax.XRefEntry{
				Offset:     off,
				Generation: gen,
				InUse:      flag == 'n',
			}
		}
		// Advance tokenizer past this subsection
		p.SetPosition(lineStart + int64(count*20))
		// Skip newline before next subsection or "trailer"
		for {
			tok, err := p.nextToken()
			if err != nil {
				break
			}
			if tok.Kind == syntax.TokenInteger {
				p.hasPeek = true
				p.lastTok = tok
				break
			}
			if tok.Kind == syntax.TokenKeyword && tok.Value == "trailer" {
				p.hasPeek = true
				p.lastTok = tok
				return entries, nil
			}
		}
	}
	return entries, nil
}

// ParseTrailer parses the trailer dictionary. Parser must be positioned at "trailer".
func (p *ParserImpl) ParseTrailer() (model.Dict, error) {
	tok, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if tok.Kind != syntax.TokenKeyword || tok.Value != "trailer" {
		return nil, fmt.Errorf("expected trailer, got %v", tok)
	}
	tok, err = p.nextToken()
	if err != nil {
		return nil, err
	}
	if tok.Kind != syntax.TokenLDict {
		return nil, fmt.Errorf("expected << after trailer, got %v", tok.Kind)
	}
	return p.parseDict()
}
