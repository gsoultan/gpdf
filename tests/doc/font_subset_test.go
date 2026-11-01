// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
)

// sfntTables returns the table directory of a TrueType program as tag -> length.
func sfntTables(t *testing.T, b []byte) map[string]int {
	t.Helper()
	if len(b) < 12 {
		t.Fatalf("font program too short: %d bytes", len(b))
	}
	n := int(binary.BigEndian.Uint16(b[4:6]))
	out := map[string]int{}
	for i := range n {
		rec := b[12+i*16:]
		out[string(rec[0:4])] = int(binary.BigEndian.Uint32(rec[12:16]))
	}
	return out
}

func TestSubsetKeepsOnlyNeededTablesAndGlyphs(t *testing.T) {
	f := loadTestFont(t)
	used := map[uint16]bool{}
	for _, r := range "Hi" {
		used[f.GlyphID(r)] = true
	}
	sub, err := f.Subset(used)
	if err != nil {
		t.Fatal(err)
	}
	tables := sfntTables(t, sub)
	for _, tag := range []string{"name", "cmap"} {
		if _, ok := tables[tag]; ok {
			t.Errorf("subset still carries %q", tag)
		}
	}
	if tables["post"] != 32 {
		t.Errorf("post table should be format 3.0 (32 bytes), got %d", tables["post"])
	}
	maxGID := 0
	for g := range used {
		maxGID = max(maxGID, int(g))
	}
	locaLen := tables["loca"]
	if want := (maxGID + 2) * 2; locaLen != want && locaLen != want*2 {
		t.Errorf("loca length %d, want %d entries", locaLen, maxGID+2)
	}
	if tables["hmtx"] > (maxGID+1)*4 {
		t.Errorf("hmtx length %d not truncated to %d glyphs", tables["hmtx"], maxGID+1)
	}
	full := 0
	if data, err := os.ReadFile("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"); err == nil {
		full = len(data)
	}
	if full > 0 && len(sub)*4 > full {
		t.Errorf("subset is %d bytes, full font %d: expected at least a 4x reduction", len(sub), full)
	}
}

func TestSubsetTextExtractsWithPoppler(t *testing.T) {
	f := loadTestFont(t)
	b := doc.New().RegisterFont(f).PageSize(300, 200).AddPage()
	b.DrawText("Subset glyphs 42", 20, 100, f.PostScriptName(), 14)
	built, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := built.Save(&buf); err != nil {
		t.Fatal(err)
	}
	pdf := buf.Bytes()
	bin, err := exec.LookPath("pdftotext")
	if err != nil {
		t.Skip("pdftotext not installed")
	}
	path := filepath.Join(t.TempDir(), "s.pdf")
	os.WriteFile(path, pdf, 0o600)
	out, err := exec.Command(bin, path, "-").CombinedOutput()
	if err != nil || !strings.Contains(string(out), "Subset glyphs 42") {
		t.Fatalf("pdftotext: %v\n%s", err, out)
	}
	if fonts, err := exec.LookPath("pdffonts"); err == nil {
		out, err := exec.Command(fonts, path).CombinedOutput()
		if err != nil || !strings.Contains(string(out), "CID TrueType") || strings.Contains(string(out), "Error") {
			t.Errorf("pdffonts: %v\n%s", err, out)
		}
	}
}
