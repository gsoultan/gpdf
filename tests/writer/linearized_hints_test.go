package writer_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/model"
)

type bitReader struct {
	data []byte
	pos  int // in bits
}

func (r *bitReader) read(n int) int64 {
	var v int64
	for range n {
		b := r.data[r.pos/8] >> (7 - r.pos%8) & 1
		v = v<<1 | int64(b)
		r.pos++
	}
	return v
}

func (r *bitReader) align() { r.pos = (r.pos + 7) / 8 * 8 }

func linValue(t *testing.T, head, key string) int64 {
	t.Helper()
	m := regexp.MustCompile(`/` + key + ` \[?(\d+)`).FindStringSubmatch(head)
	if m == nil {
		t.Fatalf("no /%s in %s", key, head)
	}
	v, _ := strconv.ParseInt(m[1], 10, 64)
	return v
}

// The primary hint stream must describe the file the way poppler (Hints.cc) and ISO
// 32000-1 F.4 read it: page object numbers and offsets derived from the page offset
// table must land on the real page objects, and shared object references must resolve.
func TestSaveLinearizedWritesUsableHintTables(t *testing.T) {
	b := doc.New().PageSize(300, 300).SetFooter(doc.PageText{Text: "{page}/{pages}"})
	for i := range 3 {
		b.AddPage()
		b.FillRect(i, 10, 10, 50, 50, doc.ColorBlue)
		b.DrawText("Body "+strconv.Itoa(i+1), 20, 200, "Helvetica", 12)
	}
	b.AddOutline("First", 0)
	built, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	var ws seekBuffer
	if err := built.SaveLinearized(&ws); err != nil {
		t.Fatal(err)
	}
	pdf := ws.data
	head := string(pdf[:200])
	hOff, hLen := linValue(t, head, "H"), func() int64 {
		m := regexp.MustCompile(`/H \[(\d+) (\d+)\]`).FindStringSubmatch(head)
		v, _ := strconv.ParseInt(m[2], 10, 64)
		return v
	}()
	firstPage := int(linValue(t, head, "O"))
	nPages := int(linValue(t, head, "N"))
	if nPages != 3 {
		t.Fatalf("/N = %d", nPages)
	}
	objOffset := func(num int) int64 {
		i := bytes.Index(pdf, []byte("\n"+strconv.Itoa(num)+" 0 obj\n"))
		if i < 0 {
			t.Fatalf("object %d not found", num)
		}
		return int64(i + 1)
	}
	// The hint stream object starts at /H offset and holds /S.
	hintObj := pdf[hOff : hOff+hLen]
	m := regexp.MustCompile(`(?s)^(\d+) 0 obj\n<< /Length (\d+) /S (\d+) >>\nstream\n`).FindSubmatch(hintObj)
	if m == nil {
		t.Fatalf("no hint stream at /H offset:\n%s", hintObj[:min(80, len(hintObj))])
	}
	length, _ := strconv.Atoi(string(m[2]))
	sharedStart, _ := strconv.Atoi(string(m[3]))
	tables := hintObj[len(m[0]) : len(m[0])+length]

	r := &bitReader{data: tables}
	leastObj := r.read(32)
	objectOffsetFirst := r.read(32)
	bObj := int(r.read(16))
	leastLen := r.read(32)
	bLen := int(r.read(16))
	r.read(32)
	bCOff := int(r.read(16))
	r.read(32)
	bCLen := int(r.read(16))
	bNumShared := int(r.read(16))
	bID := int(r.read(16))
	bNum := int(r.read(16))
	if den := r.read(16); den == 0 {
		t.Fatal("fraction denominator is 0")
	}
	if objectOffsetFirst+hLen != objOffset(firstPage) {
		t.Errorf("item 2 = %d (+%d) but the first page object is at %d", objectOffsetFirst, hLen, objOffset(firstPage))
	}
	nObj := make([]int64, nPages)
	for i := range nObj {
		nObj[i] = leastObj + r.read(bObj)
	}
	r.align()
	pageLen := make([]int64, nPages)
	for i := range pageLen {
		pageLen[i] = leastLen + r.read(bLen)
	}
	r.align()
	numShared := make([]int64, nPages)
	for i := range numShared {
		numShared[i] = r.read(bNumShared)
	}
	r.align()
	ids := make([][]int64, nPages)
	for i := range ids {
		for range numShared[i] {
			ids[i] = append(ids[i], r.read(bID))
		}
	}
	r.align()
	for i := range ids {
		for range ids[i] {
			r.read(bNum)
		}
	}
	r.align()
	for range nPages {
		r.read(bCOff)
	}
	r.align()
	for range nPages {
		r.read(bCLen)
	}
	r.align()
	if r.pos/8 != sharedStart {
		t.Errorf("page offset table is %d bytes but /S is %d", r.pos/8, sharedStart)
	}

	// Pages after the first are numbered from 1, each section consecutive (poppler's
	// pageObjectNum), and each section starts where the previous one ends.
	d := open(t, pdf, "")
	isPage := func(num int) bool {
		o, err := d.Resolve(model.Ref{ObjectNumber: num})
		if err != nil {
			return false
		}
		dict, ok := o.(model.Dict)
		return ok && dict[model.Name("Type")] == model.Name("Page")
	}
	if !isPage(firstPage) {
		t.Errorf("/O %d is not a page", firstPage)
	}
	num := 1
	off := objOffset(firstPage) + pageLen[0]
	for i := 1; i < nPages; i++ {
		if !isPage(num) {
			t.Errorf("page %d: object %d is not a page", i+1, num)
		}
		if got := objOffset(num); got != off {
			t.Errorf("page %d: section starts at %d, hint tables say %d", i+1, got, off)
		}
		num += int(nObj[i])
		off += pageLen[i]
	}

	// Shared object table: group count and first-page groups are non-zero and shared
	// identifiers referenced by pages are in range.
	s := &bitReader{data: tables[sharedStart:]}
	s.read(32)
	s.read(32)
	nFirst := s.read(32)
	nGroups := s.read(32)
	if nFirst == 0 || nGroups < nFirst {
		t.Errorf("shared groups: first page %d, total %d", nFirst, nGroups)
	}
	for i, pageIDs := range ids {
		for _, id := range pageIDs {
			if id >= nGroups {
				t.Errorf("page %d references shared group %d of %d", i+1, id, nGroups)
			}
		}
	}

	// poppler must parse the hint tables without a syntax warning and read every page.
	bin, err := exec.LookPath("pdftotext")
	if err != nil {
		return
	}
	path := filepath.Join(t.TempDir(), "lin.pdf")
	os.WriteFile(path, pdf, 0o600)
	for p := 1; p <= nPages; p++ {
		out, _ := exec.Command(bin, "-f", strconv.Itoa(p), "-l", strconv.Itoa(p), path, "-").CombinedOutput()
		if strings.Contains(string(out), "Syntax") || !strings.Contains(string(out), "Body "+strconv.Itoa(p)) {
			t.Errorf("poppler page %d: %q", p, out)
		}
	}
}
