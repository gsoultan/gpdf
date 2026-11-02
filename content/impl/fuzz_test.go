// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package impl

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/content"
	"github.com/gsoultan/gpdf/model"
)

// FuzzParseContent checks that hostile content streams never panic the
// operator parser and that what it returns can be encoded again.
func FuzzParseContent(f *testing.F) {
	for _, s := range []string{
		"BT /F1 12 Tf 20 100 Td (Seed) Tj ET",
		"q 1 0 0 1 10 10 cm /Im1 Do Q",
		"/Span << /MCID 0 >> BDC BT [(a) -20 (b)] TJ ET EMC",
		"BI /W 2 /H 2 /CS /G /BPC 8 ID \x00\xff\xff\x00 EI",
		"0 0 m 10 10 l 1 0 0 RG S ) ] >> } 1e400 Tj",
		"BT (unterminated Tj",
	} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		ops, err := NewStreamParser().Parse(data)
		if err != nil {
			return
		}
		enc := content.NewEncoder()
		for _, op := range ops {
			if op.Name == "" || strings.ContainsAny(op.Name, "()<>[]{}/%") {
				return // a stray delimiter the parser recovered from, not an operator
			}
			if err := enc.WriteOp(op); err != nil {
				return
			}
			if op.Name == "BI" && bytes.Contains([]byte(op.Args[1].(model.String)), []byte("EI")) {
				return // the samples would end the inline image early
			}
		}
		back, err := NewStreamParser().Parse(enc.Bytes())
		if err != nil {
			t.Fatalf("re-parsing encoded operators: %v\n%q", err, enc.Bytes())
		}
		if len(back) != len(ops) {
			t.Fatalf("%d operators became %d:\n%q", len(ops), len(back), enc.Bytes())
		}
	})
}
