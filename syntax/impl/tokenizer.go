package impl

import (
	"bufio"
	"bytes"
	"io"
	"strconv"

	"github.com/gsoultan/gpdf/syntax"
)

// tokenizer produces PDF tokens from a reader. It supports reset to a new position for random access.
type tokenizer struct {
	r   io.ReaderAt
	pos int64
	end int64
	buf *bufio.Reader
}

// NewTokenizer returns a tokenizer that reads from r starting at start, up to size bytes.
func NewTokenizer(r io.ReaderAt, start, size int64) syntax.Tokenizer {
	t := &tokenizer{r: r, pos: start, end: start + size}
	t.buf = bufio.NewReader(io.NewSectionReader(r, start, size))
	return t
}

// SetPosition sets the current read position (for Parser.SetPosition).
func (t *tokenizer) SetPosition(offset int64) {
	t.pos = offset
	n := max(t.end-offset, 0)
	t.buf = bufio.NewReader(io.NewSectionReader(t.r, offset, n))
}

// CurrentOffset returns the byte offset of the next character to be read.
func (t *tokenizer) CurrentOffset() int64 {
	return t.pos
}

func (t *tokenizer) readByte() (byte, error) {
	b, err := t.buf.ReadByte()
	if err == nil {
		t.pos++
	}
	return b, err
}

func (t *tokenizer) unreadByte() {
	if t.buf.UnreadByte() == nil {
		t.pos--
	}
}

func (t *tokenizer) peekByte() (byte, bool) {
	p, err := t.buf.Peek(1)
	if err != nil || len(p) == 0 {
		return 0, false
	}
	return p[0], true
}

func (t *tokenizer) discard(n int) {
	d, _ := t.buf.Discard(n)
	t.pos += int64(d)
}

func (t *tokenizer) Next() (syntax.Token, error) {
	// Skip whitespace and comments.
	for {
		b, err := t.readByte()
		if err == io.EOF {
			return syntax.Token{Kind: syntax.TokenEOF}, nil
		}
		if err != nil {
			return syntax.Token{}, err
		}
		if b == '%' {
			t.skipComment()
			continue
		}
		if isWhitespace(b) {
			continue
		}
		t.unreadByte()
		break
	}

	first, ok := t.peekByte()
	if !ok {
		return syntax.Token{Kind: syntax.TokenEOF}, nil
	}

	switch first {
	case '/':
		return t.readName()
	case '(':
		return t.readLiteralString()
	case '<':
		peek2, _ := t.buf.Peek(2)
		if len(peek2) == 2 && peek2[1] == '<' {
			t.discard(2)
			return syntax.Token{Kind: syntax.TokenLDict, Value: "<<"}, nil
		}
		return t.readHexString()
	case '>':
		peek2, _ := t.buf.Peek(2)
		if len(peek2) == 2 && peek2[1] == '>' {
			t.discard(2)
			return syntax.Token{Kind: syntax.TokenRDict, Value: ">>"}, nil
		}
		// A lone '>' is not valid PDF syntax. Consume it and report it, rather than
		// returning a zero Token (which is TokenEOF and silently truncates parsing).
		t.discard(1)
		return syntax.Token{Kind: syntax.TokenUnknown, Value: ">"}, nil
	case '[':
		t.discard(1)
		return syntax.Token{Kind: syntax.TokenLBracket}, nil
	case ']':
		t.discard(1)
		return syntax.Token{Kind: syntax.TokenRBracket}, nil
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '+', '-', '.':
		return t.readNumber()
	default:
		return t.readKeywordOrRef()
	}
}

// skipComment consumes the rest of a comment line (the '%' is already consumed).
func (t *tokenizer) skipComment() {
	for {
		b, err := t.readByte()
		if err != nil || b == '\n' || b == '\r' {
			return
		}
	}
}

func isWhitespace(b byte) bool {
	return b == 0 || b == '\t' || b == '\n' || b == '\f' || b == '\r' || b == ' '
}

func (t *tokenizer) readName() (syntax.Token, error) {
	t.discard(1) // /
	var buf bytes.Buffer
	for {
		b, err := t.readByte()
		if err != nil {
			break
		}
		if isDelimiter(b) || isWhitespace(b) {
			t.unreadByte()
			break
		}
		if b == '#' {
			hex, err := t.buf.Peek(2)
			if err == nil && len(hex) == 2 && isHex(hex[0]) && isHex(hex[1]) {
				t.discard(2)
				buf.WriteByte(hexByte(hex[0], hex[1]))
				continue
			}
		}
		buf.WriteByte(b)
	}
	return syntax.Token{Kind: syntax.TokenName, Value: buf.String()}, nil
}

func isDelimiter(b byte) bool {
	return b == '/' || b == '(' || b == ')' || b == '<' || b == '>' || b == '[' || b == ']' || b == '{' || b == '}' || b == '%'
}

func isHex(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'A' && b <= 'F') || (b >= 'a' && b <= 'f')
}

func hexByte(a, b byte) byte {
	return (hexVal(a) << 4) | hexVal(b)
}

func hexVal(b byte) byte {
	if b >= '0' && b <= '9' {
		return b - '0'
	}
	if b >= 'A' && b <= 'F' {
		return b - 'A' + 10
	}
	if b >= 'a' && b <= 'f' {
		return b - 'a' + 10
	}
	return 0
}

