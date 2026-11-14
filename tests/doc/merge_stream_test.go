// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/security"
)

// pagesPDF builds and saves a document whose page i reads "<label> page i".
func pagesPDF(t testing.TB, label string, pages int, withImage bool) []byte {
	t.Helper()
	b := doc.New().Title(label)
	for p := range pages {
		b.AddPage()
		b.DrawText(fmt.Sprintf("%s page %d", label, p+1), 72, 700, "Helvetica", 12)
		if withImage {
			b.DrawPNG(300, 100, 120, 120, benchPNG)
		}
	}
	d, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func openPDF(t testing.TB, pdf []byte, password string) doc.Document {
	t.Helper()
	d, err := doc.OpenReaderWithPassword(bytes.NewReader(pdf), int64(len(pdf)), password)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func pageTexts(t testing.TB, pdf []byte, password string) []string {
	t.Helper()
	texts, err := openPDF(t, pdf, password).ReadContentPerPage()
	if err != nil {
		t.Fatal(err)
	}
	for i := range texts {
		texts[i] = strings.TrimSpace(texts[i])
	}
	return texts
}

func TestMergeToKeepsOrderAndSelection(t *testing.T) {
	a := openPDF(t, pagesPDF(t, "A", 3, false), "")
	b := openPDF(t, pagesPDF(t, "B", 2, false), "")
	var out bytes.Buffer
	n, err := doc.MergeTo(&out, doc.CopyOptions{},
		doc.PageSource{Doc: a, Pages: []int{2, 0}},
		doc.PageSource{Doc: b},
		doc.PageSource{Doc: a, Pages: []int{1}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("MergeTo wrote %d pages, want 5", n)
	}
	want := []string{"A page 3", "A page 1", "B page 1", "B page 2", "A page 2"}
	if got := pageTexts(t, out.Bytes(), ""); !slices.Equal(got, want) {
		t.Fatalf("page texts = %q, want %q", got, want)
	}
	checkWithPoppler(t, out.Bytes(), 5)
}

func TestMergeToDropsUnselectedPagesAndInheritsResources(t *testing.T) {
	pdf := threePagePDF() // resources and MediaBox are inherited from the page tree
	src := openPDF(t, pdf, "")
	var out bytes.Buffer
	if _, err := doc.MergeTo(&out, doc.CopyOptions{}, doc.PageSource{Doc: src, Pages: []int{1}}); err != nil {
		t.Fatal(err)
	}
	if got := pageTexts(t, out.Bytes(), ""); !slices.Equal(got, []string{"Page number 2"}) {
		t.Fatalf("page texts = %q", got)
	}
	if bytes.Contains(out.Bytes(), []byte("Page number 1")) || bytes.Contains(out.Bytes(), []byte("Page number 3")) {
		t.Fatal("output holds content of pages that were not selected")
	}
}

// Identical streams across inputs (here the same image in two files) are written once,
// and copied streams keep their stored encoding.
func TestMergeToDeduplicatesIdenticalStreams(t *testing.T) {
	pdf := pagesPDF(t, "Logo", 1, true)
	var out bytes.Buffer
	if _, err := doc.MergeTo(&out, doc.CopyOptions{},
		doc.PageSource{Doc: openPDF(t, pdf, "")},
		doc.PageSource{Doc: openPDF(t, pdf, "")},
	); err != nil {
		t.Fatal(err)
	}
	if got, limit := out.Len(), len(pdf)*13/10; got > limit {
		t.Fatalf("merging a file with itself wrote %d bytes; want at most %d (source is %d)", got, limit, len(pdf))
	}
	images, err := openPDF(t, out.Bytes(), "").ReadImagesPerPage()
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 2 || len(images[0]) != 1 || len(images[1]) != 1 {
		t.Fatalf("images per page = %v, want one image on each of 2 pages", len(images))
	}
	checkWithPoppler(t, out.Bytes(), 2)
}

func TestMergeToReadsAndWritesEncryptedFiles(t *testing.T) {
	plain := openPDF(t, pagesPDF(t, "Secret", 2, false), "")
	var locked bytes.Buffer
	if err := plain.SaveWithAES256Password(&locked, "in-pass", ""); err != nil {
		t.Fatal(err)
	}
	src := openPDF(t, locked.Bytes(), "in-pass")
	var out bytes.Buffer
	if _, err := doc.MergeTo(&out, doc.CopyOptions{UserPassword: "out-pass"}, doc.PageSource{Doc: src, Pages: []int{1}}); err != nil {
		t.Fatal(err)
	}
	if _, err := doc.OpenReader(bytes.NewReader(out.Bytes()), int64(out.Len())); err == nil {
		t.Fatal("encrypted output opened without a password")
	}
	if _, err := doc.OpenReaderWithPassword(bytes.NewReader(out.Bytes()), int64(out.Len()), "in-pass"); !errors.Is(err, security.ErrWrongPassword) {
		t.Fatalf("output opened with the input password: err = %v", err)
	}
	if got := pageTexts(t, out.Bytes(), "out-pass"); !slices.Equal(got, []string{"Secret page 2"}) {
		t.Fatalf("page texts = %q", got)
	}
	if bytes.Contains(out.Bytes(), []byte("Secret page")) {
		t.Fatal("encrypted output holds plain text")
	}
	checkWithPoppler(t, out.Bytes(), 1, "-upw", "out-pass")
}

func TestMergeToWithoutObjectStreams(t *testing.T) {
	src := openPDF(t, pagesPDF(t, "Old", 2, true), "")
	var out bytes.Buffer
	if _, err := doc.MergeTo(&out, doc.CopyOptions{NoObjectStreams: true}, doc.PageSource{Doc: src}); err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(out.Bytes(), []byte("%PDF-1.4")) || bytes.Contains(out.Bytes(), []byte("/ObjStm")) {
		t.Fatal("NoObjectStreams output is not a classic 1.4 file")
	}
	if got := pageTexts(t, out.Bytes(), ""); !slices.Equal(got, []string{"Old page 1", "Old page 2"}) {
		t.Fatalf("page texts = %q", got)
	}
}

func TestMergeToRejectsBadSelections(t *testing.T) {
	src := openPDF(t, pagesPDF(t, "X", 2, false), "")
	if _, err := doc.MergeTo(io.Discard, doc.CopyOptions{}, doc.PageSource{Doc: src, Pages: []int{2}}); err == nil {
		t.Fatal("page index past the end was accepted")
	}
	if _, err := doc.MergeTo(io.Discard, doc.CopyOptions{}, doc.PageSource{Doc: src, Pages: []int{}}); !errors.Is(err, doc.ErrNoPages) {
		t.Fatalf("empty selection: err = %v, want ErrNoPages", err)
	}
	if _, err := doc.MergeTo(io.Discard, doc.CopyOptions{}); err == nil {
		t.Fatal("no sources was accepted")
	}
}

func TestSplitToWritesEachPart(t *testing.T) {
	src := openPDF(t, pagesPDF(t, "S", 5, false), "")
	var ranges [][]int
	for _, spec := range []string{"1-2", "5-3", "last"} {
		r, err := doc.ParsePageRanges(spec, 5)
		if err != nil {
			t.Fatal(err)
		}
		ranges = append(ranges, r)
	}
	parts := make([]*bytes.Buffer, len(ranges))
	err := doc.SplitTo(src, ranges, doc.CopyOptions{}, func(i int) (io.Writer, error) {
		parts[i] = new(bytes.Buffer)
		return parts[i], nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"S page 1", "S page 2"}, {"S page 5", "S page 4", "S page 3"}, {"S page 5"}}
	for i, p := range parts {
		if got := pageTexts(t, p.Bytes(), ""); !slices.Equal(got, want[i]) {
			t.Errorf("part %d texts = %q, want %q", i, got, want[i])
		}
	}
}

func TestParsePageRanges(t *testing.T) {
	cases := []struct {
		spec string
		want []int
	}{
		{"", []int{0, 1, 2, 3, 4}},
		{"all", []int{0, 1, 2, 3, 4}},
		{"1-3,5", []int{0, 1, 2, 4}},
		{" 4- ", []int{3, 4}},
		{"-2", []int{0, 1}},
		{"3-1", []int{2, 1, 0}},
		{"last-1", []int{4, 3, 2, 1, 0}},
		{"2,2", []int{1, 1}},
		{"last", []int{4}},
	}
	for _, c := range cases {
		got, err := doc.ParsePageRanges(c.spec, 5)
		if err != nil {
			t.Errorf("%q: %v", c.spec, err)
			continue
		}
		if !slices.Equal(got, c.want) {
			t.Errorf("%q = %v, want %v", c.spec, got, c.want)
		}
	}
	for _, bad := range []string{"0", "6", "1-6", "a", "1,,2", "-", "1-2-3"} {
		if _, err := doc.ParsePageRanges(bad, 5); err == nil {
			t.Errorf("%q was accepted", bad)
		}
	}
	if _, err := doc.ParsePageRanges("1", 0); err == nil {
		t.Error("a document without pages was accepted")
	}
}

// checkWithPoppler runs pdfinfo on pdf when it is installed and checks the page count.
func checkWithPoppler(t *testing.T, pdf []byte, pages int, args ...string) {
	t.Helper()
	bin, err := exec.LookPath("pdfinfo")
	if err != nil {
		return
	}
	cmd := exec.Command(bin, append(args, "-")...)
	cmd.Stdin = bytes.NewReader(pdf)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pdfinfo: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "Syntax Error") || strings.Contains(string(out), "Syntax Warning") {
		t.Fatalf("pdfinfo reported problems:\n%s", out)
	}
	if !strings.Contains(string(out), fmt.Sprintf("Pages:           %d\n", pages)) {
		t.Fatalf("pdfinfo page count, want %d:\n%s", pages, out)
	}
}
