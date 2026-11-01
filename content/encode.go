package content

import (
	"bytes"
	"fmt"
	"io"
	"slices"

	"github.com/gsoultan/gpdf/model"
)

// Encoder writes content stream operators to PDF content stream syntax.
type Encoder struct {
	w *bytes.Buffer
}

// NewEncoder returns an encoder that writes to an internal buffer.
func NewEncoder() *Encoder {
	return &Encoder{w: &bytes.Buffer{}}
}

// WriteOp appends one operator (args then name) to the stream. An inline
// image, as the parser returns it (operator BI with the parameter dictionary
// and the raw samples as its arguments), is written back in BI … ID … EI form.
func (e *Encoder) WriteOp(op Op) error {
	if op.Name == "BI" && len(op.Args) == 2 {
		if params, ok := op.Args[0].(model.Dict); ok {
			if data, ok := op.Args[1].(model.String); ok {
				return e.writeInlineImage(params, []byte(data))
			}
		}
	}
	for i, arg := range op.Args {
		if i > 0 {
			e.w.WriteByte(' ')
		}
		if err := e.writeObject(arg); err != nil {
			return err
		}
	}
	e.w.WriteByte(' ')
	e.w.WriteString(op.Name)
	e.w.WriteByte('\n')
	return nil
}

func (e *Encoder) writeInlineImage(params model.Dict, data []byte) error {
	e.w.WriteString("BI")
	for _, key := range sortedKeys(params) {
		e.w.WriteByte(' ')
		if err := e.writeObject(key); err != nil {
			return err
		}
		e.w.WriteByte(' ')
		if err := e.writeObject(params[key]); err != nil {
			return err
		}
	}
	e.w.WriteString(" ID ")
	e.w.Write(data)
	e.w.WriteString("\nEI\n")
	return nil
}

func sortedKeys(d model.Dict) []model.Name {
	keys := make([]model.Name, 0, len(d))
	for key := range d {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

// Bytes returns the encoded content stream.
func (e *Encoder) Bytes() []byte {
	return e.w.Bytes()
}

// Reset clears the buffer for reuse.
func (e *Encoder) Reset() {
	e.w.Reset()
}

func (e *Encoder) writeObject(obj model.Object) error {
	switch v := obj.(type) {
	case model.Integer:
		fmt.Fprintf(e.w, "%d", v)
	case model.Real:
		io.WriteString(e.w, model.FormatReal(float64(v)))
	case model.String:
		writeLiteralString(e.w, string(v))
	case model.HexString:
		writeHexString(e.w, []byte(v))
	case model.Name:
		fmt.Fprintf(e.w, "/%s", escapeContentName(string(v)))
	case model.Array:
		e.w.WriteByte('[')
		for i, el := range v {
			if i > 0 {
				e.w.WriteByte(' ')
			}
			if err := e.writeObject(el); err != nil {
				return err
			}
		}
		e.w.WriteByte(']')
	case model.Dict:
		e.w.WriteString("<<")
		for i, key := range sortedKeys(v) {
			if i > 0 {
				e.w.WriteByte(' ')
			}
			if err := e.writeObject(key); err != nil {
				return err
			}
			e.w.WriteByte(' ')
			if err := e.writeObject(v[key]); err != nil {
				return err
			}
		}
		e.w.WriteString(">>")
	default:
		return fmt.Errorf("content encode: unsupported type %T", obj)
	}
	return nil
}

func writeLiteralString(w io.Writer, s string) {
	io.WriteString(w, "(")
	for i := range len(s) {
		c := s[i]
		switch c {
		case '\\', '(', ')':
			fmt.Fprintf(w, "\\%c", c)
		case '\n':
			io.WriteString(w, "\\n")
		case '\r':
			io.WriteString(w, "\\r")
		case '\t':
			io.WriteString(w, "\\t")
		default:
			w.Write([]byte{c})
		}
	}
	io.WriteString(w, ")")
}

func writeHexString(w io.Writer, b []byte) {
	io.WriteString(w, "<")
	for _, c := range b {
		fmt.Fprintf(w, "%02X", c)
	}
	io.WriteString(w, ">")
}

func escapeContentName(s string) string {
	var b bytes.Buffer
	for i := range len(s) {
		c := s[i]
		if c <= ' ' || c >= 127 || c == '#' || c == '/' || c == '(' || c == ')' || c == '<' || c == '>' || c == '[' || c == ']' || c == '%' {
			fmt.Fprintf(&b, "#%02x", c)
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}

// EncodeBytes encodes a sequence of operators to content stream bytes.
func EncodeBytes(ops []Op) ([]byte, error) {
	enc := NewEncoder()
	for _, op := range ops {
		if err := enc.WriteOp(op); err != nil {
			return nil, err
		}
	}
	return enc.Bytes(), nil
}