func (t *tokenizer) readLiteralString() (syntax.Token, error) {
	t.discard(1) // (
	var buf bytes.Buffer
	depth := 1
	for depth > 0 {
		b, err := t.readByte()
		if err != nil {
			break
		}
		switch b {
		case '\\':
			n, err := t.readByte()
			if err != nil {
				break
			}
			switch n {
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case '(', ')', '\\':
				buf.WriteByte(n)
			case '\r':
				// Line continuation; a CRLF pair counts as one end-of-line marker.
				if p, ok := t.peekByte(); ok && p == '\n' {
					t.discard(1)
				}
			case '\n':
				// line continuation
			default:
				if n >= '0' && n <= '7' {
					v := n - '0'
					for range 2 {
						p, ok := t.peekByte()
						if !ok || p < '0' || p > '7' {
							break
						}
						t.discard(1)
						v = v*8 + (p - '0')
					}
					buf.WriteByte(v)
				} else {
					// Unknown escape: the backslash is ignored (ISO 32000-1 §7.3.4.2).
					buf.WriteByte(n)
				}
			}
		case '\r':
			// An unescaped end-of-line in a literal string is read as a single LF.
			if p, ok := t.peekByte(); ok && p == '\n' {
				t.discard(1)
			}
			buf.WriteByte('\n')
		case '(':
			depth++
			buf.WriteByte(b)
		case ')':
			depth--
			if depth > 0 {
				buf.WriteByte(b)
			}
		default:
			buf.WriteByte(b)
		}
	}
	return syntax.Token{Kind: syntax.TokenLiteral, Value: buf.String()}, nil
}

func (t *tokenizer) readHexString() (syntax.Token, error) {
	t.discard(1) // <
	var buf bytes.Buffer
	var hi byte
	half := false
	for {
		b, err := t.readByte()
		if err != nil || b == '>' {
			break
		}
		if !isHex(b) {
			continue // whitespace and stray characters are ignored
		}
		if !half {
			hi = hexVal(b)
			half = true
			continue
		}
		buf.WriteByte(hi<<4 | hexVal(b))
		half = false
	}
	if half {
		// An odd number of digits behaves as if a final 0 followed.
		buf.WriteByte(hi << 4)
	}
	return syntax.Token{Kind: syntax.TokenHex, Value: buf.String()}, nil
}

func (t *tokenizer) readNumber() (syntax.Token, error) {
	var buf bytes.Buffer
	for {
		b, err := t.readByte()
		if err != nil {
			break
		}
		if (b >= '0' && b <= '9') || b == '+' || b == '-' || b == '.' {
			buf.WriteByte(b)
			continue
		}
		t.unreadByte()
		break
	}
	return numberToken(buf.String()), nil
}

// numberToken converts a numeric lexeme. Malformed numbers (e.g. "--5", "1.2.3", "-")
// are read leniently the way common viewers do: a leading sign and the first run of
// digits and one decimal point are used and the rest is ignored.
func numberToken(s string) syntax.Token {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return syntax.Token{Kind: syntax.TokenInteger, Value: s, Int: i}
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return syntax.Token{Kind: syntax.TokenReal, Value: s, Float: f}
	}
	neg := false
	i := 0
	for i < len(s) && (s[i] == '+' || s[i] == '-') {
		if s[i] == '-' {
			neg = !neg
		}
		i++
	}
	j := i
	dot := false
	for j < len(s) && ((s[j] >= '0' && s[j] <= '9') || (s[j] == '.' && !dot)) {
		if s[j] == '.' {
			dot = true
		}
		j++
	}
	core := s[i:j]
	if core == "" || core == "." {
		core = "0"
	}
	if neg {
		core = "-" + core
	}
	if !dot {
		if v, err := strconv.ParseInt(core, 10, 64); err == nil {
			return syntax.Token{Kind: syntax.TokenInteger, Value: s, Int: v}
		}
	}
	f, _ := strconv.ParseFloat(core, 64)
	return syntax.Token{Kind: syntax.TokenReal, Value: s, Float: f}
}

var keywords = map[string]syntax.TokenKind{
	"obj":       syntax.TokenKeyword,
	"endobj":    syntax.TokenKeyword,
	"stream":    syntax.TokenKeyword,
	"endstream": syntax.TokenKeyword,
	"xref":      syntax.TokenKeyword,
	"trailer":   syntax.TokenKeyword,
	"startxref": syntax.TokenKeyword,
	"R":         syntax.TokenKeyword,
	"true":      syntax.TokenKeyword,
	"false":     syntax.TokenKeyword,
	"null":      syntax.TokenKeyword,
}

func (t *tokenizer) readKeywordOrRef() (syntax.Token, error) {
	var buf bytes.Buffer
	for {
		b, err := t.readByte()
		if err != nil {
			break
		}
		if isWhitespace(b) || isDelimiter(b) {
			if buf.Len() == 0 {
				// A delimiter that cannot start a token (')', '{', '}'): consume it so the
				// caller always makes progress instead of looping on the same byte.
				return syntax.Token{Kind: syntax.TokenUnknown, Value: string(b)}, nil
			}
			t.unreadByte()
			break
		}
		buf.WriteByte(b)
	}
	s := buf.String()
	if k, ok := keywords[s]; ok {
		return syntax.Token{Kind: k, Value: s}, nil
	}
	return syntax.Token{Kind: syntax.TokenUnknown, Value: s}, nil
}
